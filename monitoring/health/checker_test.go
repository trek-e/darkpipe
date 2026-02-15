// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChecker_Liveness(t *testing.T) {
	checker := NewChecker()
	status := checker.Liveness()

	if status.Status != "up" {
		t.Errorf("expected status 'up', got '%s'", status.Status)
	}

	if status.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	if len(status.Checks) != 0 {
		t.Errorf("expected no checks in liveness, got %d", len(status.Checks))
	}
}

func TestChecker_Readiness_AllPass(t *testing.T) {
	checker := NewChecker()

	// Register passing checks
	checker.RegisterCheck("test1", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "test1", Status: "ok", Message: "pass"}
	})
	checker.RegisterCheck("test2", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "test2", Status: "ok", Message: "pass"}
	})

	ctx := context.Background()
	status := checker.Readiness(ctx)

	if status.Status != "up" {
		t.Errorf("expected status 'up', got '%s'", status.Status)
	}

	if len(status.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(status.Checks))
	}

	for _, check := range status.Checks {
		if check.Status != "ok" {
			t.Errorf("check %s: expected status 'ok', got '%s'", check.Name, check.Status)
		}
		if check.Duration < 0 {
			t.Errorf("check %s: expected non-negative duration, got %v", check.Name, check.Duration)
		}
	}
}

func TestChecker_Readiness_OneFails(t *testing.T) {
	checker := NewChecker()

	checker.RegisterCheck("passing", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "passing", Status: "ok"}
	})
	checker.RegisterCheck("failing", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "failing", Status: "error", Message: "simulated failure"}
	})

	ctx := context.Background()
	status := checker.Readiness(ctx)

	if status.Status != "down" {
		t.Errorf("expected status 'down', got '%s'", status.Status)
	}

	if len(status.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(status.Checks))
	}

	// Verify the failing check has error details
	var foundFailure bool
	for _, check := range status.Checks {
		if check.Name == "failing" {
			foundFailure = true
			if check.Status != "error" {
				t.Errorf("expected failing check status 'error', got '%s'", check.Status)
			}
			if check.Message == "" {
				t.Error("expected failure message, got empty string")
			}
		}
	}

	if !foundFailure {
		t.Error("did not find failing check in results")
	}
}

func TestChecker_Readiness_ContextCancellation(t *testing.T) {
	checker := NewChecker()

	checker.RegisterCheck("slow", func(ctx context.Context) CheckResult {
		select {
		case <-time.After(100 * time.Millisecond):
			return CheckResult{Name: "slow", Status: "ok"}
		case <-ctx.Done():
			return CheckResult{Name: "slow", Status: "error", Message: ctx.Err().Error()}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	status := checker.Readiness(ctx)

	// Should complete even if check times out
	if len(status.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(status.Checks))
	}
}

func TestLivenessHandler(t *testing.T) {
	checker := NewChecker()
	handler := LivenessHandler(checker)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/health+json" {
		t.Errorf("expected Content-Type 'application/health+json', got '%s'", contentType)
	}

	var status HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if status.Status != "up" {
		t.Errorf("expected status 'up', got '%s'", status.Status)
	}
}

func TestLivenessHandler_WrongMethod(t *testing.T) {
	checker := NewChecker()
	handler := LivenessHandler(checker)

	req := httptest.NewRequest(http.MethodPost, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestReadinessHandler_Healthy(t *testing.T) {
	checker := NewChecker()
	checker.RegisterCheck("test", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "test", Status: "ok"}
	})

	handler := ReadinessHandler(checker)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/health+json" {
		t.Errorf("expected Content-Type 'application/health+json', got '%s'", contentType)
	}

	var status HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if status.Status != "up" {
		t.Errorf("expected status 'up', got '%s'", status.Status)
	}

	if len(status.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(status.Checks))
	}
}

func TestReadinessHandler_Unhealthy(t *testing.T) {
	checker := NewChecker()
	checker.RegisterCheck("failing", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "failing", Status: "error", Message: "service down"}
	})

	handler := ReadinessHandler(checker)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var status HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if status.Status != "down" {
		t.Errorf("expected status 'down', got '%s'", status.Status)
	}
}

func TestCheckResult_MarshalJSON(t *testing.T) {
	result := CheckResult{
		Name:     "test",
		Status:   "ok",
		Message:  "all good",
		Duration: 150 * time.Millisecond,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Check that duration_ms is numeric milliseconds
	durationMs, ok := decoded["duration_ms"].(float64)
	if !ok {
		t.Fatalf("duration_ms is not a number: %v", decoded["duration_ms"])
	}

	if durationMs != 150 {
		t.Errorf("expected duration_ms=150, got %v", durationMs)
	}
}
