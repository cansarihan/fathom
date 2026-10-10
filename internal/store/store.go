// Package store persists a campaign history as JSON so successive runs build a
// record of what was fuzzed, for how long, and what crashed.
package store

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/cansarihan/fathom/internal/campaign"
)

// Record is one campaign: a set of target results captured at a point in time.
type Record struct {
	ID       string            `json:"id"`
	Started  time.Time         `json:"started"`
	Finished time.Time         `json:"finished"`
	Dir      string            `json:"dir"`
	Results  []campaign.Result `json:"results"`
	Crashes  int               `json:"crashes"`
}

// Store is a history of campaigns, optionally backed by a file on disk.
type Store struct {
	mu      sync.Mutex
	path    string
	records []Record
}

// Open returns a store backed by path, loading any existing history. An empty
// path keeps the history in memory only.
func Open(path string) (*Store, error) {
	s := &Store{path: path}
	if path == "" {
		return s, nil
	}
	data, err := os.ReadFile(path) // #nosec G304 -- operator-chosen history file
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &s.records); err != nil {
		return nil, err
	}
	return s, nil
}

// Append adds a record and, when the store is file-backed, flushes to disk.
func (s *Store) Append(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record)
	return s.flushLocked()
}

// All returns the history newest first.
func (s *Store) All() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, len(s.records))
	for i, r := range s.records {
		out[len(s.records)-1-i] = r
	}
	return out
}

func (s *Store) flushLocked() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
