// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package delivery provides Postfix mail delivery status tracking via log parsing.
package delivery

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LogEntry represents a parsed Postfix mail delivery log line.
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	QueueID      string    `json:"queue_id"`
	From         string    `json:"from,omitempty"`
	To           string    `json:"to"`
	Status       string    `json:"status"` // delivered, deferred, bounced, expired
	Relay        string    `json:"relay,omitempty"`
	Delay        string    `json:"delay,omitempty"`
	DSN          string    `json:"dsn,omitempty"`
	StatusDetail string    `json:"status_detail,omitempty"`
}

// Parser parses Postfix mail.log lines for delivery status.
type Parser struct {
	// Regex patterns for extracting delivery information
	statusPattern *regexp.Regexp
	toPattern     *regexp.Regexp
	fromPattern   *regexp.Regexp
	relayPattern  *regexp.Regexp
	delayPattern  *regexp.Regexp
	dsnPattern    *regexp.Regexp
	detailPattern *regexp.Regexp
}

// NewParser creates a new Postfix log parser.
func NewParser() *Parser {
	return &Parser{
		// Match status=sent, status=deferred, status=bounced, status=expired
		statusPattern: regexp.MustCompile(`status=(\w+)`),
		// Match to=<address> or to=<address>,
		toPattern: regexp.MustCompile(`to=<([^>]+)>`),
		// Match from=<address>
		fromPattern: regexp.MustCompile(`from=<([^>]*)>`),
		// Match relay=host[ip]:port or relay=none
		relayPattern: regexp.MustCompile(`relay=([^,\s]+)`),
		// Match delay=N.N
		delayPattern: regexp.MustCompile(`delay=([\d.]+)`),
		// Match dsn=X.Y.Z
		dsnPattern: regexp.MustCompile(`dsn=([\d.]+)`),
		// Extract detail from status=X (detail) - everything inside parens
		detailPattern: regexp.MustCompile(`status=\w+\s+\(([^)]+)\)`),
	}
}

// ParseLine parses a single Postfix log line.
// Returns nil for non-delivery log lines.
func (p *Parser) ParseLine(line string) (*LogEntry, error) {
	// Check if this is a delivery status line (contains status=)
	statusMatch := p.statusPattern.FindStringSubmatch(line)
	if statusMatch == nil {
		return nil, nil // Not a delivery line
	}

	status := statusMatch[1]

	// Only process delivery-related statuses
	validStatuses := map[string]string{
		"sent":     "delivered",
		"deferred": "deferred",
		"bounced":  "bounced",
		"expired":  "expired",
	}

	normalizedStatus, ok := validStatuses[status]
	if !ok {
		return nil, nil // Not a delivery status we track
	}

	// Extract timestamp (Postfix syslog format: "Feb 14 10:23:45")
	timestamp, queueID, err := parseTimestampAndQueueID(line)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp/queue_id: %w", err)
	}

	entry := &LogEntry{
		Timestamp: timestamp,
		QueueID:   queueID,
		Status:    normalizedStatus,
	}

	// Extract recipient (required)
	if toMatch := p.toPattern.FindStringSubmatch(line); toMatch != nil {
		entry.To = toMatch[1]
	} else {
		// No recipient found - not a valid delivery line
		return nil, nil
	}

	// Extract optional fields
	if fromMatch := p.fromPattern.FindStringSubmatch(line); fromMatch != nil {
		entry.From = fromMatch[1]
	}

	if relayMatch := p.relayPattern.FindStringSubmatch(line); relayMatch != nil {
		entry.Relay = relayMatch[1]
	}

	if delayMatch := p.delayPattern.FindStringSubmatch(line); delayMatch != nil {
		entry.Delay = delayMatch[1]
	}

	if dsnMatch := p.dsnPattern.FindStringSubmatch(line); dsnMatch != nil {
		entry.DSN = dsnMatch[1]
	}

	if detailMatch := p.detailPattern.FindStringSubmatch(line); detailMatch != nil {
		entry.StatusDetail = detailMatch[1]
	}

	return entry, nil
}

// parseTimestampAndQueueID extracts timestamp and queue ID from a Postfix syslog line.
// Example: "Feb 14 10:23:45 mail postfix/smtp[1234]: ABCDEF1234: to=..."
func parseTimestampAndQueueID(line string) (time.Time, string, error) {
	// Postfix log format: "MMM DD HH:MM:SS hostname postfix/service[pid]: QUEUEID: ..."
	parts := strings.Fields(line)
	if len(parts) < 6 {
		return time.Time{}, "", fmt.Errorf("line too short: %d fields", len(parts))
	}

	// Parse timestamp (first 3 fields: month, day, time)
	// Note: syslog doesn't include year, so we use current year
	timestampStr := fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
	currentYear := time.Now().Year()
	timestampWithYear := fmt.Sprintf("%s %d", timestampStr, currentYear)

	timestamp, err := time.Parse("Jan 2 15:04:05 2006", timestampWithYear)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("parse timestamp '%s': %w", timestampWithYear, err)
	}

	// Extract queue ID (after "postfix/service[pid]:", before next ":")
	// Queue IDs are alphanumeric strings (typically 10-15 chars)
	// Example: "ABCDEF1234:" in "postfix/smtp[1234]: ABCDEF1234: to=..."
	var queueID string
	for i := 4; i < len(parts); i++ {
		if strings.HasSuffix(parts[i], ":") {
			candidate := strings.TrimSuffix(parts[i], ":")
			// Queue IDs are alphanumeric only (no slashes, brackets, etc.)
			if isAlphanumeric(candidate) && len(candidate) >= 8 && len(candidate) <= 20 {
				queueID = candidate
				break
			}
		}
	}

	if queueID == "" {
		return time.Time{}, "", fmt.Errorf("queue ID not found")
	}

	return timestamp, queueID, nil
}

// isAlphanumeric checks if a string contains only letters and numbers.
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return len(s) > 0
}
