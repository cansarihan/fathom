package progress

import (
	"testing"
	"time"
)

func TestParseProgress(t *testing.T) {
	line := "fuzz: elapsed: 3s, execs: 325017 (108336/sec), new interesting: 12 (total: 132)"
	e, ok := ParseLine(line)
	if !ok {
		t.Fatal("progress line not parsed")
	}
	if e.Kind != KindProgress {
		t.Errorf("kind = %q", e.Kind)
	}
	if e.Elapsed != 3*time.Second {
		t.Errorf("elapsed = %v", e.Elapsed)
	}
	if e.Execs != 325017 || e.ExecsPerSec != 108336 {
		t.Errorf("execs = %d (%d/sec)", e.Execs, e.ExecsPerSec)
	}
	if e.Interesting != 12 || e.CorpusTotal != 132 {
		t.Errorf("interesting=%d total=%d", e.Interesting, e.CorpusTotal)
	}
}

func TestParseBaseline(t *testing.T) {
	e, ok := ParseLine("fuzz: elapsed: 0s, gathering baseline coverage: 45/120 completed")
	if !ok || e.Kind != KindBaseline {
		t.Fatalf("baseline not parsed: %+v ok=%v", e, ok)
	}
	if e.BaselineAt != 45 || e.BaselineOf != 120 {
		t.Errorf("baseline %d/%d", e.BaselineAt, e.BaselineOf)
	}
}

func TestParseCrash(t *testing.T) {
	e, ok := ParseLine("    Failing input written to testdata/fuzz/FuzzParse/582528e")
	if !ok || e.Kind != KindCrash {
		t.Fatalf("crash not parsed: %+v", e)
	}
	if e.CrashInput != "testdata/fuzz/FuzzParse/582528e" {
		t.Errorf("crash input = %q", e.CrashInput)
	}
}

func TestParseDone(t *testing.T) {
	for _, l := range []string{"PASS", "ok  	example/pkg	10.123s"} {
		e, ok := ParseLine(l)
		if !ok || e.Kind != KindDone {
			t.Errorf("done not parsed for %q: %+v", l, e)
		}
	}
}

func TestParseIgnoresNoise(t *testing.T) {
	for _, l := range []string{"", "random log line", "=== RUN   FuzzParse"} {
		if _, ok := ParseLine(l); ok {
			t.Errorf("unexpectedly parsed %q", l)
		}
	}
}
