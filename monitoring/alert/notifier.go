// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package alert provides multi-channel alerting with rate limiting for DarkPipe monitoring events.
package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Severity represents the urgency level of an alert.
type Severity string

const (
	SeverityWarn     Severity = "warn"
	SeverityCritical Severity = "critical"
)

// AlertType categorizes the source of an alert.
type AlertType string

const (
	AlertCertExpiry       AlertType = "cert_expiry"
	AlertQueueBackup      AlertType = "queue_backup"
	AlertDeliveryFailure  AlertType = "delivery_failure"
	AlertTunnelDown       AlertType = "tunnel_down"
)

// Alert represents a monitoring alert event.
type Alert struct {
	Type      AlertType         `json:"type"`
	Severity  Severity          `json:"severity"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details"`
}

// AlertChannel is the interface for alert notification backends.
type AlertChannel interface {
	Send(ctx context.Context, alert Alert) error
	Name() string
}

// EmailChannel sends alerts via email using sendmail.
type EmailChannel struct {
	adminEmail string // Recipient email address
}

// NewEmailChannel creates an email notification channel.
// adminEmail can be provided directly or via MONITOR_ALERT_EMAIL env var.
func NewEmailChannel(adminEmail string) *EmailChannel {
	if adminEmail == "" {
		adminEmail = os.Getenv("MONITOR_ALERT_EMAIL")
	}
	return &EmailChannel{adminEmail: adminEmail}
}

func (e *EmailChannel) Name() string {
	return "email"
}

func (e *EmailChannel) Send(ctx context.Context, alert Alert) error {
	if e.adminEmail == "" {
		return fmt.Errorf("admin email not configured")
	}

	// Format email message
	subject := fmt.Sprintf("[DarkPipe %s] %s", alert.Severity, alert.Title)

	body := fmt.Sprintf(`Alert: %s
Severity: %s
Time: %s

%s

Details:
`, alert.Title, alert.Severity, alert.Timestamp.Format(time.RFC3339), alert.Message)

	for k, v := range alert.Details {
		body += fmt.Sprintf("  %s: %s\n", k, v)
	}

	// Use sendmail command (assumes Postfix/sendmail is available on mail server)
	cmd := exec.CommandContext(ctx, "sendmail", "-t")
	emailContent := fmt.Sprintf(`To: %s
Subject: %s

%s`, e.adminEmail, subject, body)
	cmd.Stdin = strings.NewReader(emailContent)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sendmail failed: %w", err)
	}

	return nil
}

// WebhookChannel sends alerts via HTTP POST to a webhook URL.
type WebhookChannel struct {
	url string
}

// NewWebhookChannel creates a webhook notification channel.
// url can be provided directly or via MONITOR_WEBHOOK_URL env var.
func NewWebhookChannel(url string) *WebhookChannel {
	if url == "" {
		url = os.Getenv("MONITOR_WEBHOOK_URL")
	}
	return &WebhookChannel{url: url}
}

func (w *WebhookChannel) Name() string {
	return "webhook"
}

func (w *WebhookChannel) Send(ctx context.Context, alert Alert) error {
	if w.url == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	// Marshal alert to JSON
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// Use curl command for HTTP POST (reusing pattern from cloud-relay/relay/notify/webhook.go)
	cmd := exec.CommandContext(ctx, "curl",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", fmt.Sprintf("X-DarkPipe-Alert: %s", alert.Type),
		"-d", string(payload),
		"--max-time", "10",
		"-f", // Fail on HTTP errors
		w.url,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("webhook POST failed: %w", err)
	}

	return nil
}

// CLIFileChannel writes alerts to a file for CLI consumption.
type CLIFileChannel struct {
	filePath string
}

// NewCLIFileChannel creates a CLI file notification channel.
// filePath can be provided directly or via MONITOR_CLI_ALERT_PATH env var.
// Default: /data/monitoring/cli-alerts.json
func NewCLIFileChannel(filePath string) *CLIFileChannel {
	if filePath == "" {
		filePath = os.Getenv("MONITOR_CLI_ALERT_PATH")
	}
	if filePath == "" {
		filePath = "/data/monitoring/cli-alerts.json"
	}
	return &CLIFileChannel{filePath: filePath}
}

func (c *CLIFileChannel) Name() string {
	return "cli-file"
}

func (c *CLIFileChannel) Send(ctx context.Context, alert Alert) error {
	// Ensure directory exists
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create alert directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open alert file: %w", err)
	}
	defer file.Close()

	// Write alert as NDJSON (newline-delimited JSON)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(alert); err != nil {
		return fmt.Errorf("failed to write alert: %w", err)
	}

	return nil
}

// AlertNotifier dispatches alerts to multiple channels with rate limiting.
type AlertNotifier struct {
	channels    []AlertChannel
	rateLimiter *RateLimiter
}

// NewAlertNotifier creates a new multi-channel alert notifier.
func NewAlertNotifier(channels []AlertChannel, limiter *RateLimiter) *AlertNotifier {
	return &AlertNotifier{
		channels:    channels,
		rateLimiter: limiter,
	}
}

// Send dispatches an alert to all configured channels.
// Rate-limited alerts are suppressed (logged but not dispatched).
// Errors from individual channels are collected but do not stop other channels.
func (a *AlertNotifier) Send(ctx context.Context, alert Alert) error {
	// Check rate limiter
	if a.rateLimiter != nil && !a.rateLimiter.ShouldSend(string(alert.Type)) {
		// Alert suppressed due to rate limiting
		return nil
	}

	// Fan-out to all channels (collect errors like Phase 2 MultiNotifier pattern)
	var firstErr error
	for _, channel := range a.channels {
		if err := channel.Send(ctx, alert); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s channel failed: %w", channel.Name(), err)
			}
			// Continue to other channels even on error
		}
	}

	return firstErr
}
