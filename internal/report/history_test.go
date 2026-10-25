package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cansarihan/fathom/internal/store"
)

func TestHistory(t *testing.T) {
	var buf bytes.Buffer
	History(&buf, []store.Record{
		{Dir: "./pkg", Started: time.Now(), Crashes: 2},
	})
	out := buf.String()
	if !strings.Contains(out, "WHEN") || !strings.Contains(out, "./pkg") {
		t.Errorf("history table missing content:\n%s", out)
	}
}

func TestHistoryEmpty(t *testing.T) {
	var buf bytes.Buffer
	History(&buf, nil)
	if !strings.Contains(buf.String(), "no campaigns") {
		t.Error("empty history should say so")
	}
}
