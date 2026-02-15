// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mobileconfig

// MobileConfigProfile represents the top-level Apple Configuration Profile.
type MobileConfigProfile struct {
	PayloadContent      []interface{} `plist:"PayloadContent"`
	PayloadDescription  string        `plist:"PayloadDescription"`
	PayloadDisplayName  string        `plist:"PayloadDisplayName"`
	PayloadIdentifier   string        `plist:"PayloadIdentifier"`
	PayloadOrganization string        `plist:"PayloadOrganization"`
	PayloadRemovalDisallowed bool     `plist:"PayloadRemovalDisallowed"`
	PayloadType         string        `plist:"PayloadType"`
	PayloadUUID         string        `plist:"PayloadUUID"`
	PayloadVersion      int           `plist:"PayloadVersion"`
}

// EmailPayload represents IMAP and SMTP email configuration.
type EmailPayload struct {
	EmailAccountDescription string `plist:"EmailAccountDescription"`
	EmailAccountName        string `plist:"EmailAccountName"`
	EmailAccountType        string `plist:"EmailAccountType"` // "EmailTypeIMAP"
	EmailAddress            string `plist:"EmailAddress"`
	IncomingMailServerAuthentication string `plist:"IncomingMailServerAuthentication"` // "EmailAuthPassword"
	IncomingMailServerHostName       string `plist:"IncomingMailServerHostName"`
	IncomingMailServerPortNumber     int    `plist:"IncomingMailServerPortNumber"`
	IncomingMailServerUseSSL         bool   `plist:"IncomingMailServerUseSSL"`
	IncomingMailServerUsername       string `plist:"IncomingMailServerUsername"`
	IncomingPassword                 string `plist:"IncomingPassword"`
	OutgoingMailServerAuthentication string `plist:"OutgoingMailServerAuthentication"` // "EmailAuthPassword"
	OutgoingMailServerHostName       string `plist:"OutgoingMailServerHostName"`
	OutgoingMailServerPortNumber     int    `plist:"OutgoingMailServerPortNumber"`
	OutgoingMailServerUseSSL         bool   `plist:"OutgoingMailServerUseSSL"`
	OutgoingMailServerUsername       string `plist:"OutgoingMailServerUsername"`
	OutgoingPassword                 string `plist:"OutgoingPassword"`
	OutgoingPasswordSameAsIncomingPassword bool `plist:"OutgoingPasswordSameAsIncomingPassword"`
	PayloadDescription               string `plist:"PayloadDescription"`
	PayloadDisplayName               string `plist:"PayloadDisplayName"`
	PayloadIdentifier                string `plist:"PayloadIdentifier"`
	PayloadOrganization              string `plist:"PayloadOrganization"`
	PayloadType                      string `plist:"PayloadType"` // "com.apple.mail.managed"
	PayloadUUID                      string `plist:"PayloadUUID"`
	PayloadVersion                   int    `plist:"PayloadVersion"`
	PreventMove                      bool   `plist:"PreventMove"`
	PreventAppSheet                  bool   `plist:"PreventAppSheet"`
	SMIMEEnabled                     bool   `plist:"SMIMEEnabled"`
}

// CalDAVPayload represents CalDAV calendar configuration.
type CalDAVPayload struct {
	CalDAVAccountDescription string `plist:"CalDAVAccountDescription"`
	CalDAVHostName           string `plist:"CalDAVHostName"`
	CalDAVPort               int    `plist:"CalDAVPort"`
	CalDAVPrincipalURL       string `plist:"CalDAVPrincipalURL"`
	CalDAVUseSSL             bool   `plist:"CalDAVUseSSL"`
	CalDAVUsername           string `plist:"CalDAVUsername"`
	CalDAVPassword           string `plist:"CalDAVPassword"`
	PayloadDescription       string `plist:"PayloadDescription"`
	PayloadDisplayName       string `plist:"PayloadDisplayName"`
	PayloadIdentifier        string `plist:"PayloadIdentifier"`
	PayloadOrganization      string `plist:"PayloadOrganization"`
	PayloadType              string `plist:"PayloadType"` // "com.apple.caldav.account"
	PayloadUUID              string `plist:"PayloadUUID"`
	PayloadVersion           int    `plist:"PayloadVersion"`
}

// CardDAVPayload represents CardDAV contacts configuration.
type CardDAVPayload struct {
	CardDAVAccountDescription string `plist:"CardDAVAccountDescription"`
	CardDAVHostName           string `plist:"CardDAVHostName"`
	CardDAVPort               int    `plist:"CardDAVPort"`
	CardDAVPrincipalURL       string `plist:"CardDAVPrincipalURL"`
	CardDAVUseSSL             bool   `plist:"CardDAVUseSSL"`
	CardDAVUsername           string `plist:"CardDAVUsername"`
	CardDAVPassword           string `plist:"CardDAVPassword"`
	PayloadDescription        string `plist:"PayloadDescription"`
	PayloadDisplayName        string `plist:"PayloadDisplayName"`
	PayloadIdentifier         string `plist:"PayloadIdentifier"`
	PayloadOrganization       string `plist:"PayloadOrganization"`
	PayloadType               string `plist:"PayloadType"` // "com.apple.carddav.account"
	PayloadUUID               string `plist:"PayloadUUID"`
	PayloadVersion            int    `plist:"PayloadVersion"`
}
