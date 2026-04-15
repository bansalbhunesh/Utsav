package otp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashCode creates a versioned OTP hash: v1:<salt-hex>:<mac-hex>.
func HashCode(secret []byte, code string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, append(secret, salt...))
	_, _ = mac.Write([]byte(strings.TrimSpace(code)))
	return fmt.Sprintf("v1:%s:%s", hex.EncodeToString(salt), hex.EncodeToString(mac.Sum(nil))), nil
}

// VerifyCode validates the OTP against either v1 HMAC format or legacy bcrypt.
func VerifyCode(secret []byte, storedHash, code string) bool {
	parts := strings.Split(strings.TrimSpace(storedHash), ":")
	if len(parts) == 3 && strings.EqualFold(parts[0], "v1") {
		salt, err := hex.DecodeString(parts[1])
		if err != nil {
			return false
		}
		expected, err := hex.DecodeString(parts[2])
		if err != nil {
			return false
		}
		mac := hmac.New(sha256.New, append(secret, salt...))
		_, _ = mac.Write([]byte(strings.TrimSpace(code)))
		got := mac.Sum(nil)
		return subtle.ConstantTimeCompare(got, expected) == 1
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(strings.TrimSpace(code))) == nil
}
