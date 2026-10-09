package campaign

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cansarihan/fathom/internal/crash"
	"github.com/cansarihan/fathom/internal/progress"
)

const defaultDuration = 10 * time.Second

// Run executes one fuzz target and streams updates to emit (which may be nil).
// It returns when the target crashes, the duration elapses, or ctx is cancelled.
func Run(ctx context.Context, opts Options, emit func(Update)) (Result, error) {
	if opts.Duration <= 0 {
		opts.Duration = defaultDuration
	}
	result := Result{Target: opts.Target}

	args := []string{
		"test",
		"-run", "^$",
		"-fuzz", "^" + opts.Target + "$",
		"-fuzztime", opts.Duration.String(),
	}
	if opts.Workers > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", opts.Workers))
	}
	args = append(args, ".")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = opts.Dir
	pipeRead, pipeWrite := io.Pipe()
	cmd.Stdout = pipeWrite
	cmd.Stderr = pipeWrite

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return result, err
	}
	go func() { _ = cmd.Wait(); _ = pipeWrite.Close() }()

	stream := newStreamer(opts, emit, &result)
	scanner := bufio.NewScanner(pipeRead)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		stream.line(scanner.Text())
	}
	result.Elapsed = time.Since(start).Round(time.Millisecond)
	return result, nil
}

// streamer carries the per-run state the line handler needs.
type streamer struct {
	opts   Options
	emit   func(Update)
	result *Result
	fault  []string
	inFail bool
}

func newStreamer(opts Options, emit func(Update), result *Result) *streamer {
	return &streamer{opts: opts, emit: emit, result: result}
}

func (s *streamer) line(raw string) {
	if strings.Contains(raw, "--- FAIL") {
		s.inFail = true
	}
	if s.inFail && len(s.fault) < 64 {
		s.fault = append(s.fault, strings.TrimSpace(raw))
	}

	event, ok := progress.ParseLine(raw)
	if !ok {
		return
	}
	update := Update{Event: event}
	switch event.Kind {
	case progress.KindProgress:
		s.result.Execs = event.Execs
		s.result.ExecsPerSec = event.ExecsPerSec
		s.result.Corpus = event.CorpusTotal
		s.result.Interesting = event.Interesting
	case progress.KindCrash:
		c := s.capture(event.CrashInput)
		s.result.Crashed = true
		s.result.Crash = c
		update.Crash = c
	}
	if s.emit != nil {
		s.emit(update)
	}
}

func (s *streamer) capture(inputRef string) *crash.Crash {
	path := inputRef
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.opts.Dir, inputRef)
	}
	data, _ := os.ReadFile(path) // #nosec G304 -- path comes from the toolchain's own output
	c := crash.New(s.opts.Target, inputRef, data, strings.Join(s.fault, "\n"))
	return &c
}
