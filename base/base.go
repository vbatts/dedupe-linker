package base

import (
	"crypto"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/vbatts/dedupe-linker/cryptomap"
	"github.com/vbatts/dedupe-linker/file"
)

func NewBase(path string, hashName string) (*Base, error) {
	root := filepath.Join(path, "dedup")
	for _, p := range []string{"blobs/" + hashName, "state", "tmp"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0755); err != nil && !os.IsExist(err) {
			return nil, err
		}
	}
	return &Base{Path: root, HashName: hashName, Hash: cryptomap.DetermineHash(hashName)}, nil
}

type Base struct {
	Path     string
	HashName string
	Hash     crypto.Hash
}

func (b Base) Stat(sum string) (os.FileInfo, error) {
	return os.Stat(b.blobPath(sum))
}

func (b Base) blobPath(sum string) string {
	if len(sum) < 3 {
		return ""
	}
	return filepath.Join(b.Path, "blobs", b.HashName, sum[0:2], sum)
}

type ReaderSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

func (b Base) SameFile(sum, path string) bool {
	var (
		bInode, dInode uint64
		err            error
	)
	if bInode, err = file.GetInode(b.blobPath(sum)); err != nil {
		return false
	}
	if dInode, err = file.GetInode(path); err != nil {
		return false
	}
	if bInode == dInode {
		return true
	}
	return false

}

// GetBlob store the content from src, for the sum and hashType
func (b Base) GetBlob(sum string) (ReaderSeekerCloser, error) {
	return os.Open(b.blobPath(sum))
}

// PutBlob store the content from src, for the sum and hashType
//
// we take the sum up front to avoid recalculation and tempfiles
func (b Base) PutBlob(src io.Reader, mode os.FileMode) (string, error) {
	fh, err := b.tmpFile()
	if err != nil {
		return "", err
	}
	defer func() {
		fh.Close()
		os.Remove(fh.Name())
	}()

	h := b.Hash.New()
	t := io.TeeReader(src, h)

	if _, err = io.Copy(fh, t); err != nil {
		return "", err
	}

	sum := fmt.Sprintf("%x", h.Sum(nil))
	fi, err := b.Stat(sum)
	if err == nil && fi.Mode().IsRegular() {
		return sum, nil
	}

	if err := os.MkdirAll(filepath.Dir(b.blobPath(sum)), 0755); err != nil && !os.IsExist(err) {
		return sum, err
	}
	destFh, err := os.Create(b.blobPath(sum))
	if err != nil {
		return sum, err
	}
	defer destFh.Close()
	_, err = fh.Seek(0, 0)
	if err != nil {
		return sum, err
	}
	if _, err = io.Copy(destFh, fh); err != nil {
		return sum, err
	}
	return sum, destFh.Chmod(mode)
}

func (b Base) tmpFile() (*os.File, error) {
	return ioutil.TempFile(filepath.Join(b.Path, "tmp"), "put")
}

// Hard link the file from src to the blob for sum
func (b Base) LinkFrom(src, sum string) error {
	if err := os.MkdirAll(filepath.Dir(b.blobPath(sum)), 0756); err != nil && !os.IsExist(err) {
		return err
	}
	return os.Link(src, b.blobPath(sum))
}

func randomString() (string, error) {
	// make a random name
	buf := make([]byte, 10)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}

// SafeLink overrides newname if it already exists. If there is an error in creating the link, the transaction is rolled back
func SafeLink(oldname, newname string) error {
	var backupName string
	// check if newname exists
	if fi, err := os.Stat(newname); err == nil && fi != nil {
		// make a random name
		buf := make([]byte, 5)
		if _, err = rand.Read(buf); err != nil {
			return err
		}
		backupName = fmt.Sprintf("%s.%x", newname, buf)
		// move newname to the random name backupName
		if err = os.Rename(newname, backupName); err != nil {
			return err
		}
	}
	// hardlink oldname to newname
	if err := os.Link(oldname, newname); err != nil {
		// if that failed, and there is a backupName
		if len(backupName) > 0 {
			// then move back the backup
			if err = os.Rename(backupName, newname); err != nil {
				return err
			}
		}
		return err
	}
	// remove the backupName
	if len(backupName) > 0 {
		os.Remove(backupName)
	}
	return nil
}

// Hard link the file for sum to the path at dest
func (b Base) LinkTo(dest, sum string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil && !os.IsExist(err) {
		return err
	}
	err := os.Link(b.blobPath(sum), dest)
	if err != nil && os.IsExist(err) {
		if !b.SameFile(sum, dest) {
			SafeLink(b.blobPath(sum), dest)
			log.Printf("dedupped %q with %q", dest, b.blobPath(sum))
		}
	} else if err != nil {
		return err
	}
	return nil
}

// HasBlob tests whether a blob with this sum exists
func (b Base) HasBlob(sum string) bool {
	fi, err := b.Stat(sum)
	return fi != nil && err == nil
}
