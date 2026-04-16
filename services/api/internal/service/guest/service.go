package guestservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	sortutil "sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/cache"
	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
)

// guestListPrefetchMax caps how many guests we load when ordering or filtering must match
// SQL-computed priority (tier filter, or priority sort — OFFSET must apply after ordering).
const guestListPrefetchMax = 10000

const relationshipOverviewCacheTTL = 5 * time.Minute

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo  guestrepo.Repository
	cache cache.Cache
}

type RelationshipScoreOverview struct {
	RankedGuests           []guestrepo.Guest `json:"ranked_guests"`
	GuestsNeedingAttention []guestrepo.Guest `json:"guests_needing_attention"`
	TierCounts             map[string]int    `json:"tier_counts"`
}

func NewService(repo guestrepo.Repository, c cache.Cache) *Service {
	return &Service{repo: repo, cache: c}
}

func relationshipOverviewCacheKey(eventID uuid.UUID) string {
	return "rel_score_overview:" + eventID.String()
}

func tierLabelFromScore(score int) string {
	switch {
	case score >= 80:
		return "Critical"
	case score >= 50:
		return "Important"
	default:
		return "Optional"
	}
}

func applyGuestPriorityFromSQL(list []guestrepo.Guest) {
	for i := range list {
		list[i].PriorityTier = tierLabelFromScore(list[i].PriorityScore)
		list[i].PriorityReasons = []string{}
	}
}

func sortGuestsByPriority(list []guestrepo.Guest, sort string) {
	sortutil.SliceStable(list, func(i, j int) bool {
		if sort == "priority_asc" {
			if list[i].PriorityScore == list[j].PriorityScore {
				return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
			}
			return list[i].PriorityScore < list[j].PriorityScore
		}
		if list[i].PriorityScore == list[j].PriorityScore {
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		}
		return list[i].PriorityScore > list[j].PriorityScore
	})
}

func paginateGuests(list []guestrepo.Guest, offset, limit int) []guestrepo.Guest {
	if offset >= len(list) {
		return []guestrepo.Guest{}
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end]
}

func normalizeListSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name_desc", "priority_asc", "priority_desc", "rsvp_desc", "shagun_desc":
		return strings.ToLower(strings.TrimSpace(sort))
	default:
		return "name_asc"
	}
}

// ListGuests returns guests and an opaque next_cursor when another page exists (keyset path only).
func (s *Service) ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int, sort, priorityTier string, cursorStr *string) ([]guestrepo.Guest, *string, *ServiceError) {
	tierFilter := strings.ToLower(strings.TrimSpace(priorityTier))
	switch tierFilter {
	case "critical", "important", "optional":
	default:
		tierFilter = ""
	}

	sortNorm := normalizeListSort(sort)

	var decoded *guestrepo.ListGuestsCursor
	if cursorStr != nil && strings.TrimSpace(*cursorStr) != "" {
		c, err := guestrepo.DecodeListGuestsCursor(*cursorStr)
		if err != nil {
			return nil, nil, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_CURSOR", Message: "Invalid or malformed guest list cursor."}
		}
		if normalizeListSort(c.Sort) != sortNorm {
			return nil, nil, &ServiceError{Status: http.StatusBadRequest, Code: "CURSOR_SORT_MISMATCH", Message: "Cursor does not match the requested sort parameter."}
		}
		decoded = &c
	}

	if tierFilter != "" {
		fetch := limit + 1
		if fetch > 10000 {
			fetch = 10000
		}
		list, err := s.repo.ListGuests(ctx, guestrepo.ListGuestsParams{
			EventID:      eventID,
			Limit:        fetch,
			Offset:       offset,
			Sort:         sort,
			Cursor:       decoded,
			PriorityTier: tierFilter,
		})
		if err != nil {
			return nil, nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
		}
		applyGuestPriorityFromSQL(list)
		hasMore := len(list) > limit
		if hasMore {
			list = list[:limit]
		}
		var next *string
		if hasMore && limit > 0 {
			last := list[len(list)-1]
			cur := guestrepo.CursorFromGuestRow(sortNorm, last)
			if enc, err := guestrepo.EncodeListGuestsCursor(cur); err == nil {
				next = &enc
			}
		}
		return list, next, nil
	}

	if sortNorm == "priority_desc" || sortNorm == "priority_asc" {
		if decoded != nil {
			return nil, nil, &ServiceError{Status: http.StatusBadRequest, Code: "CURSOR_NOT_SUPPORTED", Message: "Cursor pagination is not supported for priority sort; use offset or omit sort=priority_*."}
		}
		list, err := s.repo.ListGuests(ctx, guestrepo.ListGuestsParams{EventID: eventID, Limit: guestListPrefetchMax, Offset: 0, Sort: sort, PriorityTier: ""})
		if err != nil {
			return nil, nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
		}
		applyGuestPriorityFromSQL(list)
		sortGuestsByPriority(list, sortNorm)
		return paginateGuests(list, offset, limit), nil, nil
	}

	if decoded != nil && offset > 0 {
		offset = 0
	}

	fetch := limit + 1
	if fetch > 10000 {
		fetch = 10000
	}
	list, err := s.repo.ListGuests(ctx, guestrepo.ListGuestsParams{
		EventID:      eventID,
		Limit:        fetch,
		Offset:       offset,
		Sort:         sort,
		Cursor:       decoded,
		PriorityTier: "",
	})
	if err != nil {
		return nil, nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
	}
	applyGuestPriorityFromSQL(list)

	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	var next *string
	if hasMore && limit > 0 {
		last := list[len(list)-1]
		cur := guestrepo.CursorFromGuestRow(sortNorm, last)
		if enc, err := guestrepo.EncodeListGuestsCursor(cur); err == nil {
			next = &enc
		}
	}
	return list, next, nil
}

