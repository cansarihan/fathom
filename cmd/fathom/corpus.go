package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cansarihan/fathom/internal/corpus"
)

func runCorpus(args []string) error {
	flags := flag.NewFlagSet("corpus", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a summary")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root := filepath.Join(dir(flags.Args(), 0), "testdata", "fuzz")

	entries, err := corpus.Load(root)
	if err != nil {
		return err
	}
	stats := corpus.Summarize(entries)

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Dir     string         `json:"dir"`
			Stats   corpus.Stats   `json:"stats"`
			Entries []corpus.Entry `json:"entries"`
		}{root, stats, entries})
	}

	fmt.Printf("%s\n%d files · %d unique · %d bytes · largest %d\n",
		root, stats.Count, stats.Unique, stats.Bytes, stats.Largest)
	return nil
}
