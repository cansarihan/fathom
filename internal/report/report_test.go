package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/crash"
)

func sample() Summary {
	return Summary{
		Dir: "./pkg",
		Results: []campaign.Result{
			{Target: "FuzzParse", Execs: 120000, ExecsPerSec: 40000, Corpus: 57, Interesting: 7, Crashed: true},
			{Target: "FuzzChecksum", Execs: 90000, ExecsPerSec: 30000, Corpus: 40},
		},
		Crashes: []crash.Crash{{Target: "FuzzParse", Count: 3, Signature: "index out of range"}},
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	Table(&buf, sample())
	out := buf.String()
	for _, want := range []string{"FuzzParse", "CRASH", "1 crashed", "index out of range", "TARGET"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\n%s", want, out)
		}
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	var back Summary
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatal(err)
	}
	if len(back.Results) != 2 || len(back.Crashes) != 1 {
		t.Errorf("round trip lost data: %+v", back)
	}
}
