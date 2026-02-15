// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package ephemeral provides verification that no mail content persists on disk after forwarding.
//
// The cloud relay must be ephemeral per RELAY-02: no mail should remain on the filesystem
// after successful forwarding to the home device. This package scans Postfix queue directories
// and detects violations.
package ephemeral

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ViolationType indicates what kind of persisted mail was found.
type ViolationType string

const (
	QueueFile ViolationType = "queue_file"
	DataFile  ViolationType = "data_file"
	TempFile  ViolationType = "temp_file"
)

// Violation represents a single instance of persisted mail content.
type Violation struct {
	Path    string        // Full path to the offending file
	Size    int64         // File size in bytes
	ModTime time.Time     // Last modification time
	Type    ViolationType // Type of violation
}

// VerifyResult contains the result of ephemeral storage verification.
type VerifyResult struct {
	Clean        bool        // True if no violations found
	Violations   []Violation // List of persisted mail files found
	ScannedPaths []string    // Paths that were scanned
	ScannedAt    time.Time   // When the scan completed
}

// VerifyNoPersistedMail scans Postfix queue directories for lingering mail files.
//
// After successful forwarding, all queue directories should be empty except for
// Postfix control files (pid files, locks, etc.). This function detects mail data
// files that indicate forwarding failure or queue buildup.
//
// Scanned directories:
//   - incoming/ - New mail waiting for processing
//   - active/ - Currently being processed
//   - deferred/ - Failed delivery, will retry
//   - hold/ - Administratively held
//   - corrupt/ - Unreadable messages
//
// Returns VerifyResult with violations list if mail files are found.
func VerifyNoPersistedMail(postfixQueueDir string) (*VerifyResult, error) {
	result := &VerifyResult{
		Clean:        true,
		Violations:   []Violation{},
		ScannedPaths: []string{},
		ScannedAt:    time.Now().UTC(),
	}

	// Queue directories that should be empty after successful forwarding
	queueDirs := []string{
		"incoming",
		"active",
		"deferred",
		"hold",
		"corrupt",
	}

	for _, dir := range queueDirs {
		path := filepath.Join(postfixQueueDir, dir)
		result.ScannedPaths = append(result.ScannedPaths, path)

		// Check if directory exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue // Directory doesn't exist, skip
		}

		// Scan directory for files
		err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip the directory itself
			if filePath == path {
				return nil
			}

			// Skip directories (only flag files)
			if d.IsDir() {
				return nil
			}

			// Skip Postfix control files that are expected to exist
			if isControlFile(d.Name()) {
				return nil
			}

			// This is a violation - mail data file found
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("stat %s: %w", filePath, err)
			}

			violation := Violation{
				Path:    filePath,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Type:    detectViolationType(dir, d.Name()),
			}

			result.Violations = append(result.Violations, violation)
			result.Clean = false

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", path, err)
		}
	}

	return result, nil
}

// WatchAndVerify runs periodic verification and calls alertFunc when violations are found.
//
// This function blocks until the context is cancelled. It runs VerifyNoPersistedMail
// every interval and invokes alertFunc if violations are detected. This allows integration
// with notification systems to alert operators of queue buildup.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	go WatchAndVerify(ctx, 60*time.Second, "/var/spool/postfix", func(result VerifyResult) {
//	    log.Printf("ALERT: %d mail files persisted", len(result.Violations))
//	})
func WatchAndVerify(ctx context.Context, interval time.Duration, queueDir string, alertFunc func(VerifyResult)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := VerifyNoPersistedMail(queueDir)
			if err != nil {
				// On error, create a synthetic result indicating the problem
				alertFunc(VerifyResult{
					Clean:      false,
					ScannedAt:  time.Now().UTC(),
					Violations: []Violation{},
				})
				continue
			}

			if !result.Clean {
				alertFunc(*result)
			}
		}
	}
}

// isControlFile returns true if the filename is a Postfix control file that should
// be ignored during verification. These files are expected to exist and don't indicate
// persisted mail content.
func isControlFile(name string) bool {
	// Postfix control files to ignore
	controlFiles := []string{
		"pid",
		"master.lock",
		"public",
		"private",
		"maildrop",
	}

	for _, cf := range controlFiles {
		if name == cf {
			return true
		}
	}

	// Ignore dotfiles (hidden files)
	if strings.HasPrefix(name, ".") {
		return true
	}

	return false
}

// detectViolationType classifies the type of violation based on the queue directory
// and filename.
func detectViolationType(queueDir, filename string) ViolationType {
	switch queueDir {
	case "incoming", "active":
		return QueueFile
	case "deferred", "hold", "corrupt":
		return DataFile
	default:
		return TempFile
	}
}
