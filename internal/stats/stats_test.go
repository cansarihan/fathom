package stats

import (
	"strings"
	"testing"
)

func TestSeriesRingEviction(t *testing.T) {
	s := NewSeries(3)
	for _, v := range []float64{1, 2, 3, 4, 5} {
		s.Push(v)
	}
	got := s.Snapshot()
	if len(got) != 3 || got[0] != 3 || got[2] != 5 {
		t.Fatalf("snapshot = %v, want [3 4 5]", got)
	}
	if s.Last() != 5 {
		t.Errorf("last = %v, want 5", s.Last())
	}
	if s.Max() != 5 {
		t.Errorf("max = %v, want 5", s.Max())
	}
}

func TestSparklineWidthAndShape(t *testing.T) {
	out := Sparkline([]float64{0, 1, 2, 3, 4, 5, 6, 7}, 8)
	if len([]rune(out)) != 8 {
		t.Fatalf("width = %d, want 8", len([]rune(out)))
	}
	if !strings.HasPrefix(out, "▁") || !strings.HasSuffix(out, "█") {
		t.Errorf("sparkline = %q, want rising ramp", out)
	}
	if got := Sparkline(nil, 4); len([]rune(got)) != 4 {
		t.Errorf("empty sparkline width = %d, want 4", len([]rune(got)))
	}
}
