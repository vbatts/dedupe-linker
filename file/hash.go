package file

import (
	"crypto"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type FileHashInfo struct {
	HashType crypto.Hash
	Hash     string
	Path     string
	ModTime  time.Time
	Err      error
}

func HashFileGetter(path string, hash crypto.Hash, workers int, done <-chan struct{}) <-chan FileHashInfo {
	out := make(chan FileHashInfo, workers)
	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() {
				fhi := hashFile(path, hash, info)
				select {
				case out <- *fhi:
				case <-done:
				}
				wg.Done()
			}()
			select {
			case <-done:
				return fmt.Errorf("walk canceled")
			default:
				return nil
			}
		})
		if err != nil {
			out <- FileHashInfo{Err: err}
		}
		go func() {
			wg.Wait()
			close(out)
		}()
	}()
	return out
}

func hashFile(path string, hash crypto.Hash, info os.FileInfo) *FileHashInfo {
	fhi := FileHashInfo{HashType: hash, Path: path, ModTime: info.ModTime()}
	h := hash.New()
	fh, err := os.Open(path)
	if err != nil {
		fhi.Err = err
		return &fhi
	}
	if _, err = io.Copy(h, fh); err != nil {
		fhi.Err = err
		return &fhi
	}
	fh.Close()
	fhi.Hash = fmt.Sprintf("%x", h.Sum(nil))
	return &fhi
}

// SameInodePaths checks whether path1 and path2 are the same inode
func SameInodePaths(path1, path2 string) (match bool, err error) {
	var inode1, inode2 uint64
	if inode1, err = GetInode(path1); err != nil {
		return false, err
	}
	if inode2, err = GetInode(path2); err != nil {
		return false, err
	}
	return inode1 == inode2, nil
}

// SameInodePaths checks whether path1 and path2 are on the same device
func SameDevPaths(path1, path2 string) (match bool, err error) {
	var dev1, dev2 uint64
	if dev1, err = GetDev(path1); err != nil {
		return false, err
	}
	if dev2, err = GetDev(path2); err != nil {
		return false, err
	}
	return dev1 == dev2, nil
}

func FormatDev(stat *syscall.Stat_t) string {
	return fmt.Sprintf("%d:%d", MajorDev(stat.Dev), MinorDev(stat.Dev))
}

func MajorDev(dev uint64) uint64 {
	return (((dev >> 8) & 0xfff) | ((dev >> 32) & ^uint64(0xfff)))
}

func MinorDev(dev uint64) uint64 {
	return ((dev & 0xff) | ((dev >> 12) & ^uint64(0xff)))
}

func GetStat(path string) (*syscall.Stat_t, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return fi.Sys().(*syscall.Stat_t), nil
}

// GetInode returns the inode for path
func GetInode(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Ino, nil
}

// GetDev returns the device for path
func GetDev(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Dev, nil
}

// GetNlink returns the number of links for path
func GetNlink(path string) (uint64, error) {
	stat, err := GetStat(path)
	if err != nil {
		return 0, err
	}
	return stat.Nlink, nil
}
