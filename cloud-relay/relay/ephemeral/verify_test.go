// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package ephemeral

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifyNoPersistedMail_CleanQueue(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory structure
	queueDir := t.TempDir()
	queueDirs := []string{"incoming", "active", "deferred", "hold", "corrupt"}
	for _, dir := range queueDirs {
		if err := os.Mkdir(filepath.Join(queueDir, dir), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	// Verify clean queue
	result, err := VerifyNoPersistedMail(queueDir)
	if err != nil {
		t.Fatalf("VerifyNoPersistedMail: %v", err)
	}

	if !result.Clean {
		t.Errorf("expected Clean=true, got false")
	}

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}

	if len(result.ScannedPaths) != len(queueDirs) {
		t.Errorf("expected %d scanned paths, got %d", len(queueDirs), len(result.ScannedPaths))
	}

	if result.ScannedAt.IsZero() {
		t.Errorf("ScannedAt should not be zero")
	}
}

func TestVerifyNoPersistedMail_IncomingViolation(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory with file in incoming/
	queueDir := t.TempDir()
	incomingDir := filepath.Join(queueDir, "incoming")
	if err := os.Mkdir(incomingDir, 0755); err != nil {
		t.Fatalf("mkdir incoming: %v", err)
	}

	// Create a mail file
	mailFile := filepath.Join(incomingDir, "ABCD1234")
	if err := os.WriteFile(mailFile, []byte("test mail content"), 0644); err != nil {
		t.Fatalf("write mail file: %v", err)
	}

	// Verify should detect violation
	result, err := VerifyNoPersistedMail(queueDir)
	if err != nil {
		t.Fatalf("VerifyNoPersistedMail: %v", err)
	}

	if result.Clean {
		t.Errorf("expected Clean=false, got true")
	}

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	violation := result.Violations[0]
	if violation.Path != mailFile {
		t.Errorf("violation path = %s, want %s", violation.Path, mailFile)
	}

	if violation.Size != 17 {
		t.Errorf("violation size = %d, want 17", violation.Size)
	}

	if violation.Type != QueueFile {
		t.Errorf("violation type = %s, want %s", violation.Type, QueueFile)
	}

	if violation.ModTime.IsZero() {
		t.Errorf("violation ModTime should not be zero")
	}
}

func TestVerifyNoPersistedMail_DeferredViolation(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory with file in deferred/
	queueDir := t.TempDir()
	deferredDir := filepath.Join(queueDir, "deferred")
	if err := os.Mkdir(deferredDir, 0755); err != nil {
		t.Fatalf("mkdir deferred: %v", err)
	}

	// Create a deferred mail file
	mailFile := filepath.Join(deferredDir, "DEFERRED123")
	if err := os.WriteFile(mailFile, []byte("deferred mail"), 0644); err != nil {
		t.Fatalf("write mail file: %v", err)
	}

	// Verify should detect violation
	result, err := VerifyNoPersistedMail(queueDir)
	if err != nil {
		t.Fatalf("VerifyNoPersistedMail: %v", err)
	}

	if result.Clean {
		t.Errorf("expected Clean=false, got true")
	}

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	violation := result.Violations[0]
	if violation.Type != DataFile {
		t.Errorf("violation type = %s, want %s", violation.Type, DataFile)
	}
}

func TestVerifyNoPersistedMail_IgnoresControlFiles(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory with control files
	queueDir := t.TempDir()
	activeDir := filepath.Join(queueDir, "active")
	if err := os.Mkdir(activeDir, 0755); err != nil {
		t.Fatalf("mkdir active: %v", err)
	}

	// Create control files that should be ignored
	controlFiles := []string{"pid", "master.lock", ".hidden"}
	for _, name := range controlFiles {
		if err := os.WriteFile(filepath.Join(activeDir, name), []byte("control"), 0644); err != nil {
			t.Fatalf("write control file %s: %v", name, err)
		}
	}

	// Verify should ignore control files
	result, err := VerifyNoPersistedMail(queueDir)
	if err != nil {
		t.Fatalf("VerifyNoPersistedMail: %v", err)
	}

	if !result.Clean {
		t.Errorf("expected Clean=true (control files should be ignored), got false")
	}

	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations (control files should be ignored), got %d", len(result.Violations))
	}
}

