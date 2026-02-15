// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package tls provides TLS monitoring and strict mode configuration for Postfix.
package tls

import (
	"bufio"
	"context"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/notify"
)

// TLSMonitor monitors Postfix logs for TLS connection quality and emits notifications.
type TLSMonitor struct {
	notifier  notify.Notifier
	logReader io.Reader
}

// Postfix log patterns for TLS detection
var (
	// Pattern for successful TLS connections
	tlsSuccessPattern = regexp.MustCompile(`Anonymous TLS connection established`)

	// Pattern for TLS required but not offered (strict mode violation)
	tlsRequiredPattern = regexp.MustCompile(`TLS is required, but was not offered`)

	// Pattern for certificate verification failures
	tlsUntrustedPattern = regexp.MustCompile(`untrusted issuer|certificate verification failed`)

	// Pattern for TLS handshake failures
	tlsHandshakeFailPattern = regexp.MustCompile(`Cannot start TLS|TLS handshake failed`)

	// Pattern to extract domain from log lines: to=<user@domain>
	domainExtractPattern = regexp.MustCompile(`to=<[^@]+@([^>]+)>`)

	// Alternative pattern for connection info: connect from domain[ip]
	connectFromPattern = regexp.MustCompile(`connect from ([^\[]+)\[`)
)

// NewTLSMonitor creates a new TLS monitor that reads from the provided log stream.
func NewTLSMonitor(logReader io.Reader, notifier notify.Notifier) *TLSMonitor {
	return &TLSMonitor{
		notifier:  notifier,
		logReader: logReader,
	}
}

// Monitor reads lines from the log stream and detects TLS-related events.
// It blocks until the context is cancelled or the log stream ends.
func (m *TLSMonitor) Monitor(ctx context.Context) error {
	scanner := bufio.NewScanner(m.logReader)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if !scanner.Scan() {
			// End of stream or error
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		line := scanner.Text()
		m.processLogLine(ctx, line)
	}
}

// processLogLine analyzes a single log line for TLS events.
func (m *TLSMonitor) processLogLine(ctx context.Context, line string) {
	// Successful TLS connection (info only, no notification)
	if tlsSuccessPattern.MatchString(line) {
		log.Printf("INFO: Successful TLS connection detected")
		return
	}

	// TLS required but not offered (strict mode violation)
	if tlsRequiredPattern.MatchString(line) {
		domain := extractDomain(line)
		event := notify.Event{
			Type:      "tls_failure",
			Domain:    domain,
			Message:   "TLS is required, but remote server did not offer TLS",
			Timestamp: time.Now(),
			Details: map[string]string{
				"log_line": line,
			},
		}
		if err := m.notifier.Send(ctx, event); err != nil {
			log.Printf("ERROR: Failed to send tls_failure notification: %v", err)
		}
		return
	}

	// Certificate verification failures
	if tlsUntrustedPattern.MatchString(line) {
		domain := extractDomain(line)
		event := notify.Event{
			Type:      "tls_warning",
			Domain:    domain,
			Message:   "TLS certificate verification failed (untrusted issuer)",
			Timestamp: time.Now(),
			Details: map[string]string{
				"log_line": line,
			},
		}
		if err := m.notifier.Send(ctx, event); err != nil {
			log.Printf("ERROR: Failed to send tls_warning notification: %v", err)
		}
		return
	}

	// TLS handshake failures
	if tlsHandshakeFailPattern.MatchString(line) {
		domain := extractDomain(line)
		event := notify.Event{
			Type:      "tls_warning",
			Domain:    domain,
			Message:   "TLS handshake failed",
			Timestamp: time.Now(),
			Details: map[string]string{
				"log_line": line,
			},
		}
		if err := m.notifier.Send(ctx, event); err != nil {
			log.Printf("ERROR: Failed to send tls_warning notification: %v", err)
		}
		return
	}
}

// extractDomain attempts to extract the domain from a Postfix log line.
func extractDomain(line string) string {
	// Try to extract from to=<user@domain>
	if matches := domainExtractPattern.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}

	// Try to extract from connect from domain[ip]
	if matches := connectFromPattern.FindStringSubmatch(line); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Fallback: return "unknown"
	return "unknown"
}
