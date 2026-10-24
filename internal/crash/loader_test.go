package crash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadInputs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "FuzzX")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "abc"), []byte("go test fuzz v1\n[]byte(\"boom\")\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notcorpus"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	inputs, err := LoadInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 1 {
		t.Fatalf("inputs = %d, want 1 (non-corpus skipped)", len(inputs))
	}
	if inputs[0].Target != "FuzzX" || len(inputs[0].Record.Args) != 1 {
		t.Errorf("input = %+v", inputs[0])
	}
}

func TestLoadInputsMissing(t *testing.T) {
	inputs, err := LoadInputs(filepath.Join(t.TempDir(), "absent"))
	if err != nil || len(inputs) != 0 {
		t.Errorf("missing dir: inputs=%v err=%v", inputs, err)
	}
}
