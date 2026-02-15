// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package status

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/cert"
	"github.com/darkpipe/darkpipe/monitoring/delivery"
	"github.com/darkpipe/darkpipe/monitoring/health"
	"github.com/darkpipe/darkpipe/monitoring/queue"
)

// Mock implementations for testing

type mockHealthChecker struct {
	status *health.HealthStatus
}

func (m *mockHealthChecker) RegisterCheck(name string, fn health.CheckFunc) {
	// No-op for mock
}

func (m *mockHealthChecker) Liveness() health.HealthStatus {
	return health.HealthStatus{Status: "up"}
}

func (m *mockHealthChecker) Readiness(ctx context.Context) health.HealthStatus {
	return *m.status
}

type mockCertWatcher struct {
	certs []cert.CertInfo
	err   error
}

func (m *mockCertWatcher) CheckCert(path string) (*cert.CertInfo, error) {
	return nil, nil
}

func (m *mockCertWatcher) CheckAll() ([]cert.CertInfo, error) {
	return m.certs, m.err
}

func TestGetStatusAllHealthy(t *testing.T) {
	// Create mock dependencies - all healthy
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
				{Name: "imap", Status: "ok"},
			},
		},
	}

	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{
			Depth:    5,
			Deferred: 0,
			Stuck:    0,
		}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)
	deliveryTracker.Record(&delivery.LogEntry{
		QueueID:   "ABC123",
		Status:    "delivered",
		Timestamp: time.Now(),
	})

	certWatcher := &mockCertWatcher{
		certs: []cert.CertInfo{
			{
				Subject:  "relay.example.com",
				NotAfter: time.Now().Add(62 * 24 * time.Hour), // 62 days
				DaysLeft: 62,
			},
		},
	}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "healthy" {
		t.Errorf("Expected overall status 'healthy', got %q", status.OverallStatus)
	}

	if status.Health.Status != "up" {
		t.Errorf("Expected health status 'up', got %q", status.Health.Status)
	}

	if status.Queue.Depth != 5 {
		t.Errorf("Expected queue depth 5, got %d", status.Queue.Depth)
	}

	if status.Delivery.Delivered != 1 {
		t.Errorf("Expected 1 delivered, got %d", status.Delivery.Delivered)
	}

	if len(status.Certificates.Certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(status.Certificates.Certificates))
	}

	if status.Certificates.DaysUntilExpiry < 60 || status.Certificates.DaysUntilExpiry > 63 {
		t.Errorf("Expected ~62 days until expiry, got %d", status.Certificates.DaysUntilExpiry)
	}
}

func TestGetStatusFailingHealthCheck(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "down",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "error"},
			},
		},
	}

	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{Depth: 0}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)

	certWatcher := &mockCertWatcher{
		certs: []cert.CertInfo{},
	}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy' for failed health check, got %q", status.OverallStatus)
	}
}

func TestGetStatusCertWarning(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
	}

	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{Depth: 0}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)

	// Certificate expiring in 10 days (should trigger warning)
	certWatcher := &mockCertWatcher{
		certs: []cert.CertInfo{
			{
				Subject:  "relay.example.com",
				NotAfter: time.Now().Add(10 * 24 * time.Hour),
				DaysLeft: 10,
			},
		},
	}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "degraded" {
		t.Errorf("Expected overall status 'degraded' for cert warning (10 days), got %q", status.OverallStatus)
	}
}

func TestGetStatusCertCritical(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
	}

	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{Depth: 0}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)

	// Certificate expiring in 5 days (should trigger critical)
	certWatcher := &mockCertWatcher{
		certs: []cert.CertInfo{
			{
				Subject:  "relay.example.com",
				NotAfter: time.Now().Add(5 * 24 * time.Hour),
				DaysLeft: 5,
			},
		},
	}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy' for cert critical (5 days), got %q", status.OverallStatus)
	}
}

func TestGetStatusQueueBackup(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
	}

	// High queue depth should trigger degraded
	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{
			Depth:    100,
			Deferred: 0,
			Stuck:    0,
		}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)
	certWatcher := &mockCertWatcher{certs: []cert.CertInfo{}}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "degraded" {
		t.Errorf("Expected overall status 'degraded' for high queue depth, got %q", status.OverallStatus)
	}
}

func TestGetStatusStuckMessages(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
	}

	// Stuck messages should trigger unhealthy
	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{
			Depth:    10,
			Deferred: 0,
			Stuck:    2,
		}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)
	certWatcher := &mockCertWatcher{certs: []cert.CertInfo{}}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.OverallStatus != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy' for stuck messages, got %q", status.OverallStatus)
	}
}

func TestGetStatusQueueError(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
	}

	// Queue function returns error - should handle gracefully
	queueFunc := func() (*queue.QueueStats, error) {
		return nil, errors.New("postqueue not available")
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)
	certWatcher := &mockCertWatcher{certs: []cert.CertInfo{}}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus should not fail when queue unavailable: %v", err)
	}

	// Should still return status with zero queue values
	if status.Queue.Depth != 0 {
		t.Errorf("Expected queue depth 0 on error, got %d", status.Queue.Depth)
	}
}

func TestGetStatusMultipleCerts(t *testing.T) {
	healthChecker := &mockHealthChecker{
		status: &health.HealthStatus{Status: "up"},
	}

	queueFunc := func() (*queue.QueueStats, error) {
		return &queue.QueueStats{Depth: 0}, nil
	}

	deliveryTracker := delivery.NewDeliveryTracker(100)

	// Multiple certs - should pick the one expiring soonest
	certWatcher := &mockCertWatcher{
		certs: []cert.CertInfo{
			{
				Subject:  "relay.example.com",
				NotAfter: time.Now().Add(60 * 24 * time.Hour),
				DaysLeft: 60,
			},
			{
				Subject:  "internal-ca",
				NotAfter: time.Now().Add(30 * 24 * time.Hour), // Expires sooner
				DaysLeft: 30,
			},
		},
	}

	aggregator := NewStatusAggregator(healthChecker, queueFunc, deliveryTracker, certWatcher)

	status, err := aggregator.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	// Should use the 30-day cert as next expiry
	if status.Certificates.DaysUntilExpiry < 29 || status.Certificates.DaysUntilExpiry > 31 {
		t.Errorf("Expected ~30 days until expiry (soonest cert), got %d", status.Certificates.DaysUntilExpiry)
	}
}
