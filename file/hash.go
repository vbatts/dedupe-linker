package file

import (
	"crypto"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HashInfo for tracking the information regarding a file, it's checksum
// and status.
// If Err is set then the caller must take an appropriate action.
type HashInfo struct {
	HashType crypto.Hash
	Hash     string
	Path     string
	Size     int64
	ModTime  time.Time
	Err      error
}

// HashFileGetter walks the provided `path` with `workers` number of threads.
// The channel of HashInfo are for each regular file encountered.
func HashFileGetter(path string, hash crypto.Hash, ignoreSuffixes []string, workers int, done <-chan struct{}) <-chan HashInfo {
	out := make(chan HashInfo, workers)
	go func() {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			for _, suff := range ignoreSuffixes {
				if os.Getenv("DEBUG") != "" {
					fmt.Printf("[DEBUG] path: %q ; suff: %q\n", filepath.Clean(path), filepath.Clean(suff))
				}
				if strings.HasSuffix(filepath.Clean(path), filepath.Clean(suff)) {
					return filepath.SkipDir
				}
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			fhi := hashFile(path, hash, info)
			out <- *fhi
			select {
			case <-done:
				return fmt.Errorf("walk canceled")
			default:
				return nil
			}
		})
		if err != nil {
			out <- HashInfo{Err: err}
		}
		close(out)
	}()
	return out
}

func hashFile(path string, hash crypto.Hash, info os.FileInfo) *HashInfo {
	fhi := HashInfo{HashType: hash, Path: path, ModTime: info.ModTime(), Size: info.Size()}
	h := hash.New()
	fh, err := os.Open(path)
	if err != nil {
		fhi.Err = err
		return &fhi
	}
	if _, err = io.Copy(h, fh); err != nil {
		fhi.Err = err
		fh.Close()
		return &fhi
	}
	fh.Close()
	fhi.Hash = fmt.Sprintf("%x", h.Sum(nil))
	return &fhi
}
