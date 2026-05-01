// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package status

import (
	"context"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/cert"
	"github.com/darkpipe/darkpipe/monitoring/delivery"
	"github.com/darkpipe/darkpipe/monitoring/health"
	"github.com/darkpipe/darkpipe/monitoring/queue"
	"github.com/darkpipe/darkpipe/monitoring/status/healtheval"
)

// SystemStatus represents the complete monitoring state of the DarkPipe system
type SystemStatus struct {
	Health         HealthSummary    `json:"health"`
	Queue          QueueSummary     `json:"queue"`
	Delivery       DeliverySummary  `json:"delivery"`
	Certificates   CertSummary      `json:"certificates"`
	Timestamp      time.Time        `json:"timestamp"`
	OverallStatus  string           `json:"overall_status"` // "healthy", "degraded", "unhealthy"
	OverallReasons []string         `json:"overall_reasons,omitempty"`
	TriggeredRules []healtheval.TriggeredRule `json:"triggered_rules,omitempty"`
}

// HealthSummary contains health check status
type HealthSummary struct {
	Status string                 `json:"status"` // "up", "down"
	Checks []health.CheckResult   `json:"checks"`
}

// QueueSummary contains mail queue metrics
type QueueSummary struct {
	Depth    int `json:"depth"`
	Deferred int `json:"deferred"`
	Stuck    int `json:"stuck"`
}

// DeliverySummary contains delivery statistics
type DeliverySummary struct {
	Delivered     int                   `json:"delivered"`
	Deferred      int                   `json:"deferred"`
	Bounced       int                   `json:"bounced"`
	Total         int                   `json:"total"`
	RecentEntries []delivery.LogEntry   `json:"recent_entries,omitempty"`
}

// CertSummary contains certificate expiry information
type CertSummary struct {
	Certificates   []cert.CertInfo `json:"certificates"`
	NextExpiry     time.Time       `json:"next_expiry"`
	DaysUntilExpiry int            `json:"days_until_expiry"`
}

// HealthChecker interface for health check operations
type HealthChecker interface {
	Readiness(ctx context.Context) health.HealthStatus
}

// CertWatcher interface for certificate monitoring
type CertWatcher interface {
	CheckAll() ([]cert.CertInfo, error)
}

// DeliveryTracker interface for delivery statistics
type DeliveryTracker interface {
	GetStats() delivery.DeliveryStats
	GetRecent(n int) []delivery.LogEntry
}

// StatusAggregator collects monitoring data from all sources
type StatusAggregator struct {
	health   HealthChecker
	queue    func() (*queue.QueueStats, error)
	delivery DeliveryTracker
	certs    CertWatcher
	evaluator healtheval.Module
}

// NewStatusAggregator creates a new status aggregator with dependency injection
func NewStatusAggregator(
	healthChecker HealthChecker,
	queueFunc func() (*queue.QueueStats, error),
	deliveryTracker DeliveryTracker,
	certWatcher CertWatcher,
) *StatusAggregator {
	return &StatusAggregator{
		health:   healthChecker,
		queue:    queueFunc,
		delivery: deliveryTracker,
		certs:    certWatcher,
		evaluator: healtheval.New(healtheval.Policy{}),
	}
}

// GetStatus collects monitoring data from all sources and computes overall status
func (s *StatusAggregator) GetStatus(ctx context.Context) (*SystemStatus, error) {
	now := time.Now()

	// Collect health status
	healthStatus := s.health.Readiness(ctx)
	healthSummary := HealthSummary{
		Status: healthStatus.Status,
		Checks: healthStatus.Checks,
	}

	// Collect queue stats
	queueStats, err := s.queue()
	queueSummary := QueueSummary{}
	if err == nil && queueStats != nil {
		queueSummary.Depth = queueStats.Depth
		queueSummary.Deferred = queueStats.Deferred
		queueSummary.Stuck = queueStats.Stuck
	}

	// Collect delivery stats
	deliveryStats := s.delivery.GetStats()
	deliverySummary := DeliverySummary{
		Delivered:     deliveryStats.Delivered,
		Deferred:      deliveryStats.Deferred,
		Bounced:       deliveryStats.Bounced,
		Total:         deliveryStats.Total,
		RecentEntries: s.delivery.GetRecent(10), // Last 10 deliveries
	}

	// Collect certificate info
	certInfos, err := s.certs.CheckAll()
	certSummary := CertSummary{
		Certificates: certInfos,
	}
	if err != nil {
		// Log error but continue - certInfos will be empty slice
		_ = err
	}

	// Find next expiring certificate
	if len(certInfos) > 0 {
		nextExpiry := certInfos[0].NotAfter
		for _, c := range certInfos {
			if c.NotAfter.Before(nextExpiry) {
				nextExpiry = c.NotAfter
			}
		}
		certSummary.NextExpiry = nextExpiry
		certSummary.DaysUntilExpiry = int(time.Until(nextExpiry).Hours() / 24)
	}

	// Compute overall status via evaluation module
	eval := s.evaluator.Evaluate(healtheval.Snapshot{
		HealthStatus: healthSummary.Status,
		Checks: toEvalChecks(healthSummary.Checks),
		QueueDepth: queueSummary.Depth,
		QueueStuck: queueSummary.Stuck,
		Delivered: deliverySummary.Delivered,
		Deferred: deliverySummary.Deferred,
		Bounced: deliverySummary.Bounced,
		Total: deliverySummary.Total,
		DaysUntilCertExpiry: certSummary.DaysUntilExpiry,
	})

	return &SystemStatus{
		Health:        healthSummary,
		Queue:         queueSummary,
		Delivery:      deliverySummary,
		Certificates:  certSummary,
		Timestamp:     now,
		OverallStatus: eval.Status,
		OverallReasons: eval.Reasons,
		TriggeredRules: eval.Triggered,
	}, nil
}

func toEvalChecks(in []health.CheckResult) []healtheval.Check {
	out := make([]healtheval.Check, 0, len(in))
	for _, c := range in {
		out = append(out, healtheval.Check{Name: c.Name, Status: c.Status, Message: c.Message})
	}
	return out
}
