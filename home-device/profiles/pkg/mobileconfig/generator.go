// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mobileconfig

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/micromdm/plist"
)

// ProfileConfig contains all configuration needed to generate a .mobileconfig profile.
type ProfileConfig struct {
	Domain       string // e.g., example.com
	MailHostname string // e.g., mail.example.com
	Email        string // e.g., user@example.com
	AppPassword  string // Generated app password
	CalDAVURL    string // Empty if no CalDAV (optional)
	CardDAVURL   string // Empty if no CardDAV (optional)
	CalDAVPort   int    // 5232 for Radicale, 443 for Stalwart
	CardDAVPort  int    // 5232 for Radicale, 443 for Stalwart
}

// ProfileGenerator generates Apple .mobileconfig profiles.
type ProfileGenerator struct{}

// GenerateProfile creates an Apple Configuration Profile with Email, CalDAV, and CardDAV payloads.
func (g *ProfileGenerator) GenerateProfile(cfg ProfileConfig) ([]byte, error) {
	profileUUID := uuid.New().String()

	payloads := []interface{}{}

	// Email payload (always included)
	emailPayload := EmailPayload{
		EmailAccountDescription: fmt.Sprintf("%s Email", cfg.Domain),
		EmailAccountName:        cfg.Email,
		EmailAccountType:        "EmailTypeIMAP",
		EmailAddress:            cfg.Email,
		IncomingMailServerAuthentication: "EmailAuthPassword",
		IncomingMailServerHostName:       cfg.MailHostname,
		IncomingMailServerPortNumber:     993,
		IncomingMailServerUseSSL:         true,
		IncomingMailServerUsername:       cfg.Email,
		IncomingPassword:                 cfg.AppPassword,
		OutgoingMailServerAuthentication: "EmailAuthPassword",
		OutgoingMailServerHostName:       cfg.MailHostname,
		OutgoingMailServerPortNumber:     587,
		OutgoingMailServerUseSSL:         false, // STARTTLS, not SSL
		OutgoingMailServerUsername:       cfg.Email,
		OutgoingPassword:                 cfg.AppPassword,
		OutgoingPasswordSameAsIncomingPassword: true,
		PayloadDescription:               "Email configuration",
		PayloadDisplayName:               "Email",
		PayloadIdentifier:                fmt.Sprintf("com.darkpipe.%s.email", cfg.Domain),
		PayloadOrganization:              "DarkPipe",
		PayloadType:                      "com.apple.mail.managed",
		PayloadUUID:                      uuid.New().String(),
		PayloadVersion:                   1,
		PreventMove:                      false,
		PreventAppSheet:                  false,
		SMIMEEnabled:                     false,
	}
	payloads = append(payloads, emailPayload)

	// CalDAV payload (optional)
	if cfg.CalDAVURL != "" {
		calDAVPayload := CalDAVPayload{
			CalDAVAccountDescription: fmt.Sprintf("%s Calendar", cfg.Domain),
			CalDAVHostName:           cfg.MailHostname,
			CalDAVPort:               cfg.CalDAVPort,
			CalDAVPrincipalURL:       cfg.CalDAVURL,
			CalDAVUseSSL:             cfg.CalDAVPort == 443,
			CalDAVUsername:           cfg.Email,
			CalDAVPassword:           cfg.AppPassword,
			PayloadDescription:       "CalDAV configuration",
			PayloadDisplayName:       "CalDAV",
			PayloadIdentifier:        fmt.Sprintf("com.darkpipe.%s.caldav", cfg.Domain),
			PayloadOrganization:      "DarkPipe",
			PayloadType:              "com.apple.caldav.account",
			PayloadUUID:              uuid.New().String(),
			PayloadVersion:           1,
		}
		payloads = append(payloads, calDAVPayload)
	}

	// CardDAV payload (optional)
	if cfg.CardDAVURL != "" {
		cardDAVPayload := CardDAVPayload{
			CardDAVAccountDescription: fmt.Sprintf("%s Contacts", cfg.Domain),
			CardDAVHostName:           cfg.MailHostname,
			CardDAVPort:               cfg.CardDAVPort,
			CardDAVPrincipalURL:       cfg.CardDAVURL,
			CardDAVUseSSL:             cfg.CardDAVPort == 443,
			CardDAVUsername:           cfg.Email,
			CardDAVPassword:           cfg.AppPassword,
			PayloadDescription:        "CardDAV configuration",
			PayloadDisplayName:        "CardDAV",
			PayloadIdentifier:         fmt.Sprintf("com.darkpipe.%s.carddav", cfg.Domain),
			PayloadOrganization:       "DarkPipe",
			PayloadType:               "com.apple.carddav.account",
			PayloadUUID:               uuid.New().String(),
			PayloadVersion:            1,
		}
		payloads = append(payloads, cardDAVPayload)
	}

	// Top-level profile
	profile := MobileConfigProfile{
		PayloadContent:      payloads,
		PayloadDescription:  fmt.Sprintf("Email and groupware configuration for %s", cfg.Domain),
		PayloadDisplayName:  fmt.Sprintf("%s Configuration", cfg.Domain),
		PayloadIdentifier:   fmt.Sprintf("com.darkpipe.%s", cfg.Domain),
		PayloadOrganization: "DarkPipe",
		PayloadRemovalDisallowed: false,
		PayloadType:         "Configuration",
		PayloadUUID:         profileUUID,
		PayloadVersion:      1,
	}

	// Serialize to plist XML
	data, err := plist.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal plist: %w", err)
	}

	return data, nil
}
