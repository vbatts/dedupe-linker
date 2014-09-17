package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"./base"
	"./cryptomap"
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

	// TODO the *flCipher has not been checked yet, and would cause the directory to get created
	ourbase, err := base.NewBase(*flVarBase, *flCipher)
	if err != nil {
		log.Fatal(err)
	}

	var (
		hash = cryptomap.DetermineHash(*flCipher)
		//infos = []*file.FileHashInfo{}
		//results := make(chan file.FileHashInfo, 2)
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
			if ourbase.HasBlob(fi.Hash) {
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
