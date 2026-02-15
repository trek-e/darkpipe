// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package providers

import (
	"context"
	"fmt"

	"github.com/emersion/go-sasl"
	"golang.org/x/oauth2"
)

// OAuthConfig holds OAuth2 device authorization grant flow configuration
type OAuthConfig struct {
	ProviderName  string
	ClientID      string
	ClientSecret  string
	Scopes        []string
	Endpoint      oauth2.Endpoint
	DeviceAuthURL string
}

// RunDeviceFlow executes OAuth2 device authorization grant (RFC 8628)
// The display callback receives verification URL and user code for display to user
func RunDeviceFlow(ctx context.Context, config *OAuthConfig, display func(verificationURL, userCode string)) (*oauth2.Token, error) {
	// Create oauth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
		Endpoint:     config.Endpoint,
	}

	// Step 1: Request device code
	response, err := oauthConfig.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("device auth request failed: %w", err)
	}

	// Step 2: Display verification URL and user code to user
	display(response.VerificationURI, response.UserCode)

	// Step 3: Poll for token (blocks until user completes auth or timeout)
	token, err := oauthConfig.DeviceAccessToken(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("device access token failed: %w", err)
	}

	return token, nil
}

// TokenSourceFromToken creates an auto-refreshing token source
func TokenSourceFromToken(config *OAuthConfig, token *oauth2.Token) oauth2.TokenSource {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
		Endpoint:     config.Endpoint,
	}

	return oauthConfig.TokenSource(context.Background(), token)
}

// XOAUTH2Token formats XOAUTH2 SASL string for IMAP authentication
// Format: user=<email>\x01auth=Bearer <token>\x01\x01
func XOAUTH2Token(email, accessToken string) string {
	return fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", email, accessToken)
}

// xoauth2Client implements sasl.Client for XOAUTH2 mechanism
type xoauth2Client struct {
	email       string
	accessToken string
}

// NewXOAUTH2Client creates a SASL client for XOAUTH2 authentication
func NewXOAUTH2Client(email, accessToken string) sasl.Client {
	return &xoauth2Client{
		email:       email,
		accessToken: accessToken,
	}
}

func (c *xoauth2Client) Start() (mech string, ir []byte, err error) {
	mech = "XOAUTH2"
	ir = []byte(XOAUTH2Token(c.email, c.accessToken))
	return
}

func (c *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	// XOAUTH2 is a one-step mechanism - no further responses needed
	return nil, fmt.Errorf("unexpected server challenge")
}
