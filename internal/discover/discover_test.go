package discover

import (
	"context"
	"testing"
)

func TestDiscoverSample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping toolchain-backed test in short mode")
	}
	targets, err := Discover(context.Background(), "../../deploy/sample")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tg := range targets {
		names[tg.Name] = true
	}
	for _, want := range []string{"FuzzParseRecords", "FuzzChecksum"} {
		if !names[want] {
			t.Errorf("missing target %q in %v", want, targets)
		}
	}
}
