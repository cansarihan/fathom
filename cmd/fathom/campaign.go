package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/crash"
	"github.com/cansarihan/fathom/internal/discover"
	"github.com/cansarihan/fathom/internal/report"
	"github.com/cansarihan/fathom/internal/schedule"
	"github.com/cansarihan/fathom/internal/store"
)

func runCampaign(args []string) error {
	flags := flag.NewFlagSet("campaign", flag.ContinueOnError)
	total := flags.Duration("t", time.Minute, "total fuzz time across all targets")
	rounds := flags.Int("rounds", 1, "interleave this many passes over the targets")
	history := flags.String("history", "", "append the campaign to a JSON history file")
	asJSON := flags.Bool("json", false, "print JSON instead of a table")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root := dir(flags.Args(), 0)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	targets, err := discover.Discover(ctx, root)
	if err != nil {
		return err
	}
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.Name
	}
	slots := schedule.Even(names, *total)
	if *rounds > 1 {
		slots = schedule.Rounds(names, *total, *rounds)
	}

	start := time.Now()
	var results []campaign.Result
	var crashes []crash.Crash
	for _, slot := range slots {
		result, _ := campaign.Run(ctx, campaign.Options{Dir: root, Target: slot.Target, Duration: slot.Duration}, nil)
		results = append(results, result)
		if result.Crash != nil {
			crashes = append(crashes, *result.Crash)
		}
	}

	summary := report.Summary{
		Dir:     root,
		Elapsed: time.Since(start).Round(time.Second),
		Results: results,
		Crashes: crash.Group(crashes),
	}

	if *history != "" {
		if s, err := store.Open(*history); err == nil {
			_ = s.Append(store.Record{
				ID: time.Now().Format("20060102T150405"), Started: start, Finished: time.Now(),
				Dir: root, Results: results, Crashes: len(summary.Crashes),
			})
		}
	}

	if *asJSON {
		return report.JSON(os.Stdout, summary)
	}
	report.Table(os.Stdout, summary)
	return nil
}
