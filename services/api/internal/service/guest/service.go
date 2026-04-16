package guestservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bhune/utsav/services/api/pkg/cache"
	phoneutil "github.com/bhune/utsav/services/api/pkg/phone"
	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
)

// relationshipOverviewCacheTTL matches guest list page TTL so dashboards stay consistent.
const relationshipOverviewCacheTTL = 45 * time.Second

// guestListPageCacheTTL bounds staleness for organiser guest list pages (cache-aside).
const guestListPageCacheTTL = 45 * time.Second

type guestListPageCached struct {
	Guests []guestrepo.Guest `json:"guests"`
	Next   *string           `json:"next,omitempty"`
}

func guestListPageCacheKey(eventID uuid.UUID, sortNorm, tier string, limit, offset int, nsVersion int64, cursorStr *string) string {
	cTag := "-"
	if cursorStr != nil && strings.TrimSpace(*cursorStr) != "" {
		sum := sha256.Sum256([]byte(strings.TrimSpace(*cursorStr)))
		cTag = hex.EncodeToString(sum[:10])
	}
	tierK := tier
	if tierK == "" {
		tierK = "-"
	}
	return "guestlist:" + eventID.String() + ":" + strconv.FormatInt(nsVersion, 10) + ":" + sortNorm + ":" + tierK + ":" + strconv.Itoa(limit) + ":" + strconv.Itoa(offset) + ":" + cTag
}

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type repository interface {
	ListGuests(ctx context.Context, p guestrepo.ListGuestsParams) ([]guestrepo.Guest, error)
	UpsertGuest(ctx context.Context, eventID uuid.UUID, input guestrepo.GuestInput) (string, error)
	UpsertGuestTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input guestrepo.GuestInput) (string, error)
	GuestIDByEventPhoneTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, phone string) (string, error)
	ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, src io.Reader) (*guestrepo.ImportResult, error)
}

type Service struct {
	repo  repository
	cache cache.Cache
}

type RelationshipScoreOverview struct {
	RankedGuests           []guestrepo.Guest `json:"ranked_guests"`
	GuestsNeedingAttention []guestrepo.Guest `json:"guests_needing_attention"`
	TierCounts             map[string]int    `json:"tier_counts"`
}

func NewService(repo repository, c cache.Cache) *Service {
	return &Service{repo: repo, cache: c}
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

func normalizeListSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name_desc", "priority_asc", "priority_desc", "rsvp_desc", "shagun_desc":
		return strings.ToLower(strings.TrimSpace(sort))
	default:
		return "name_asc"
	}
}

// ListGuests returns guests and an opaque next_cursor when another page exists.
// Priority sorts use SQL ORDER BY + LIMIT/OFFSET or keyset (cursor); no full-table prefetch.
func (s *Service) ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int, sort, priorityTier string, cursorStr *string) ([]guestrepo.Guest, *string, *ServiceError) {
	qctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

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
		// SQL ignores OFFSET when cursor is set; align cache key so the same page does not fragment across keys.
		offKey := offset
		if decoded != nil {
			offKey = 0
		}
		nsVersion := cache.GuestListNamespaceVersion(qctx, s.cache, eventID)
		cacheKey := guestListPageCacheKey(eventID, sortNorm, tierFilter, limit, offKey, nsVersion, cursorStr)
		if s.cache != nil {
			if raw, err := s.cache.Get(qctx, cacheKey); err == nil {
				var p guestListPageCached
				if json.Unmarshal(raw, &p) == nil {
					return p.Guests, p.Next, nil
				}
			}
		}
		fetch := limit + 1
		list, err := s.repo.ListGuests(qctx, guestrepo.ListGuestsParams{
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
		if s.cache != nil {
			if raw, err := json.Marshal(guestListPageCached{Guests: list, Next: next}); err == nil {
				_ = s.cache.Set(qctx, cacheKey, raw, guestListPageCacheTTL)
			}
		}
		return list, next, nil
	}

	if decoded != nil && offset > 0 {
		offset = 0
	}

	nsVersion := cache.GuestListNamespaceVersion(qctx, s.cache, eventID)
	cacheKey := guestListPageCacheKey(eventID, sortNorm, "", limit, offset, nsVersion, cursorStr)
	if s.cache != nil {
		if raw, err := s.cache.Get(qctx, cacheKey); err == nil {
			var p guestListPageCached
			if json.Unmarshal(raw, &p) == nil {
				return p.Guests, p.Next, nil
			}
		}
	}

	fetch := limit + 1
	list, err := s.repo.ListGuests(qctx, guestrepo.ListGuestsParams{
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
	if s.cache != nil {
		if raw, err := json.Marshal(guestListPageCached{Guests: list, Next: next}); err == nil {
			_ = s.cache.Set(qctx, cacheKey, raw, guestListPageCacheTTL)
		}
	}
	return list, next, nil
}

func (s *Service) RelationshipScoreOverview(ctx context.Context, eventID uuid.UUID) (*RelationshipScoreOverview, *ServiceError) {
	if s.cache != nil {
		key := cache.KeyRelationshipScoreOverview(eventID)
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
		key := cache.KeyRelationshipScoreOverview(eventID)
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
	cache.InvalidateGuestListForEvent(ctx, s.cache, eventID)
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cache.KeyRelationshipScoreOverview(eventID))
}

func normalizeGuestInput(input guestrepo.GuestInput) (guestrepo.GuestInput, *ServiceError) {
	input.Name = strings.TrimSpace(input.Name)
	phoneNorm, err := phoneutil.NormalizeE164(input.Phone)
	if input.Name == "" || err != nil {
		return input, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Name and phone are required."}
	}
	input.Phone = phoneNorm
	return input, nil
}

func (s *Service) UpsertGuest(ctx context.Context, eventID uuid.UUID, input guestrepo.GuestInput) (string, *ServiceError) {
	normalized, svcErr := normalizeGuestInput(input)
	if svcErr != nil {
		return "", svcErr
	}
	guestID, err := s.repo.UpsertGuest(ctx, eventID, normalized)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to save guest."}
	}
	s.InvalidateRelationshipOverview(ctx, eventID)
	return guestID, nil
}

// UpsertGuestTx is the same as UpsertGuest but runs on an existing transaction (caller commits).
func (s *Service) UpsertGuestTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input guestrepo.GuestInput) (string, *ServiceError) {
	normalized, svcErr := normalizeGuestInput(input)
	if svcErr != nil {
		return "", svcErr
	}
	guestID, err := s.repo.UpsertGuestTx(ctx, tx, eventID, normalized)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to save guest."}
	}
	return guestID, nil
}

// GuestIDByEventPhoneTx returns the guest id for an event/phone pair within tx (idempotent replay path).
func (s *Service) GuestIDByEventPhoneTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, phone string) (string, *ServiceError) {
	phoneNorm, err := phoneutil.NormalizeE164(phone)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone is required."}
	}
	guestID, err := s.repo.GuestIDByEventPhoneTx(ctx, tx, eventID, phoneNorm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_REPLAY_INCOMPLETE", Message: "Unable to resolve guest for idempotent replay."}
		}
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to load guest."}
	}
	return guestID, nil
}

func (s *Service) ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, src io.Reader) (*guestrepo.ImportResult, *ServiceError) {
	if src == nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "EMPTY_CSV", Message: "CSV payload cannot be empty."}
	}
	result, err := s.repo.ImportGuestsCSV(ctx, eventID, src)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "CSV_IMPORT_FAILED", Message: "Unable to import guest CSV."}
	}
	s.InvalidateRelationshipOverview(ctx, eventID)
	return result, nil
}
