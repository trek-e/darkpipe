// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package dkim

import (
	"fmt"
	"time"
)

// GenerateSelector creates a time-based DKIM selector name.
// Format: {prefix}-{YYYY}q{Q} (e.g., "darkpipe-2026q1")
// This naming convention enables automated quarterly key rotation.
func GenerateSelector(prefix string, t time.Time) string {
	year := t.Year()
	quarter := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%s-%dq%d", prefix, year, quarter)
}

// GetCurrentSelector returns the selector for the current quarter.
func GetCurrentSelector(prefix string) string {
	return GenerateSelector(prefix, time.Now())
}

// GetNextSelector returns the selector for the next quarter.
// Used for rotation planning: publish new selector before switching signing key.
func GetNextSelector(prefix string) string {
	now := time.Now()
	nextQuarter := now.AddDate(0, 3, 0)
	return GenerateSelector(prefix, nextQuarter)
}

// ShouldRotate returns true if the current time is in a different quarter
// than the lastRotation time, indicating a key rotation is due.
func ShouldRotate(lastRotation time.Time) bool {
	now := time.Now()
	nowQuarter := (int(now.Month())-1)/3 + 1
	lastQuarter := (int(lastRotation.Month())-1)/3 + 1

	// Different year or different quarter in same year
	return now.Year() != lastRotation.Year() || nowQuarter != lastQuarter
}
