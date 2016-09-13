package base

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHasPermission(t *testing.T) {
	if !hasPermission("/tmp") {
		t.Error("expected to have permission to /tmp, but did not")
	}

	if hasPermission("/") {
		t.Error("expected to not have permission to /, but did")
	}
}

func TestSameDev(t *testing.T) {
	file1, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()
	file2, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()

	stat1, err := file1.Stat()
	if err != nil {
		t.Fatal(err)
	}
	stat2, err := file2.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if !sameDevice(stat1, stat2) {
		t.Errorf("expected the two files to be on same device. But %q and %q are not", file1.Name(), file2.Name())
	} else {
		os.Remove(stat1.Name())
		os.Remove(stat2.Name())
	}
}

func TestNotSameDev(t *testing.T) {
	file1, err := ioutil.TempFile("/tmp", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()
	file2, err := ioutil.TempFile(os.Getenv("HOME"), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()

	stat1, err := file1.Stat()
	if err != nil {
		t.Fatal(err)
	}
	stat2, err := file2.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if sameDevice(stat1, stat2) {
		t.Errorf("expected the two files _not_ to be on same device. But %q and %q are not", file1.Name(), file2.Name())
	} else {
		os.Remove(stat1.Name())
		os.Remove(stat2.Name())
	}
}
