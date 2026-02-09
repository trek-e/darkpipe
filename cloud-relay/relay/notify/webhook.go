// Package notify provides a notification system for TLS events and warnings.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebhookNotifier sends notification events to a webhook endpoint via HTTP POST.
type WebhookNotifier struct {
	url        string
	httpClient *http.Client
	// Rate limiting: track last notification time per domain
	lastNotified sync.Map // domain -> time.Time
	dedupeWindow time.Duration
}

// NewWebhookNotifier creates a new webhook notifier with the specified URL.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		url: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		dedupeWindow: 1 * time.Hour, // Deduplicate notifications for same domain within 1 hour
	}
}

// Send sends the event as JSON to the webhook URL.
// Rate limits notifications to prevent spam from domains that consistently fail TLS.
func (w *WebhookNotifier) Send(ctx context.Context, event Event) error {
	// Check rate limiting: deduplicate notifications for same domain within dedupeWindow
	if lastTime, ok := w.lastNotified.Load(event.Domain); ok {
		if time.Since(lastTime.(time.Time)) < w.dedupeWindow {
			// Notification suppressed due to rate limiting
			return nil
		}
	}

	// Update last notification time for this domain
	w.lastNotified.Store(event.Domain, time.Now())

	// Marshal event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DarkPipe-Event", event.Type)

	// Send request (best-effort, do not block mail flow on failure)
	resp, err := w.httpClient.Do(req)
	if err != nil {
		// Log failure but do not propagate error (notifications are best-effort)
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the webhook notifier and releases resources.
func (w *WebhookNotifier) Close() error {
	// HTTP client does not require explicit cleanup
	return nil
}
