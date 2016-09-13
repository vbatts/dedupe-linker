// Package walker is a work-in-progress
package walker

import (
	"github.com/vbatts/dedupe-linker/base"
)

type walker struct {
	Base *base.Base
}

func (w walker) Walk(path string, quit chan int) error {
	// XXX what is going on here?
	select {
	case <-quit:
		close(quit)
	}
	return nil
}
