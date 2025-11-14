package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyRequest_Success(t *testing.T) {
	// Mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer mockServer.Close()

	// Execute
	status, body, headers, err := ProxyRequest("GET", mockServer.URL, nil, http.Header{})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}
	if string(body) != `{"message":"success"}` {
		t.Errorf("Expected body '{\"message\":\"success\"}', got '%s'", string(body))
	}
	if headers.Get("X-Test-Header") != "test-value" {
		t.Errorf("Expected header 'test-value', got '%s'", headers.Get("X-Test-Header"))
	}
}

func TestProxyRequest_WithBody(t *testing.T) {
	// Mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test payload" {
			t.Errorf("Expected body 'test payload', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	// Execute
	status, _, _, err := ProxyRequest("POST", mockServer.URL, bytes.NewReader([]byte("test payload")), http.Header{})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", status)
	}
}

func TestPostEvent_Success(t *testing.T) {
	// Mock event bus server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" {
			t.Errorf("Expected path '/events', got '%s'", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer mockServer.Close()

	// Execute
	event := map[string]interface{}{
		"type": "test.event",
		"payload": map[string]string{"key": "value"},
	}
	err := PostEvent(mockServer.URL, event)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestPostEvent_EmptyURL(t *testing.T) {
	// Execute
	event := map[string]interface{}{"type": "test"}
	err := PostEvent("", event)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for empty URL, got %v", err)
	}
}