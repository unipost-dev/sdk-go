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
	if client.Workspace == nil {
		t.Error("expected Workspace service")
	}
	if client.APIKeys == nil {
		t.Error("expected APIKeys service")
	}
	if client.Webhooks == nil {
		t.Error("expected Webhooks service")
	}
	if client.PlatformCredentials == nil {
		t.Error("expected PlatformCredentials service")
	}
	if client.DeliveryJobs == nil {
		t.Error("expected DeliveryJobs service")
	}
	if client.Logs == nil {
		t.Error("expected Logs service")
	}
}

func TestNewClient_FromEnv(t *testing.T) {
	old := os.Getenv("UNIPOST_API_KEY")
	os.Setenv("UNIPOST_API_KEY", "up_test_env_key")
	defer os.Setenv("UNIPOST_API_KEY", old)

	client := NewClient()
	if client.apiKey != "up_test_env_key" {
		t.Errorf("expected api key from env, got %q", client.apiKey)
	}
}
