package server

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	state := s.find(r.PathValue("id"))
	if state == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "campaign not found"})
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming is not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe, history := state.feed.subscribe()
	defer unsubscribe()
	for _, event := range history {
		writeEvent(w, event)
	}
	flusher.Flush()

	if r.URL.Query().Get("stream") == "off" {
		return
	}

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			writeEvent(w, event)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": keep alive\n\n")
			flusher.Flush()
		}
	}
}

func writeEvent(w http.ResponseWriter, data []byte) {
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}
