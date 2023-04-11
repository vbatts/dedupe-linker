package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vbatts/dedupe-linker/base"
	"github.com/vbatts/dedupe-linker/cryptomap"
	"github.com/vbatts/dedupe-linker/file"
)

var (
	varBaseDir = filepath.Join(os.Getenv("HOME"), ".dedupe-linker/")

	flVarBase = flag.String("b", varBaseDir, "base directory where files are duplicated")
	flCipher  = flag.String("c", cryptomap.DefaultCipher, "block cipher to use (sha1, or sha256)")
	flWorkers = flag.Int("w", 2, "workers to do summing")
	flNoop    = flag.Bool("noop", false, "don't do any moving or linking")
	flDebug   = flag.Bool("debug", false, "enable debug output")
)

func init() {
	// give ourselves a little wiggle room
	if runtime.NumCPU() > 1 && len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(2)
	}

	if os.Getenv("VARBASEDIR") != "" {
		varBaseDir = filepath.Clean(os.Getenv("VARBASEDIR"))
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\nThis app will make hardlinks of a directory tree, to a content addressible collection in the 'basedir'\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	if os.Getenv("DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "VARBASEDIR=%q\n", *flVarBase)
	}

	// TODO the *flCipher has not been checked yet, and would cause the directory to get created
	ourbase, err := base.NewBase(*flVarBase, *flCipher)
	if err != nil {
		log.Fatal(err)
	}

	var (
		hash = cryptomap.DetermineHash(*flCipher)
		//infos = []*file.HashInfo{}
		//results := make(chan file.HashInfo, 2)
	)

	for _, arg := range flag.Args() {
		if !*flNoop {
			if m, err := file.SameDevPaths(*flVarBase, arg); err != nil {
				log.Fatal(err)
			} else if !m {
				log.Printf("SKIPPING: %q is not on the same device as %q", arg, *flVarBase)
				continue
			}
		}
		done := make(chan struct{})
		infos := file.HashFileGetter(arg, hash, *flWorkers, done)
		for fi := range infos {
			if fi.Err != nil {
				log.Println(fi.Err)
				//done <- struct{}{}
			}
			if *flNoop {
				fmt.Printf("%s  [%d]  %s\n", fi.Hash, fi.Size, fi.Path)
			} else {
				if os.Getenv("DEBUG") != "" {
					fmt.Printf("[%s] exists? %t (%q)\n", fi.Hash, ourbase.HasBlob(fi.Hash), fi.Path)
				}
				if ourbase.HasBlob(fi.Hash) && !ourbase.SameFile(fi.Hash, fi.Path) {
					if err := ourbase.LinkTo(fi.Path, fi.Hash); err != nil {
						log.Println("ERROR-1", err)
					}
				} else if !ourbase.HasBlob(fi.Hash) {
					if err := ourbase.LinkFrom(fi.Path, fi.Hash); err != nil {
						log.Println("ERROR-2", err)
					}
				}
			}
		}
	}
	//if len(infos) > 0 {
	//fmt.Println("collected", len(infos), "sums")
	//}
}
