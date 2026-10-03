package stats

import "strings"

var ramp = []rune("▁▂▃▄▅▆▇█")

// Sparkline renders a series as a single row of block characters scaled to the
// maximum sample, trimmed or left-padded to width columns.
func Sparkline(values []float64, width int) string {
	if width < 1 {
		width = 1
	}
	if len(values) > width {
		values = values[len(values)-width:]
	}
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	var b strings.Builder
	for i := 0; i < width-len(values); i++ {
		b.WriteRune(' ')
	}
	for _, v := range values {
		if max <= 0 {
			b.WriteRune(ramp[0])
			continue
		}
		idx := int(v / max * float64(len(ramp)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(ramp) {
			idx = len(ramp) - 1
		}
		b.WriteRune(ramp[idx])
	}
	return b.String()
}
