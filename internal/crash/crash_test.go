package crash

import (
	"strings"
	"testing"
)

func TestDecodeRecord(t *testing.T) {
	data := []byte("go test fuzz v1\n[]byte(\"he\\x00llo\")\nint(42)\n")
	rec, err := DecodeRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Args) != 2 {
		t.Fatalf("args = %d, want 2", len(rec.Args))
	}
	if rec.Args[0].Type != "[]byte" || string(rec.Args[0].Bytes) != "he\x00llo" {
		t.Errorf("arg0 = %+v", rec.Args[0])
	}
	if rec.Args[1].Type != "int" || rec.Args[1].Value != "42" {
		t.Errorf("arg1 = %+v", rec.Args[1])
	}
}

func TestDecodeRejectsNonCorpus(t *testing.T) {
	if _, err := DecodeRecord([]byte("hello world")); err == nil {
		t.Fatal("expected error for non-corpus data")
	}
}

func TestSignatureStable(t *testing.T) {
	a := Signature("panic: runtime error: index out of range [5] with length 3\n\tmain.go:12 +0x1a")
	b := Signature("panic: runtime error: index out of range [5] with length 3\n\tmain.go:12 +0x9f")
	if a != b {
		t.Errorf("signatures should match after address stripping:\n%q\n%q", a, b)
	}
	if a == "" || a == "unknown" {
		t.Errorf("unexpected signature %q", a)
	}
}

func TestSignatureExtractsPanicLine(t *testing.T) {
	text := "--- FAIL: FuzzParse (0.01s)\n    testing.go:1927: panic: runtime error: slice bounds out of range [:9] with capacity 8\n    panic({0x104, 0x53})"
	sig := Signature(text)
	if strings.Contains(sig, "FAIL") || strings.Contains(sig, "testing.go") {
		t.Errorf("signature kept scaffolding: %q", sig)
	}
	if !strings.Contains(sig, "slice bounds out of range") {
		t.Errorf("signature lost the fault: %q", sig)
	}
	if strings.Contains(sig, "[:9]") || strings.Contains(sig, "capacity 8") {
		t.Errorf("signature not normalized: %q", sig)
	}
}

func TestGroupCollapses(t *testing.T) {
	crashes := []Crash{
		New("FuzzA", "p1", []byte("x"), "panic: boom at 0x1"),
		New("FuzzA", "p2", []byte("y"), "panic: boom at 0x2"),
		New("FuzzA", "p3", []byte("z"), "panic: other"),
	}
	groups := Group(crashes)
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}
	if groups[0].Count != 2 {
		t.Errorf("top group count = %d, want 2", groups[0].Count)
	}
}
