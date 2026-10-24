package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cansarihan/fathom/internal/crash"
)

func runCrashes(args []string) error {
	flags := flag.NewFlagSet("crashes", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	root := filepath.Join(dir(flags.Args(), 0), "testdata", "fuzz")

	inputs, err := crash.LoadInputs(root)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		fmt.Printf("no recorded crashes under %s\n", root)
		return nil
	}
	for _, in := range inputs {
		fmt.Printf("%s/%s\n  %s\n", in.Target, filepath.Base(in.Path), preview(in.Record))
	}
	return nil
}

func preview(rec crash.Record) string {
	parts := make([]string, 0, len(rec.Args))
	for _, a := range rec.Args {
		v := a.Value
		if len(v) > 48 {
			v = v[:48] + "…"
		}
		parts = append(parts, a.Type+"("+v+")")
	}
	return strings.Join(parts, ", ")
}
