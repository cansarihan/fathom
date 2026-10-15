package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cansarihan/fathom/internal/logging"
	"github.com/cansarihan/fathom/internal/server"
)

func runServe(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	listen := flags.String("listen", "127.0.0.1:8080", "listen address")
	root := flags.String("dir", ".", "package directory to fuzz")
	duration := flags.Duration("t", 10*time.Second, "fuzz time per target")
	token := flags.String("token", os.Getenv("FATHOM_TOKEN"), "admin token")
	noUI := flags.Bool("no-ui", false, "disable the panel")
	if err := flags.Parse(args); err != nil {
		return err
	}

	logger := logging.New("info", "text", os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler := server.New(server.Options{
		Dir:      *root,
		Duration: *duration,
		Token:    *token,
		UI:       !*noUI,
	}, server.BuildInfo{
		Version:   version,
		Commit:    commit,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}).Handler()

	httpServer := &http.Server{Addr: *listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		logger.Info("fathom listening", "address", *listen, "dir", *root, "ui", !*noUI)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdown)
}
