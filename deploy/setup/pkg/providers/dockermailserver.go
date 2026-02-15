// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// DockerMailServerProvider implements Provider for docker-mailserver
type DockerMailServerProvider struct {
	IMAPHost string
	Username string
	Password string
}

func init() {
	// Register docker-mailserver provider in global registry
	DefaultRegistry.Register(&DockerMailServerProvider{})
}

func (p *DockerMailServerProvider) Name() string {
	return "docker-mailserver"
}

func (p *DockerMailServerProvider) Slug() string {
	return "dockermailserver"
}

func (p *DockerMailServerProvider) ConnectIMAP(ctx context.Context) (*imapclient.Client, error) {
	// Dial TLS to IMAPHost:993
	conn, err := tls.Dial("tcp", p.IMAPHost+":993", &tls.Config{
		ServerName: p.IMAPHost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial docker-mailserver IMAP: %w", err)
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

func (p *DockerMailServerProvider) ConnectCalDAV(ctx context.Context) (*caldav.Client, error) {
	// docker-mailserver has no CalDAV
	return nil, fmt.Errorf("docker-mailserver does not support CalDAV - use file import")
}

func (p *DockerMailServerProvider) ConnectCardDAV(ctx context.Context) (*carddav.Client, error) {
	// docker-mailserver has no CardDAV
	return nil, fmt.Errorf("docker-mailserver does not support CardDAV - use file import")
}

func (p *DockerMailServerProvider) SupportsLabels() bool {
	return false
}

func (p *DockerMailServerProvider) SupportsAPI() bool {
	return false
}

func (p *DockerMailServerProvider) SupportsCalDAV() bool {
	return false
}

func (p *DockerMailServerProvider) SupportsCardDAV() bool {
	return false
}

func (p *DockerMailServerProvider) GetFolderMapping() map[string]string {
	// docker-mailserver uses standard IMAP folder names
	return map[string]string{}
}

func (p *DockerMailServerProvider) GetSkipFolders() map[string]bool {
	return map[string]bool{}
}

func (p *DockerMailServerProvider) IMAPEndpoint() string {
	return p.IMAPHost + ":993"
}

func (p *DockerMailServerProvider) CalDAVEndpoint() string {
	return ""
}

func (p *DockerMailServerProvider) CardDAVEndpoint() string {
	return ""
}

func (p *DockerMailServerProvider) WizardPrompts() []WizardPrompt {
	return []WizardPrompt{
		{
			Type:     "input",
			Label:    "IMAP Host",
			Field:    "imap_host",
			HelpText: "IMAP server hostname or IP address",
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
