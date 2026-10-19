package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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
	records := s.All()
	if len(records) == 0 {
		fmt.Println("no campaigns recorded")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "WHEN\tDIR\tTARGETS\tCRASHES")
	for _, r := range records {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n",
			r.Started.Format(time.RFC3339), r.Dir, len(r.Results), r.Crashes)
	}
	return tw.Flush()
}
