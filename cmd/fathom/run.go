package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/progress"
)

func runRun(args []string) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	duration := flags.Duration("t", 30*time.Second, "fuzz time for the target")
	workers := flags.Int("p", 0, "parallel workers (0 lets the toolchain decide)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() < 2 {
		return errors.New("run needs a directory and a target: fathom run ./pkg FuzzName")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	result, err := campaign.Run(ctx, campaign.Options{
		Dir:      flags.Arg(0),
		Target:   flags.Arg(1),
		Duration: *duration,
		Workers:  *workers,
	}, func(u campaign.Update) {
		e := u.Event
		switch e.Kind {
		case progress.KindProgress:
			fmt.Fprintf(os.Stderr, "\r%-6s execs %-10d %6d/s  corpus %-5d new %-3d",
				e.Elapsed.Round(time.Second), e.Execs, e.ExecsPerSec, e.CorpusTotal, e.Interesting)
		case progress.KindCrash:
			fmt.Fprintf(os.Stderr, "\ncrash: %s\n", e.CrashInput)
		}
	})
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return err
	}

	if result.Crashed && result.Crash != nil {
		fmt.Printf("CRASH %s\n  input: %s\n  %s\n", result.Target, result.Crash.InputPath, result.Crash.Signature)
		return nil
	}
	fmt.Printf("clean %s · %d execs · %d corpus\n", result.Target, result.Execs, result.Corpus)
	return nil
}
