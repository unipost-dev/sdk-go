package unipost

import (
	"os"
	"testing"
)

func TestNewClient_WithAPIKey(t *testing.T) {
	client := NewClient(WithAPIKey("up_test_xxx"))
	if client.Posts == nil {
		t.Error("expected Posts service")
	}
	if client.Accounts == nil {
		t.Error("expected Accounts service")
	}
	if client.Media == nil {
		t.Error("expected Media service")
	}
	if client.Analytics == nil {
		t.Error("expected Analytics service")
	}
	if client.Connect == nil {
		t.Error("expected Connect service")
	}
	if client.Users == nil {
		t.Error("expected Users service")
	}
}

func TestNewClient_PanicsWithoutKey(t *testing.T) {
	// Ensure env var is not set
	old := os.Getenv("UNIPOST_API_KEY")
	os.Unsetenv("UNIPOST_API_KEY")
	defer func() {
		if old != "" {
			os.Setenv("UNIPOST_API_KEY", old)
		}
	}()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic without API key")
		}
	}()
	NewClient()
}
