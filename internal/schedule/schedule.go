// Package schedule divides a campaign's total time budget across its fuzz
// targets. An even split is the default; a weighted split lets a target that is
// still finding coverage earn a longer turn on the next round.
package schedule

import "time"

// Slot is one target's turn in a campaign round.
type Slot struct {
	Target   string        `json:"target"`
	Duration time.Duration `json:"duration"`
}

// Even splits total equally across the targets, handing any sub-nanosecond
// remainder to the first slot so the slots always sum to exactly total.
func Even(targets []string, total time.Duration) []Slot {
	if len(targets) == 0 || total <= 0 {
		return nil
	}
	each := total / time.Duration(len(targets))
	remainder := total - each*time.Duration(len(targets))
	slots := make([]Slot, len(targets))
	for i, t := range targets {
		d := each
		if i == 0 {
			d += remainder
		}
		slots[i] = Slot{Target: t, Duration: d}
	}
	return slots
}

// Weighted splits total in proportion to each target's weight. Targets missing
// from weights, or with a non-positive weight, fall back to an equal share so a
// new target is never starved.
func Weighted(targets []string, weights map[string]float64, total time.Duration) []Slot {
	if len(targets) == 0 || total <= 0 {
		return nil
	}
	sum := 0.0
	effective := make([]float64, len(targets))
	for i, t := range targets {
		w := weights[t]
		if w <= 0 {
			w = 1
		}
		effective[i] = w
		sum += w
	}
	slots := make([]Slot, len(targets))
	assigned := time.Duration(0)
	for i, t := range targets {
		d := time.Duration(float64(total) * effective[i] / sum)
		slots[i] = Slot{Target: t, Duration: d}
		assigned += d
	}
	slots[0].Duration += total - assigned
	return slots
}
