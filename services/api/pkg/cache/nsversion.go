package cache

import (
	"context"

	"github.com/google/uuid"
)

// guestListNSReader is implemented by RedisCache (ReadIntKey).
type guestListNSReader interface {
	ReadIntKey(ctx context.Context, key string) (int64, error)
}

// guestListNSBumper is implemented by RedisCache (BumpKey / INCR).
type guestListNSBumper interface {
	BumpKey(ctx context.Context, key string) (int64, error)
}

// GuestListNamespaceVersion returns the namespace counter embedded in guest list cache keys
// (Redis: guestlist_nsver:{eventID}). Missing or unreadable => 0.
func GuestListNamespaceVersion(ctx context.Context, c Cache, eventID uuid.UUID) int64 {
	if c == nil {
		return 0
	}
	r, ok := c.(guestListNSReader)
	if !ok {
		return 0
	}
	v, err := r.ReadIntKey(ctx, KeyGuestListNamespaceVersion(eventID))
	if err != nil || v < 0 {
		return 0
	}
	return v
}

// BumpGuestListNamespaceVersion invalidates all guest-list page keys for an event in O(1)
// via INCR on guestlist_nsver:{eventID} (no SCAN).
func BumpGuestListNamespaceVersion(ctx context.Context, c Cache, eventID uuid.UUID) {
	if c == nil {
		return
	}
	b, ok := c.(guestListNSBumper)
	if !ok {
		return
	}
	_, _ = b.BumpKey(ctx, KeyGuestListNamespaceVersion(eventID))
}
