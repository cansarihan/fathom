package progress

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var progressRe = regexp.MustCompile(
	`elapsed:\s*([0-9hms.]+),\s*execs:\s*(\d+)\s*\((\d+)/sec\),\s*new interesting:\s*(\d+)\s*\(total:\s*(\d+)\)`)

var baselineRe = regexp.MustCompile(
	`elapsed:\s*([0-9hms.]+),\s*gathering baseline coverage:\s*(\d+)/(\d+)`)

var crashRe = regexp.MustCompile(`Failing input written to (\S+)`)

var minimizeRe = regexp.MustCompile(`minimizing`)

// ParseLine turns one line of `go test -fuzz` output into an event. The second
// result is false for lines that carry no status Fathom tracks.
func ParseLine(line string) (Event, bool) {
	line = strings.TrimSpace(line)
	if m := progressRe.FindStringSubmatch(line); m != nil {
		return Event{
			Kind:        KindProgress,
			Elapsed:     parseDuration(m[1]),
			Execs:       atoi64(m[2]),
			ExecsPerSec: atoi64(m[3]),
			Interesting: int(atoi64(m[4])),
			CorpusTotal: int(atoi64(m[5])),
		}, true
	}
	if m := baselineRe.FindStringSubmatch(line); m != nil {
		return Event{
			Kind:       KindBaseline,
			Elapsed:    parseDuration(m[1]),
			BaselineAt: int(atoi64(m[2])),
			BaselineOf: int(atoi64(m[3])),
		}, true
	}
	if m := crashRe.FindStringSubmatch(line); m != nil {
		return Event{Kind: KindCrash, CrashInput: m[1]}, true
	}
	if minimizeRe.MatchString(line) {
		return Event{Kind: KindMinimizing}, true
	}
	if line == "PASS" || strings.HasPrefix(line, "ok ") {
		return Event{Kind: KindDone}, true
	}
	return Event{}, false
}

func parseDuration(s string) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return 0
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
