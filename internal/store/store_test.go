package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
)

func TestAppendPersistsAndReloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"one", "two"} {
		err := s.Append(Record{
			ID:      id,
			Started: time.Now(),
			Results: []campaign.Result{{Target: "FuzzX", Execs: int64(i + 1)}},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	all := reopened.All()
	if len(all) != 2 {
		t.Fatalf("records = %d, want 2", len(all))
	}
	if all[0].ID != "two" {
		t.Errorf("newest-first ordering broken: %q", all[0].ID)
	}
}

func TestOpenMissingFile(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.All()) != 0 {
		t.Error("expected empty history")
	}
}
