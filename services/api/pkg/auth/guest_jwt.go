package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GuestClaims struct {
	jwt.RegisteredClaims
}

// Guest token subject format: "<eventUUID>|<E164 phone>"
func SignGuestToken(eventID uuid.UUID, phone string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	sub := eventID.String() + "|" + phone
	claims := GuestClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    "utsav-guest",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return t.SignedString(secret)
}

func ParseGuestToken(token string, secret []byte) (eventID uuid.UUID, phone string, err error) {
	parsed, err := jwt.ParseWithClaims(token, &GuestClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	claims, ok := parsed.Claims.(*GuestClaims)
	if !ok || !parsed.Valid || claims.Issuer != "utsav-guest" {
		return uuid.Nil, "", fmt.Errorf("invalid guest claims")
	}
	parts := strings.SplitN(claims.Subject, "|", 2)
	if len(parts) != 2 {
		return uuid.Nil, "", fmt.Errorf("invalid guest subject")
	}
	eid, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", err
	}
	return eid, parts[1], nil
}
