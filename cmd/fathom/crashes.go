package main

import (
	"flag"
	"fmt"
	"os"
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

	found := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path) // #nosec G304 -- walking the target's own testdata
		if readErr != nil {
			return nil
		}
		rec, decodeErr := crash.DecodeRecord(data)
		if decodeErr != nil {
			return nil
		}
		target := filepath.Base(filepath.Dir(path))
		found++
		fmt.Printf("%s/%s\n  %s\n", target, filepath.Base(path), preview(rec))
		return nil
	})
	if err != nil {
		return err
	}
	if found == 0 {
		fmt.Printf("no recorded crashes under %s\n", root)
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
