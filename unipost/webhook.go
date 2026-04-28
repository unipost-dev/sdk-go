package unipost

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a webhook request.
//
// payload is the raw request body, signature is the X-UniPost-Signature header
// value (with or without a "sha256=" prefix), and secret is your webhook secret.
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	signature = strings.TrimSpace(signature)
	signature = strings.TrimPrefix(strings.ToLower(signature), "sha256=")
	if signature == "" || secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
