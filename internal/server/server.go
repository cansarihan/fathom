// Package server wraps the campaign engine in an HTTP API and serves the
// embedded panel. It fuzzes the targets in one fixed package directory chosen at
// startup, so a request can never point it at an arbitrary path on the host.
package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cansarihan/fathom/internal/crash"
)

// BuildInfo is the version metadata reported on the status endpoint.
type BuildInfo struct {
	Version   string
	Commit    string
	GoVersion string
	Platform  string
}

// Options configures a server instance.
type Options struct {
	Dir        string
	Duration   time.Duration
	Token      string
	UI         bool
	MaxHistory int
}

// TargetRun is one target's live state within a campaign.
type TargetRun struct {
	Target      string `json:"target"`
	Status      string `json:"status"`
	Execs       int64  `json:"execs"`
	ExecsPerSec int64  `json:"execs_per_sec"`
	Corpus      int    `json:"corpus"`
	Interesting int    `json:"interesting"`
	Crashed     bool   `json:"crashed"`
}

// Campaign is one run over every target in the directory, growing as it streams.
type Campaign struct {
	ID         string        `json:"id"`
	Dir        string        `json:"dir"`
	Status     string        `json:"status"`
	Started    time.Time     `json:"started"`
	Finished   time.Time     `json:"finished,omitempty"`
	Duration   string        `json:"duration"`
	Targets    []TargetRun   `json:"targets"`
	Crashes    []crash.Crash `json:"crashes"`
	TotalExecs int64         `json:"total_execs"`
	Rate       []float64     `json:"rate"`
	Err        string        `json:"err,omitempty"`
}

// Status is the response of the status endpoint.
type Status struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	GoVersion string    `json:"go_version"`
	Platform  string    `json:"platform"`
	Dir       string    `json:"dir"`
	Duration  string    `json:"duration"`
	StartedAt time.Time `json:"started_at"`
	Uptime    string    `json:"uptime"`
	Campaigns int       `json:"campaigns"`
	Running   int       `json:"running"`
}

// Server holds the campaign history and build metadata.
type Server struct {
	options Options
	info    BuildInfo
	started time.Time

	mu        sync.Mutex
	campaigns []*campaignState
}

type campaignState struct {
	mu     sync.Mutex
	record Campaign
	feed   *feed
}

// New builds a server over one target directory.
func New(options Options, info BuildInfo) *Server {
	if options.MaxHistory <= 0 {
		options.MaxHistory = 25
	}
	if options.Duration <= 0 {
		options.Duration = 10 * time.Second
	}
	return &Server{options: options, info: info, started: time.Now()}
}

// Handler returns the HTTP handler for the API and, when enabled, the panel.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/v1/status", s.guard(s.handleStatus))
	mux.HandleFunc("GET /api/v1/targets", s.guard(s.handleTargets))
	mux.HandleFunc("GET /api/v1/campaigns", s.guard(s.handleList))
	mux.HandleFunc("POST /api/v1/campaigns", s.guard(s.handleCreate))
	mux.HandleFunc("GET /api/v1/campaigns/{id}", s.guard(s.handleGet))
	mux.HandleFunc("GET /api/v1/campaigns/{id}/stream", s.guard(s.handleStream))
	if s.options.UI {
		mux.Handle("GET /", uiHandler())
	}
	return mux
}

func (s *Server) guard(next http.HandlerFunc) http.HandlerFunc {
	token := s.options.Token
	return func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if provided == "" {
				provided = r.URL.Query().Get("token")
			}
			if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	total := len(s.campaigns)
	running := 0
	for _, c := range s.campaigns {
		if c.snapshot().Status == "running" {
			running++
		}
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, Status{
		Version:   s.info.Version,
		Commit:    s.info.Commit,
		GoVersion: s.info.GoVersion,
		Platform:  s.info.Platform,
		Dir:       s.options.Dir,
		Duration:  s.options.Duration.String(),
		StartedAt: s.started,
		Uptime:    time.Since(s.started).Round(time.Second).String(),
		Campaigns: total,
		Running:   running,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newID() string {
	buffer := make([]byte, 5)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
