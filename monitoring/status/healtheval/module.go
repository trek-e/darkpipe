// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package healtheval

type Check struct {
	Name    string
	Status  string
	Message string
}

type Snapshot struct {
	HealthStatus       string
	Checks             []Check
	QueueDepth         int
	QueueStuck         int
	Delivered          int
	Deferred           int
	Bounced            int
	Total              int
	DaysUntilCertExpiry int
}

type Policy struct {
	QueueDepthWarn      int
	QueueDepthCritical  int
	CertDaysWarn        int
	CertDaysCritical    int
	BounceRateCritical  float64
}

type TriggeredRule struct {
	ID       string `json:"id"`
	Severity string `json:"severity"` // degraded|unhealthy
	Reason   string `json:"reason"`
}

type Evaluation struct {
	Status    string          `json:"status"`
	Reasons   []string        `json:"reasons"`
	Triggered []TriggeredRule `json:"triggered"`
}

type Module interface { Evaluate(s Snapshot) Evaluation }

type DefaultModule struct{ policy Policy }

func New(policy Policy) Module {
	if policy.QueueDepthWarn == 0 { policy.QueueDepthWarn = 50 }
	if policy.QueueDepthCritical == 0 { policy.QueueDepthCritical = 200 }
	if policy.CertDaysWarn == 0 { policy.CertDaysWarn = 14 }
	if policy.CertDaysCritical == 0 { policy.CertDaysCritical = 7 }
	if policy.BounceRateCritical == 0 { policy.BounceRateCritical = 0.5 }
	return &DefaultModule{policy: policy}
}

func (m *DefaultModule) Evaluate(s Snapshot) Evaluation {
	out := Evaluation{Status: "healthy"}
	add := func(id, sev, reason string) {
		out.Triggered = append(out.Triggered, TriggeredRule{ID: id, Severity: sev, Reason: reason})
		out.Reasons = append(out.Reasons, reason)
		if sev == "unhealthy" {
			out.Status = "unhealthy"
		} else if sev == "degraded" && out.Status == "healthy" {
			out.Status = "degraded"
		}
	}

	if s.HealthStatus == "down" { add("health.down", "unhealthy", "health status is down") }
	if s.QueueStuck > 0 { add("queue.stuck", "unhealthy", "queue has stuck messages") }
	if s.QueueDepth > m.policy.QueueDepthCritical { add("queue.depth.critical", "unhealthy", "queue depth above critical threshold") }
	if s.Total > 0 {
		bounce := float64(s.Bounced) / float64(s.Total)
		if bounce > m.policy.BounceRateCritical { add("delivery.bounce.critical", "unhealthy", "bounce rate above critical threshold") }
	}
	if s.DaysUntilCertExpiry > 0 && s.DaysUntilCertExpiry <= m.policy.CertDaysCritical {
		add("cert.expiry.critical", "unhealthy", "certificate expiry in critical window")
	}

	if s.QueueDepth > m.policy.QueueDepthWarn { add("queue.depth.warn", "degraded", "queue depth above warning threshold") }
	if s.DaysUntilCertExpiry > 0 && s.DaysUntilCertExpiry <= m.policy.CertDaysWarn {
		add("cert.expiry.warn", "degraded", "certificate expiry in warning window")
	}
	for _, c := range s.Checks {
		if c.Status != "ok" {
			add("health.check.warn", "degraded", "one or more health checks not ok")
			break
		}
	}

	return out
}
