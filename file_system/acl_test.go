package fileSystem

import (
	"testing"
)

func TestACLRead(t *testing.T) {
	fs := New()

	user := "admin"

	// setting up the tree and relevant acl

	_, err := fs.Create("/ACL/acl_name/r/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Create("/sample/gao", "zhengao", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	n, err := fs.InternalGet("/sample/gao", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	n.ACL = "acl_name"

	// begin testing

	err = fs.hasPerm(n, "r", true)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.hasPerm(n, "r", false)
	if err != nil {
		t.Fatal(err)
	}

	e, err := fs.Get("/sample/gao", false, false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "zhengao" {
		t.Fatal(err)
	}

}

func TestACLRecurRead(t *testing.T) {

	fs := New()

	user := "admin"

	// setting up the tree and relevant acl

	_, err := fs.Create("/ACL/acl_name/r/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Create("/ACL/acl_name/w/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.CreateDir("/sample", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	d, err := fs.InternalGet("/sample", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	d.ACL = "acl_name"

	// */sample/gao* now inherits parent */sample*
	_, err = fs.Create("/sample/gao/gao2", "zhengao", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	// begin testing

	d, err = fs.InternalGet("/sample/gao", 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.hasPerm(d, "r", true)
	if err != nil {
		t.Fatal(err)
	}

	n, err := fs.InternalGet("/sample/gao/gao2", 1, 1)
	err = fs.hasPerm(n, "r", false)
	if err != nil {
		t.Fatal(err)
	}

	e, err := fs.Get("/sample/gao/gao2", false, false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "zhengao" {
		t.Fatal(err)
	}

}

func TestCreate(t *testing.T) {
	fs := New()

	user := "admin"

	// setting up the tree and relevant acl

        _, err := fs.Create("/ACL/acl_name/r/" + user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
        _, err = fs.Create("/ACL/acl_name/w/" + user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	fs.Root.ACL = "acl_name"

        // begin testing

        _, err = fs.Create("/a/b/c", "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

        _, err = fs.CreateDir("/a/b2", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
        _, err = fs.Create("/a/b3", "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}
