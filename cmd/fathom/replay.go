package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/cansarihan/fathom/internal/crash"
)

func runReplay(args []string) error {
	flags := flag.NewFlagSet("replay", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("replay needs one corpus file: fathom replay testdata/fuzz/FuzzX/abc")
	}

	data, err := os.ReadFile(flags.Arg(0)) // #nosec G304 -- operator-supplied crash input
	if err != nil {
		return err
	}
	rec, err := crash.DecodeRecord(data)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n%s · %d argument(s)\n", flags.Arg(0), rec.Version, len(rec.Args))
	for i, a := range rec.Args {
		fmt.Printf("  [%d] %s = %q\n", i, a.Type, a.Value)
		if len(a.Bytes) > 0 {
			fmt.Printf("      hex: %s\n", hex.EncodeToString(a.Bytes))
		}
	}
	return nil
}
