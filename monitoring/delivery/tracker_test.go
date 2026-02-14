package delivery

import (
	"sync"
	"testing"
	"time"
)

func TestNewDeliveryTracker(t *testing.T) {
	tracker := NewDeliveryTracker(100)

	if tracker.maxEntries != 100 {
		t.Errorf("expected maxEntries 100, got %d", tracker.maxEntries)
	}

	if len(tracker.entries) != 100 {
		t.Errorf("expected entries slice length 100, got %d", len(tracker.entries))
	}
}

func TestDeliveryTracker_Record(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	entry := &LogEntry{
		QueueID:   "TEST123",
		To:        "user@example.com",
		Status:    "delivered",
		Timestamp: time.Now(),
	}

	tracker.Record(entry)

	recent := tracker.GetRecent(1)
	if len(recent) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(recent))
	}

	if recent[0].QueueID != "TEST123" {
		t.Errorf("expected queue_id TEST123, got %s", recent[0].QueueID)
	}
}

func TestDeliveryTracker_Record_Nil(t *testing.T) {
	tracker := NewDeliveryTracker(10)
	tracker.Record(nil) // Should not panic

	recent := tracker.GetRecent(10)
	if len(recent) != 0 {
		t.Errorf("expected 0 entries, got %d", len(recent))
	}
}

func TestDeliveryTracker_RingBuffer_Wraparound(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	// Add 15 entries to a 10-capacity tracker
	for i := 0; i < 15; i++ {
		entry := &LogEntry{
			QueueID:   "TEST" + intToStringSimple(i),
			To:        "user@example.com",
			Status:    "delivered",
			Timestamp: time.Now(),
		}
		tracker.Record(entry)
	}

	// Should have only the last 10 entries
	recent := tracker.GetRecent(20) // Ask for more than capacity
	if len(recent) != 10 {
		t.Fatalf("expected 10 entries, got %d", len(recent))
	}

	// Most recent should be TEST14
	if recent[0].QueueID != "TEST14" {
		t.Errorf("expected most recent TEST14, got %s", recent[0].QueueID)
	}

	// Oldest should be TEST5 (entries 0-4 were overwritten)
	if recent[9].QueueID != "TEST5" {
		t.Errorf("expected oldest TEST5, got %s", recent[9].QueueID)
	}
}

func TestDeliveryTracker_GetRecent_NewestFirst(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	// Add entries with increasing queue IDs
	for i := 0; i < 5; i++ {
		entry := &LogEntry{
			QueueID:   "TEST" + intToStringSimple(i),
			Timestamp: time.Now(),
		}
		tracker.Record(entry)
	}

	recent := tracker.GetRecent(3)
	if len(recent) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(recent))
	}

	// Should be in reverse order (newest first)
	if recent[0].QueueID != "TEST4" {
		t.Errorf("expected TEST4 first, got %s", recent[0].QueueID)
	}
	if recent[1].QueueID != "TEST3" {
		t.Errorf("expected TEST3 second, got %s", recent[1].QueueID)
	}
	if recent[2].QueueID != "TEST2" {
		t.Errorf("expected TEST2 third, got %s", recent[2].QueueID)
	}
}

func TestDeliveryTracker_GetRecent_Empty(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	recent := tracker.GetRecent(10)
	if len(recent) != 0 {
		t.Errorf("expected 0 entries, got %d", len(recent))
	}
}

func TestDeliveryTracker_GetRecent_Zero(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	// Add some entries
	tracker.Record(&LogEntry{QueueID: "TEST1", Timestamp: time.Now()})
	tracker.Record(&LogEntry{QueueID: "TEST2", Timestamp: time.Now()})

	// Requesting 0 should return all
	recent := tracker.GetRecent(0)
	if len(recent) != 2 {
		t.Errorf("expected 2 entries when requesting 0, got %d", len(recent))
	}
}

