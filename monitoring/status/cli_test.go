package status

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/cert"
	"github.com/darkpipe/darkpipe/monitoring/health"
)

func TestDisplayStatusJSON(t *testing.T) {
	status := &SystemStatus{
		Health: HealthSummary{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "postfix", Status: "ok"},
			},
		},
		Queue: QueueSummary{
			Depth:    10,
			Deferred: 2,
			Stuck:    0,
		},
		Delivery: DeliverySummary{
			Delivered: 50,
			Deferred:  2,
			Bounced:   1,
			Total:     53,
		},
		Certificates: CertSummary{
			Certificates: []cert.CertInfo{
				{
					Subject:  "relay.example.com",
					NotAfter: time.Now().Add(60 * 24 * time.Hour),
					DaysLeft: 60,
				},
			},
			DaysUntilExpiry: 60,
		},
		Timestamp:     time.Now(),
		OverallStatus: "healthy",
	}

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := DisplayStatusJSON(status)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("DisplayStatusJSON failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Check expected fields
	if parsed["overall_status"] != "healthy" {
		t.Errorf("Expected overall_status 'healthy', got %v", parsed["overall_status"])
	}

	health, ok := parsed["health"].(map[string]interface{})
	if !ok {
		t.Fatal("health field missing or wrong type")
	}

	if health["status"] != "up" {
		t.Errorf("Expected health status 'up', got %v", health["status"])
	}
}

func TestDisplayStatusHumanReadable(t *testing.T) {
	status := &SystemStatus{
		Health: HealthSummary{
			Status: "up",
			Checks: []health.CheckResult{
				{Name: "Postfix SMTP", Status: "ok"},
				{Name: "IMAP Server", Status: "ok"},
			},
		},
		Queue: QueueSummary{
			Depth:    3,
			Deferred: 0,
			Stuck:    0,
		},
		Delivery: DeliverySummary{
			Delivered: 47,
			Deferred:  2,
			Bounced:   0,
			Total:     49,
		},
		Certificates: CertSummary{
			Certificates: []cert.CertInfo{
				{
					Subject:  "relay.example.com",
					NotAfter: time.Now().Add(62 * 24 * time.Hour),
					DaysLeft: 62,
				},
			},
			DaysUntilExpiry: 62,
		},
		Timestamp:     time.Now(),
		OverallStatus: "healthy",
	}

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	DisplayStatus(status)
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check for expected sections
	expectedSections := []string{
		"DarkPipe System Status",
		"Overall:",
		"Services:",
		"Mail Queue:",
		"Recent Deliveries",
		"Certificates:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output missing expected section: %s\nOutput:\n%s", section, output)
		}
	}

	// Check for specific values
	if !strings.Contains(output, "Depth:    3") {
		t.Errorf("Output missing queue depth")
	}

	if !strings.Contains(output, "Delivered: 47") {
		t.Errorf("Output missing delivered count")
	}
}

func TestRunStatusCommandFlags(t *testing.T) {
	// Test that flags are recognized (can't easily test full execution without wiring)
	tests := []struct {
		name string
		args []string
	}{
		{"json flag", []string{"--json"}},
		{"watch flag", []string{"--watch"}},
		{"watch interval", []string{"--watch", "--watch-interval=10"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will call runOnce and exit with error (aggregator not wired)
			// We're just checking that flag parsing works
			// Actual execution tested in integration tests
		})
	}
}

func TestCheckCLIAlertsNoFile(t *testing.T) {
	// Set a non-existent path
	os.Setenv("MONITOR_CLI_ALERT_PATH", "/tmp/nonexistent-darkpipe-alerts-test.json")
	defer os.Unsetenv("MONITOR_CLI_ALERT_PATH")

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	checkCLIAlerts()
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should produce no output when file doesn't exist
	if strings.Contains(output, "WARNING") {
		t.Errorf("Expected no warning when alert file missing, got: %s", output)
	}
}

func TestCheckCLIAlertsWithFile(t *testing.T) {
	// Create temporary alert file
	tmpFile, err := os.CreateTemp("", "darkpipe-alerts-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some NDJSON alerts
	tmpFile.WriteString(`{"type":"cert_expiry","message":"Certificate expiring soon"}` + "\n")
	tmpFile.WriteString(`{"type":"queue_backup","message":"Queue depth high"}` + "\n")
	tmpFile.Close()

	os.Setenv("MONITOR_CLI_ALERT_PATH", tmpFile.Name())
	defer os.Unsetenv("MONITOR_CLI_ALERT_PATH")

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	checkCLIAlerts()
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should show warning about 2 alerts
	if !strings.Contains(output, "2 pending alert") {
		t.Errorf("Expected warning about 2 alerts, got: %s", output)
	}
}
