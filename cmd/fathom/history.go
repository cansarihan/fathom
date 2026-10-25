package main

import (
	"errors"
	"flag"
	"os"

	"github.com/cansarihan/fathom/internal/report"
	"github.com/cansarihan/fathom/internal/store"
)

func runHistory(args []string) error {
	flags := flag.NewFlagSet("history", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("history needs a file: fathom history campaigns.json")
	}
	s, err := store.Open(flags.Arg(0))
	if err != nil {
		return err
	}
	report.History(os.Stdout, s.All())
	return nil
}
