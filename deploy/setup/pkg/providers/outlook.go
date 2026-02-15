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
	"golang.org/x/oauth2/microsoft"
)

// OutlookProvider implements Provider for Outlook/Microsoft 365 with OAuth2
type OutlookProvider struct {
	Email    string
	Token    *oauth2.Token
	ClientID string
}

func init() {
	// Register Outlook provider in global registry
	DefaultRegistry.Register(&OutlookProvider{})
}

func (p *OutlookProvider) Name() string {
	return "Outlook"
}

func (p *OutlookProvider) Slug() string {
	return "outlook"
}

func (p *OutlookProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to outlook.office365.com:993
	conn, err := tls.Dial("tcp", "outlook.office365.com:993", &tls.Config{
		ServerName: "outlook.office365.com",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial Outlook IMAP: %w", err)
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

func (p *OutlookProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// Outlook CalDAV is not widely available
	// Users should export .ics from Outlook web
	return nil, fmt.Errorf("Outlook CalDAV not supported - export .ics files from Outlook web")
}

func (p *OutlookProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// Outlook CardDAV is not widely available
	// Users should export .vcf from Outlook web
	return nil, fmt.Errorf("Outlook CardDAV not supported - export .vcf files from Outlook web")
}

func (p *OutlookProvider) SupportsLabels() bool {
	return false
}

func (p *OutlookProvider) SupportsAPI() bool {
	return false
}

func (p *OutlookProvider) SupportsCalDAV() bool {
	return false
}

func (p *OutlookProvider) SupportsCardDAV() bool {
	return false
}

func (p *OutlookProvider) GetFolderMapping() map[string]string {
	// Outlook-specific folder mappings
	return map[string]string{
		"Deleted Items": "Trash",
		"Sent Items":    "Sent",
		"Junk Email":    "Junk",
		"Clutter":       "", // Skip
	}
}

func (p *OutlookProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{
		"Clutter": true,
	}
}

func (p *OutlookProvider) IMAPEndpoint() string {
	return "outlook.office365.com:993"
}

func (p *OutlookProvider) CalDAVEndpoint() string {
	return ""
}

func (p *OutlookProvider) CardDAVEndpoint() string {
	return ""
}

func (p *OutlookProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "oauth",
			Label:    "Outlook OAuth2 Authentication",
			Field:    "oauth",
			HelpText: "You will need to create an app registration in Azure AD. Set OUTLOOK_CLIENT_ID environment variable. Note: Microsoft requires OAuth2 for IMAP after April 2026.",
		},
		{
			Type:     "input",
			Label:    "Outlook Email Address",
			Field:    "email",
			HelpText: "Your full Outlook address (e.g., user@outlook.com or user@organization.com)",
		},
	}
}

// GetOutlookOAuthConfig returns OAuth2 config for Outlook
// Requires OUTLOOK_CLIENT_ID environment variable
func GetOutlookOAuthConfig() (*OAuthConfig, error) {
	clientID := os.Getenv("OUTLOOK_CLIENT_ID")

	if clientID == "" {
		return nil, fmt.Errorf("OUTLOOK_CLIENT_ID environment variable required")
	}

	return &OAuthConfig{
		ProviderName: "Outlook",
		ClientID:     clientID,
		ClientSecret: "", // Not required for device flow with public client
		Scopes: []string{
			"https://outlook.office.com/IMAP.AccessAsUser.All",
			"offline_access",
		},
		Endpoint:      microsoft.AzureADEndpoint("common"),
		DeviceAuthURL: "https://login.microsoftonline.com/common/oauth2/v2.0/devicecode",
	}, nil
}
