// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// RunQRCommand implements the CLI QR code generation command
func RunQRCommand(args []string) {
	fs := flag.NewFlagSet("qr", flag.ExitOnError)
	pngPath := fs.String("png", "", "Save QR code as PNG file instead of terminal display")
	profileServerURL := fs.String("server", os.Getenv("PROFILE_SERVER_URL"), "Profile server URL (default: from PROFILE_SERVER_URL env var)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s qr <email@domain> [--png <file>] [--server <url>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generates a QR code for device setup.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  email       Email address for device setup\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Display QR code in terminal\n")
		fmt.Fprintf(os.Stderr, "  %s qr user@example.com\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Save as PNG file\n")
		fmt.Fprintf(os.Stderr, "  %s qr user@example.com --png setup-qr.png\n", os.Args[0])
	}

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(1)
	}

	email := fs.Arg(0)

	// If profile server URL is provided, use HTTP API to create token
	// Otherwise, create token locally using in-memory store
	var token string
	var tokenExpiry time.Time

	if *profileServerURL != "" {
		// TODO: Implement HTTP API call to profile server for token creation
		// For now, fall back to standalone mode
		log.Println("HTTP API mode not yet implemented, using standalone mode")
		tokenStore := qrcode.NewMemoryTokenStore()
		tokenExpiry = time.Now().Add(15 * time.Minute)
		var err error
		token, err = tokenStore.Create(email, tokenExpiry)
		if err != nil {
			log.Fatalf("Failed to create token: %v", err)
		}
	} else {
		// Standalone mode: create token directly
		tokenStore := qrcode.NewMemoryTokenStore()
		tokenExpiry = time.Now().Add(15 * time.Minute)
		var err error
		token, err = tokenStore.Create(email, tokenExpiry)
		if err != nil {
			log.Fatalf("Failed to create token: %v", err)
		}
	}

	// Get hostname from environment or use default
	hostname := os.Getenv("MAIL_HOSTNAME")
	if hostname == "" {
		hostname = "mail.example.com"
		log.Printf("WARNING: MAIL_HOSTNAME not set, using default: %s", hostname)
	}

	// Generate URL
	profileURL := fmt.Sprintf("https://%s/profile/download?token=%s", hostname, token)

	// Generate QR code
	if *pngPath != "" {
		// Save as PNG file
		pngData, err := qrcode.GenerateQRCodePNG(profileURL, 256)
		if err != nil {
			log.Fatalf("Failed to generate QR code PNG: %v", err)
		}
		// Write to file
		if err := os.WriteFile(*pngPath, pngData, 0644); err != nil {
			log.Fatalf("Failed to write QR code file: %v", err)
		}
		fmt.Printf("QR code saved to: %s\n", *pngPath)
		fmt.Printf("URL: %s\n", profileURL)
		fmt.Printf("Token expires: %s (%s)\n", tokenExpiry.Format(time.RFC3339), time.Until(tokenExpiry).Round(time.Second))
	} else {
		// Display in terminal as ASCII art
		terminalQR, err := qrcode.GenerateQRCodeTerminal(profileURL)
		if err != nil {
			log.Fatalf("Failed to generate terminal QR code: %v", err)
		}

		fmt.Println()
		fmt.Println(terminalQR)
		fmt.Println()
		fmt.Printf("URL: %s\n", profileURL)
		fmt.Printf("Token expires: %s (%s)\n", tokenExpiry.Format(time.RFC3339), time.Until(tokenExpiry).Round(time.Second))
		fmt.Println()
		fmt.Println("Instructions:")
		fmt.Println("  1. Scan this QR code with your phone camera")
		fmt.Println("  2. Follow the prompts to install the mail profile")
		fmt.Println("  3. Your device will be configured automatically")
		fmt.Println()
		fmt.Println("Note: This token can only be used once and expires in 15 minutes")
	}
}

// If running as standalone CLI tool (not HTTP server), check for 'qr' subcommand
func init() {
	// This will be called before main() if imported as a library
	// The actual CLI dispatch should be done in main.go if we want to support both modes
}
