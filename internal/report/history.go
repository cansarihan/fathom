package report

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/cansarihan/fathom/internal/store"
)

// History renders a campaign history, newest first, as a table.
func History(w io.Writer, records []store.Record) {
	if len(records) == 0 {
		fmt.Fprintln(w, "no campaigns recorded")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "WHEN\tDIR\tTARGETS\tCRASHES")
	for _, r := range records {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n",
			r.Started.Format(time.RFC3339), r.Dir, len(r.Results), r.Crashes)
	}
	_ = tw.Flush()
}
