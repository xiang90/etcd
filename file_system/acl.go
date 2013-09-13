package fileSystem

import (
	"path"
	"strings"

	etcdErr "github.com/coreos/etcd/error"
)

func getUser() string {
	return "admin"
}

// checkPerm function checks whether the given acl-name has permission for
// current user.
// If it has, then return nil.
// Otherwise, return error with code permission denied.
func (fs *FileSystem) checkPerm(aclName string, perm string) error {

	user := getUser()

	// Enumerate the permissions
	for _, char := range perm {
		_, err := fs.InternalGet(path.Join("/ACL", aclName, string(char), user), fs.Index, fs.Term)

		if err != nil {
			return etcdErr.NewError(etcdErr.EcodePermissionDenied, perm)
		}
	}

	return nil

}

// hasPerm function is a higher level function wrapping checkPerm so
// acl_stringas to provide recursive functionality
func (fs *FileSystem) hasPerm(n *Node, perm string, recursive bool) error {
	err := fs.checkPerm(n.ACL, perm)
	if err != nil {
		return err
	}

	if n.IsDir() && recursive {
		children, _ := n.List()

		for _, child := range children {

			if child.IsHidden() { // get will not list hidden node
				continue
			}

			err = fs.hasPerm(child, perm, recursive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// hasPermOnParent function will check the permission based on the nodePath
// passed in. It will disregard the last one name in the node path and check
// permission on the closest parent directory node.
func (fs *FileSystem) hasPermOnParent(nodePath string, perm string) error {
	curNode := fs.Root

	components := strings.Split(nodePath, "/")

	// ignore the last node name. We are checking parent directory only
	for i := 1; i < len(components)-1; i++ {
		nodeName := components[i]
		child, ok := curNode.Children[nodeName]

		// We are checking closest parent only, since there's no further node
		// name and directories will be created automatically and ACL will be
		// passed down to those nodes.
		if !ok {
			err := fs.checkPerm(curNode.ACL, perm)
			return err
		}
		curNode = child

	}

	return nil
}
