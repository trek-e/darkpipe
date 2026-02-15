// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package health provides a unified health check framework with liveness
// and readiness separation for monitoring all DarkPipe mail services.
package health

import (
	"context"
	"encoding/json"
	"time"
)

// CheckFunc is a function that performs a single health check.
type CheckFunc func(ctx context.Context) CheckResult

// CheckResult holds the result of a single health check.
type CheckResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "ok" or "error"
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration_ms"`
}

// MarshalJSON implements custom JSON marshaling for CheckResult
// to convert Duration to milliseconds.
func (cr CheckResult) MarshalJSON() ([]byte, error) {
	type Alias CheckResult
	return json.Marshal(&struct {
		*Alias
		Duration int64 `json:"duration_ms"`
	}{
		Alias:    (*Alias)(&cr),
		Duration: cr.Duration.Milliseconds(),
	})
}

// HealthStatus represents the overall health status with individual check results.
type HealthStatus struct {
	Status    string        `json:"status"` // "up" or "down"
	Checks    []CheckResult `json:"checks,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Checker manages and executes health checks.
type Checker struct {
	checks map[string]CheckFunc
}

// NewChecker creates a new health checker.
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]CheckFunc),
	}
}

// RegisterCheck adds a named health check function.
func (c *Checker) RegisterCheck(name string, fn CheckFunc) {
	c.checks[name] = fn
}

// Liveness performs a cheap liveness check (process is alive).
// This always returns "up" as long as the process is running.
func (c *Checker) Liveness() HealthStatus {
	return HealthStatus{
		Status:    "up",
		Timestamp: time.Now(),
	}
}

// Readiness performs deep health checks on all registered services.
// Returns "down" if any check fails.
func (c *Checker) Readiness(ctx context.Context) HealthStatus {
	results := make([]CheckResult, 0, len(c.checks))
	overallStatus := "up"

	for name, checkFn := range c.checks {
		start := time.Now()
		result := checkFn(ctx)
		result.Duration = time.Since(start)

		// Ensure name is set
		if result.Name == "" {
			result.Name = name
		}

		results = append(results, result)

		if result.Status == "error" {
			overallStatus = "down"
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Checks:    results,
		Timestamp: time.Now(),
	}
}
