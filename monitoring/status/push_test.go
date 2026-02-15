// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package status

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPingGET(t *testing.T) {
	// Create test server
	pinged := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		pinged = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 1*time.Minute)

	err := pinger.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if !pinged {
		t.Error("Server was not pinged")
	}
}

func TestPingWithStatusPOST(t *testing.T) {
	// Create test server
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 1*time.Minute)

	status := &SystemStatus{
		OverallStatus: "healthy",
		Timestamp:     time.Now(),
	}

	err := pinger.PingWithStatus(context.Background(), status)
	if err != nil {
		t.Fatalf("PingWithStatus failed: %v", err)
	}

	// Verify JSON was sent
	if !strings.Contains(receivedBody, `"overall_status":"healthy"`) {
		t.Errorf("Expected JSON with overall_status, got: %s", receivedBody)
	}
}

func TestPingServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 1*time.Minute)

	err := pinger.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for 500 response, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error mentioning status 500, got: %v", err)
	}
}

func TestPingDisabled(t *testing.T) {
	// Empty URL means disabled
	pinger := NewHealthchecksPinger("", 1*time.Minute)

	err := pinger.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping should be no-op when disabled, got error: %v", err)
	}

	status := &SystemStatus{OverallStatus: "healthy"}
	err = pinger.PingWithStatus(context.Background(), status)
	if err != nil {
		t.Errorf("PingWithStatus should be no-op when disabled, got error: %v", err)
	}
}

func TestPingContextCancellation(t *testing.T) {
	// Server that takes too long
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 1*time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := pinger.Ping(ctx)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestRunCancellation(t *testing.T) {
	pingCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	getStatus := func() (*SystemStatus, error) {
		return &SystemStatus{OverallStatus: "healthy"}, nil
	}

	// Run in background
	go pinger.Run(ctx, getStatus)

	// Wait for a couple pings
	time.Sleep(120 * time.Millisecond)

	// Cancel and verify it stops
	cancel()
	time.Sleep(100 * time.Millisecond)

	finalCount := pingCount

	// Wait a bit more - count should not increase
	time.Sleep(100 * time.Millisecond)

	if pingCount != finalCount {
		t.Errorf("Pinger did not stop after context cancellation (count changed from %d to %d)", finalCount, pingCount)
	}

	if pingCount < 2 {
		t.Errorf("Expected at least 2 pings before cancellation, got %d", pingCount)
	}
}

func TestRunDisabled(t *testing.T) {
	pinger := NewHealthchecksPinger("", 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	getStatus := func() (*SystemStatus, error) {
		return &SystemStatus{OverallStatus: "healthy"}, nil
	}

	// Should return immediately when disabled
	pinger.Run(ctx, getStatus)
	// If it doesn't hang, test passes
}

func TestNewHealthchecksPingerDefaultInterval(t *testing.T) {
	pinger := NewHealthchecksPinger("http://example.com", 0)

	if pinger.interval != 5*time.Minute {
		t.Errorf("Expected default interval 5 minutes, got %v", pinger.interval)
	}
}

func TestPingFallbackOnStatusError(t *testing.T) {
	// Track request types
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method)

		// Fail POST but succeed GET
		if r.Method == "POST" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	pinger := NewHealthchecksPinger(server.URL, 1*time.Minute)

	getStatus := func() (*SystemStatus, error) {
		return &SystemStatus{OverallStatus: "healthy"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Send one ping
	pinger.sendPing(ctx, getStatus)

	// Should have tried POST (with status) then fallen back to GET
	if len(requests) < 2 {
		t.Errorf("Expected fallback to GET after POST failure, got %d requests", len(requests))
	}
}