func TestDeliveryTracker_GetByQueueID(t *testing.T) {
	tracker := NewDeliveryTracker(10)

	// Add multiple entries for the same queue ID (multi-recipient)
	tracker.Record(&LogEntry{
		QueueID:   "MULTI1",
		To:        "user1@example.com",
		Timestamp: time.Now(),
	})
	tracker.Record(&LogEntry{
		QueueID:   "OTHER1",
		To:        "other@example.com",
		Timestamp: time.Now(),
	})
	tracker.Record(&LogEntry{
		QueueID:   "MULTI1",
		To:        "user2@example.com",
		Timestamp: time.Now(),
	})

	matches := tracker.GetByQueueID("MULTI1")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	// Verify both recipients are present
	recipients := make(map[string]bool)
	for _, entry := range matches {
		recipients[entry.To] = true
	}

	if !recipients["user1@example.com"] || !recipients["user2@example.com"] {
		t.Error("expected both user1 and user2 recipients")
	}
}

func TestDeliveryTracker_GetByQueueID_NotFound(t *testing.T) {
	tracker := NewDeliveryTracker(10)
	tracker.Record(&LogEntry{QueueID: "TEST1", Timestamp: time.Now()})

	matches := tracker.GetByQueueID("NOTFOUND")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestDeliveryTracker_GetStats(t *testing.T) {
	tracker := NewDeliveryTracker(100)

	// Add various delivery statuses
	tracker.Record(&LogEntry{Status: "delivered", Timestamp: time.Now()})
	tracker.Record(&LogEntry{Status: "delivered", Timestamp: time.Now()})
	tracker.Record(&LogEntry{Status: "deferred", Timestamp: time.Now()})
	tracker.Record(&LogEntry{Status: "bounced", Timestamp: time.Now()})
	tracker.Record(&LogEntry{Status: "expired", Timestamp: time.Now()})

	stats := tracker.GetStats()

	if stats.Delivered != 2 {
		t.Errorf("expected 2 delivered, got %d", stats.Delivered)
	}
	if stats.Deferred != 1 {
		t.Errorf("expected 1 deferred, got %d", stats.Deferred)
	}
	if stats.Bounced != 1 {
		t.Errorf("expected 1 bounced, got %d", stats.Bounced)
	}
	if stats.Expired != 1 {
		t.Errorf("expected 1 expired, got %d", stats.Expired)
	}
	if stats.Total != 5 {
		t.Errorf("expected 5 total, got %d", stats.Total)
	}
}

func TestDeliveryTracker_GetStats_Empty(t *testing.T) {
	tracker := NewDeliveryTracker(100)
	stats := tracker.GetStats()

	if stats.Total != 0 {
		t.Errorf("expected 0 total, got %d", stats.Total)
	}
	if stats.Period != 0 {
		t.Errorf("expected 0 period, got %v", stats.Period)
	}
}

func TestDeliveryTracker_GetStats_Period(t *testing.T) {
	tracker := NewDeliveryTracker(100)

	now := time.Now()
	tracker.Record(&LogEntry{Status: "delivered", Timestamp: now.Add(-10 * time.Minute)})
	tracker.Record(&LogEntry{Status: "delivered", Timestamp: now})

	stats := tracker.GetStats()

	// Period should be approximately 10 minutes
	if stats.Period < 9*time.Minute || stats.Period > 11*time.Minute {
		t.Errorf("expected period ~10 minutes, got %v", stats.Period)
	}
}

func TestDeliveryTracker_ThreadSafety(t *testing.T) {
	tracker := NewDeliveryTracker(1000)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				entry := &LogEntry{
					QueueID:   "TEST" + intToStringSimple(id*100+j),
					Status:    "delivered",
					Timestamp: time.Now(),
				}
				tracker.Record(entry)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = tracker.GetRecent(10)
				_ = tracker.GetStats()
			}
		}()
	}

	wg.Wait()

	// Verify tracker is in valid state
	stats := tracker.GetStats()
	if stats.Total != 1000 {
		t.Errorf("expected 1000 total entries, got %d", stats.Total)
	}
}

func TestGetDeliveryStatus(t *testing.T) {
	tracker := NewDeliveryTracker(100)

	tracker.Record(&LogEntry{Status: "delivered", Timestamp: time.Now()})
	tracker.Record(&LogEntry{Status: "deferred", Timestamp: time.Now()})

	status := GetDeliveryStatus(tracker, 5)

	if status.Stats.Total != 2 {
		t.Errorf("expected 2 total, got %d", status.Stats.Total)
	}

	if len(status.RecentEntries) != 2 {
		t.Errorf("expected 2 recent entries, got %d", len(status.RecentEntries))
	}

	if status.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

// Helper function for tests
func intToStringSimple(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
