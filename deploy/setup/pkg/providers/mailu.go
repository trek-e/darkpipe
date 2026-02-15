// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// MailuProvider implements Provider for Mailu with API access
type MailuProvider struct {
	APIURL   string
	APIKey   string
	IMAPHost string
	Username string
	Password string
}

// MailuUser represents a user from Mailu API
type MailuUser struct {
	Email    string `json:"email"`
	Name     string `json:"comment"`
	Enabled  bool   `json:"enabled"`
}

// MailuAlias represents an alias from Mailu API
type MailuAlias struct {
	Email       string   `json:"email"`
	Destination []string `json:"destination"`
	Enabled     bool     `json:"enabled"`
}

// MailuDomain represents a domain from Mailu API
type MailuDomain struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func init() {
	// Register Mailu provider in global registry
	DefaultRegistry.Register(&MailuProvider{})
}

func (p *MailuProvider) Name() string {
	return "Mailu"
}

func (p *MailuProvider) Slug() string {
	return "mailu"
}

func (p *MailuProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to IMAPHost:993
	conn, err := tls.Dial("tcp", p.IMAPHost+":993", &tls.Config{
		ServerName: p.IMAPHost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial Mailu IMAP: %w", err)
	}

	// Create IMAP client
	client := imapclient.New(conn, nil)

	// Login with credentials
	if err := client.Login(p.Username, p.Password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return client, nil
}

func (p *MailuProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// Mailu doesn't have built-in CalDAV
	return nil, fmt.Errorf("Mailu does not support CalDAV - use separate CalDAV server or file import")
}

func (p *MailuProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// Mailu doesn't have built-in CardDAV
	return nil, fmt.Errorf("Mailu does not support CardDAV - use separate CardDAV server or file import")
}

func (p *MailuProvider) SupportsLabels() bool {
	return false
}

func (p *MailuProvider) SupportsAPI() bool {
	return true
}

func (p *MailuProvider) SupportsCalDAV() bool {
	return false
}

func (p *MailuProvider) SupportsCardDAV() bool {
	return false
}

func (p *MailuProvider) GetFolderMapping() map[string]string {
	// Mailu uses standard IMAP folder names
	return map[string]string{}
}

func (p *MailuProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{}
}

func (p *MailuProvider) IMAPEndpoint() string {
	return p.IMAPHost + ":993"
}

func (p *MailuProvider) CalDAVEndpoint() string {
	return ""
}

func (p *MailuProvider) CardDAVEndpoint() string {
	return ""
}

func (p *MailuProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "input",
			Label:    "Mailu API URL",
			Field:    "api_url",
			HelpText: "Your Mailu instance URL (e.g., https://mail.example.com). Leave empty to skip API and use IMAP only.",
		},
		{
			Type:     "input",
			Label:    "Mailu API Key (optional)",
			Field:    "api_key",
			HelpText: "API key from Mailu admin panel. Leave empty for IMAP-only mode.",
		},
		{
			Type:     "input",
			Label:    "IMAP Host",
			Field:    "imap_host",
			HelpText: "IMAP server hostname (usually same as API URL domain)",
		},
		{
			Type:     "input",
			Label:    "Email Address",
			Field:    "username",
			HelpText: "Your mailbox email address",
		},
		{
			Type:     "input",
			Label:    "Password",
			Field:    "password",
			HelpText: "Your mailbox password",
		},
	}
}

// GetUsers retrieves all users from Mailu API
func (p *MailuProvider) GetUsers(ctx context.Context) ([]MailuUser, error) {
	if p.APIURL == "" || p.APIKey == "" {
		return nil, fmt.Errorf("API not configured - use IMAP-only mode")
	}

	url := p.APIURL + "/api/v1/user"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed (fallback to IMAP-only): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d (fallback to IMAP-only): %s", resp.StatusCode, string(body))
	}

	var users []MailuUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// GetAliases retrieves all aliases from Mailu API
func (p *MailuProvider) GetAliases(ctx context.Context) ([]MailuAlias, error) {
	if p.APIURL == "" || p.APIKey == "" {
		return nil, fmt.Errorf("API not configured - use IMAP-only mode")
	}

	url := p.APIURL + "/api/v1/alias"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var aliases []MailuAlias
	if err := json.NewDecoder(resp.Body).Decode(&aliases); err != nil {
		return nil, fmt.Errorf("failed to decode aliases: %w", err)
	}

	return aliases, nil
}

// GetDomains retrieves all domains from Mailu API
func (p *MailuProvider) GetDomains(ctx context.Context) ([]MailuDomain, error) {
	if p.APIURL == "" || p.APIKey == "" {
		return nil, fmt.Errorf("API not configured - use IMAP-only mode")
	}

	url := p.APIURL + "/api/v1/domain"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var domains []MailuDomain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, fmt.Errorf("failed to decode domains: %w", err)
	}

	return domains, nil
}
