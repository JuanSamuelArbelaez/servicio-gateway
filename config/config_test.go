package config

import (
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Setup
	os.Setenv("SECURITY_URL", "http://test-security:8080")
	os.Setenv("PROFILE_URL", "http://test-profile:8087")
	os.Setenv("EVENT_BUS_URL", "http://test-eventbus:8085")
	os.Setenv("PORT", "9999")
	defer func() {
		os.Unsetenv("SECURITY_URL")
		os.Unsetenv("PROFILE_URL")
		os.Unsetenv("EVENT_BUS_URL")
		os.Unsetenv("PORT")
	}()

	// Execute
	cfg := LoadConfigFromEnv()

	// Assert
	if cfg.SecurityURL != "http://test-security:8080" {
		t.Errorf("Expected SecurityURL 'http://test-security:8080', got '%s'", cfg.SecurityURL)
	}
	if cfg.ProfileURL != "http://test-profile:8087" {
		t.Errorf("Expected ProfileURL 'http://test-profile:8087', got '%s'", cfg.ProfileURL)
	}
	if cfg.EventBusURL != "http://test-eventbus:8085" {
		t.Errorf("Expected EventBusURL 'http://test-eventbus:8085', got '%s'", cfg.EventBusURL)
	}
	if cfg.Port != "9999" {
		t.Errorf("Expected Port '9999', got '%s'", cfg.Port)
	}
}

func TestLoadConfigFromEnv_DefaultPort(t *testing.T) {
	// Setup - no PORT env var
	os.Unsetenv("PORT")

	// Execute
	cfg := LoadConfigFromEnv()

	// Assert
	if cfg.Port != "8088" {
		t.Errorf("Expected default Port '8088', got '%s'", cfg.Port)
	}
}