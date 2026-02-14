package queue

import (
	"time"
)

// QueueStats represents aggregated Postfix queue statistics.
type QueueStats struct {
	Depth          int           `json:"depth"`           // Total messages in queue
	Active         int           `json:"active"`          // Messages in active delivery
	Deferred       int           `json:"deferred"`        // Messages deferred for retry
	Hold           int           `json:"hold"`            // Messages on hold
	Stuck          int           `json:"stuck"`           // Messages older than threshold
	OldestMessage  time.Time     `json:"oldest_message"`  // Timestamp of oldest message
	Timestamp      time.Time     `json:"timestamp"`       // When stats were collected
	StuckThreshold time.Duration `json:"-"`               // Threshold for stuck detection (not serialized)
}

// QueueSnapshot provides a detailed view of the queue including individual messages.
type QueueSnapshot struct {
	QueueStats         // Embedded stats
	Messages   []QueueMessage `json:"messages"` // Full message list
}

// GetDetailedQueue returns a complete queue snapshot with all message details.
func GetDetailedQueue() (*QueueSnapshot, error) {
	return GetDetailedQueueWithExecutor(defaultExecutor)
}

// GetDetailedQueueWithExecutor returns a queue snapshot using the provided executor.
func GetDetailedQueueWithExecutor(executor PostqueueExecutor) (*QueueSnapshot, error) {
	output, err := executor.Execute()
	if err != nil {
		// Handle empty queue
		if len(output) == 0 {
			return &QueueSnapshot{
				QueueStats: QueueStats{
					Depth:     0,
					Timestamp: time.Now(),
				},
				Messages: []QueueMessage{},
			}, nil
		}
		return nil, err
	}

	messages, err := parsePostqueueJSON(output)
	if err != nil {
		return nil, err
	}

	stats := calculateStats(messages)

	return &QueueSnapshot{
		QueueStats: *stats,
		Messages:   messages,
	}, nil
}
