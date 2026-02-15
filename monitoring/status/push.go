// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package status

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HealthchecksPinger sends push-based pings to external uptime monitoring services
// Implements the Dead Man's Switch pattern - service alerts if pings stop arriving
type HealthchecksPinger struct {
	checkURL string
	client   *http.Client
	interval time.Duration
}

// NewHealthchecksPinger creates a new pinger for push-based monitoring
// Default interval is 5 minutes
func NewHealthchecksPinger(checkURL string, interval time.Duration) *HealthchecksPinger {
	if interval == 0 {
		interval = 5 * time.Minute
	}

	return &HealthchecksPinger{
		checkURL: checkURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		interval: interval,
	}
}

// Ping sends a simple GET request to the check URL
func (p *HealthchecksPinger) Ping(ctx context.Context) error {
	if p.checkURL == "" {
		return nil // Disabled
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.checkURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ping failed with status %d", resp.StatusCode)
	}

	return nil
}

// PingWithStatus sends a POST request with the full system status as JSON
// Healthchecks.io and similar services support this for richer monitoring
func (p *HealthchecksPinger) PingWithStatus(ctx context.Context, status *SystemStatus) error {
	if p.checkURL == "" {
		return nil // Disabled
	}

	body, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("marshal status: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.checkURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ping failed with status %d", resp.StatusCode)
	}

	return nil
}

// Run starts the background ping loop
// Errors on individual pings are logged but don't stop the loop
// The Dead Man's Switch pattern means the external service alerts if pings stop
func (p *HealthchecksPinger) Run(ctx context.Context, getStatus func() (*SystemStatus, error)) {
	if p.checkURL == "" {
		return // Disabled
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Send initial ping immediately
	p.sendPing(ctx, getStatus)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.sendPing(ctx, getStatus)
		}
	}
}

func (p *HealthchecksPinger) sendPing(ctx context.Context, getStatus func() (*SystemStatus, error)) {
	// Try to get current status and send with details
	status, err := getStatus()
	if err == nil && status != nil {
		// Send detailed status if available
		if err := p.PingWithStatus(ctx, status); err != nil {
			// Fall back to simple ping on error
			_ = p.Ping(ctx)
		}
	} else {
		// Send simple ping if status unavailable
		_ = p.Ping(ctx)
	}
}
