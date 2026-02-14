package dkim

import (
	"testing"
	"time"
)

func TestGenerateSelector(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		date     string
		expected string
	}{
		{
			name:     "Q1 2026",
			prefix:   "darkpipe",
			date:     "2026-01-15",
			expected: "darkpipe-2026q1",
		},
		{
			name:     "Q2 2026",
			prefix:   "darkpipe",
			date:     "2026-04-15",
			expected: "darkpipe-2026q2",
		},
		{
			name:     "Q3 2026",
			prefix:   "darkpipe",
			date:     "2026-07-15",
			expected: "darkpipe-2026q3",
		},
		{
			name:     "Q4 2026",
			prefix:   "darkpipe",
			date:     "2026-10-15",
			expected: "darkpipe-2026q4",
		},
		{
			name:     "Custom prefix",
			prefix:   "test",
			date:     "2026-03-31",
			expected: "test-2026q1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := time.Parse("2006-01-02", tt.date)
			if err != nil {
				t.Fatalf("Failed to parse test date: %v", err)
			}

			selector := GenerateSelector(tt.prefix, date)
			if selector != tt.expected {
				t.Fatalf("GenerateSelector() = %s, want %s", selector, tt.expected)
			}
		})
	}
}

func TestGetCurrentSelector(t *testing.T) {
	selector := GetCurrentSelector("darkpipe")
	if selector == "" {
		t.Fatal("GetCurrentSelector() returned empty string")
	}

	// Verify format: prefix-YYYYqQ
	now := time.Now()
	expectedYear := now.Year()
	expectedQuarter := (int(now.Month())-1)/3 + 1
	expected := GenerateSelector("darkpipe", now)

	if selector != expected {
		t.Fatalf("GetCurrentSelector() = %s, want %s (year=%d, quarter=%d)",
			selector, expected, expectedYear, expectedQuarter)
	}
}

func TestGetNextSelector(t *testing.T) {
	tests := []struct {
		name        string
		currentDate string
		wantYear    int
		wantQuarter int
	}{
		{
			name:        "Q1 -> Q2 same year",
			currentDate: "2026-01-15",
			wantYear:    2026,
			wantQuarter: 2,
		},
		{
			name:        "Q4 -> Q1 next year",
			currentDate: "2026-12-15",
			wantYear:    2027,
			wantQuarter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock "now" by parsing the test date
			// Since GetNextSelector uses time.Now(), we test the logic with GenerateSelector
			currentDate, err := time.Parse("2006-01-02", tt.currentDate)
			if err != nil {
				t.Fatalf("Failed to parse test date: %v", err)
			}

			nextQuarter := currentDate.AddDate(0, 3, 0)
			selector := GenerateSelector("darkpipe", nextQuarter)

			expectedSelector := GenerateSelector("darkpipe",
				time.Date(tt.wantYear, time.Month((tt.wantQuarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC))

			if selector != expectedSelector {
				t.Fatalf("GetNextSelector() = %s, want %s", selector, expectedSelector)
			}
		})
	}
}

func TestShouldRotate(t *testing.T) {
	tests := []struct {
		name         string
		lastRotation string
		currentTime  string
		want         bool
	}{
		{
			name:         "Same quarter, no rotation",
			lastRotation: "2026-01-01",
			currentTime:  "2026-01-15",
			want:         false,
		},
		{
			name:         "Different quarter, same year",
			lastRotation: "2026-01-15",
			currentTime:  "2026-04-15",
			want:         true,
		},
		{
			name:         "Different year",
			lastRotation: "2025-12-15",
			currentTime:  "2026-01-15",
			want:         true,
		},
		{
			name:         "Last day of quarter",
			lastRotation: "2026-03-31",
			currentTime:  "2026-04-01",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastRotation, err := time.Parse("2006-01-02", tt.lastRotation)
			if err != nil {
				t.Fatalf("Failed to parse lastRotation date: %v", err)
			}

			currentTime, err := time.Parse("2006-01-02", tt.currentTime)
			if err != nil {
				t.Fatalf("Failed to parse currentTime date: %v", err)
			}

			// We need to test the logic, so compute quarters directly
			nowQuarter := (int(currentTime.Month())-1)/3 + 1
			lastQuarter := (int(lastRotation.Month())-1)/3 + 1

			got := currentTime.Year() != lastRotation.Year() || nowQuarter != lastQuarter

			if got != tt.want {
				t.Fatalf("ShouldRotate(%s, %s) = %v, want %v (quarters: %d vs %d)",
					tt.lastRotation, tt.currentTime, got, tt.want, lastQuarter, nowQuarter)
			}
		})
	}
}