func TestVerifyNoPersistedMail_MultipleViolations(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory with multiple violations
	queueDir := t.TempDir()
	incomingDir := filepath.Join(queueDir, "incoming")
	deferredDir := filepath.Join(queueDir, "deferred")

	if err := os.Mkdir(incomingDir, 0755); err != nil {
		t.Fatalf("mkdir incoming: %v", err)
	}
	if err := os.Mkdir(deferredDir, 0755); err != nil {
		t.Fatalf("mkdir deferred: %v", err)
	}

	// Create mail files in different queues
	if err := os.WriteFile(filepath.Join(incomingDir, "MAIL1"), []byte("mail 1"), 0644); err != nil {
		t.Fatalf("write mail1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(incomingDir, "MAIL2"), []byte("mail 2"), 0644); err != nil {
		t.Fatalf("write mail2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deferredDir, "DEFERRED1"), []byte("deferred"), 0644); err != nil {
		t.Fatalf("write deferred: %v", err)
	}

	// Verify should detect all violations
	result, err := VerifyNoPersistedMail(queueDir)
	if err != nil {
		t.Fatalf("VerifyNoPersistedMail: %v", err)
	}

	if result.Clean {
		t.Errorf("expected Clean=false, got true")
	}

	if len(result.Violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(result.Violations))
	}
}

func TestWatchAndVerify(t *testing.T) {
	t.Parallel()

	// Create temporary queue directory with a violation
	queueDir := t.TempDir()
	incomingDir := filepath.Join(queueDir, "incoming")
	if err := os.Mkdir(incomingDir, 0755); err != nil {
		t.Fatalf("mkdir incoming: %v", err)
	}

	mailFile := filepath.Join(incomingDir, "TESTMAIL")
	if err := os.WriteFile(mailFile, []byte("test"), 0644); err != nil {
		t.Fatalf("write mail file: %v", err)
	}

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Track alerts
	alertChan := make(chan VerifyResult, 1)
	alertFunc := func(result VerifyResult) {
		select {
		case alertChan <- result:
		default:
		}
	}

	// Start watcher in background
	go WatchAndVerify(ctx, 50*time.Millisecond, queueDir, alertFunc)

	// Wait for alert
	select {
	case result := <-alertChan:
		if result.Clean {
			t.Errorf("expected Clean=false in alert, got true")
		}
		if len(result.Violations) != 1 {
			t.Errorf("expected 1 violation in alert, got %d", len(result.Violations))
		}
	case <-ctx.Done():
		t.Errorf("timeout waiting for alert")
	}
}

func TestWatchAndVerify_Cancellation(t *testing.T) {
	t.Parallel()

	// Create clean queue directory
	queueDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(queueDir, "incoming"), 0755); err != nil {
		t.Fatalf("mkdir incoming: %v", err)
	}

	// Set up context
	ctx, cancel := context.WithCancel(context.Background())

	// Alert function should not be called for clean queue
	alertCount := 0
	alertFunc := func(result VerifyResult) {
		alertCount++
	}

	// Start watcher
	done := make(chan struct{})
	go func() {
		WatchAndVerify(ctx, 50*time.Millisecond, queueDir, alertFunc)
		close(done)
	}()

	// Cancel after a short time
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for goroutine to exit
	select {
	case <-done:
		// Success - goroutine exited on cancellation
	case <-time.After(500 * time.Millisecond):
		t.Errorf("WatchAndVerify did not exit after context cancellation")
	}

	if alertCount != 0 {
		t.Errorf("expected 0 alerts for clean queue, got %d", alertCount)
	}
}

func TestIsControlFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"pid file", "pid", true},
		{"master lock", "master.lock", true},
		{"hidden file", ".hidden", true},
		{"public", "public", true},
		{"private", "private", true},
		{"maildrop", "maildrop", true},
		{"mail file", "ABCD1234", false},
		{"deferred", "DEFERRED123", false},
		{"regular file", "somefile.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isControlFile(tt.filename)
			if got != tt.want {
				t.Errorf("isControlFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectViolationType(t *testing.T) {
	tests := []struct {
		queueDir string
		filename string
		want     ViolationType
	}{
		{"incoming", "MAIL123", QueueFile},
		{"active", "ACTIVE456", QueueFile},
		{"deferred", "DEFERRED789", DataFile},
		{"hold", "HOLD123", DataFile},
		{"corrupt", "CORRUPT456", DataFile},
		{"unknown", "FILE123", TempFile},
	}

	for _, tt := range tests {
		t.Run(tt.queueDir+"/"+tt.filename, func(t *testing.T) {
			got := detectViolationType(tt.queueDir, tt.filename)
			if got != tt.want {
				t.Errorf("detectViolationType(%q, %q) = %v, want %v", tt.queueDir, tt.filename, got, tt.want)
			}
		})
	}
}
