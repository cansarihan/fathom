package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cansarihan/fathom/internal/discover"
)

func runTargets(args []string) error {
	flags := flag.NewFlagSet("targets", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a list")
	if err := flags.Parse(args); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	targets, err := discover.Discover(ctx, dir(flags.Args(), 0))
	if err != nil {
		return err
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(targets)
	}
	for _, t := range targets {
		fmt.Println(t.Name)
	}
	return nil
}
