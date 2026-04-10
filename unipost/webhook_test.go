package unipost

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"event":"post.published"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhookSignature(payload, sig, secret) {
		t.Error("expected valid signature")
	}
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	if VerifyWebhookSignature(
		[]byte(`{"event":"post.published"}`),
		"0000000000000000000000000000000000000000000000000000000000000000",
		"whsec_test",
	) {
		t.Error("expected invalid signature")
	}
}
