// Package discover lists the fuzz targets in a package by asking the Go
// toolchain, so Fathom works against any module without being told what to run.
package discover

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Target is one fuzz function in a package.
type Target struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

// Discover returns the fuzz targets defined in the package at dir. It runs
// `go test -list`, which compiles the test binary but executes nothing.
func Discover(ctx context.Context, dir string) ([]Target, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-list", "^Fuzz", "-run", "^$", ".")
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go test -list in %s: %w: %s", dir, err, strings.TrimSpace(errOut.String()))
	}

	var targets []Target
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(name, "Fuzz") {
			targets = append(targets, Target{Dir: dir, Name: name})
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets, scanner.Err()
}
