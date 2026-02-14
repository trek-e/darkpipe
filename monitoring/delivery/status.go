package delivery

import (
	"time"
)

// DeliveryStats represents aggregate delivery statistics.
type DeliveryStats struct {
	Delivered int           `json:"delivered"` // Successfully delivered
	Deferred  int           `json:"deferred"`  // Temporarily deferred
	Bounced   int           `json:"bounced"`   // Permanently bounced
	Expired   int           `json:"expired"`   // TTL exceeded
	Total     int           `json:"total"`     // Total tracked
	Period    time.Duration `json:"period_ms"` // Time span of data
}

// MarshalJSON implements custom JSON marshaling for DeliveryStats
// to convert Period to milliseconds.
func (ds DeliveryStats) MarshalJSON() ([]byte, error) {
	type Alias DeliveryStats
	return []byte(`{"delivered":` + intToString(ds.Delivered) +
		`,"deferred":` + intToString(ds.Deferred) +
		`,"bounced":` + intToString(ds.Bounced) +
		`,"expired":` + intToString(ds.Expired) +
		`,"total":` + intToString(ds.Total) +
		`,"period_ms":` + intToString(int(ds.Period.Milliseconds())) + `}`), nil
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	// Simple int to string conversion
	var result []byte
	isNeg := n < 0
	if isNeg {
		n = -n
	}
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	if isNeg {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}

// DeliveryStatus provides a complete delivery status snapshot.
type DeliveryStatus struct {
	Stats         DeliveryStats `json:"stats"`
	RecentEntries []LogEntry    `json:"recent_entries"`
	Timestamp     time.Time     `json:"timestamp"`
}

// GetDeliveryStatus returns a complete delivery status snapshot.
func GetDeliveryStatus(tracker *DeliveryTracker, recentCount int) *DeliveryStatus {
	return &DeliveryStatus{
		Stats:         tracker.GetStats(),
		RecentEntries: tracker.GetRecent(recentCount),
		Timestamp:     time.Now(),
	}
}
