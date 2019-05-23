package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	v1 "github.com/bitum-project/bitumtime/api/v1"
	"github.com/bitum-project/bitumtime/merkle"
	"github.com/bitum-project/bitumtime/util"
)

var (
	proof       = flag.String("p", "", "Proof file")
	file        = flag.String("f", "", "Original file")
	bitumdataHost = flag.String("h", "", "bitumdata host")
	testnet     = flag.Bool("testnet", false, "Use testnet port")
	verbose     = flag.Bool("v", false, "Verbose")
)

func _main() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "bitumtime_checker [-h {bitumdatahost}|"+
			"-testnet|-v] -f {file} -p {proof}\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// require -f
	if *file == "" {
		return fmt.Errorf("must provide -f")
	}

	// require -p
	if *proof == "" {
		return fmt.Errorf("must provide -p")
	}

	// Ensure proof looks correct
	fProof, err := os.Open(*proof)
	if err != nil {
		return err
	}
	var vr v1.VerifyReply
	decoder := json.NewDecoder(fProof)
	if err := decoder.Decode(&vr); err != nil {
		return fmt.Errorf("Could node decode VerifyReply: %v", err)
	}

	// Handle bitumtime host
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
	// Get file digest
	d, err := util.DigestFile(*file)
	if err != nil {
		return err
	}

	if *verbose {
		fmt.Printf("%v  %v\n", d, *file)
	}

	// Ensure file digest exists in the proof and that the saved answer was
	// correct
	found := -1
	for k, v := range vr.Digests {
		if v.Digest != d {
			continue
		}

		found = k
		break
	}
	if found == -1 {
		return fmt.Errorf("file digest not found in proof")
	}
	v := vr.Digests[found]

	// Verify result of matching digest
	if _, ok := v1.Result[v.Result]; !ok {
		return fmt.Errorf("%v invalid error code %v\n", v.Digest,
			v.Result)
	}

	// Verify merkle path.
	root, err := merkle.VerifyAuthPath(&v.ChainInformation.MerklePath)
	if err != nil {
		if err != merkle.ErrEmpty {
			return fmt.Errorf("%v invalid auth path %v\n",
				v.Digest, err)
		}
		return fmt.Errorf("%v Not anchored\n", v.Digest)
	}

	// Verify merkle root.
	merkleRoot, err := hex.DecodeString(v.ChainInformation.MerkleRoot)
	if err != nil {
		return fmt.Errorf("invalid merkle root: %v\n", err)
	}
	// This is silly since we check against returned root.
	if !bytes.Equal(root[:], merkleRoot) {
		return fmt.Errorf("%v invalid merkle root\n", v.Digest)
	}

	// If we made it here we have a valid proof
	if *verbose {
		fmt.Printf("%v  Proof  OK\n", d)
	}

	// Verify against bitumdata
	err = util.VerifyAnchor(*bitumdataHost,
		vr.Digests[found].ChainInformation.Transaction, root[:])
	if err != nil {
		return err
	}

	if *verbose {
		fmt.Printf("%v  Anchor OK\n", d)
	}

	return nil
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
