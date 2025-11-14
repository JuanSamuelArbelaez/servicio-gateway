package main

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestExtractToken_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")

	token, err := extractToken(req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token != "test-token-123" {
		t.Errorf("Expected 'test-token-123', got '%s'", token)
	}
}

func TestExtractToken_MissingHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	_, err := extractToken(req)

	if err == nil {
		t.Error("Expected error for missing Authorization header")
	}
}

func TestExtractToken_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	_, err := extractToken(req)

	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	// Setup secret
	os.Setenv("JWT_SECRET", "test-secret")
	jwtSecret = []byte("test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Create valid token
	claims := jwt.MapClaims{
		"userId": "123",
		"email":  "test@example.com",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	// Execute
	resultClaims, err := validateJWT(tokenString)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resultClaims["userId"] != "123" {
		t.Errorf("Expected userId '123', got '%v'", resultClaims["userId"])
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	jwtSecret = []byte("test-secret")
	defer os.Unsetenv("JWT_SECRET")

	_, err := validateJWT("invalid.token.here")

	if err == nil {
		t.Error("Expected error for invalid token")
	}
}