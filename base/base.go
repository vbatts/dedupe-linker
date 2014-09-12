package base

import (
	"crypto"
	"os"
	"path/filepath"
)

func InitVarBase(base string) error {
	for _, path := range []string{"dedup/blobs", "dedup/state"} {
		if err := os.MkdirAll(filepath.Join(base, path), 0755); err != nil {
			return err
		}
	}
	return nil
}

type Base struct {
	Path string
}

func (b Base) HasBlob(hashType crypto.Hash, sum string) bool {
	return true
}
