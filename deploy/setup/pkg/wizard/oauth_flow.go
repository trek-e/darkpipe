// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package wizard

import (
	"context"
	"fmt"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/pterm/pterm"
	"golang.org/x/oauth2"
)

// RunOAuthDeviceFlow executes OAuth2 device flow with rich UI feedback
func RunOAuthDeviceFlow(ctx context.Context, config *providers.OAuthConfig) (*oauth2.Token, error) {
	pterm.Info.Printf("Initiating OAuth2 device flow for %s...\n", config.ProviderName)

	// Run device flow with custom display callback
	token, err := providers.RunDeviceFlow(ctx, config, func(verificationURL, userCode string) {
		// Display verification URL in a highlighted box
		fmt.Println()
		pterm.DefaultBox.
			WithTitle("Authorization Required").
			WithTitleTopCenter().
			WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
			Println(
				fmt.Sprintf("1. Visit: %s\n\n2. Enter code: %s",
					verificationURL, userCode),
			)
		fmt.Println()

		// Display user code in large, bold text for easy reading
		pterm.DefaultBigText.
			WithLetters(pterm.NewLettersFromString(userCode)).
			Render()

		// Show spinner while waiting for authorization
		spinner, _ := pterm.DefaultSpinner.
			WithText("Waiting for authorization... (complete in your browser)").
			Start()

		// Store spinner in context for cleanup (called by parent)
		// Note: RunDeviceFlow blocks until authorization completes
		defer spinner.Stop()
	})

	if err != nil {
		pterm.Error.Printf("OAuth2 authorization failed: %v\n", err)
		return nil, err
	}

	// Success feedback
	pterm.Success.Println("Authorization successful!")
	fmt.Println()

	return token, nil
}
