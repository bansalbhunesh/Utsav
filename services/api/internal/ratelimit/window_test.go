package ratelimit

import (
	"testing"
	"time"
)

func TestWindowAllow(t *testing.T) {
	w := New(2, 100*time.Millisecond)
	if !w.Allow("k") || !w.Allow("k") {
		t.Fatal("expected first two allowed")
	}
	if w.Allow("k") {
		t.Fatal("expected third blocked")
	}
	time.Sleep(110 * time.Millisecond)
	if !w.Allow("k") {
		t.Fatal("expected allowed after window")
	}
}
