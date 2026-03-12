// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logutil provides logging utilities for PII redaction.
package logutil

import "strings"

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

// RedactEmails applies RedactEmail to each element of a string slice.
func RedactEmails(addrs []string) []string {
	if addrs == nil {
		return nil
	}
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = RedactEmail(a)
	}
	return out
}
