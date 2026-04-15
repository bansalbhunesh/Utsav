package ratelimit

import (
	"sync"
	"time"
)

// Window is a fixed-window-style limiter: at most max hits per key within window duration.
type Window struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
}

func New(max int, window time.Duration) *Window {
	return &Window{
		max:    max,
		window: window,
		hits:   make(map[string][]time.Time),
	}
}

// Allow returns false if the key has reached max within the sliding window.
func (w *Window) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-w.window)

	w.mu.Lock()
	defer w.mu.Unlock()

	xs := w.hits[key]
	kept := xs[:0]
	for _, t := range xs {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= w.max {
		w.hits[key] = kept
		return false
	}
	if len(kept) == 0 {
		delete(w.hits, key)
	}
	kept = append(kept, now)
	w.hits[key] = kept
	return true
}
