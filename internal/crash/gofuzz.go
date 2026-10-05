// Package crash reads the failing inputs a fuzzing run records and groups them
// by the fault they trigger, so a campaign that rediscovers one bug a hundred
// times is reported as one crash.
package crash

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Arg is one decoded argument from a go-fuzz corpus file.
type Arg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Bytes []byte `json:"-"`
}

// Record is a decoded go-fuzz corpus file: a version header and the argument
// tuple the engine will replay into the fuzz target.
type Record struct {
	Version string `json:"version"`
	Args    []Arg  `json:"args"`
}

const header = "go test fuzz v1"

// DecodeRecord parses the `go test fuzz v1` corpus format. Each line after the
// header is a Go-literal argument such as `[]byte("...")` or `string("...")`.
func DecodeRecord(data []byte) (Record, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	if !scanner.Scan() {
		return Record{}, fmt.Errorf("empty corpus file")
	}
	if strings.TrimSpace(scanner.Text()) != header {
		return Record{}, fmt.Errorf("not a go-fuzz corpus file")
	}
	rec := Record{Version: header}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rec.Args = append(rec.Args, decodeArg(line))
	}
	return rec, scanner.Err()
}

func decodeArg(line string) Arg {
	switch {
	case strings.HasPrefix(line, "[]byte("):
		if b, ok := unquoteInner(line, "[]byte("); ok {
			return Arg{Type: "[]byte", Value: string(b), Bytes: b}
		}
	case strings.HasPrefix(line, "string("):
		if b, ok := unquoteInner(line, "string("); ok {
			return Arg{Type: "string", Value: string(b), Bytes: b}
		}
	}
	if open := strings.IndexByte(line, '('); open > 0 && strings.HasSuffix(line, ")") {
		return Arg{Type: line[:open], Value: line[open+1 : len(line)-1]}
	}
	return Arg{Type: "raw", Value: line}
}

func unquoteInner(line, prefix string) ([]byte, bool) {
	if !strings.HasSuffix(line, ")") {
		return nil, false
	}
	inner := line[len(prefix) : len(line)-1]
	s, err := strconv.Unquote(inner)
	if err != nil {
		return nil, false
	}
	return []byte(s), true
}
