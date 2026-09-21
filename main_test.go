package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	generateKeys()

	if validKey.PrivateKey == nil {
		t.Error("valid private key was not generated")
	}

	if expiredKey.PrivateKey == nil {
		t.Error("expired private key was not generated")
	}

	if validKey.Kid == "" || expiredKey.Kid == "" {
		t.Error("key IDs were not generated")
	}
}

func TestJWKSHandler(t *testing.T) {
	generateKeys()

	request := httptest.NewRequest(
		http.MethodGet,
		"/.well-known/jwks.json",
		nil,
	)

	recorder := httptest.NewRecorder()

	jwksHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	var response map[string]interface{}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	if err != nil {
		t.Fatal("response was not valid JSON")
	}

	keys, exists := response["keys"]

	if !exists {
		t.Fatal("JWKS response does not contain keys")
	}

	if strings.Contains(recorder.Body.String(), "expired-key") {
		t.Error("JWKS should not contain expired keys")
	}

	if keys == nil {
		t.Error("JWKS keys should not be nil")
	}
}

func TestAuthHandler(t *testing.T) {
	generateKeys()

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth",
		nil,
	)

	recorder := httptest.NewRecorder()

	authHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Body.Len() == 0 {
		t.Error("expected a JWT")
	}
}

func TestExpiredAuth(t *testing.T) {
	generateKeys()

	request := httptest.NewRequest(
		http.MethodPost,
		"/auth?expired=true",
		nil,
	)

	recorder := httptest.NewRecorder()

	authHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Body.Len() == 0 {
		t.Error("expected an expired JWT")
	}
}
func TestSetupServer(t *testing.T) {
	generateKeys()

	server := setupServer()

	request := httptest.NewRequest(
		http.MethodGet,
		"/.well-known/jwks.json",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}
