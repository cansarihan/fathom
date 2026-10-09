package campaign

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// tempModule copies the sample package into a throwaway module so a real run
// writes any generated corpus and crash files under t.TempDir(), never into the
// repository.
func tempModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fathomsample\n\ngo 1.25.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"parser.go", "parser_test.go"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "deploy", "sample", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRunCleanTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping toolchain-backed test in short mode")
	}
	dir := tempModule(t)
	res, err := Run(context.Background(), Options{Dir: dir, Target: "FuzzChecksum", Duration: 2 * time.Second}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Crashed {
		t.Errorf("checksum target should not crash: %+v", res.Crash)
	}
	if res.Execs == 0 {
		t.Error("expected a positive exec count")
	}
}

func TestRunFindsCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping toolchain-backed test in short mode")
	}
	dir := tempModule(t)
	res, err := Run(context.Background(), Options{Dir: dir, Target: "FuzzParseRecords", Duration: 20 * time.Second}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Crashed || res.Crash == nil {
		t.Fatalf("parser target should crash: %+v", res)
	}
	if res.Crash.Signature == "" {
		t.Error("crash should carry a signature")
	}
}
