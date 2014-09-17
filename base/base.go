package base

import (
	"io"
	"os"
	"path/filepath"
)

func NewBase(path string, hashName string) (*Base, error) {
	root := filepath.Join(path, "dedup")
	for _, p := range []string{"blobs/" + hashName, "state"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0755); err != nil {
			return nil, err
		}
	}
	return &Base{Path: root, HashName: hashName}, nil
}

type Base struct {
	Path     string
	HashName string
}

func (b Base) blobPath(sum string) string {
	if len(sum) < 3 {
		return ""
	}
	return filepath.Join(b.Path, "blobs", b.HashName, sum[0:2], sum)
}

// GetBlob store the content from src, for the sum and hashType
func (b Base) GetBlob(sum string) (io.Reader, error) {
	// XXX
	return nil, nil
}

// PutBlob store the content from src, for the sum and hashType
//
// we take the sum up front to avoid recalculation and tempfiles
func (b Base) PutBlob(sum string, src io.Reader) error {
	// XXX
	return nil
}

// HasBlob tests whether a blob with this sum exists
func (b Base) HasBlob(sum string) bool {
	// XXX
	return true
}
