// Package corpus inventories the inputs a fuzz target keeps: the seed corpus in
// the package's testdata and the generated corpus the engine caches. It hashes
// each entry so duplicates across locations collapse, and summarizes the set.
package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
)

// Entry is one corpus file.
type Entry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// Stats summarizes a corpus.
type Stats struct {
	Count   int   `json:"count"`
	Unique  int   `json:"unique"`
	Bytes   int64 `json:"bytes"`
	Largest int64 `json:"largest"`
}

// Load walks a directory and returns one entry per regular file, hashed by
// content. A missing directory yields an empty corpus rather than an error, so
// a target with no generated corpus yet is not a failure.
func Load(dir string) ([]Entry, error) {
	var entries []Entry
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path) // #nosec G304 -- walking an operator-supplied corpus directory
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(data)
		entries = append(entries, Entry{Path: path, Hash: hex.EncodeToString(sum[:]), Size: int64(len(data))})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

// Dedup drops entries whose content hash was already seen, keeping first paths.
func Dedup(entries []Entry) []Entry {
	seen := make(map[string]struct{}, len(entries))
	var out []Entry
	for _, e := range entries {
		if _, ok := seen[e.Hash]; ok {
			continue
		}
		seen[e.Hash] = struct{}{}
		out = append(out, e)
	}
	return out
}

// Summarize counts the corpus and its unique inputs by content.
func Summarize(entries []Entry) Stats {
	s := Stats{Count: len(entries)}
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		s.Bytes += e.Size
		if e.Size > s.Largest {
			s.Largest = e.Size
		}
		if _, ok := seen[e.Hash]; !ok {
			seen[e.Hash] = struct{}{}
			s.Unique++
		}
	}
	return s
}
