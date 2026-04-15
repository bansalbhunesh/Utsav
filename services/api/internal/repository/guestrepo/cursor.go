package guestrepo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var ErrInvalidGuestListCursor = errors.New("invalid guest list cursor")

// ListGuestsCursor is an opaque pagination token for keyset (seek) queries.
// Sort field must match the active list sort mode.
type ListGuestsCursor struct {
	Sort          string `json:"s"`
	ID            string `json:"id"`
	Name          string `json:"n,omitempty"`
	PriorityScore int    `json:"ps,omitempty"`
	RSVPYes       int    `json:"rsvp,omitempty"`
	ShagunPaise   int64  `json:"sg,omitempty"`
}

func EncodeListGuestsCursor(c ListGuestsCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func DecodeListGuestsCursor(s string) (ListGuestsCursor, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ListGuestsCursor{}, ErrInvalidGuestListCursor
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ListGuestsCursor{}, ErrInvalidGuestListCursor
	}
	var c ListGuestsCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return ListGuestsCursor{}, ErrInvalidGuestListCursor
	}
	if c.ID == "" || c.Sort == "" {
		return ListGuestsCursor{}, ErrInvalidGuestListCursor
	}
	return c, nil
}

func CursorFromGuestRow(sort string, g Guest) ListGuestsCursor {
	c := ListGuestsCursor{Sort: sort, ID: g.ID, Name: g.Name}
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "priority_desc", "priority_asc":
		c.PriorityScore = g.PriorityScore
	case "rsvp_desc":
		c.RSVPYes = g.RSVPYesCount
	case "shagun_desc":
		c.ShagunPaise = g.ShagunPaise
	}
	return c
}
