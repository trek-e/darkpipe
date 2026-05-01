package healtheval

import "testing"

func TestEvaluateHealthy(t *testing.T) {
	m := New(Policy{})
	res := m.Evaluate(Snapshot{HealthStatus: "up", Checks: []Check{{Name: "imap", Status: "ok"}}})
	if res.Status != "healthy" {
		t.Fatalf("want healthy, got %s", res.Status)
	}
}

func TestEvaluateCriticalWins(t *testing.T) {
	m := New(Policy{})
	res := m.Evaluate(Snapshot{HealthStatus: "down", QueueDepth: 100, Checks: []Check{{Name: "imap", Status: "error"}}})
	if res.Status != "unhealthy" {
		t.Fatalf("want unhealthy, got %s", res.Status)
	}
	if len(res.Triggered) == 0 {
		t.Fatal("expected triggered rules")
	}
}

func TestEvaluateDegraded(t *testing.T) {
	m := New(Policy{})
	res := m.Evaluate(Snapshot{HealthStatus: "up", QueueDepth: 60})
	if res.Status != "degraded" {
		t.Fatalf("want degraded, got %s", res.Status)
	}
}
