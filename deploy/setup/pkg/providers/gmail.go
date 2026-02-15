// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GmailProvider implements Provider for Gmail with OAuth2
type GmailProvider struct {
	Email        string
	Token        *oauth2.Token
	ClientID     string
	ClientSecret string
}

func init() {
	// Register Gmail provider in global registry
	DefaultRegistry.Register(&GmailProvider{})
}

func (p *GmailProvider) Name() string {
	return "Gmail"
}

func (p *GmailProvider) Slug() string {
	return "gmail"
}

func (p *GmailProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to imap.gmail.com:993
	conn, err := tls.Dial("tcp", "imap.gmail.com:993", &tls.Config{
		ServerName: "imap.gmail.com",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial Gmail IMAP: %w", err)
	}

	// Create IMAP client
	client := imapclient.New(conn, nil)

	// Authenticate with XOAUTH2 SASL
	saslClient := NewXOAUTH2Client(p.Email, p.Token.AccessToken)
	if err := client.Authenticate(saslClient); err != nil {
		client.Close()
		return nil, fmt.Errorf("XOAUTH2 authentication failed: %w", err)
	}

	return client, nil
}

func (p *GmailProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// Create HTTP client with OAuth2 bearer auth
	config := &OAuthConfig{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint:     google.Endpoint,
	}
	tokenSource := TokenSourceFromToken(config, p.Token)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	// Create CalDAV client
	client, err := caldav.NewClient(httpClient, "https://www.googleapis.com/caldav/v2/")
	if err != nil {
		return nil, fmt.Errorf("failed to create CalDAV client: %w", err)
	}

	return client, nil
}

func (p *GmailProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// Create HTTP client with OAuth2 bearer auth
	config := &OAuthConfig{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint:     google.Endpoint,
	}
	tokenSource := TokenSourceFromToken(config, p.Token)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	// Create CardDAV client
	client, err := carddav.NewClient(httpClient, "https://www.googleapis.com/.well-known/carddav")
	if err != nil {
		return nil, fmt.Errorf("failed to create CardDAV client: %w", err)
	}

	return client, nil
}

func (p *GmailProvider) SupportsLabels() bool {
	return true
}

func (p *GmailProvider) SupportsAPI() bool {
	return false
}

func (p *GmailProvider) SupportsCalDAV() bool {
	return true
}

func (p *GmailProvider) SupportsCardDAV() bool {
	return true
}

func (p *GmailProvider) GetFolderMapping() map[string]string {
	// Gmail-specific folder mappings per locked decision
	return map[string]string{
		"[Gmail]/Sent Mail":     "Sent",
		"[Gmail]/Drafts":        "Drafts",
		"[Gmail]/Trash":         "Trash",
		"[Gmail]/Spam":          "Junk",
		"[Gmail]/All Mail":      "", // Skip
		"[Gmail]/Important":     "", // Skip
		"[Gmail]/Starred":       "", // Skip
	}
}

func (p *GmailProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{
		"[Gmail]/All Mail":  true,
		"[Gmail]/Important": true,
		"[Gmail]/Starred":   true,
	}
}

func (p *GmailProvider) IMAPEndpoint() string {
	return "imap.gmail.com:993"
}

func (p *GmailProvider) CalDAVEndpoint() string {
	return "https://www.googleapis.com/caldav/v2/"
}

func (p *GmailProvider) CardDAVEndpoint() string {
	return "https://www.googleapis.com/.well-known/carddav"
}

func (p *GmailProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "oauth",
			Label:    "Gmail OAuth2 Authentication",
			Field:    "oauth",
			HelpText: "You will need to create OAuth2 credentials in Google Cloud Console. Set GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables.",
		},
		{
			Type:     "input",
			Label:    "Gmail Email Address",
			Field:    "email",
			HelpText: "Your full Gmail address (e.g., user@gmail.com)",
		},
	}
}

// GetGmailOAuthConfig returns OAuth2 config for Gmail
// Requires GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables
func GetGmailOAuthConfig() (*OAuthConfig, error) {
	clientID := os.Getenv("GMAIL_CLIENT_ID")
	clientSecret := os.Getenv("GMAIL_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables required")
	}

	return &OAuthConfig{
		ProviderName: "Gmail",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"https://mail.google.com/",                    // Full IMAP access
			"https://www.googleapis.com/auth/calendar",    // CalDAV
			"https://www.googleapis.com/auth/contacts",    // CardDAV
		},
		Endpoint: google.Endpoint,
	}, nil
}
