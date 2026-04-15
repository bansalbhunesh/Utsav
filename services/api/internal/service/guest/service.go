package guestservice

import (
	"context"
	"math"
	"net/http"
	sortutil "sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
)

// guestListPrefetchMax caps how many guests we load when ordering or filtering must match
// scoreGuestPriority() (tier filter, or priority sort — OFFSET must apply after Go ordering).
const guestListPrefetchMax = 10000

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo guestrepo.Repository
}

type RelationshipScoreOverview struct {
	RankedGuests           []guestrepo.Guest `json:"ranked_guests"`
	GuestsNeedingAttention []guestrepo.Guest `json:"guests_needing_attention"`
	TierCounts             map[string]int    `json:"tier_counts"`
}

func NewService(repo guestrepo.Repository) *Service {
	return &Service{repo: repo}
}

func applyGuestPriorityScores(list []guestrepo.Guest) {
	for i := range list {
		score, tier, reasons := scoreGuestPriority(list[i])
		list[i].PriorityScore = score
		list[i].PriorityTier = tier
		list[i].PriorityReasons = reasons
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

func (s *Service) ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int, sort, priorityTier string) ([]guestrepo.Guest, *ServiceError) {
	tierFilter := strings.ToLower(strings.TrimSpace(priorityTier))
	switch tierFilter {
	case "critical", "important", "optional":
	default:
		tierFilter = ""
	}

	if tierFilter != "" {
		list, err := s.repo.ListGuests(ctx, eventID, guestListPrefetchMax, 0, sort)
		if err != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
		}
		applyGuestPriorityScores(list)
		if sort == "priority_desc" || sort == "priority_asc" {
			sortGuestsByPriority(list, sort)
		}
		filtered := make([]guestrepo.Guest, 0, len(list))
		for _, g := range list {
			if strings.ToLower(g.PriorityTier) == tierFilter {
				filtered = append(filtered, g)
			}
		}
		return paginateGuests(filtered, offset, limit), nil
	}

	if sort == "priority_desc" || sort == "priority_asc" {
		list, err := s.repo.ListGuests(ctx, eventID, guestListPrefetchMax, 0, sort)
		if err != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
		}
		applyGuestPriorityScores(list)
		sortGuestsByPriority(list, sort)
		return paginateGuests(list, offset, limit), nil
	}

	list, err := s.repo.ListGuests(ctx, eventID, limit, offset, sort)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
	}
	applyGuestPriorityScores(list)
	return list, nil
}

func (s *Service) RelationshipScoreOverview(ctx context.Context, eventID uuid.UUID) (*RelationshipScoreOverview, *ServiceError) {
	ranked, svcErr := s.ListGuests(ctx, eventID, 200, 0, "priority_desc", "")
	if svcErr != nil {
		return nil, svcErr
	}
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
	}, nil
}

func scoreGuestPriority(g guestrepo.Guest) (int, string, []string) {
	reasons := make([]string, 0, 6)

	relationshipWeight := relationshipScore(g.Relationship)
	rsvpCommitment := clamp01(float64(g.RSVPYesCount) / 3.0)
	responseSpeed := 0.0
	if g.LastRSVPAt != nil {
		ageDays := time.Since(*g.LastRSVPAt).Hours() / 24.0
		responseSpeed = clamp01(1.0 - (ageDays / 14.0))
	}
	eventCoverage := 0.0
	if g.SubEventTotal > 0 {
		eventCoverage = clamp01(float64(g.RSVPTotal) / float64(g.SubEventTotal))
	}
	historicalReliability := 0.0
	if g.RSVPTotal > 0 {
		historicalReliability = clamp01(float64(g.RSVPYesCount) / float64(g.RSVPTotal))
	}
	hostOverride := hostOverrideFromTags(g.Tags)

	base := (0.30 * relationshipWeight) +
		(0.20 * rsvpCommitment) +
		(0.15 * responseSpeed) +
		(0.15 * eventCoverage) +
		(0.10 * historicalReliability) +
		(0.10 * hostOverride)

	decay := 0.90
	if g.LastRSVPAt != nil {
		ageDays := time.Since(*g.LastRSVPAt).Hours() / 24.0
		decay = 0.75 + (0.25 * math.Exp(-ageDays/30.0))
	}
	missing := 0
	if g.RSVPTotal == 0 {
		missing += 3
	}
	if g.SubEventTotal == 0 {
		missing++
	}
	uncertainty := 1.0 - (0.08 * float64(missing))
	if uncertainty < 0.70 {
		uncertainty = 0.70
	}
	score := int(math.Round(100.0 * base * decay * uncertainty))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	if relationshipWeight >= 0.7 {
		reasons = append(reasons, "strong relationship weight")
	}
	if rsvpCommitment >= 0.66 {
		reasons = append(reasons, "high RSVP commitment ("+strconv.Itoa(g.RSVPYesCount)+")")
	}
	if eventCoverage >= 0.66 {
		reasons = append(reasons, "high event coverage")
	}
	if hostOverride > 0 {
		reasons = append(reasons, "host override applied")
	}
	if uncertainty < 0.9 {
		reasons = append(reasons, "score has uncertainty due to missing data")
	}

	tier := "Optional"
	switch {
	case score >= 80:
		tier = "Critical"
	case score >= 50:
		tier = "Important"
	}
	return score, tier, reasons
}

func relationshipScore(relationship string) float64 {
	switch strings.TrimSpace(strings.ToLower(relationship)) {
	case "close_family", "immediate_family":
		return 1.0
	case "family", "relative", "relatives":
		return 0.85
	case "friend", "friends":
		return 0.65
	case "colleague", "coworker":
		return 0.45
	case "":
		return 0.20
	default:
		return 0.35
	}
}

func hostOverrideFromTags(tags []string) float64 {
	for _, t := range tags {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt == "vip" || tt == "priority" || tt == "must_call" {
			return 1.0
		}
	}
	return 0.0
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
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
	return result, nil
}
