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

// iCloudProvider implements Provider for iCloud with app-specific passwords
type iCloudProvider struct {
	Email       string
	AppPassword string
}

func init() {
	// Register iCloud provider in global registry
	DefaultRegistry.Register(&iCloudProvider{})
}

func (p *iCloudProvider) Name() string {
	return "iCloud"
}

func (p *iCloudProvider) Slug() string {
	return "icloud"
}

func (p *iCloudProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to imap.mail.me.com:993
	conn, err := tls.Dial("tcp", "imap.mail.me.com:993", &tls.Config{
		ServerName: "imap.mail.me.com",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial iCloud IMAP: %w", err)
	}

	// Create IMAP client
	client := imapclient.New(conn, nil)

	// Login with email and app password
	if err := client.Login(p.Email, p.AppPassword).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return client, nil
}

func (p *iCloudProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// Create HTTP client with basic auth
	httpClient := &http.Client{
		Transport: &basicAuthTransport{
			Username: p.Email,
			Password: p.AppPassword,
		},
	}

	// Create CalDAV client
	client, err := caldav.NewClient(httpClient, "https://caldav.icloud.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to create CalDAV client: %w", err)
	}

	return client, nil
}

func (p *iCloudProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// Create HTTP client with basic auth
	httpClient := &http.Client{
		Transport: &basicAuthTransport{
			Username: p.Email,
			Password: p.AppPassword,
		},
	}

	// Create CardDAV client
	client, err := carddav.NewClient(httpClient, "https://contacts.icloud.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to create CardDAV client: %w", err)
	}

	return client, nil
}

func (p *iCloudProvider) SupportsLabels() bool {
	return false
}

func (p *iCloudProvider) SupportsAPI() bool {
	return false
}

func (p *iCloudProvider) SupportsCalDAV() bool {
	return true
}

func (p *iCloudProvider) SupportsCardDAV() bool {
	return true
}

func (p *iCloudProvider) GetFolderMapping() map[string]string {
	// iCloud uses standard folder names
	return map[string]string{}
}

func (p *iCloudProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{}
}

func (p *iCloudProvider) IMAPEndpoint() string {
	return "imap.mail.me.com:993"
}

func (p *iCloudProvider) CalDAVEndpoint() string {
	return "https://caldav.icloud.com/"
}

func (p *iCloudProvider) CardDAVEndpoint() string {
	return "https://contacts.icloud.com/"
}

func (p *iCloudProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:  "info",
			Label: "iCloud App-Specific Password Setup",
			HelpText: `iCloud requires an app-specific password for IMAP/CalDAV/CardDAV access.

Steps to create an app-specific password:
1. Go to https://appleid.apple.com/
2. Sign in with your Apple ID
3. Navigate to Security section
4. Click "Generate Password" under App-Specific Passwords
5. Enter a label (e.g., "DarkPipe Migration")
6. Copy the generated password (format: xxxx-xxxx-xxxx-xxxx)

Note: You must have two-factor authentication enabled on your Apple ID.`,
		},
		{
			Type:     "input",
			Label:    "iCloud Email Address",
			Field:    "email",
			HelpText: "Your iCloud email (e.g., user@icloud.com or user@me.com)",
		},
		{
			Type:     "input",
			Label:    "App-Specific Password",
			Field:    "app_password",
			HelpText: "The app-specific password you generated (format: xxxx-xxxx-xxxx-xxxx)",
		},
	}
}

// basicAuthTransport implements http.RoundTripper with basic auth
type basicAuthTransport struct {
	Username string
	Password string
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return http.DefaultTransport.RoundTrip(req)
}
