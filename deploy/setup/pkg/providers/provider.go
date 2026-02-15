// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// Provider abstracts authentication and capabilities for email/calendar/contact providers
type Provider interface {
	// Name returns display name ("Gmail", "Outlook", etc.)
	Name() string

	// Slug returns CLI identifier ("gmail", "outlook", etc.)
	Slug() string

	// ConnectIMAP authenticates and returns IMAP client
	ConnectIMAP(ctx context.Context) (*imapclient.Client, error)

	// ConnectCalDAV returns CalDAV client or nil if unsupported
	ConnectCalDAV(ctx context.Context) (*caldav.Client, error)

	// ConnectCardDAV returns CardDAV client or nil if unsupported
	ConnectCardDAV(ctx context.Context) (*carddav.Client, error)

	// SupportsLabels returns true for Gmail-style labels vs folders
	SupportsLabels() bool

	// SupportsAPI returns true if provider has API (MailCow/Mailu)
	SupportsAPI() bool

	// SupportsCalDAV returns true if CalDAV is available
	SupportsCalDAV() bool

	// SupportsCardDAV returns true if CardDAV is available
	SupportsCardDAV() bool

	// GetFolderMapping returns provider-specific folder mappings
	GetFolderMapping() map[string]string

	// GetSkipFolders returns folders to skip
	GetSkipFolders() map[string]bool

	// IMAPEndpoint returns IMAP server:port
	IMAPEndpoint() string

	// CalDAVEndpoint returns CalDAV base URL
	CalDAVEndpoint() string

	// CardDAVEndpoint returns CardDAV base URL
	CardDAVEndpoint() string

	// WizardPrompts returns provider-specific prompts for wizard
	WizardPrompts() []WizardPrompt
}

// WizardPrompt represents a wizard prompt for provider configuration
type WizardPrompt struct {
	Type     string // "oauth", "input", "info"
	Label    string
	Field    string
	HelpText string
}

// Registry manages provider instances
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) {
	r.providers[p.Slug()] = p
}

// Get retrieves a provider by slug
func (r *Registry) Get(slug string) (Provider, error) {
	p, ok := r.providers[slug]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", slug)
	}
	return p, nil
}

// List returns all registered provider slugs
func (r *Registry) List() []string {
	slugs := make([]string, 0, len(r.providers))
	for slug := range r.providers {
		slugs = append(slugs, slug)
	}
	return slugs
}

// DefaultRegistry is the global provider registry
var DefaultRegistry = NewRegistry()
