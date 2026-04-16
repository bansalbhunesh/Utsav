package cache

import (
	"context"

	"github.com/google/uuid"
)

// PrefixDeleter is implemented by Redis-backed caches for bulk invalidation.
type PrefixDeleter interface {
	DeleteKeysWithPrefix(ctx context.Context, prefix string) error
}

// NamespaceVersionBumper bumps a namespace version key to invalidate stale cache entries atomically.
type NamespaceVersionBumper interface {
	BumpKey(ctx context.Context, key string) (int64, error)
}

// InvalidateGuestListForEvent removes all guest-list page cache entries for an event.
func InvalidateGuestListForEvent(ctx context.Context, c Cache, eventID uuid.UUID) {
	if c == nil {
		return
	}
	if b, ok := c.(NamespaceVersionBumper); ok {
		_, _ = b.BumpKey(ctx, KeyGuestListNamespaceVersion(eventID))
		return
	}
	if d, ok := c.(PrefixDeleter); ok {
		_ = d.DeleteKeysWithPrefix(ctx, PrefixGuestListForEvent(eventID))
	}
}
