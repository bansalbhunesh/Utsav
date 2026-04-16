package cache

import "github.com/google/uuid"

// KeyRelationshipScoreOverview is the Redis key for guest relationship overview payload.
func KeyRelationshipScoreOverview(eventID uuid.UUID) string {
	return "rel_score_overview:" + eventID.String()
}

// KeyGuestListNamespaceVersion is the Redis counter key (INCR) for guest list cache namespace.
func KeyGuestListNamespaceVersion(eventID uuid.UUID) string {
	return "guestlist_nsver:" + eventID.String()
}
