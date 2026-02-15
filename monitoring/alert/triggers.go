// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package alert

import (
	"fmt"
	"time"
)

// TriggerConfig holds thresholds for alert evaluation.
type TriggerConfig struct {
	// Certificate expiry alerts
	CertWarnDays     int // Days before expiry to send warning (default: 14)
	CertCriticalDays int // Days before expiry to send critical (default: 7)

	// Queue health alerts
	QueueDepthWarn     int // Queue depth threshold for warning (default: 50)
	QueueDepthCritical int // Queue depth threshold for critical (default: 200)

	// Delivery health alerts
	DeliveryFailureRate float64 // Bounce rate threshold (default: 0.5 = 50%)

	// Tunnel health alerts
	TunnelDownEnabled bool // Enable tunnel down alerts (default: true)
}

// DefaultTriggerConfig returns the default trigger configuration.
func DefaultTriggerConfig() TriggerConfig {
	return TriggerConfig{
		CertWarnDays:        14,
		CertCriticalDays:    7,
		QueueDepthWarn:      50,
		QueueDepthCritical:  200,
		DeliveryFailureRate: 0.5,
		TunnelDownEnabled:   true,
	}
}

// EvaluateCertExpiry evaluates certificate expiry and returns an alert if needed.
// Returns:
//   - Critical alert if daysLeft <= CertCriticalDays
//   - Warning alert if daysLeft <= CertWarnDays
//   - nil if no alert needed
func EvaluateCertExpiry(daysLeft int, config TriggerConfig) *Alert {
	if daysLeft <= config.CertCriticalDays {
		return &Alert{
			Type:      AlertCertExpiry,
			Severity:  SeverityCritical,
			Title:     "Certificate Expiring Soon",
			Message:   fmt.Sprintf("Certificate expires in %d days (critical threshold: %d days)", daysLeft, config.CertCriticalDays),
			Timestamp: time.Now(),
			Details: map[string]string{
				"days_left":          fmt.Sprintf("%d", daysLeft),
				"critical_threshold": fmt.Sprintf("%d", config.CertCriticalDays),
			},
		}
	}

	if daysLeft <= config.CertWarnDays {
		return &Alert{
			Type:      AlertCertExpiry,
			Severity:  SeverityWarn,
			Title:     "Certificate Expiring Soon",
			Message:   fmt.Sprintf("Certificate expires in %d days (warning threshold: %d days)", daysLeft, config.CertWarnDays),
			Timestamp: time.Now(),
			Details: map[string]string{
				"days_left":       fmt.Sprintf("%d", daysLeft),
				"warn_threshold": fmt.Sprintf("%d", config.CertWarnDays),
			},
		}
	}

	return nil
}

// EvaluateQueueHealth evaluates queue health and returns an alert if needed.
// Returns:
//   - Critical alert if stuck > 0 (messages stuck in queue)
//   - Warning alert if depth > QueueDepthWarn
//   - Critical alert if depth > QueueDepthCritical
//   - nil if no alert needed
func EvaluateQueueHealth(depth, stuck int, config TriggerConfig) *Alert {
	// Stuck messages are always critical (mail flow blocked)
	if stuck > 0 {
		return &Alert{
			Type:      AlertQueueBackup,
			Severity:  SeverityCritical,
			Title:     "Queue Messages Stuck",
			Message:   fmt.Sprintf("%d messages stuck in queue (not making progress)", stuck),
			Timestamp: time.Now(),
			Details: map[string]string{
				"stuck_count": fmt.Sprintf("%d", stuck),
				"queue_depth": fmt.Sprintf("%d", depth),
			},
		}
	}

	// High queue depth
	if depth > config.QueueDepthCritical {
		return &Alert{
			Type:      AlertQueueBackup,
			Severity:  SeverityCritical,
			Title:     "Queue Depth Critical",
			Message:   fmt.Sprintf("Queue depth is %d (critical threshold: %d)", depth, config.QueueDepthCritical),
			Timestamp: time.Now(),
			Details: map[string]string{
				"queue_depth":        fmt.Sprintf("%d", depth),
				"critical_threshold": fmt.Sprintf("%d", config.QueueDepthCritical),
			},
		}
	}

	if depth > config.QueueDepthWarn {
		return &Alert{
			Type:      AlertQueueBackup,
			Severity:  SeverityWarn,
			Title:     "Queue Depth Elevated",
			Message:   fmt.Sprintf("Queue depth is %d (warning threshold: %d)", depth, config.QueueDepthWarn),
			Timestamp: time.Now(),
			Details: map[string]string{
				"queue_depth":    fmt.Sprintf("%d", depth),
				"warn_threshold": fmt.Sprintf("%d", config.QueueDepthWarn),
			},
		}
	}

	return nil
}

// EvaluateDeliveryHealth evaluates delivery health based on bounce rate.
// Returns:
//   - Critical alert if bounceRate > DeliveryFailureRate threshold
//   - nil if no alert needed
func EvaluateDeliveryHealth(bounceRate float64, config TriggerConfig) *Alert {
	if bounceRate > config.DeliveryFailureRate {
		return &Alert{
			Type:      AlertDeliveryFailure,
			Severity:  SeverityCritical,
			Title:     "High Delivery Failure Rate",
			Message:   fmt.Sprintf("Bounce rate is %.1f%% (threshold: %.1f%%)", bounceRate*100, config.DeliveryFailureRate*100),
			Timestamp: time.Now(),
			Details: map[string]string{
				"bounce_rate":        fmt.Sprintf("%.2f", bounceRate),
				"failure_threshold": fmt.Sprintf("%.2f", config.DeliveryFailureRate),
			},
		}
	}

	return nil
}

// EvaluateTunnelHealth evaluates tunnel connectivity.
// Returns:
//   - Critical alert if tunnel is down
//   - nil if tunnel is healthy
func EvaluateTunnelHealth(healthy bool, config TriggerConfig) *Alert {
	if !config.TunnelDownEnabled {
		return nil
	}

	if !healthy {
		return &Alert{
			Type:      AlertTunnelDown,
			Severity:  SeverityCritical,
			Title:     "Tunnel Connection Down",
			Message:   "WireGuard/mTLS tunnel is not responding",
			Timestamp: time.Now(),
			Details: map[string]string{
				"status": "down",
			},
		}
	}

	return nil
}
