package httpserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestTierPricePaise(t *testing.T) {
	if got := tierPricePaise("pro"); got != 99000 {
		t.Fatalf("pro price mismatch: %d", got)
	}
	if got := tierPricePaise("elite"); got != 249000 {
		t.Fatalf("elite price mismatch: %d", got)
	}
	if got := tierPricePaise("free"); got != 0 {
		t.Fatalf("free price mismatch: %d", got)
	}
}

func TestVerifyRazorpayWebhookSignature(t *testing.T) {
	secret := "abc123"
	payload := []byte(`{"event":"payment.captured"}`)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(payload)
	sig := hex.EncodeToString(m.Sum(nil))
	if !verifyRazorpayWebhookSignature(secret, payload, sig) {
		t.Fatal("expected signature to verify")
	}
	if verifyRazorpayWebhookSignature(secret, payload, "deadbeef") {
		t.Fatal("expected signature verification to fail")
	}
}

