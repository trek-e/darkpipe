// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package delivery

import (
	"os"
	"strconv"
	"sync"
	"time"
)

// DeliveryTracker maintains an in-memory ring buffer of recent delivery events.
type DeliveryTracker struct {
	mu         sync.RWMutex
	entries    []LogEntry
	maxEntries int
	writeIndex int
	full       bool // true once we've wrapped around
}

// NewDeliveryTracker creates a tracker with the specified capacity.
// Default is 1000 entries, configurable via MONITOR_DELIVERY_HISTORY env var.
func NewDeliveryTracker(maxEntries int) *DeliveryTracker {
	if maxEntries <= 0 {
		maxEntries = getDefaultCapacity()
	}

	return &DeliveryTracker{
		entries:    make([]LogEntry, maxEntries),
		maxEntries: maxEntries,
		writeIndex: 0,
		full:       false,
	}
}

// getDefaultCapacity returns the configured capacity from environment or 1000.
func getDefaultCapacity() int {
	if env := os.Getenv("MONITOR_DELIVERY_HISTORY"); env != "" {
		if capacity, err := strconv.Atoi(env); err == nil && capacity > 0 {
			return capacity
		}
	}
	return 1000
}

// Record adds a delivery entry to the ring buffer.
// Thread-safe for concurrent access.
func (dt *DeliveryTracker) Record(entry *LogEntry) {
	if entry == nil {
		return
	}

	dt.mu.Lock()
	defer dt.mu.Unlock()

	dt.entries[dt.writeIndex] = *entry
	dt.writeIndex++

	if dt.writeIndex >= dt.maxEntries {
		dt.writeIndex = 0
		dt.full = true
	}
}

// GetRecent returns the last N entries, newest first.
// If n is 0 or exceeds available entries, returns all available entries.
func (dt *DeliveryTracker) GetRecent(n int) []LogEntry {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	count := dt.count()
	if n <= 0 || n > count {
		n = count
	}

	if n == 0 {
		return []LogEntry{}
	}

	result := make([]LogEntry, 0, n)

	// Start from most recent entry and work backwards
	for i := 0; i < n; i++ {
		idx := dt.getIndexFromEnd(i)
		result = append(result, dt.entries[idx])
	}

	return result
}

// GetByQueueID returns all entries matching the specified queue ID.
func (dt *DeliveryTracker) GetByQueueID(queueID string) []LogEntry {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	var result []LogEntry
	count := dt.count()

	for i := 0; i < count; i++ {
		idx := dt.getIndexFromEnd(i)
		if dt.entries[idx].QueueID == queueID {
			result = append(result, dt.entries[idx])
		}
	}

	return result
}

// GetStats returns aggregate statistics over all tracked entries.
func (dt *DeliveryTracker) GetStats() DeliveryStats {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	stats := DeliveryStats{}
	count := dt.count()

	if count == 0 {
		return stats
	}

	var oldestTime, newestTime time.Time

	for i := 0; i < count; i++ {
		idx := dt.getIndex(i)
		entry := dt.entries[idx]

		// Count by status
		switch entry.Status {
		case "delivered":
			stats.Delivered++
		case "deferred":
			stats.Deferred++
		case "bounced":
			stats.Bounced++
		case "expired":
			stats.Expired++
		}

		stats.Total++

		// Track time range
		if oldestTime.IsZero() || entry.Timestamp.Before(oldestTime) {
			oldestTime = entry.Timestamp
		}
		if newestTime.IsZero() || entry.Timestamp.After(newestTime) {
			newestTime = entry.Timestamp
		}
	}

	if !oldestTime.IsZero() && !newestTime.IsZero() {
		stats.Period = newestTime.Sub(oldestTime)
	}

	return stats
}

// count returns the number of valid entries in the buffer.
func (dt *DeliveryTracker) count() int {
	if dt.full {
		return dt.maxEntries
	}
	return dt.writeIndex
}

// getIndex returns the actual slice index for the i-th oldest entry.
func (dt *DeliveryTracker) getIndex(i int) int {
	if !dt.full {
		return i
	}
	return (dt.writeIndex + i) % dt.maxEntries
}

// getIndexFromEnd returns the actual slice index for the i-th newest entry.
// i=0 returns the most recent, i=1 returns second most recent, etc.
func (dt *DeliveryTracker) getIndexFromEnd(i int) int {
	count := dt.count()
	if i >= count {
		i = count - 1
	}
	return dt.getIndex(count - 1 - i)
}
