package otp

import (
	"crypto/rand"
	"fmt"
)

// GenerateNumericCode creates a zero-padded 6-digit OTP code.
func GenerateNumericCode() (string, error) {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	n := int(raw[0])<<24 | int(raw[1])<<16 | int(raw[2])<<8 | int(raw[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000), nil
}