func (s *Service) RelationshipScoreOverview(ctx context.Context, eventID uuid.UUID) (*RelationshipScoreOverview, *ServiceError) {
	if s.cache != nil {
		key := relationshipOverviewCacheKey(eventID)
		b, err := s.cache.Get(ctx, key)
		if err == nil {
			var o RelationshipScoreOverview
			if json.Unmarshal(b, &o) == nil {
				return &o, nil
			}
		} else if !errors.Is(err, cache.ErrMiss) {
			// fall through to DB
		}
	}

	ranked, _, svcErr := s.ListGuests(ctx, eventID, 200, 0, "priority_desc", "", nil)
	if svcErr != nil {
		return nil, svcErr
	}
	out := buildRelationshipOverview(ranked)

	if s.cache != nil {
		key := relationshipOverviewCacheKey(eventID)
		if raw, err := json.Marshal(out); err == nil {
			_ = s.cache.Set(ctx, key, raw, relationshipOverviewCacheTTL)
		}
	}
	return out, nil
}

func buildRelationshipOverview(ranked []guestrepo.Guest) *RelationshipScoreOverview {
	top := ranked
	if len(top) > 15 {
		top = top[:15]
	}
	attention := make([]guestrepo.Guest, 0, 10)
	for _, g := range ranked {
		if g.PriorityScore < 50 {
			attention = append(attention, g)
			if len(attention) == 10 {
				break
			}
		}
	}
	counts := map[string]int{
		"critical":  0,
		"important": 0,
		"optional":  0,
	}
	for _, g := range ranked {
		switch strings.ToLower(g.PriorityTier) {
		case "critical":
			counts["critical"]++
		case "important":
			counts["important"]++
		default:
			counts["optional"]++
		}
	}
	return &RelationshipScoreOverview{
		RankedGuests:           top,
		GuestsNeedingAttention: attention,
		TierCounts:             counts,
	}
}

// InvalidateRelationshipOverview drops cached dashboard intelligence for an event.
func (s *Service) InvalidateRelationshipOverview(ctx context.Context, eventID uuid.UUID) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, relationshipOverviewCacheKey(eventID))
}

func (s *Service) UpsertGuest(ctx context.Context, eventID uuid.UUID, input guestrepo.GuestInput) (string, *ServiceError) {
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Name == "" || input.Phone == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Name and phone are required."}
	}
	guestID, err := s.repo.UpsertGuest(ctx, eventID, input)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to save guest."}
	}
	s.InvalidateRelationshipOverview(ctx, eventID)
	return guestID, nil
}

func (s *Service) ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, csv string) (*guestrepo.ImportResult, *ServiceError) {
	if strings.TrimSpace(csv) == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "EMPTY_CSV", Message: "CSV payload cannot be empty."}
	}
	result, err := s.repo.ImportGuestsCSV(ctx, eventID, csv)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "CSV_IMPORT_FAILED", Message: "Unable to import guest CSV."}
	}
	s.InvalidateRelationshipOverview(ctx, eventID)
	return result, nil
}
