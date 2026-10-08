package schedule

import (
	"testing"
	"time"
)

func sum(slots []Slot) time.Duration {
	var total time.Duration
	for _, s := range slots {
		total += s.Duration
	}
	return total
}

func TestEvenSumsToTotal(t *testing.T) {
	slots := Even([]string{"a", "b", "c"}, 10*time.Second)
	if len(slots) != 3 {
		t.Fatalf("slots = %d", len(slots))
	}
	if sum(slots) != 10*time.Second {
		t.Errorf("sum = %v, want 10s", sum(slots))
	}
}

func TestWeightedProportions(t *testing.T) {
	slots := Weighted([]string{"a", "b"}, map[string]float64{"a": 3, "b": 1}, 8*time.Second)
	if sum(slots) != 8*time.Second {
		t.Errorf("sum = %v, want 8s", sum(slots))
	}
	if slots[0].Duration <= slots[1].Duration {
		t.Errorf("weighted slot a (%v) should exceed b (%v)", slots[0].Duration, slots[1].Duration)
	}
}

func TestWeightedFallsBackToEqual(t *testing.T) {
	slots := Weighted([]string{"a", "b"}, nil, 8*time.Second)
	if slots[0].Duration != slots[1].Duration {
		t.Errorf("missing weights should split evenly: %v vs %v", slots[0].Duration, slots[1].Duration)
	}
}

func TestEmptyInputs(t *testing.T) {
	if Even(nil, time.Second) != nil || Even([]string{"a"}, 0) != nil {
		t.Error("empty inputs should yield no slots")
	}
}
