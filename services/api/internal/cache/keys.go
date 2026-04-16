package cache

import "github.com/google/uuid"

// KeyRelationshipScoreOverview is the Redis key for guest relationship overview payload.
func KeyRelationshipScoreOverview(eventID uuid.UUID) string {
	return "rel_score_overview:" + eventID.String()
}

// PrefixGuestListForEvent matches all guest list page keys for an event (for SCAN invalidation).
func PrefixGuestListForEvent(eventID uuid.UUID) string {
	return "guestlist:" + eventID.String() + ":"
}

// KeyGuestListNamespaceVersion is a monotonic namespace version for guest list cache keys.
func KeyGuestListNamespaceVersion(eventID uuid.UUID) string {
	return "guestlist_nsver:" + eventID.String()
}
