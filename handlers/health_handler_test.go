package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse JSON response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify status field
	if response["status"] != "UP" {
		t.Errorf("Expected status 'UP', got '%v'", response["status"])
	}

	// Verify version field exists
	if _, ok := response["version"]; !ok {
		t.Error("Expected 'version' field in response")
	}

	// Verify uptime field exists
	if _, ok := response["uptime"]; !ok {
		t.Error("Expected 'uptime' field in response")
	}
}