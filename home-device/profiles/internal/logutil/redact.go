// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logutil provides logging utilities for PII redaction.
package logutil

import (
	"net/url"
	"strings"
)

// RedactEmail masks the local-part of an email address, preserving the domain.
// Examples:
//
//	"sender@example.com" → "s***r@example.com"
//	"ab@example.com"     → "a*@example.com"
//	"a@example.com"      → "*@example.com"
//	""                   → ""
//	"no-at-sign"         → "no-at-sign"
func RedactEmail(addr string) string {
	if addr == "" {
		return ""
	}

	at := strings.LastIndex(addr, "@")
	if at < 0 {
		return addr
	}

	local := addr[:at]
	domain := addr[at:] // includes "@"

	switch len(local) {
	case 0:
		return domain
	case 1:
		return "*" + domain
	case 2:
		return string(local[0]) + "*" + domain
	default:
		return string(local[0]) + "***" + string(local[len(local)-1]) + domain
	}
}

// sensitiveKeys lists query parameter keys whose values should be redacted.
var sensitiveKeys = map[string]bool{
	"emailaddress": true,
	"email":        true,
	"token":        true,
}

// RedactQueryParams parses a raw query string and replaces values for
// sensitive keys (emailaddress, email, token) with "[REDACTED]".
// Non-sensitive parameters are preserved as-is.
func RedactQueryParams(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	params, err := url.ParseQuery(rawQuery)
	if err != nil {
		// If unparseable, redact the entire thing to be safe.
		return "[REDACTED]"
	}

	for key := range params {
		if sensitiveKeys[strings.ToLower(key)] {
			params.Set(key, "[REDACTED]")
		}
	}

	return params.Encode()
}
