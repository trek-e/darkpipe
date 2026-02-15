// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// GenericProvider implements Provider for any IMAP/CalDAV/CardDAV server
type GenericProvider struct {
	IMAPHost    string
	IMAPPort    int
	Username    string
	Password    string
	CalDAVURL   string
	CardDAVURL  string
	UseTLS      bool
}

func init() {
	// Register generic provider in global registry
	DefaultRegistry.Register(&GenericProvider{})
}

func (p *GenericProvider) Name() string {
	return "Generic IMAP/CalDAV/CardDAV"
}

func (p *GenericProvider) Slug() string {
	return "generic"
}

func (p *GenericProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Default port if not specified
	port := p.IMAPPort
	if port == 0 {
		if p.UseTLS {
			port = 993
		} else {
			port = 143
		}
	}

	endpoint := fmt.Sprintf("%s:%d", p.IMAPHost, port)

	var client *imapclient.Client

	if p.UseTLS {
		// Direct TLS connection (IMAPS on port 993)
		conn, err := tls.Dial("tcp", endpoint, &tls.Config{
			ServerName: p.IMAPHost,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dial IMAP with TLS: %w", err)
		}

		client = imapclient.New(conn, nil)
	} else {
		// STARTTLS connection (IMAP on port 143)
		// Note: go-imap v2 handles STARTTLS differently
		// For now, we'll use direct TLS connection
		// TODO: Implement STARTTLS support if needed
		return nil, fmt.Errorf("STARTTLS not yet implemented - use port 993 with UseTLS=true")
	}

	// Login with credentials
	if err := client.Login(p.Username, p.Password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return client, nil
}

func (p *GenericProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	if p.CalDAVURL == "" {
		return nil, fmt.Errorf("CalDAV URL not configured")
	}

	// Create HTTP client with basic auth
	httpClient := &http.Client{
		Transport: &basicAuthTransport{
			Username: p.Username,
			Password: p.Password,
		},
	}

	// Create CalDAV client
	client, err := caldav.NewClient(httpClient, p.CalDAVURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create CalDAV client: %w", err)
	}

	return client, nil
}

func (p *GenericProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	if p.CardDAVURL == "" {
		return nil, fmt.Errorf("CardDAV URL not configured")
	}

	// Create HTTP client with basic auth
	httpClient := &http.Client{
		Transport: &basicAuthTransport{
			Username: p.Username,
			Password: p.Password,
		},
	}

	// Create CardDAV client
	client, err := carddav.NewClient(httpClient, p.CardDAVURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create CardDAV client: %w", err)
	}

	return client, nil
}

func (p *GenericProvider) SupportsLabels() bool {
	return false
}

func (p *GenericProvider) SupportsAPI() bool {
	return false
}

func (p *GenericProvider) SupportsCalDAV() bool {
	return p.CalDAVURL != ""
}

func (p *GenericProvider) SupportsCardDAV() bool {
	return p.CardDAVURL != ""
}

func (p *GenericProvider) GetFolderMapping() map[string]string {
	// No provider-specific mappings - user can override with --folder-map
	return map[string]string{}
}

func (p *GenericProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{}
}

func (p *GenericProvider) IMAPEndpoint() string {
	port := p.IMAPPort
	if port == 0 {
		if p.UseTLS {
			port = 993
		} else {
			port = 143
		}
	}
	return fmt.Sprintf("%s:%d", p.IMAPHost, port)
}

func (p *GenericProvider) CalDAVEndpoint() string {
	return p.CalDAVURL
}

func (p *GenericProvider) CardDAVEndpoint() string {
	return p.CardDAVURL
}

func (p *GenericProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "input",
			Label:    "IMAP Host",
			Field:    "imap_host",
			HelpText: "IMAP server hostname or IP address",
		},
		{
			Type:     "input",
			Label:    "IMAP Port (default: 993 for TLS, 143 for STARTTLS)",
			Field:    "imap_port",
			HelpText: "IMAP port number. Leave empty for default.",
		},
		{
			Type:     "input",
			Label:    "Use TLS (true/false)",
			Field:    "use_tls",
			HelpText: "Use direct TLS connection (port 993) vs STARTTLS (port 143). Recommended: true",
		},
		{
			Type:     "input",
			Label:    "Username",
			Field:    "username",
			HelpText: "Your email address or username",
		},
		{
			Type:     "input",
			Label:    "Password",
			Field:    "password",
			HelpText: "Your password",
		},
		{
			Type:     "input",
			Label:    "CalDAV URL (optional)",
			Field:    "caldav_url",
			HelpText: "CalDAV server URL (e.g., https://cal.example.com/caldav). Leave empty to skip calendar sync.",
		},
		{
			Type:     "input",
			Label:    "CardDAV URL (optional)",
			Field:    "carddav_url",
			HelpText: "CardDAV server URL (e.g., https://contacts.example.com/carddav). Leave empty to skip contact sync.",
		},
	}
}
