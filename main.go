package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"./base"
	"./file"
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
	if err := base.InitVarBase(*flVarBase); err != nil {
		log.Fatal(err)
	}

	var (
		hash    = DetermineHash(*flCipher)
		ourbase = base.Base{Path: *flVarBase}
		//infos = []*file.FileHashInfo{}
		//mu    = sync.Mutex{}
		//results := make(chan file.FileHashInfo, 2)
		//wg := sync.WaitGroup{}
	)

	for _, arg := range flag.Args() {
		if m, err := file.SameDevPaths(*flVarBase, arg); err != nil {
			log.Fatal(err)
		} else if !m {
			log.Printf("SKIPPING: %q is not on the same device as %q", arg, *flVarBase)
			continue
		}
		done := make(chan struct{})
		infos := file.HashFileGetter(arg, hash, *flWorkers, done)
		for fi := range infos {
			if fi.Err != nil {
				log.Println(fi.Err)
				done <- struct{}{}
			}
			fmt.Printf("%s  %s\n", fi.Hash, fi.Path)
			if ourbase.HasBlob(fi.HashType, fi.Hash) {
				// TODO check if they have the same Inode
				// if not, then clobber
			} else {
				// TODO hard link to blobs
			}
		}
	}
	//if len(infos) > 0 {
	//fmt.Println("collected", len(infos), "sums")
	//}
}
