package fileSystem

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	etcdErr "github.com/coreos/etcd/error"
)

type FileSystem struct {
	Root       *Node
	WatcherHub *watcherHub
	Index      uint64
	Term       uint64
}

func New() *FileSystem {
	return &FileSystem{
		Root:       newDir("/", 0, 0, nil, "", Permanent),
		WatcherHub: newWatchHub(1000),
	}

}

func (fs *FileSystem) Get(nodePath string, recursive, sorted bool, index uint64, term uint64) (*Event, error) {
	n, err := fs.InternalGet(nodePath, index, term)

	if err != nil {
		return nil, err
	}

	e := newEvent(Get, nodePath, index, term)

	if n.IsDir() { // node is dir
		e.Dir = true

		children, _ := n.List()
		e.KVPairs = make([]KeyValuePair, len(children))

		// we do not use the index in the children slice directly
		// we need to skip the hidden one
		i := 0

		for _, child := range children {

			if child.IsHidden() { // get will not list hidden node
				continue
			}

			e.KVPairs[i] = child.Pair(recursive, sorted)

			i++
		}

		// eliminate hidden nodes
		e.KVPairs = e.KVPairs[:i]

		rootPairs := KeyValuePair{
			KVPairs: e.KVPairs,
		}

		if sorted {
			sort.Sort(rootPairs)
		}

	} else { // node is file
		e.Value = n.Value
	}

	return e, nil
}

// CreateDir function is wrapper to create directory node.
func (fs *FileSystem) CreateDir(nodePath string, expireTime time.Time, index uint64, term uint64) (*Event, error) {
	return fs.Create(nodePath, "", expireTime, index, term)
}

// Create function creates the Node at nodePath. Create will help to create intermediate directories with no ttl.
// If the node has already existed, create will fail.
// If any node on the path is a file, create will fail.
func (fs *FileSystem) Create(nodePath string, value string, expireTime time.Time, index uint64, term uint64) (*Event, error) {
	nodePath = path.Clean("/" + nodePath)

	// make sure we can create the node
	_, err := fs.InternalGet(nodePath, index, term)

	if err == nil { // key already exists
		return nil, etcdErr.NewError(etcdErr.EcodeNodeExist, nodePath)
	}

	etcdError, _ := err.(etcdErr.Error)

	if etcdError.ErrorCode == 104 { // we cannot create the key due to meet a file while walking
		return nil, err
	}

	dir, _ := path.Split(nodePath)

	// walk through the nodePath, create dirs and get the last directory node
	d, err := fs.walk(dir, fs.checkDir)

	if err != nil {
		return nil, err
	}

	e := newEvent(Create, nodePath, fs.Index, fs.Term)

	var n *Node

	if len(value) != 0 { // create file
		e.Value = value

		n = newFile(nodePath, value, fs.Index, fs.Term, d, "", expireTime)

	} else { // create directory
		e.Dir = true

		n = newDir(nodePath, fs.Index, fs.Term, d, "", expireTime)

	}

	err = d.Add(n)

	if err != nil {
		return nil, err
	}

	// Node with TTL
	if expireTime != Permanent {
		n.Expire()
		e.Expiration = &n.ExpireTime
		e.TTL = int64(expireTime.Sub(time.Now()) / time.Second)
	}

	fs.WatcherHub.notify(e)
	return e, nil
}

// Update function updates the value/ttl of the node.
// If the node is a file, the value and the ttl can be updated.
// If the node is a directory, only the ttl can be updated.
func (fs *FileSystem) Update(nodePath string, value string, expireTime time.Time, index uint64, term uint64) (*Event, error) {
	n, err := fs.InternalGet(nodePath, index, term)

	if err != nil { // if the node does not exist, return error
		return nil, err
	}

	e := newEvent(Update, nodePath, fs.Index, fs.Term)

	if n.IsDir() { // if the node is a directory, we can only update ttl

		if len(value) != 0 {
			return nil, etcdErr.NewError(etcdErr.EcodeNotFile, nodePath)
		}

	} else { // if the node is a file, we can update value and ttl
		e.PrevValue = n.Value

		if len(value) != 0 {
			e.Value = value
		}

		n.Write(value, index, term)
	}

	// update ttl
	if !n.IsPermanent() && expireTime != Permanent {
		n.stopExpire <- true
	}

	if expireTime.Sub(Permanent) != 0 {
		n.ExpireTime = expireTime
		n.Expire()
		e.Expiration = &n.ExpireTime
		e.TTL = int64(expireTime.Sub(time.Now()) / time.Second)
	}

	fs.WatcherHub.notify(e)
	return e, nil
}

