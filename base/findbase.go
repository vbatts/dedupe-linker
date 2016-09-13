package base

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// FindBase steps up the directory tree to find the top-level that is still on
// the same device as the path provided
func FindBase(path string) (string, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return findBaseInfo(stat)
	}

	return FindBase(filepath.Dir(path))
}

func findBaseInfo(stat os.FileInfo) (string, error) {
	dirstat, err := os.Lstat(filepath.Dir(stat.Name()))
	if err != nil {
		return "", err
	}
	if stat.Name() == dirstat.Name() {
		return stat.Name(), nil
	}

	if sameDevice(stat, dirstat) {
		return findBaseInfo(dirstat)
	}
	return stat.Name(), nil
}

func hasPermission(path string) bool {
	stat, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if !stat.IsDir() {
		path = filepath.Dir(path)
	}
	fh, err := ioutil.TempFile(path, "perm.test.")
	if err != nil {
		return false
	}
	os.Remove(fh.Name())
	return true
}

func sameDevice(file1, file2 os.FileInfo) bool {
	sys1 := file1.Sys().(*syscall.Stat_t)
	sys2 := file2.Sys().(*syscall.Stat_t)
	return ((major(sys1.Dev) == major(sys2.Dev)) && (minor(sys1.Dev) == minor(sys2.Dev)))
}

func major(n uint64) uint64 {
	return uint64(n / 256)
}

func minor(n uint64) uint64 {
	return uint64(n % 256)
}
