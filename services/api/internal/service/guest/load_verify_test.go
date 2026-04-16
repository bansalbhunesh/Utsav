package guestservice

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
)

// recordingRepo captures the maximum ListGuests Limit observed under concurrent load.
type recordingRepo struct {
	mu      sync.Mutex
	maxLimit int
}

func (r *recordingRepo) ListGuests(_ context.Context, p guestrepo.ListGuestsParams) ([]guestrepo.Guest, error) {
	r.mu.Lock()
	if p.Limit > r.maxLimit {
		r.maxLimit = p.Limit
	}
	r.mu.Unlock()
	return []guestrepo.Guest{}, nil
}

func (r *recordingRepo) UpsertGuest(context.Context, uuid.UUID, guestrepo.GuestInput) (string, error) {
	return "", nil
}

func (r *recordingRepo) ImportGuestsCSV(context.Context, uuid.UUID, string) (*guestrepo.ImportResult, error) {
	return &guestrepo.ImportResult{}, nil
}

// TestGuestListPriority_Concurrent10k_UsesPagedSQL verifies the post-refactor path never asks the repo
// for more than limit+1 rows for priority sort (no 10k prefetch).
func TestGuestListPriority_Concurrent10k_UsesPagedSQL(t *testing.T) {
	const concurrent = 10_000
	const pageLimit = 50

	repo := &recordingRepo{}
	svc := NewService(repo, nil)
	eid := uuid.MustParse("00000000-0000-4000-8000-000000000001")

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			_, _, err := svc.ListGuests(context.Background(), eid, pageLimit, 0, "priority_desc", "", nil)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("ListGuests: %v", err)
	}

	max := repo.maxLimit
	// Unified path uses fetch := limit+1 (capped at 10000); for limit=50 expect 51.
	if max > pageLimit+1 {
		t.Fatalf("expected repo ListGuests Limit <= %d under priority sort, got %d (regression: full-table prefetch?)", pageLimit+1, max)
	}
	t.Logf("10k concurrent priority_desc requests: max ListGuests Limit observed = %d (want <= %d)", max, pageLimit+1)
}