func (fs *FileSystem) TestAndSet(nodePath string, prevValue string, prevIndex uint64,
	value string, expireTime time.Time, index uint64, term uint64) (*Event, error) {

	f, err := fs.InternalGet(nodePath, index, term)

	if err != nil {

		return nil, err
	}

	if f.IsDir() { // can only test and set file
		return nil, etcdErr.NewError(etcdErr.EcodeNotFile, nodePath)
	}

	if f.Value == prevValue || f.ModifiedIndex == prevIndex {
		// if test succeed, write the value
		e := newEvent(TestAndSet, nodePath, index, term)
		e.PrevValue = f.Value
		e.Value = value
		f.Write(value, index, term)

		fs.WatcherHub.notify(e)

		return e, nil
	}

	cause := fmt.Sprintf("[%v/%v] [%v/%v]", prevValue, f.Value, prevIndex, f.ModifiedIndex)
	return nil, etcdErr.NewError(etcdErr.EcodeTestFailed, cause)
}

// Delete function deletes the node at the given path.
// If the node is a directory, recursive must be true to delete it.
func (fs *FileSystem) Delete(nodePath string, recursive bool, index uint64, term uint64) (*Event, error) {
	n, err := fs.InternalGet(nodePath, index, term)

	if err != nil { // if the node does not exist, return error
		return nil, err
	}

	e := newEvent(Delete, nodePath, index, term)

	if n.IsDir() {
		e.Dir = true
	} else {
		e.PrevValue = n.Value
	}

	callback := func(path string) { // notify function
		fs.WatcherHub.notifyWithPath(e, path, true)
	}

	err = n.Remove(recursive, callback)

	if err != nil {
		return nil, err
	}

	fs.WatcherHub.notify(e)

	return e, nil
}

// walk function walks all the nodePath and apply the walkFunc on each directory
func (fs *FileSystem) walk(nodePath string, walkFunc func(prev *Node, component string) (*Node, error)) (*Node, error) {
	components := strings.Split(nodePath, "/")

	curr := fs.Root

	var err error
	for i := 1; i < len(components); i++ {
		if len(components[i]) == 0 { // ignore empty string
			return curr, nil
		}

		curr, err = walkFunc(curr, components[i])
		if err != nil {
			return nil, err
		}

	}

	return curr, nil
}

// InternalGet function get the node of the given nodePath.
func (fs *FileSystem) InternalGet(nodePath string, index uint64, term uint64) (*Node, error) {
	nodePath = path.Clean("/" + nodePath)

	// update file system known index and term
	fs.Index, fs.Term = index, term

	walkFunc := func(parent *Node, name string) (*Node, error) {

		if !parent.IsDir() {
			return nil, etcdErr.NewError(etcdErr.EcodeNotDir, parent.Path)
		}

		child, ok := parent.Children[name]
		if ok {
			return child, nil
		}

		return nil, etcdErr.NewError(etcdErr.EcodeKeyNotFound, path.Join(parent.Path, name))
	}

	f, err := fs.walk(nodePath, walkFunc)

	if err != nil {
		return nil, err
	}

	return f, nil
}

// checkDir function will check whether the component is a directory under parent node.
// If it is a directory, this function will return the pointer to that node.
// If it does not exist, this function will create a new directory and return the pointer to that node.
// If it is a file, this function will return error.
func (fs *FileSystem) checkDir(parent *Node, dirName string) (*Node, error) {
	subDir, ok := parent.Children[dirName]

	if ok {
		return subDir, nil
	}

	n := newDir(path.Join(parent.Path, dirName), fs.Index, fs.Term, parent, parent.ACL, Permanent)

	parent.Children[dirName] = n

	return n, nil
}

// Save function saves the static state of the store system.
// Save function will not be able to save the state of watchers.
// Save function will not save the parent field of the node. Or there will
// be cyclic dependencies issue for the json package.
func (fs *FileSystem) Save() []byte {
	cloneFs := New()
	cloneFs.Root = fs.Root.Clone()

	b, err := json.Marshal(fs)

	if err != nil {
		panic(err)
	}

	return b
}

// recovery function recovery the store system from a static state.
// It needs to recovery the parent field of the nodes.
// It needs to delete the expired nodes since the saved time and also
// need to create monitor go routines.
func (fs *FileSystem) Recover(state []byte) {
	err := json.Unmarshal(state, fs)

	if err != nil {
		panic(err)
	}

	fs.Root.recoverAndclean()

}
