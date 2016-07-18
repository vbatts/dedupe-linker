package walker

import (
	"github.com/vbatts/dedupe-linker/base"
)

type Walker struct {
	Base *base.Base
}

func (w Walker) Walk(path string, quit chan int) error {

	select {
	case <-quit:
		return nil
	}
	return nil
}
