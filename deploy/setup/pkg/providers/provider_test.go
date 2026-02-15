// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	// Register a test provider
	testProvider := &GmailProvider{}
	registry.Register(testProvider)

	// Test Get
	provider, err := registry.Get("gmail")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if provider.Name() != "Gmail" {
		t.Errorf("Provider.Name() = %q, want %q", provider.Name(), "Gmail")
	}

	// Test Get non-existent provider
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get should return error for non-existent provider")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&GmailProvider{})
	registry.Register(&OutlookProvider{})

	slugs := registry.List()

	if len(slugs) != 2 {
		t.Errorf("List returned %d slugs, want 2", len(slugs))
	}

	// Check that slugs contain expected values
	found := make(map[string]bool)
	for _, slug := range slugs {
		found[slug] = true
	}

	if !found["gmail"] {
		t.Error("List should contain 'gmail'")
	}

	if !found["outlook"] {
		t.Error("List should contain 'outlook'")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Default registry should have all providers registered via init()
	slugs := DefaultRegistry.List()

	expectedProviders := []string{
		"gmail",
		"outlook",
		"icloud",
		"mailcow",
		"mailu",
		"dockermailserver",
		"generic",
	}

	if len(slugs) != len(expectedProviders) {
		t.Errorf("DefaultRegistry has %d providers, want %d", len(slugs), len(expectedProviders))
	}

	for _, expected := range expectedProviders {
		provider, err := DefaultRegistry.Get(expected)
		if err != nil {
			t.Errorf("DefaultRegistry missing provider %q: %v", expected, err)
		}

		if provider.Slug() != expected {
			t.Errorf("Provider slug = %q, want %q", provider.Slug(), expected)
		}
	}
}

func TestGmailProvider_FolderMapping(t *testing.T) {
	provider := &GmailProvider{}

	mapping := provider.GetFolderMapping()

	tests := []struct {
		source string
		want   string
	}{
		{"[Gmail]/Sent Mail", "Sent"},
		{"[Gmail]/Drafts", "Drafts"},
		{"[Gmail]/Trash", "Trash"},
		{"[Gmail]/Spam", "Junk"},
		{"[Gmail]/All Mail", ""}, // Skip
		{"[Gmail]/Important", ""}, // Skip
		{"[Gmail]/Starred", ""}, // Skip
	}

	for _, tt := range tests {
		got := mapping[tt.source]
		if got != tt.want {
			t.Errorf("GetFolderMapping()[%q] = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestGmailProvider_SkipFolders(t *testing.T) {
	provider := &GmailProvider{}

	skip := provider.GetSkipFolders()

	expectedSkip := []string{
		"[Gmail]/All Mail",
		"[Gmail]/Important",
		"[Gmail]/Starred",
	}

	for _, folder := range expectedSkip {
		if !skip[folder] {
			t.Errorf("GetSkipFolders() should skip %q", folder)
		}
	}
}

func TestGmailProvider_Capabilities(t *testing.T) {
	provider := &GmailProvider{}

	if !provider.SupportsLabels() {
		t.Error("Gmail should support labels")
	}

	if provider.SupportsAPI() {
		t.Error("Gmail provider should not support API (uses IMAP)")
	}

	if !provider.SupportsCalDAV() {
		t.Error("Gmail should support CalDAV")
	}

	if !provider.SupportsCardDAV() {
		t.Error("Gmail should support CardDAV")
	}
}

func TestOutlookProvider_FolderMapping(t *testing.T) {
	provider := &OutlookProvider{}

	mapping := provider.GetFolderMapping()

	tests := []struct {
		source string
		want   string
	}{
		{"Deleted Items", "Trash"},
		{"Sent Items", "Sent"},
		{"Junk Email", "Junk"},
		{"Clutter", ""}, // Skip
	}

	for _, tt := range tests {
		got := mapping[tt.source]
		if got != tt.want {
			t.Errorf("GetFolderMapping()[%q] = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestOutlookProvider_Capabilities(t *testing.T) {
	provider := &OutlookProvider{}

	if provider.SupportsLabels() {
		t.Error("Outlook should not support labels")
	}

	if provider.SupportsCalDAV() {
		t.Error("Outlook should not support CalDAV")
	}

	if provider.SupportsCardDAV() {
		t.Error("Outlook should not support CardDAV")
	}
}

func TestICloudProvider_Capabilities(t *testing.T) {
	provider := &iCloudProvider{}

	if provider.SupportsLabels() {
		t.Error("iCloud should not support labels")
	}

	if !provider.SupportsCalDAV() {
		t.Error("iCloud should support CalDAV")
	}

	if !provider.SupportsCardDAV() {
		t.Error("iCloud should support CardDAV")
	}
}

func TestMailCowProvider_Capabilities(t *testing.T) {
	provider := &MailCowProvider{}

	if !provider.SupportsAPI() {
		t.Error("MailCow should support API")
	}

	if provider.SupportsCalDAV() {
		t.Error("MailCow should not support CalDAV")
	}

	if provider.SupportsCardDAV() {
		t.Error("MailCow should not support CardDAV")
	}
}

func TestMailuProvider_Capabilities(t *testing.T) {
	provider := &MailuProvider{}

	if !provider.SupportsAPI() {
		t.Error("Mailu should support API")
	}

	if provider.SupportsCalDAV() {
		t.Error("Mailu should not support CalDAV")
	}
}

func TestDockerMailServerProvider_Capabilities(t *testing.T) {
	provider := &DockerMailServerProvider{}

	if provider.SupportsAPI() {
		t.Error("docker-mailserver should not support API")
	}

	if provider.SupportsCalDAV() {
		t.Error("docker-mailserver should not support CalDAV")
	}
}

func TestGenericProvider_Capabilities(t *testing.T) {
	// Without CalDAV/CardDAV URLs
	provider := &GenericProvider{
		CalDAVURL:  "",
		CardDAVURL: "",
	}

	if provider.SupportsCalDAV() {
		t.Error("Generic provider without CalDAV URL should not support CalDAV")
	}

	if provider.SupportsCardDAV() {
		t.Error("Generic provider without CardDAV URL should not support CardDAV")
	}

	// With CalDAV/CardDAV URLs
	providerWithDAV := &GenericProvider{
		CalDAVURL:  "https://cal.example.com/caldav",
		CardDAVURL: "https://contacts.example.com/carddav",
	}

	if !providerWithDAV.SupportsCalDAV() {
		t.Error("Generic provider with CalDAV URL should support CalDAV")
	}

	if !providerWithDAV.SupportsCardDAV() {
		t.Error("Generic provider with CardDAV URL should support CardDAV")
	}
}

func TestXOAUTH2Token(t *testing.T) {
	email := "user@example.com"
	token := "ya29.a0AfH6SMBx..."

	result := XOAUTH2Token(email, token)

	expected := "user=user@example.com\x01auth=Bearer ya29.a0AfH6SMBx...\x01\x01"

	if result != expected {
		t.Errorf("XOAUTH2Token() = %q, want %q", result, expected)
	}
}

func TestMailCowAPI_GetMailboxes(t *testing.T) {
	// Create test server with canned response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API key header
		if r.Header.Get("X-API-Key") != "test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return canned mailboxes
		mailboxes := []MailCowMailbox{
			{Username: "user1@example.com", Name: "User One", QuotaUsed: 1024, Messages: 10, Active: true},
			{Username: "user2@example.com", Name: "User Two", QuotaUsed: 2048, Messages: 20, Active: true},
		}

		json.NewEncoder(w).Encode(mailboxes)
	}))
	defer server.Close()

	provider := &MailCowProvider{
		APIURL: server.URL,
		APIKey: "test-key",
	}

	mailboxes, err := provider.GetMailboxes(context.Background())
	if err != nil {
		t.Fatalf("GetMailboxes failed: %v", err)
	}

	if len(mailboxes) != 2 {
		t.Errorf("GetMailboxes returned %d mailboxes, want 2", len(mailboxes))
	}

	if mailboxes[0].Username != "user1@example.com" {
		t.Errorf("mailboxes[0].Username = %q, want %q", mailboxes[0].Username, "user1@example.com")
	}
}

func TestMailCowAPI_GetAliases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		aliases := []MailCowAlias{
			{Address: "alias@example.com", GoTo: "user@example.com", Active: true},
		}

		json.NewEncoder(w).Encode(aliases)
	}))
	defer server.Close()

	provider := &MailCowProvider{
		APIURL: server.URL,
		APIKey: "test-key",
	}

	aliases, err := provider.GetAliases(context.Background())
	if err != nil {
		t.Fatalf("GetAliases failed: %v", err)
	}

	if len(aliases) != 1 {
		t.Errorf("GetAliases returned %d aliases, want 1", len(aliases))
	}

	if aliases[0].Address != "alias@example.com" {
		t.Errorf("aliases[0].Address = %q, want %q", aliases[0].Address, "alias@example.com")
	}
}

