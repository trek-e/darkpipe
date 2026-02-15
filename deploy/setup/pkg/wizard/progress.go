// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package wizard

import (
	"fmt"
	"os"
	"sync"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// MigrationProgress manages migration progress display with per-folder progress bars
type MigrationProgress struct {
	mu               sync.Mutex
	progressBars     map[string]*pterm.ProgressbarPrinter
	overallBar       *pterm.ProgressbarPrinter
	totalFolders     int
	completedFolders int
	useFallback      bool // true if terminal doesn't support multi-printer
}

// NewMigrationProgress creates a new migration progress tracker
func NewMigrationProgress() *MigrationProgress {
	// Detect if terminal supports multi-printer
	useFallback := !isTerminal() || os.Getenv("TERM") == "dumb"

	return &MigrationProgress{
		progressBars: make(map[string]*pterm.ProgressbarPrinter),
		useFallback:  useFallback,
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Start initializes the progress display
func (p *MigrationProgress) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.useFallback {
		// Fallback mode: simple text output
		fmt.Println("Starting migration...")
		return nil
	}

	// Multi-printer mode would be initialized here if needed
	// For now, we'll use individual progress bars
	return nil
}

// SetOverall sets the total number of folders for overall progress
func (p *MigrationProgress) SetOverall(totalFolders int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalFolders = totalFolders

	if p.useFallback {
		fmt.Printf("Total folders to migrate: %d\n", totalFolders)
		return
	}

	// Create overall progress bar
	p.overallBar, _ = pterm.DefaultProgressbar.
		WithTotal(totalFolders).
		WithTitle("Overall Progress").
		WithShowCount(true).
		Start()
}

// StartFolder creates a progress bar for a specific folder
func (p *MigrationProgress) StartFolder(folder string, totalMessages int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.useFallback {
		fmt.Printf("[%d/%d] Starting folder: %s (%d messages)\n",
			p.completedFolders+1, p.totalFolders, folder, totalMessages)
		return
	}

	// Create folder-specific progress bar
	bar, _ := pterm.DefaultProgressbar.
		WithTotal(totalMessages).
		WithTitle(fmt.Sprintf("Migrating %s", folder)).
		WithShowCount(true).
		Start()

	p.progressBars[folder] = bar
}

// UpdateFolder updates the progress for a specific folder
func (p *MigrationProgress) UpdateFolder(folder string, current, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.useFallback {
		// Fallback: print periodic updates (every 50 messages)
		if current%50 == 0 || current == total {
			fmt.Printf("  %s: %d/%d messages processed\n", folder, current, total)
		}
		return
	}

	// Update folder progress bar (set to current count, not increment)
	if bar, exists := p.progressBars[folder]; exists {
		bar.Current = current
	}
}

// CompleteFolder marks a folder as done and increments overall progress
func (p *MigrationProgress) CompleteFolder(folder string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.completedFolders++

	if p.useFallback {
		fmt.Printf("[%d/%d] Completed folder: %s\n",
			p.completedFolders, p.totalFolders, folder)
		return
	}

	// Stop folder progress bar
	if bar, exists := p.progressBars[folder]; exists {
		bar.Stop()
		delete(p.progressBars, folder)
	}

	// Update overall progress
	if p.overallBar != nil {
		p.overallBar.Add(1)
	}
}

// Stop cleans up the progress display
func (p *MigrationProgress) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.useFallback {
		fmt.Println("Migration complete.")
		return
	}

	// Stop all remaining progress bars
	for _, bar := range p.progressBars {
		bar.Stop()
	}

	if p.overallBar != nil {
		p.overallBar.Stop()
	}
}
