// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrationsource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/pterm/pterm"
	"golang.org/x/oauth2"
)

type Capability string

const (
	CapabilityLabels  Capability = "labels"
	CapabilityCalDAV  Capability = "caldav"
	CapabilityCardDAV Capability = "carddav"
	CapabilityAPI     Capability = "api"
)

type SourceAdapters struct {
	IMAP    *imapclient.Client
	CalDAV  *caldav.Client
	CardDAV *carddav.Client
}

type SourceMetadata struct {
	ProviderSlug string
	ProviderName string
	Capabilities map[Capability]bool
	Endpoints    map[string]string
	Counts       map[string]int
}

type Module interface {
	Authenticate(ctx context.Context, provider providers.Provider) error
	OpenAdapters(ctx context.Context, provider providers.Provider) (*SourceAdapters, error)
	Capabilities(provider providers.Provider) map[Capability]bool
	DiscoverMetadata(ctx context.Context, provider providers.Provider) (*SourceMetadata, error)
}

// Typed interface errors.
type ErrAuthFailed struct{ Cause error }

func (e ErrAuthFailed) Error() string { return fmt.Sprintf("authentication failed: %v", e.Cause) }

type ErrMissingScope struct{ Scope string }

func (e ErrMissingScope) Error() string { return fmt.Sprintf("missing required scope: %s", e.Scope) }

type ErrRateLimited struct{ RetryAfterSeconds int }

func (e ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limited; retry after %d seconds", e.RetryAfterSeconds)
}

type ErrUnsupportedCapability struct{ Capability Capability }

func (e ErrUnsupportedCapability) Error() string {
	return fmt.Sprintf("unsupported capability: %s", e.Capability)
}

type ErrTemporaryNetwork struct{ Cause error }

func (e ErrTemporaryNetwork) Error() string { return fmt.Sprintf("temporary network error: %v", e.Cause) }

type DefaultModule struct{}

func New() Module { return &DefaultModule{} }

func (m *DefaultModule) Authenticate(ctx context.Context, provider providers.Provider) error {
	prompts := provider.WizardPrompts()
	for _, prompt := range prompts {
		switch prompt.Type {
		case "oauth":
			oauthCfg, err := oauthConfigFor(provider)
			if err != nil {
				return classifyError(err)
			}
			token, err := runOAuthDeviceFlow(ctx, oauthCfg)
			if err != nil {
				return classifyError(err)
			}
			if err := setProviderToken(provider, token); err != nil {
				return classifyError(err)
			}
		case "input":
			var value string
			inputPrompt := &survey.Input{Message: prompt.Label, Help: prompt.HelpText}
			if strings.Contains(strings.ToLower(prompt.Field), "password") {
				if err := survey.AskOne(&survey.Password{Message: prompt.Label, Help: prompt.HelpText}, &value); err != nil {
					return classifyError(err)
				}
			} else if err := survey.AskOne(inputPrompt, &value); err != nil {
				return classifyError(err)
			}
			if err := setProviderField(provider, prompt.Field, value); err != nil {
				return classifyError(err)
			}
		case "info":
			pterm.Info.Println(prompt.Label)
			if prompt.HelpText != "" {
				fmt.Println(prompt.HelpText)
				fmt.Println()
			}
		}
	}
	return nil
}

func (m *DefaultModule) OpenAdapters(ctx context.Context, provider providers.Provider) (*SourceAdapters, error) {
	imapClient, err := provider.ConnectIMAP(ctx)
	if err != nil {
		return nil, classifyError(fmt.Errorf("IMAP connection failed: %w", err))
	}

	adapters := &SourceAdapters{IMAP: imapClient}
	if provider.SupportsCalDAV() {
		if c, err := provider.ConnectCalDAV(ctx); err == nil {
			adapters.CalDAV = c
		} else {
			pterm.Warning.Printf("CalDAV connection failed (skipping calendars): %v\n", err)
		}
	}
	if provider.SupportsCardDAV() {
		if c, err := provider.ConnectCardDAV(ctx); err == nil {
			adapters.CardDAV = c
		} else {
			pterm.Warning.Printf("CardDAV connection failed (skipping contacts): %v\n", err)
		}
	}

	return adapters, nil
}

func (m *DefaultModule) Capabilities(provider providers.Provider) map[Capability]bool {
	return map[Capability]bool{
		CapabilityLabels:  provider.SupportsLabels(),
		CapabilityCalDAV:  provider.SupportsCalDAV(),
		CapabilityCardDAV: provider.SupportsCardDAV(),
		CapabilityAPI:     provider.SupportsAPI(),
	}
}

