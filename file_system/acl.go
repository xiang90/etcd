package fileSystem

import (
	"path"

	etcdErr "github.com/coreos/etcd/error"
)

func getUserByCertSubj() string {
	return "test"
}

// check_perm function checks whether the given acl-name has permission for
// current user.
// If it has, then return nil.
// Otherwise, return error with code permission denied.
func (fs *FileSystem) check_perm(acl string, perm string) error {

	user := getUserByCertSubj()

	// Enumerate the permissions
	for _, char := range perm {
		_, err := fs.InternalGet(path.Join(acl, string(char), user), fs.Index, fs.Term)

		if err != nil {
			return etcdErr.NewError(etcdErr.EcodePermissionDenied, perm)
		}
	}

	return nil

}

// has_perm function is a higher level function wrapping check_perm so
// acl_stringas to provide recursive functionality
func (fs *FileSystem) has_perm(n *Node, perm string, recursive bool) error {
	err := fs.check_perm(n.ACL, perm)
	if err != nil {
		return err
	}

	if n.IsDir() && recursive {
		children, _ := n.List()

		for _, child := range children {

			if child.IsHidden() { // get will not list hidden node
				continue
			}

			err = fs.has_perm(child, perm, recursive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
