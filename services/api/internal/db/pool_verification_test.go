package db

import (
	"sync"
	"testing"
	"time"
)

// TestSimulate10kConcurrent_WithPoolSize20_Queues demonstrates that 10k in-flight
// "users" still serialize through a small connection pool: wall time scales with N/C * queryDuration.
// Tuning DB_MAX_CONNS or adding replicas reduces this queue; the guest-list refactor reduces queryDuration per slot.
func TestSimulate10kConcurrent_WithPoolSize20_Queues(t *testing.T) {
	const (
		nUsers       = 10_000
		poolSlots    = 20
		queryDuration = 5 * time.Millisecond // simulated per-connection work (after fixes: guest page ~ms-scale vs former multi-second risk)
	)

	sem := make(chan struct{}, poolSlots)
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(nUsers)
	for i := 0; i < nUsers; i++ {
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			time.Sleep(queryDuration)
			<-sem
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	minExpected := time.Duration(nUsers/poolSlots) * queryDuration
	if elapsed+time.Millisecond*50 < minExpected {
		t.Fatalf("expected elapsed ~>= %v, got %v", minExpected, elapsed)
	}
	t.Logf("10k concurrent × %s work, pool=%d: elapsed=%v (theoretical floor ~%v)",
		queryDuration, poolSlots, elapsed, minExpected)

	// Throughput ceiling at saturation: poolSlots / queryDuration "sessions" completing per second.
	maxRPS := float64(poolSlots) / queryDuration.Seconds()
	t.Logf("Saturated DB-bound RPS ceiling (this model): %.0f", maxRPS)
}
