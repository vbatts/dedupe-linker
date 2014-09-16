package base

import (
	"io"
	"os"
	"path/filepath"
)

func NewBase(path string, hashName string) (*Base, error) {
	for _, p := range []string{"dedup/blobs" + hashName, "dedup/state"} {
		if err := os.MkdirAll(filepath.Join(path, p), 0755); err != nil {
			return nil, err
		}
	}
	return &Base{Path: path, HashName: hashName}, nil
}

type Base struct {
	Path     string
	HashName string
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
