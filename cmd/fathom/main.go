// Command fathom orchestrates Go fuzzing campaigns: it discovers targets, runs
// them for a bounded time, tracks execs and coverage, and groups the crashes
// they find — on the command line or behind a live panel.
package main

import (
	"errors"
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fathom: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return usage()
	}
	switch os.Args[1] {
	case "targets":
		return runTargets(os.Args[2:])
	case "run":
		return runRun(os.Args[2:])
	case "campaign":
		return runCampaign(os.Args[2:])
	case "corpus":
		return runCorpus(os.Args[2:])
	case "version":
		fmt.Printf("fathom %s (commit %s, built %s)\n", version, commit, date)
		return nil
	default:
		return usage()
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "usage: fathom <targets|run|campaign|corpus|crashes|replay|serve|version> [args]")
	fmt.Fprintln(os.Stderr, "  fathom targets ./pkg")
	fmt.Fprintln(os.Stderr, "  fathom campaign ./pkg -t 30s")
	fmt.Fprintln(os.Stderr, "  fathom serve -dir ./pkg -listen 127.0.0.1:8080")
	return errors.New("no command given")
}

func dir(args []string, index int) string {
	if len(args) > index {
		return args[index]
	}
	return "."
}
