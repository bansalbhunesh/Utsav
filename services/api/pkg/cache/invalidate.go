package cache

import (
	"context"

	"github.com/google/uuid"
)

// InvalidateGuestListForEvent bumps guestlist_nsver:{eventID} so all list keys miss (O(1)).
func InvalidateGuestListForEvent(ctx context.Context, c Cache, eventID uuid.UUID) {
	BumpGuestListNamespaceVersion(ctx, c, eventID)
}