func (m *DefaultModule) DiscoverMetadata(ctx context.Context, provider providers.Provider) (*SourceMetadata, error) {
	meta := &SourceMetadata{
		ProviderSlug: provider.Slug(),
		ProviderName: provider.Name(),
		Capabilities: m.Capabilities(provider),
		Endpoints: map[string]string{
			"imap":    provider.IMAPEndpoint(),
			"caldav":  provider.CalDAVEndpoint(),
			"carddav": provider.CardDAVEndpoint(),
		},
		Counts: map[string]int{},
	}

	switch p := provider.(type) {
	case *providers.MailCowProvider:
		if mailboxes, err := p.GetMailboxes(ctx); err == nil {
			meta.Counts["mailboxes"] = len(mailboxes)
		}
		if aliases, err := p.GetAliases(ctx); err == nil {
			meta.Counts["aliases"] = len(aliases)
		}
		if domains, err := p.GetDomains(ctx); err == nil {
			meta.Counts["domains"] = len(domains)
		}
	case *providers.MailuProvider:
		if users, err := p.GetUsers(ctx); err == nil {
			meta.Counts["users"] = len(users)
		}
		if aliases, err := p.GetAliases(ctx); err == nil {
			meta.Counts["aliases"] = len(aliases)
		}
		if domains, err := p.GetDomains(ctx); err == nil {
			meta.Counts["domains"] = len(domains)
		}
	}

	return meta, nil
}

func oauthConfigFor(provider providers.Provider) (*providers.OAuthConfig, error) {
	switch provider.Slug() {
	case "gmail":
		return providers.GetGmailOAuthConfig()
	case "outlook":
		return providers.GetOutlookOAuthConfig()
	default:
		return nil, ErrUnsupportedCapability{Capability: CapabilityAPI}
	}
}

func runOAuthDeviceFlow(ctx context.Context, config *providers.OAuthConfig) (*oauth2.Token, error) {
	pterm.Info.Printf("Initiating OAuth2 device flow for %s...\n", config.ProviderName)
	token, err := providers.RunDeviceFlow(ctx, config, func(verificationURL, userCode string) {
		fmt.Println()
		pterm.DefaultBox.WithTitle("OAuth2 Authorization Required").WithTitleTopCenter().Println(
			fmt.Sprintf("1. Open this URL in your browser:\n   %s\n\n2. Enter this code:\n   %s\n", verificationURL, userCode),
		)
	})
	if err != nil {
		return nil, err
	}
	pterm.Success.Printf("✓ %s authentication successful\n", config.ProviderName)
	return token, nil
}

func setProviderToken(provider providers.Provider, token *oauth2.Token) error {
	switch p := provider.(type) {
	case *providers.GmailProvider:
		p.Token = token
		return nil
	case *providers.OutlookProvider:
		p.Token = token
		return nil
	default:
		return nil
	}
}

func setProviderField(provider providers.Provider, field, value string) error {
	switch p := provider.(type) {
	case *providers.GmailProvider:
		switch field {
		case "email":
			p.Email = value
		}
	case *providers.OutlookProvider:
		switch field {
		case "email":
			p.Email = value
		}
	case *providers.GenericProvider:
		switch field {
		case "imap_host":
			p.IMAPHost = value
		case "imap_port":
			if value != "" {
				n, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				p.IMAPPort = n
			}
		case "use_tls":
			p.UseTLS = strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
		case "username":
			p.Username = value
		case "password":
			p.Password = value
		case "caldav_url":
			p.CalDAVURL = value
		case "carddav_url":
			p.CardDAVURL = value
		}
	case *providers.MailuProvider:
		switch field {
		case "api_url":
			p.APIURL = value
		case "api_key":
			p.APIKey = value
		case "imap_host":
			p.IMAPHost = value
		case "username":
			p.Username = value
		case "password":
			p.Password = value
		}
	case *providers.MailCowProvider:
		switch field {
		case "api_url":
			p.APIURL = value
		case "api_key":
			p.APIKey = value
		case "imap_host":
			p.IMAPHost = value
		case "username":
			p.Username = value
		case "password":
			p.Password = value
		}
	case *providers.DockerMailServerProvider:
		switch field {
		case "imap_host":
			p.IMAPHost = value
		case "username":
			p.Username = value
		case "password":
			p.Password = value
		}
	}
	return nil
}

func classifyError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "scope") {
		return ErrMissingScope{Scope: "unknown"}
	}
	if strings.Contains(msg, "rate") || strings.Contains(msg, "too many") {
		return ErrRateLimited{RetryAfterSeconds: 60}
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary") || strings.Contains(msg, "connection") || strings.Contains(msg, "dial") {
		return ErrTemporaryNetwork{Cause: err}
	}
	if strings.Contains(msg, "oauth") || strings.Contains(msg, "auth") || strings.Contains(msg, "login failed") {
		return ErrAuthFailed{Cause: err}
	}
	return err
}
