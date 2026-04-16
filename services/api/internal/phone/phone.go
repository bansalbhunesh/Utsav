package phone

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidPhone is returned when a value cannot be normalized to E.164.
	ErrInvalidPhone = errors.New("invalid phone")
	e164Regex       = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
)

const defaultCountryCode = "91"

// NormalizeE164 converts common user-entered formats into canonical E.164.
// It accepts separators like spaces/hyphens/parentheses and returns +<digits>.
func NormalizeE164(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrInvalidPhone
	}

	// Convert 00-country prefix to "+".
	if strings.HasPrefix(s, "00") {
		s = "+" + strings.TrimPrefix(s, "00")
	}

	plus := false
	digits := make([]rune, 0, len(s))
	for i, r := range s {
		switch {
		case r == '+' && i == 0:
			plus = true
		case r >= '0' && r <= '9':
			digits = append(digits, r)
		case r == ' ' || r == '-' || r == '(' || r == ')' || r == '.':
			continue
		default:
			return "", ErrInvalidPhone
		}
	}
	if len(digits) == 0 {
		return "", ErrInvalidPhone
	}

	d := string(digits)
	if !plus {
		// Common local Indian mobile forms like 09876543210 or 9876543210.
		for strings.HasPrefix(d, "0") {
			d = strings.TrimPrefix(d, "0")
		}
		if len(d) == 10 {
			d = defaultCountryCode + d
		}
	}

	out := "+" + d
	if !e164Regex.MatchString(out) {
		return "", ErrInvalidPhone
	}
	return out, nil
}
