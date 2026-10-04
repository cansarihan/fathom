package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAndSummarize(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a", "hello")
	write(t, dir, "b", "hello") // duplicate content
	write(t, dir, "c", "a longer input here")

	entries, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("loaded %d entries, want 3", len(entries))
	}

	s := Summarize(entries)
	if s.Count != 3 || s.Unique != 2 {
		t.Errorf("summary count=%d unique=%d, want 3/2", s.Count, s.Unique)
	}
	if s.Largest != int64(len("a longer input here")) {
		t.Errorf("largest = %d", s.Largest)
	}

	if got := Dedup(entries); len(got) != 2 {
		t.Errorf("dedup kept %d, want 2", len(got))
	}
}

func TestLoadMissingDir(t *testing.T) {
	entries, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("missing dir should not error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("want empty corpus, got %d", len(entries))
	}
}
