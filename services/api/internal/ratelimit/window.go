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

// StartPeriodicCleanup removes expired timestamps and empty keys so one-off IPs do not
// stay in memory forever. Call once per Window (e.g. from NewInMemoryLimiter).
func (w *Window) StartPeriodicCleanup(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			w.pruneExpiredKeys()
		}
	}()
}

func (w *Window) pruneExpiredKeys() {
	cutoff := time.Now().Add(-w.window)
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, xs := range w.hits {
		kept := xs[:0]
		for _, hit := range xs {
			if hit.After(cutoff) {
				kept = append(kept, hit)
			}
		}
		if len(kept) == 0 {
			delete(w.hits, k)
		} else {
			w.hits[k] = kept
		}
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
