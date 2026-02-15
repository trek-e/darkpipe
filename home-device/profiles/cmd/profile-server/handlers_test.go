// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// mockAppPasswordStore is a simple in-memory store for testing.
type mockAppPasswordStore struct {
	passwords map[string][]apppassword.AppPassword
}

func newMockAppPasswordStore() *mockAppPasswordStore {
	return &mockAppPasswordStore{
		passwords: make(map[string][]apppassword.AppPassword),
	}
}

func (m *mockAppPasswordStore) Create(email, deviceName, plainPassword string) (*apppassword.AppPassword, error) {
	ap := &apppassword.AppPassword{
		ID:         "test-id",
		Email:      email,
		DeviceName: deviceName,
		CreatedAt:  time.Now(),
	}
	m.passwords[email] = append(m.passwords[email], *ap)
	return ap, nil
}

func (m *mockAppPasswordStore) List(email string) ([]apppassword.AppPassword, error) {
	return m.passwords[email], nil
}

func (m *mockAppPasswordStore) Revoke(id string) error {
	return nil
}

func (m *mockAppPasswordStore) Verify(email, plainPassword string) (bool, error) {
	return true, nil
}

func setupTestHandler() *ProfileHandler {
	return &ProfileHandler{
		ProfileGen:   &mobileconfig.ProfileGenerator{},
		TokenStore:   qrcode.NewMemoryTokenStore(),
		AppPassStore: newMockAppPasswordStore(),
		Config: ServerConfig{
			Domain:      "example.com",
			Hostname:    "mail.example.com",
			CalDAVURL:   "https://mail.example.com/caldav",
			CardDAVURL:  "https://mail.example.com/carddav",
			CalDAVPort:  443,
			CardDAVPort: 443,
			AdminUser:   "admin",
			AdminPass:   "testpass",
		},
	}
}

func TestHandleProfileDownloadWithValidToken(t *testing.T) {
	handler := setupTestHandler()

	// Create a valid token
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/x-apple-aspen-config" {
		t.Errorf("Expected Content-Type 'application/x-apple-aspen-config', got '%s'", contentType)
	}

	// Check Content-Disposition header
	contentDisposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "darkpipe-mail.mobileconfig") {
		t.Errorf("Expected Content-Disposition to contain filename, got '%s'", contentDisposition)
	}

	// Check body is not empty
	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestHandleProfileDownloadWithInvalidToken(t *testing.T) {
	handler := setupTestHandler()

	// Use invalid token
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token=invalid-token", nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleProfileDownloadWithExpiredToken(t *testing.T) {
	handler := setupTestHandler()

	// Create expired token
	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Minute) // Already expired
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleProfileDownloadSingleUseEnforcement(t *testing.T) {
	handler := setupTestHandler()

	// Create valid token
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w1 := httptest.NewRecorder()
	handler.HandleProfileDownload(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request expected status 200, got %d", w1.Code)
	}

	// Second request with same token should fail (single-use)
	req2 := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w2 := httptest.NewRecorder()
	handler.HandleProfileDownload(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Second request expected status 401, got %d", w2.Code)
	}
}

func TestHandleProfileDownloadMissingToken(t *testing.T) {
	handler := setupTestHandler()

	// Request without token parameter
	req := httptest.NewRequest(http.MethodGet, "/profile/download", nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAutoconfig(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/mail/config-v1.1.xml?emailaddress=test@example.com", nil)
	w := httptest.NewRecorder()

	handler.HandleAutoconfig(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/xml") {
		t.Errorf("Expected Content-Type to contain 'application/xml', got '%s'", contentType)
	}

	// Verify it's valid XML
	var result interface{}
	if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Response is not valid XML: %v", err)
	}

	// Check body contains expected elements
	body := w.Body.String()
	if !strings.Contains(body, "mail.example.com") {
		t.Error("Expected response to contain mail hostname")
	}
}

func TestHandleAutodiscover(t *testing.T) {
	handler := setupTestHandler()

	// Test GET request (simpler case)
	req := httptest.NewRequest(http.MethodGet, "/autodiscover/autodiscover.xml", nil)
	w := httptest.NewRecorder()

	handler.HandleAutodiscover(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/xml") {
		t.Errorf("Expected Content-Type to contain 'application/xml', got '%s'", contentType)
	}

	// Verify it's valid XML
	var result interface{}
	if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Response is not valid XML: %v", err)
	}
}

func TestHandleHealth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestHandleQRGenerateWithAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/generate?email=test@example.com", nil)
	req.SetBasicAuth("admin", "testpass")
	w := httptest.NewRecorder()

	handler.HandleQRGenerate(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Expected Content-Type 'image/png', got '%s'", contentType)
	}

	// Check PNG magic number
	body := w.Body.Bytes()
	if len(body) < 8 {
		t.Fatal("Response body too short")
	}

	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if body[i] != pngMagic[i] {
			t.Errorf("PNG magic number mismatch at byte %d", i)
			break
		}
	}
}

func TestHandleQRGenerateWithoutAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/generate?email=test@example.com", nil)
	w := httptest.NewRecorder()

	handler.HandleQRGenerate(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleQRImageWithAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/image?email=test@example.com", nil)
	req.SetBasicAuth("admin", "testpass")
	w := httptest.NewRecorder()

	handler.HandleQRImage(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Expected Content-Type 'image/png', got '%s'", contentType)
	}

	// Should NOT have Content-Disposition (inline)
	contentDisposition := w.Header().Get("Content-Disposition")
	if contentDisposition != "" {
		t.Errorf("Expected no Content-Disposition for inline image, got '%s'", contentDisposition)
	}
}

func TestExtractEmailFromAutodiscoverXML(t *testing.T) {
	xmlBody := `<?xml version="1.0"?>
	<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
		<Request>
			<EMailAddress>test@example.com</EMailAddress>
			<AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
		</Request>
	</Autodiscover>`

	email := extractEmailFromAutodiscoverXML(xmlBody)
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}
}
