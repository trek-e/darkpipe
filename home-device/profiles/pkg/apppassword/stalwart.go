// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package apppassword

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StalwartStore implements Store using Stalwart's REST API.
type StalwartStore struct {
	BaseURL      string
	AdminUser    string
	AdminPassword string
	Client       *http.Client
}

// NewStalwartStore creates a Stalwart backend store.
func NewStalwartStore() *StalwartStore {
	baseURL := os.Getenv("STALWART_BASE_URL")
	if baseURL == "" {
		baseURL = "http://stalwart:8080"
	}

	return &StalwartStore{
		BaseURL:       baseURL,
		AdminUser:     os.Getenv("STALWART_ADMIN_USER"),
		AdminPassword: os.Getenv("STALWART_ADMIN_PASSWORD"),
		Client:        &http.Client{Timeout: 10 * time.Second},
	}
}

// stalwartAppPasswordFormat formats app passwords as Stalwart expects.
// Format: $app$<device-name>$<bcrypt-hash>
func stalwartAppPasswordFormat(deviceName, hashedPassword string) string {
	// Sanitize device name (remove special chars)
	deviceName = strings.ReplaceAll(deviceName, "$", "-")
	return fmt.Sprintf("$app$%s$%s", deviceName, hashedPassword)
}

func (s *StalwartStore) Create(email, deviceName, plainPassword string) (*AppPassword, error) {
	if s.AdminUser == "" || s.AdminPassword == "" {
		return nil, fmt.Errorf("STALWART_ADMIN_USER and STALWART_ADMIN_PASSWORD must be set")
	}

	hash, err := HashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	appPassword := &AppPassword{
		ID:             uuid.New().String(),
		Email:          email,
		DeviceName:     deviceName,
		HashedPassword: hash,
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Time{},
	}

	// Store in Stalwart via REST API
	// PUT /api/principal/{email}/app-password/{device-name}
	url := fmt.Sprintf("%s/api/principal/%s/app-password/%s", s.BaseURL, email, deviceName)

	payload := map[string]string{
		"password": stalwartAppPasswordFormat(deviceName, hash),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(s.AdminUser, s.AdminPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stalwart api error: %s (status %d)", string(body), resp.StatusCode)
	}

	return appPassword, nil
}

func (s *StalwartStore) List(email string) ([]AppPassword, error) {
	if s.AdminUser == "" || s.AdminPassword == "" {
		return nil, fmt.Errorf("STALWART_ADMIN_USER and STALWART_ADMIN_PASSWORD must be set")
	}

	// GET /api/principal/{email}/app-passwords
	url := fmt.Sprintf("%s/api/principal/%s/app-passwords", s.BaseURL, email)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(s.AdminUser, s.AdminPassword)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []AppPassword{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stalwart api error: %s (status %d)", string(body), resp.StatusCode)
	}

	// Parse response (simplified - actual Stalwart API may differ)
	var passwords []AppPassword
	if err := json.NewDecoder(resp.Body).Decode(&passwords); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return passwords, nil
}

func (s *StalwartStore) Revoke(id string) error {
	if s.AdminUser == "" || s.AdminPassword == "" {
		return fmt.Errorf("STALWART_ADMIN_USER and STALWART_ADMIN_PASSWORD must be set")
	}

	// DELETE /api/app-password/{id}
	url := fmt.Sprintf("%s/api/app-password/%s", s.BaseURL, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(s.AdminUser, s.AdminPassword)

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stalwart api error: %s (status %d)", string(body), resp.StatusCode)
	}

	return nil
}

func (s *StalwartStore) Verify(email, plainPassword string) (bool, error) {
	// Stalwart handles verification internally during IMAP/SMTP auth
	// This method would typically be called by an external auth script
	// For now, we check against stored passwords
	passwords, err := s.List(email)
	if err != nil {
		return false, fmt.Errorf("list passwords: %w", err)
	}

	for _, pw := range passwords {
		if VerifyPassword(plainPassword, pw.HashedPassword) {
			// Update LastUsedAt (would need separate API call in real implementation)
			return true, nil
		}
	}

	return false, nil
}
