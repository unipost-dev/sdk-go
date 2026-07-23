package unipost

import (
	"os"
	"strings"
	"testing"
)

func TestReleaseVersion(t *testing.T) {
	if sdkVersion != "0.6.0" {
		t.Fatalf("sdkVersion = %q, want 0.6.0", sdkVersion)
	}
	if userAgent != "unipost-go/0.6.0" {
		t.Fatalf("userAgent = %q, want unipost-go/0.6.0", userAgent)
	}
}

func TestReleaseDocumentationCoversInboxContract(t *testing.T) {
	readme := readReleaseDocument(t, "../README.md")
	requireReleaseMarkers(t, "README.md", readme, []string{
		"Latest release: v0.6.0",
		"go get github.com/unipost-dev/sdk-go@v0.6.0",
		"Production Inbox Integration",
		"ManagedUser(",
		"Workspace()",
		"WithIdempotencyKey(",
		"XOutboundStatus(",
		"WebSocketConnectionDetails()",
		"SyncXBackfill(",
		"XInboxBackfillConfirmationRequired",
		"XInboxBackfillInProgress",
		"XInboxBackfillCompleted",
		"errors.As(err, &apiErr)",
		"*unipost.APIError",
	})
	for _, obsoleteType := range []string{
		"unipost.AuthError",
		"unipost.RateLimitError",
		"unipost.ValidationError",
		"unipost.UniPostError",
	} {
		if strings.Contains(readme, obsoleteType) {
			t.Errorf("README.md references nonexistent %q", obsoleteType)
		}
	}

	changelog := readReleaseDocument(t, "../CHANGELOG.md")
	requireReleaseMarkers(t, "CHANGELOG.md", changelog, []string{
		"## [0.6.0]",
		"## [0.5.0]",
		"managed-user",
		"idempotency",
		"WebSocket",
		"X backfill",
	})
}

func readReleaseDocument(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}

func requireReleaseMarkers(t *testing.T, name, document string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		if !strings.Contains(document, marker) {
			t.Errorf("%s is missing %q", name, marker)
		}
	}
}
