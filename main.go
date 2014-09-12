package main

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	flVarBase = flag.String("b", filepath.Join(os.Getenv("HOME"), "var"), "base directory where files are duplicated")
	flCipher  = flag.String("c", "sha1", "block cipher to use (sha1, or sha256)")
	flWorkers = flag.Int("w", 2, "workers to do summing")
)

func init() {
	// give ourselves a little wiggle room
	if runtime.NumCPU() > 1 && len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(2)
	}
}

func main() {
	flag.Parse()
	if err := InitVarBase(*flVarBase); err != nil {
		log.Fatal(err)
	}

	var (
		hash crypto.Hash
		//infos = []*FileHashInfo{}
		//mu    = sync.Mutex{}
		//results := make(chan FileHashInfo, 2)
		//wg := sync.WaitGroup{}
	)

	switch strings.ToLower(*flCipher) {
	case "md5":
		hash = crypto.MD5
	case "sha1":
		hash = crypto.SHA1
	case "sha224":
		hash = crypto.SHA224
	case "sha256":
		hash = crypto.SHA256
	case "sha384":
		hash = crypto.SHA384
	case "sha512":
		hash = crypto.SHA512
	default:
		log.Fatalf("ERROR: unknown cipher %q", *flCipher)
	}

	for _, arg := range flag.Args() {
		if m, err := SameDevPaths(*flVarBase, arg); err != nil {
			log.Fatal(err)
		} else if !m {
			log.Printf("SKIPPING: %q is not on the same device as %q", arg, *flVarBase)
			continue
		}
		done := make(chan struct{})
		infos := HashFileGetter(arg, hash, *flWorkers, done)
		for fi := range infos {
			if fi.Err != nil {
				log.Println(fi.Err)
				done <- struct{}{}
			}
			fmt.Printf("%x  %s\n", fi.Hash, fi.Path)
		}
	}
	//if len(infos) > 0 {
	//fmt.Println("collected", len(infos), "sums")
	//}
}

func InitVarBase(base string) error {
	for _, path := range []string{"dedup/blobs", "dedup/state"} {
		if err := os.MkdirAll(filepath.Join(base, path), 0755); err != nil {
			return err
		}
	}
	return nil
}

type FileHashInfo struct {
	HashType crypto.Hash
	Hash     []byte
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
	fhi.Hash = h.Sum(nil)
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
