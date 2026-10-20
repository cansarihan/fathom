package campaign

import (
	"os"
	"path/filepath"
	"testing"
)

// feeding canned toolchain output exercises the line handler without running go.
func TestStreamerHandlesProgressAndCrash(t *testing.T) {
	dir := t.TempDir()
	inputRel := "testdata/fuzz/FuzzX/abc123"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, inputRel)), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, inputRel), []byte("go test fuzz v1\n[]byte(\"boom\")\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var updates []Update
	result := Result{Target: "FuzzX"}
	s := newStreamer(Options{Dir: dir, Target: "FuzzX"}, func(u Update) { updates = append(updates, u) }, &result)

	for _, line := range []string{
		"fuzz: elapsed: 0s, gathering baseline coverage: 0/50 completed",
		"fuzz: elapsed: 3s, execs: 120000 (40000/sec), new interesting: 7 (total: 57)",
		"--- FAIL: FuzzX (0.01s)",
		"    panic: runtime error: index out of range [9] with length 8",
		"    Failing input written to " + inputRel,
	} {
		s.line(line)
	}

	if result.Execs != 120000 || result.ExecsPerSec != 40000 || result.Corpus != 57 {
		t.Errorf("progress not tracked: %+v", result)
	}
	if !result.Crashed || result.Crash == nil {
		t.Fatal("crash not captured")
	}
	if result.Crash.Signature == "" || result.Crash.Panic == "" {
		t.Errorf("crash lacks signature/panic: %+v", result.Crash)
	}
	if len(updates) == 0 {
		t.Error("no updates emitted")
	}
}
