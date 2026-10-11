// Package report renders a campaign's results for the terminal or as JSON.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/crash"
)

// Summary is a whole campaign: the per-target results and the deduplicated
// crashes found across them.
type Summary struct {
	Dir     string            `json:"dir"`
	Elapsed time.Duration     `json:"elapsed"`
	Results []campaign.Result `json:"results"`
	Crashes []crash.Crash     `json:"crashes"`
}

// JSON writes the summary as indented JSON.
func JSON(w io.Writer, summary Summary) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

// Table writes a per-target table and, when present, a crash digest.
func Table(w io.Writer, summary Summary) {
	crashed := 0
	for _, r := range summary.Results {
		if r.Crashed {
			crashed++
		}
	}
	fmt.Fprintf(w, "%s\n%d targets · %s elapsed · %d crashed\n\n",
		summary.Dir, len(summary.Results), summary.Elapsed.Round(time.Second), crashed)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TARGET\tEXECS\tEXEC/S\tCORPUS\tNEW\tRESULT")
	for _, r := range summary.Results {
		result := "clean"
		if r.Crashed {
			result = "CRASH"
		} else if r.Err != "" {
			result = "error"
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%s\n",
			r.Target, r.Execs, r.ExecsPerSec, r.Corpus, r.Interesting, result)
	}
	_ = tw.Flush()

	if len(summary.Crashes) > 0 {
		fmt.Fprintf(w, "\n%d distinct crash(es):\n", len(summary.Crashes))
		for _, c := range summary.Crashes {
			fmt.Fprintf(w, "  [%s ×%d] %s\n", c.Target, c.Count, c.Signature)
		}
	}
}
