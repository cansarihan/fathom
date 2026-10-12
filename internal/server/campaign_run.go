package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
	"github.com/cansarihan/fathom/internal/crash"
	"github.com/cansarihan/fathom/internal/discover"
)

func (s *Server) handleTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := discover.Discover(r.Context(), s.options.Dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, targets)
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	out := make([]Campaign, 0, len(s.campaigns))
	for i := len(s.campaigns) - 1; i >= 0; i-- {
		out = append(out, s.campaigns[i].snapshot())
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if state := s.find(r.PathValue("id")); state != nil {
		writeJSON(w, http.StatusOK, state.snapshot())
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "campaign not found"})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	state := &campaignState{
		feed: newFeed(16384),
		record: Campaign{
			ID:       newID(),
			Dir:      s.options.Dir,
			Status:   "running",
			Started:  time.Now(),
			Duration: s.options.Duration.String(),
		},
	}
	s.mu.Lock()
	s.campaigns = append(s.campaigns, state)
	if len(s.campaigns) > s.options.MaxHistory {
		s.campaigns = s.campaigns[len(s.campaigns)-s.options.MaxHistory:]
	}
	s.mu.Unlock()

	go s.execute(context.WithoutCancel(r.Context()), state)
	writeJSON(w, http.StatusAccepted, state.snapshot())
}

func (s *Server) find(id string) *campaignState {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, state := range s.campaigns {
		if state.record.ID == id {
			return state
		}
	}
	return nil
}

func (s *Server) execute(ctx context.Context, state *campaignState) {
	work, cleanup, err := workdir(s.options.Dir)
	if err == nil {
		defer cleanup()
	} else {
		work = s.options.Dir // fall back to fuzzing in place
	}

	targets, err := discover.Discover(ctx, work)
	if err != nil {
		state.fail(err.Error())
		state.feed.publish([]byte(`{"type":"done"}`))
		state.feed.close()
		return
	}
	state.initTargets(targets)

	var crashes []crash.Crash
	for _, tg := range targets {
		state.setStatus(tg.Name, "running")
		result, _ := campaign.Run(ctx, campaign.Options{
			Dir: work, Target: tg.Name, Duration: s.options.Duration,
		}, func(u campaign.Update) {
			state.apply(tg.Name, u)
			if payload, mErr := json.Marshal(map[string]any{"type": "update", "target": tg.Name, "update": u}); mErr == nil {
				state.feed.publish(payload)
			}
		})
		state.setResult(tg.Name, result)
		if result.Crash != nil {
			crashes = append(crashes, *result.Crash)
		}
	}

	state.finish(crash.Group(crashes))
	state.feed.publish([]byte(`{"type":"done"}`))
	state.feed.close()
}

// workdir copies the Go files of a self-contained package into a throwaway
// module so a run's generated corpus and crash inputs land in a temp directory
// instead of the served source tree.
func workdir(src string) (string, func(), error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return "", nil, err
	}
	dir, err := os.MkdirTemp("", "fathom-work-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fathomwork\n\ngo 1.25.0\n"), 0o600); err != nil {
		cleanup()
		return "", nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(src, e.Name())) // #nosec G304 -- served package directory
		if readErr != nil {
			cleanup()
			return "", nil, readErr
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0o600); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return dir, cleanup, nil
}
