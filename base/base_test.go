package base

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSumPath(t *testing.T) {
	expected := "/var/dedup/blobs/sha1/de/deadbeef"
	b := Base{Path: "/var/dedup", HashName: "sha1"}
	if bp := b.blobPath("deadbeef"); bp != expected {
		t.Errorf("expected %q, got %q", expected, bp)
	}
}

func TestRand(t *testing.T) {
	randmap := map[string]bool{}
	for i := 0; i < 100; i++ {
		r, err := randomString()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := randmap[r]; ok {
			t.Errorf("expected no duplicates, but %q is a dup random string", r)
		}
		randmap[r] = true
	}
}

func TestGetPut(t *testing.T) {
	var (
		srcDir, destDir string
		err             error
	)
	if srcDir, err = ioutil.TempDir("", "dedupe-linker-src"); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)
	if destDir, err = ioutil.TempDir("", "dedupe-linker-dest"); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	b, err := NewBase(destDir, "sha1")
	if err != nil {
		t.Fatal(err)
	}

	rHash := "8f074e76e82ae6156c451019840a6f857bbe5157"
	rMsg := "this is the dead beef"

	r := bytes.NewReader([]byte(rMsg))
	sum, err := b.PutBlob(r, 0666)
	if err != nil {
		t.Error(err)
	}
	if sum != rHash {
		t.Errorf("expected %q; got %q", rHash, sum)
	}

	fi, err := b.Stat(rHash)
	if err != nil {
		t.Error(err)
	}
	if fi == nil {
		t.Fatal("did not find the blob " + rHash)
	}
	//fmt.Printf("%#v\n", fi.Sys())

	if err = b.LinkTo(path.Join(srcDir, "beef1.txt"), rHash); err != nil {
		t.Error(err)
	}
	fi2, err := os.Stat(path.Join(srcDir, "beef1.txt"))
	if err != nil {
		t.Error(err)
	}
	if fi2 == nil {
		t.Fatal("did not find the linked file " + path.Join(srcDir, "beef1.txt"))
	}
	//fmt.Printf("%#v\n", fi2.Sys())

	if err = b.LinkTo(path.Join(srcDir, "beef1.txt"), rHash); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	if rHash != sum {
		t.Errorf("expected %s; got %s", rHash, sum)
	}
}