func TestMailuAPI_GetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Bearer token
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		users := []MailuUser{
			{Email: "user@example.com", Name: "User", Enabled: true},
		}

		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	provider := &MailuProvider{
		APIURL: server.URL,
		APIKey: "test-token",
	}

	users, err := provider.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("GetUsers returned %d users, want 1", len(users))
	}

	if users[0].Email != "user@example.com" {
		t.Errorf("users[0].Email = %q, want %q", users[0].Email, "user@example.com")
	}
}

func TestGenericProvider_IMAPEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		provider GenericProvider
		want     string
	}{
		{
			name:     "Default TLS port",
			provider: GenericProvider{IMAPHost: "imap.example.com", UseTLS: true, IMAPPort: 0},
			want:     "imap.example.com:993",
		},
		{
			name:     "Default STARTTLS port",
			provider: GenericProvider{IMAPHost: "imap.example.com", UseTLS: false, IMAPPort: 0},
			want:     "imap.example.com:143",
		},
		{
			name:     "Custom port",
			provider: GenericProvider{IMAPHost: "imap.example.com", IMAPPort: 9993},
			want:     "imap.example.com:9993",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.provider.IMAPEndpoint()
			if got != tt.want {
				t.Errorf("IMAPEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWizardPrompts(t *testing.T) {
	providers := []Provider{
		&GmailProvider{},
		&OutlookProvider{},
		&iCloudProvider{},
		&MailCowProvider{},
		&MailuProvider{},
		&DockerMailServerProvider{},
		&GenericProvider{},
	}

	for _, p := range providers {
		t.Run(p.Name(), func(t *testing.T) {
			prompts := p.WizardPrompts()
			if len(prompts) == 0 {
				t.Errorf("%s has no wizard prompts", p.Name())
			}

			// Check that each prompt has required fields
			for i, prompt := range prompts {
				if prompt.Type == "" {
					t.Errorf("Prompt %d has empty Type", i)
				}

				if prompt.Label == "" {
					t.Errorf("Prompt %d has empty Label", i)
				}

				// Type should be one of: oauth, input, info
				validTypes := map[string]bool{"oauth": true, "input": true, "info": true}
				if !validTypes[prompt.Type] {
					t.Errorf("Prompt %d has invalid Type %q", i, prompt.Type)
				}
			}
		})
	}
}
