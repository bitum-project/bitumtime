package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitum-project/bitumd/chaincfg"
	"github.com/bitum-project/bitumd/bitumutil"
	"github.com/bitum-project/bitumtime/bitumtimed/backend"
	"github.com/bitum-project/bitumtime/bitumtimed/backend/filesystem"
)

var (
	defaultHomeDir = bitumutil.AppDataDir("bitumtimed", false)

	file        = flag.String("file", "", "journal of modifications if used (will be written despite -fix)")
	fix         = flag.Bool("fix", false, "Try to correct correctable failures")
	bitumdataHost = flag.String("host", "", "bitumdata block explorer")
	printHashes = flag.Bool("printhashes", false, "Print all hashes")
	fsRoot      = flag.String("source", "", "Source directory")
	testnet     = flag.Bool("testnet", false, "Use testnet port")
	verbose     = flag.Bool("v", false, "Print more information during run")
)

func _main() error {
	flag.Parse()

	root := *fsRoot
	if root == "" {
		root = filepath.Join(defaultHomeDir, "data")
		if *testnet {
			root = filepath.Join(root, chaincfg.TestNetParams.Name)
		} else {
			root = filepath.Join(root, chaincfg.MainNetParams.Name)
		}
	}

	if *bitumdataHost == "" {
		if *testnet {
			*bitumdataHost = "https://testnet.bitum.io/api/tx/"
		} else {
			*bitumdataHost = "https://explorer.bitum.io/api/tx/"
		}
	} else {
		if !strings.HasSuffix(*bitumdataHost, "/") {
			*bitumdataHost += "/"
		}
	}

	fmt.Printf("=== Root: %v\n", root)

	fs, err := filesystem.NewDump(root)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.Fsck(&backend.FsckOptions{
		Verbose:     *verbose,
		PrintHashes: *printHashes,
		Fix:         *fix,
		URL:         *bitumdataHost,
		File:        *file,
	})
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
