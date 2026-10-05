package crash

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
)

// Crash is one failing input and the fault it triggers.
type Crash struct {
	Target    string `json:"target"`
	InputPath string `json:"input_path"`
	InputHash string `json:"input_hash"`
	Panic     string `json:"panic"`
	Signature string `json:"signature"`
	Count     int    `json:"count"`
}

var (
	hexAddr  = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	numParen = regexp.MustCompile(`\(\d+\)`)
)

// Signature reduces a panic message to a stable key: the first meaningful line
// with addresses and counts stripped, so the same fault from different inputs
// collapses to one group.
func Signature(panicText string) string {
	for _, line := range strings.Split(panicText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "panic: ")
		line = hexAddr.ReplaceAllString(line, "0x?")
		line = numParen.ReplaceAllString(line, "(N)")
		return line
	}
	return "unknown"
}

// New builds a crash from a target, input bytes and the panic text it caused.
func New(target, inputPath string, input []byte, panicText string) Crash {
	sum := sha256.Sum256(input)
	return Crash{
		Target:    target,
		InputPath: inputPath,
		InputHash: hex.EncodeToString(sum[:]),
		Panic:     strings.TrimSpace(panicText),
		Signature: Signature(panicText),
		Count:     1,
	}
}

// Group collapses crashes that share a signature, summing their counts and
// keeping the first input as the representative. Groups are returned most
// frequent first.
func Group(crashes []Crash) []Crash {
	index := map[string]*Crash{}
	var order []string
	for _, c := range crashes {
		if existing, ok := index[c.Signature]; ok {
			existing.Count += c.Count
			continue
		}
		clone := c
		index[c.Signature] = &clone
		order = append(order, c.Signature)
	}
	out := make([]Crash, 0, len(order))
	for _, sig := range order {
		out = append(out, *index[sig])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}
