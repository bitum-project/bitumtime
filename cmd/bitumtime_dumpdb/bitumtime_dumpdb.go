package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitum-project/bitumd/chaincfg"
	"github.com/bitum-project/bitumd/bitumutil"
	"github.com/bitum-project/bitumtime/bitumtimed/backend/filesystem"
)

var (
	defaultHomeDir = bitumutil.AppDataDir("bitumtimed", false)

	destination = flag.String("destination", "", "Restore destination")
	dumpJSON    = flag.Bool("json", false, "Dump JSON")
	restore     = flag.Bool("restore", false, "Restore backend, -destination is required")
	fsRoot      = flag.String("source", "", "Source directory")
	testnet     = flag.Bool("testnet", false, "Use testnet port")
)

func _main() error {
	flag.Parse()

	if *restore {
		if *destination == "" {
			return fmt.Errorf("-destination must be set")
		}

		fs, err := filesystem.NewRestore(*destination)
		if err != nil {
			return err
		}
		defer fs.Close()

		return fs.Restore(os.Stdin, true, *destination)
	}

	root := *fsRoot
	if root == "" {
		root = filepath.Join(defaultHomeDir, "data")
		if *testnet {
			root = filepath.Join(root, chaincfg.TestNetParams.Name)
		} else {
			root = filepath.Join(root, chaincfg.MainNetParams.Name)
		}
	}

	// Dump

	fmt.Printf("=== Root: %v\n", root)

	fs, err := filesystem.NewDump(root)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.Dump(os.Stdout, !*dumpJSON)
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
