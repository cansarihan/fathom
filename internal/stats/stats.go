// Package stats keeps the time series a running campaign produces — execution
// rate over time, coverage growth — in a fixed-size ring, and renders a compact
// sparkline for the terminal and the panel.
package stats

import "sync"

// Series is a bounded ring of samples. The oldest sample is dropped once the
// ring is full, so a long campaign keeps a constant footprint.
type Series struct {
	mu     sync.Mutex
	values []float64
	cap    int
	head   int
	size   int
}

// NewSeries returns a series that remembers the most recent capacity samples.
func NewSeries(capacity int) *Series {
	if capacity < 1 {
		capacity = 1
	}
	return &Series{values: make([]float64, capacity), cap: capacity}
}

// Push appends a sample.
func (s *Series) Push(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[s.head] = v
	s.head = (s.head + 1) % s.cap
	if s.size < s.cap {
		s.size++
	}
}

// Snapshot returns the samples in chronological order.
func (s *Series) Snapshot() []float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]float64, s.size)
	start := (s.head - s.size + s.cap) % s.cap
	for i := 0; i < s.size; i++ {
		out[i] = s.values[(start+i)%s.cap]
	}
	return out
}

// Last returns the most recent sample, or zero when the series is empty.
func (s *Series) Last() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.size == 0 {
		return 0
	}
	return s.values[(s.head-1+s.cap)%s.cap]
}

// Max returns the largest sample seen in the ring.
func (s *Series) Max() float64 {
	max := 0.0
	for _, v := range s.Snapshot() {
		if v > max {
			max = v
		}
	}
	return max
}

// Mean returns the average of the samples in the ring, or zero when empty.
func (s *Series) Mean() float64 {
	values := s.Snapshot()
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
