// Package campaign runs a fuzz target through the Go toolchain for a bounded
// time, turns the engine's live output into a stream of events, and captures
// the failing input when the target crashes.
package campaign

import (
	"time"

	"github.com/cansarihan/fathom/internal/crash"
	"github.com/cansarihan/fathom/internal/progress"
)

// Options configures a single target's run.
type Options struct {
	Dir      string        // package directory to run go test in
	Target   string        // fuzz function name, e.g. FuzzParse
	Duration time.Duration // -fuzztime; zero means a short default
	Workers  int           // -parallel; zero lets the toolchain decide
}

// Update is one streamed step of a run: a parsed status event and, when the run
// just crashed, the captured crash.
type Update struct {
	Event progress.Event `json:"event"`
	Crash *crash.Crash   `json:"crash,omitempty"`
}

// Result is the outcome of a run.
type Result struct {
	Target      string        `json:"target"`
	Elapsed     time.Duration `json:"elapsed"`
	Execs       int64         `json:"execs"`
	ExecsPerSec int64         `json:"execs_per_sec"`
	Corpus      int           `json:"corpus"`
	Interesting int           `json:"interesting"`
	Crashed     bool          `json:"crashed"`
	Crash       *crash.Crash  `json:"crash,omitempty"`
	Err         string        `json:"err,omitempty"`
}
