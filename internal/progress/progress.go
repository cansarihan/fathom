// Package progress parses the live output a `go test -fuzz` run writes while it
// works. Each status line becomes a typed event the orchestrator can aggregate
// and the panel can draw, so Fathom can follow a campaign without owning the
// fuzzing engine itself.
package progress

import "time"

// Kind classifies a parsed line.
type Kind string

const (
	// KindBaseline is emitted while the engine measures coverage of the seed corpus.
	KindBaseline Kind = "baseline"
	// KindProgress is a periodic execs/coverage status line.
	KindProgress Kind = "progress"
	// KindCrash is the line announcing a failing input was found.
	KindCrash Kind = "crash"
	// KindDone marks the campaign finishing cleanly.
	KindDone Kind = "done"
	// KindMinimizing marks the engine shrinking a failing input before it reports.
	KindMinimizing Kind = "minimizing"
)

// Event is one parsed status line from a fuzzing run.
type Event struct {
	Kind        Kind          `json:"kind"`
	Elapsed     time.Duration `json:"elapsed"`
	Execs       int64         `json:"execs"`
	ExecsPerSec int64         `json:"execs_per_sec"`
	Interesting int           `json:"interesting"`
	CorpusTotal int           `json:"corpus_total"`
	BaselineAt  int           `json:"baseline_at"`
	BaselineOf  int           `json:"baseline_of"`
	CrashInput  string        `json:"crash_input,omitempty"`
}
