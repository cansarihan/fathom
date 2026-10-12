package server

import "sync"

// feed is a fan-out of campaign events to any number of server-sent-event
// subscribers, with a bounded replay history so a late subscriber still sees
// the run from the start.
type feed struct {
	limit   int
	mu      sync.Mutex
	history [][]byte
	subs    map[int]chan []byte
	seq     int
	closed  bool
}

func newFeed(limit int) *feed {
	return &feed{limit: limit, subs: map[int]chan []byte{}}
}

func (f *feed) publish(data []byte) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	if len(f.history) < f.limit {
		f.history = append(f.history, data)
	}
	targets := make([]chan []byte, 0, len(f.subs))
	for _, channel := range f.subs {
		targets = append(targets, channel)
	}
	f.mu.Unlock()
	for _, channel := range targets {
		select {
		case channel <- data:
		default:
		}
	}
}

func (f *feed) subscribe() (<-chan []byte, func(), [][]byte) {
	channel := make(chan []byte, 512)
	f.mu.Lock()
	history := append([][]byte(nil), f.history...)
	if f.closed {
		f.mu.Unlock()
		close(channel)
		return channel, func() {}, history
	}
	f.seq++
	id := f.seq
	f.subs[id] = channel
	f.mu.Unlock()
	return channel, func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if existing, ok := f.subs[id]; ok {
			delete(f.subs, id)
			close(existing)
		}
	}, history
}

func (f *feed) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return
	}
	f.closed = true
	for id, channel := range f.subs {
		delete(f.subs, id)
		close(channel)
	}
}
