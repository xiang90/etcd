package fileSystem

import (
	"testing"
)

func TestReadPerm(t *testing.T) {
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
		t.Fatal("Get is wrong")
	}

}

func TestRecurReadPerm(t *testing.T) {

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

func TestCreatePerm(t *testing.T) {
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

func TestUpdatePerm(t *testing.T) {
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
	e, err := fs.Get("/sample/gao", false, false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "zhengao" {
		t.Fatal("Get is wrong")
	}

	e, err = fs.Update("/sample/gao", "gaozhen", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "gaozhen" {
		t.Fatal("Update is wrong")
	}
}

func TestDeletePerm(t *testing.T) {
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

	n, err := fs.InternalGet("/sample/", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	n.ACL = "acl_name"

	// begin testing
	_, err = fs.Delete("/sample/gao", true, 1, 1)
	if err == nil {
		t.Fatal(err)
	}
	_, err = fs.Create("/ACL/acl_name/w/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Delete("/sample/gao", true, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecurDeletePerm(t *testing.T) {
	fs := New()
	user := "admin"

	// setting up the tree and relevant acl

	_, err := fs.Create("/ACL/acl_name/r/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.CreateDir("/sample/", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	n, err := fs.InternalGet("/sample/", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	n.ACL = "acl_name"

	_, err = fs.Create("/sample/gao/mao", "zhengao", Permanent, 1, 1)
	if err == nil {
		t.Fatal("expect to get an error")
	}
	_, err = fs.InternalCreate("/sample/gao/mao", "zhengao", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	// begin testing
	_, err = fs.Delete("/sample/gao", true, 1, 1)
	if err == nil {
		t.Fatal(err)
	}
	e, err := fs.Get("/sample/gao/mao", false, false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "zhengao" {
		t.Fatal("/sample/gao/mao value is wrong")
	}

	_, err = fs.Create("/ACL/acl_name/w/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Delete("/sample/gao", true, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	e, err = fs.Get("/sample/gao/mao", false, false, 1, 1)
	if err == nil {
		t.Fatal("except to get an error here")
	}
}

func TestTestAndSetPerm(t *testing.T) {
	fs := New()
	user := "admin"

	// setting up the tree and relevant acl

	_, err := fs.Create("/ACL/acl_name/r/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	fs.Create("/foo", "bar", Permanent, 1, 1)

	n, err := fs.InternalGet("/foo", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	n.ACL = "acl_name"

	_, err = fs.TestAndSet("/foo", "bar", 0, "car", Permanent, 2, 1)
	if err == nil {
		t.Fatal("test and set should fail without write permission")
	}

	_, err = fs.Create("/ACL/acl_name/w/"+user, "1", Permanent, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.TestAndSet("/foo", "bar", 0, "car", Permanent, 2, 1)
	if err != nil {
		t.Fatal(err)
	}

}
