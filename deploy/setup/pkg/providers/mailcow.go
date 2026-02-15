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

// MailCowProvider implements Provider for MailCow with API access
type MailCowProvider struct {
	APIURL   string
	APIKey   string
	IMAPHost string
	Username string
	Password string
}

// MailCowMailbox represents a mailbox from MailCow API
type MailCowMailbox struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	QuotaUsed  int64  `json:"quota_used"`
	Messages   int    `json:"messages"`
	Active     bool   `json:"active"`
}

// MailCowAlias represents an alias from MailCow API
type MailCowAlias struct {
	Address string `json:"address"`
	GoTo    string `json:"goto"`
	Active  bool   `json:"active"`
}

// MailCowDomain represents a domain from MailCow API
type MailCowDomain struct {
	DomainName string `json:"domain_name"`
	Active     bool   `json:"active"`
}

// MailCowExport aggregates all API data for migration preview
type MailCowExport struct {
	Mailboxes []MailCowMailbox
	Aliases   []MailCowAlias
	Domains   []MailCowDomain
}

func init() {
	// Register MailCow provider in global registry
	DefaultRegistry.Register(&MailCowProvider{})
}

func (p *MailCowProvider) Name() string {
	return "MailCow"
}

func (p *MailCowProvider) Slug() string {
	return "mailcow"
}

func (p *MailCowProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to IMAPHost:993
	conn, err := tls.Dial("tcp", p.IMAPHost+":993", &tls.Config{
		ServerName: p.IMAPHost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial MailCow IMAP: %w", err)
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

func (p *MailCowProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// MailCow doesn't have built-in CalDAV
	return nil, fmt.Errorf("MailCow does not support CalDAV - use separate CalDAV server or file import")
}

func (p *MailCowProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// MailCow doesn't have built-in CardDAV
	return nil, fmt.Errorf("MailCow does not support CardDAV - use separate CardDAV server or file import")
}

func (p *MailCowProvider) SupportsLabels() bool {
	return false
}

func (p *MailCowProvider) SupportsAPI() bool {
	return true
}

func (p *MailCowProvider) SupportsCalDAV() bool {
	return false
}

func (p *MailCowProvider) SupportsCardDAV() bool {
	return false
}

func (p *MailCowProvider) GetFolderMapping() map[string]string {
	// MailCow uses standard IMAP folder names
	return map[string]string{}
}

func (p *MailCowProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{}
}

func (p *MailCowProvider) IMAPEndpoint() string {
	return p.IMAPHost + ":993"
}

func (p *MailCowProvider) CalDAVEndpoint() string {
	return ""
}

func (p *MailCowProvider) CardDAVEndpoint() string {
	return ""
}

func (p *MailCowProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "input",
			Label:    "MailCow API URL",
			Field:    "api_url",
			HelpText: "Your MailCow instance URL (e.g., https://mail.example.com)",
		},
		{
			Type:     "input",
			Label:    "MailCow API Key",
			Field:    "api_key",
			HelpText: "API key from MailCow admin panel (Configuration > Access > API)",
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

// GetMailboxes retrieves all mailboxes from MailCow API
func (p *MailCowProvider) GetMailboxes(ctx context.Context) ([]MailCowMailbox, error) {
	url := p.APIURL + "/api/v1/get/mailbox/all"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mailboxes []MailCowMailbox
	if err := json.NewDecoder(resp.Body).Decode(&mailboxes); err != nil {
		return nil, fmt.Errorf("failed to decode mailboxes: %w", err)
	}

	return mailboxes, nil
}

// GetAliases retrieves all aliases from MailCow API
func (p *MailCowProvider) GetAliases(ctx context.Context) ([]MailCowAlias, error) {
	url := p.APIURL + "/api/v1/get/alias/all"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var aliases []MailCowAlias
	if err := json.NewDecoder(resp.Body).Decode(&aliases); err != nil {
		return nil, fmt.Errorf("failed to decode aliases: %w", err)
	}

	return aliases, nil
}

// GetDomains retrieves all domains from MailCow API
func (p *MailCowProvider) GetDomains(ctx context.Context) ([]MailCowDomain, error) {
	url := p.APIURL + "/api/v1/get/domain/all"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var domains []MailCowDomain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, fmt.Errorf("failed to decode domains: %w", err)
	}

	return domains, nil
}

// ExportConfig aggregates all API data for migration preview
func (p *MailCowProvider) ExportConfig(ctx context.Context) (*MailCowExport, error) {
	mailboxes, err := p.GetMailboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailboxes: %w", err)
	}

	aliases, err := p.GetAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %w", err)
	}

	domains, err := p.GetDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}

	return &MailCowExport{
		Mailboxes: mailboxes,
		Aliases:   aliases,
		Domains:   domains,
	}, nil
}
