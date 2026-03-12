# Phase 09: Monitoring & Observability - Research

**Researched:** 2026-02-14
**Domain:** Mail server monitoring, health checks, certificate lifecycle automation
**Confidence:** HIGH

## Summary

Phase 09 builds a comprehensive monitoring and observability layer on top of the existing DarkPipe infrastructure. The phase focuses on four key visibility areas: mail queue health (depth, stuck messages), delivery status tracking (sent/deferred/bounced), certificate lifecycle management (expiry monitoring, automated rotation), and container health checks (deep readiness probes for Docker and external uptime services).

The research reveals strong ecosystem support for all required capabilities. Postfix provides native JSON queue output (postqueue -j), Go's crypto/x509 package handles certificate expiry inspection without external dependencies, Docker health checks support both liveness and readiness patterns, and external push-based monitoring (Healthchecks.io, UptimeRobot) avoids exposing additional inbound ports. Let's Encrypt's timeline for moving to 45-day certificates (opt-in May 2026, default February 2028) validates the user's decision to renew at 2/3 lifetime and design for shorter certificate lifetimes from day one.

**Primary recommendation:** Build monitoring as native Go packages integrated into existing relay and mail server components, expose health data via JSON HTTP endpoints, use systemd path units for certificate rotation automation, and implement push-based external monitoring for security-first uptime tracking.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Health visibility & status display:**
- Both CLI (`darkpipe status`) and web dashboard for system health
- Four metric categories: mail queue depth & stuck messages, recent delivery status, certificate expiry countdown, tunnel/transport health
- JSON output via `--json` flag for scripting and Home Assistant/monitoring tool integration
- CLI for power users, web dashboard for household members

**Alert & notification behavior:**
- All delivery methods: email to admin, webhook (HTTP POST), CLI warning on next command
- All four trigger conditions: certificate expiry approaching, queue backup threshold, delivery failure spike, transport tunnel down
- Rate-limit per alert type (same alert type at most once per hour) to prevent notification storms during extended outages
- Certificate alerts at 14 days and 7 days before expiry (CERT-04 requirement)

**Certificate lifecycle management:**
- Renew at 2/3 of certificate lifetime (future-proof for Let's Encrypt moving from 90-day to 45-day certs through 2026-2028)
- For 90-day LE certs: renew at 60 days. For 45-day: renew at 30 days. For step-ca internal: configurable
- DKIM key rotation automated quarterly (matches Phase 4 selector format {prefix}-{YYYY}q{Q})
- Retry with exponential backoff on renewal failure (3 retries), then alert admin. Keep using old cert until actual expiry
- Let's Encrypt timeline awareness: 90-day default now, 45-day opt-in May 2026, 64-day default Feb 2027, 45-day default Feb 2028

**Container health checks & endpoints:**
- Deep readiness checks (actual service health: can Postfix accept mail? Is IMAP responding? Is tunnel connected?)
- Both per-container healthcheck endpoints (for Docker) AND unified aggregation endpoint (for user-facing status)
- Public health endpoint via Caddy with Basic Auth for remote monitoring
- Push-based pings to external uptime services (Healthchecks.io, UptimeRobot) via outbound HTTP — no inbound exposure needed

### Claude's Discretion

- Web dashboard location (profile server /status vs separate container)
- CLI auto-refresh behavior (one-shot vs watch mode)
- Per-user vs system-wide delivery stats
- Inbound vs outbound delivery tracking scope
- Delivery history retention period
- Alert severity levels (warn/critical vs single level)
- Certificate rotation service interruption strategy (hot reload vs brief restart)
- Health check aggregation architecture

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope

</user_constraints>

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib crypto/x509 | (stdlib) | Certificate expiry inspection | Built-in certificate parsing, NotAfter field as time.Time for expiry checks, zero external deps |
| Go stdlib net/http | (stdlib) | Health check HTTP endpoints | Standard server, JSON response encoding, /health and /ready endpoint patterns |
| github.com/cenkalti/backoff/v4 | v4.x | Exponential backoff for cert renewal retries | Already a project dependency (from Phase 01-02 mTLS), clean Retry function with permanent error handling |
| gopkg.in/yaml.v3 | v3.x | Configuration file parsing | Already in project (Phase 07-02), standard for DarkPipe config files |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/RussellLuo/slidingwindow | Latest | Alert rate limiting (1-hour dedup window) | Distributed sliding window counters for per-alert-type rate limiting |
| systemd path units | (system) | Certificate file change monitoring | Native systemd integration, triggers reload when cert files modified |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom health check lib | github.com/alexliesenfeld/health | More features (versioning, caching), but adds dependency. Stdlib net/http sufficient for DarkPipe's needs |
| Custom queue parser | github.com/alexjurkiewicz/apq | Postfix 3.1+ has native JSON output (postqueue -j), parsing directly avoids external tool dependency |
| Grafana/Prometheus | Embedded web dashboard | Grafana is overkill for 4 metrics + certificate countdown. Lightweight Go template or HTMX dashboard keeps UX-02 (minimal footprint) |

**Installation:**

```bash
# Already in project (no new external deps needed):
# - cenkalti/backoff/v4 (Phase 01-02)
# - gopkg.in/yaml.v3 (Phase 07-02)

# Add sliding window rate limiter if using distributed approach:
go get github.com/RussellLuo/slidingwindow
```

---

## Architecture Patterns

### Recommended Project Structure

```
monitoring/
├── health/              # Health check aggregator
│   ├── checker.go       # Unified health check interface
│   ├── postfix.go       # Postfix SMTP readiness check
│   ├── imap.go          # IMAP server readiness check
│   ├── tunnel.go        # WireGuard/mTLS transport health
│   └── server.go        # HTTP /health and /ready endpoints
├── queue/               # Mail queue monitoring
│   ├── mailq.go         # Postfix queue depth via postqueue -j
│   └── stats.go         # Queue statistics (depth, deferred, stuck)
├── delivery/            # Delivery status tracking
│   ├── parser.go        # Postfix log parsing (sent/deferred/bounced)
│   ├── tracker.go       # Recent delivery history with retention
│   └── status.go        # Per-message delivery status lookup
├── cert/                # Certificate lifecycle management
│   ├── watcher.go       # Certificate expiry monitoring
│   ├── rotator.go       # Automated renewal with backoff retry
│   ├── dkim.go          # Quarterly DKIM key rotation
│   └── reload.go        # Service reload orchestration (Postfix/IMAP)
├── alert/               # Alert and notification system
│   ├── notifier.go      # Multi-channel notification (reuse cloud-relay/relay/notify)
│   ├── ratelimit.go     # Per-alert-type deduplication (1-hour window)
│   └── triggers.go      # Alert conditions (cert expiry, queue backup, etc.)
└── status/              # Status aggregation and CLI/dashboard
    ├── aggregator.go    # Collect health/queue/cert/delivery data
    ├── cli.go           # darkpipe status command with --json flag
    └── dashboard.go     # Web dashboard handler (embedded templates)
```

### Pattern 1: Health Check Separation (Liveness vs Readiness)

**What:** Docker healthchecks expose two distinct endpoints: `/health/live` (process running) and `/health/ready` (dependencies reachable). Liveness checks are cheap and only verify the process is alive. Readiness checks are deep and verify critical services (Postfix SMTP port 25, IMAP port 993, WireGuard tunnel connected).

**When to use:** Always. Kubernetes and Docker Compose both support this pattern. Prevents false positives from transient dependency hiccups.

**Example:**

```go
// Source: Docker health check best practices (HIGH confidence)
// https://oneuptime.com/blog/post/2026-01-30-docker-health-check-best-practices/view

package health

import (
	"context"
	"net"
	"time"
)

// Liveness check - process alive?
func (c *Checker) Liveness(ctx context.Context) error {
	// Cheap check: can we respond?
	return nil // If this runs, process is alive
}

// Readiness check - can we serve traffic?
func (c *Checker) Readiness(ctx context.Context) error {
	// Deep check: are dependencies healthy?
	if err := c.CheckPostfix(ctx); err != nil {
		return err
	}
	if err := c.CheckIMAP(ctx); err != nil {
		return err
	}
	if err := c.CheckTunnel(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Checker) CheckPostfix(ctx context.Context) error {
	// Attempt TCP connection to Postfix port 25
	conn, err := net.DialTimeout("tcp", "localhost:25", 2*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
```

### Pattern 2: Postfix Queue Monitoring via JSON Output

**What:** Postfix 3.1+ supports `postqueue -j` which outputs queue data as newline-delimited JSON. Parse this to extract queue depth, deferred message count, and stuck message detection.

**When to use:** All Postfix queue monitoring (MON-01). Avoids parsing human-readable mailq output.

**Example:**

```go
// Source: Postfix postqueue JSON parsing (HIGH confidence)
// https://www.postfix.org/postqueue.1.html
// https://github.com/alexjurkiewicz/apq

package queue

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"time"
)

type QueueMessage struct {
	QueueID      string    `json:"queue_id"`
	QueueName    string    `json:"queue_name"` // "active", "deferred", "incoming"
	ArrivalTime  time.Time `json:"arrival_time"`
	MessageSize  int64     `json:"message_size"`
	Recipients   []Recipient `json:"recipients"`
}

type Recipient struct {
	Address      string `json:"address"`
	DelayReason  string `json:"delay_reason,omitempty"`
}

func GetQueueStats() (depth int, deferred int, stuck int, err error) {
	cmd := exec.Command("postqueue", "-j")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, 0, 0, err
	}

	if err := cmd.Start(); err != nil {
		return 0, 0, 0, err
	}

	scanner := bufio.NewScanner(stdout)
	now := time.Now()
	stuckThreshold := 24 * time.Hour // Messages older than 24h = stuck

	for scanner.Scan() {
		var msg QueueMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue // Skip malformed lines
		}

		depth++
		if msg.QueueName == "deferred" {
			deferred++
		}
		if now.Sub(msg.ArrivalTime) > stuckThreshold {
			stuck++
		}
	}

	if err := cmd.Wait(); err != nil {
		return 0, 0, 0, err
	}

	return depth, deferred, stuck, nil
}
```

### Pattern 3: Certificate Expiry Monitoring with crypto/x509

**What:** Use Go's stdlib crypto/x509 to parse certificate files and check NotAfter field. No external dependencies required.

**When to use:** All certificate expiry monitoring (CERT-04). Works for Let's Encrypt public certs and step-ca internal certs.

**Example:**

```go
// Source: Go crypto/x509 certificate parsing (HIGH confidence)
// https://pkg.go.dev/crypto/x509

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

func GetCertExpiry(certPath string) (time.Time, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("no PEM block found")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}

	return cert.NotAfter, nil
}

func CheckExpiry(certPath string, warnThreshold time.Duration) (daysLeft int, shouldAlert bool, err error) {
	expiry, err := GetCertExpiry(certPath)
	if err != nil {
		return 0, false, err
	}

	timeLeft := time.Until(expiry)
	daysLeft = int(timeLeft.Hours() / 24)
	shouldAlert = timeLeft <= warnThreshold

	return daysLeft, shouldAlert, nil
}
```

### Pattern 4: Certificate Rotation with Exponential Backoff Retry

**What:** Use cenkalti/backoff/v4 (already a project dependency) to retry certificate renewal with exponential backoff (3 retries). Wrap permanent errors (e.g., ACME account issues) to stop retrying.

**When to use:** All certificate renewal automation (CERT-03). Matches existing project pattern from Phase 01-02 mTLS reconnection.

**Example:**

```go
// Source: cenkalti/backoff Retry function (HIGH confidence)
// https://github.com/cenkalti/backoff/blob/v5/README.md

package cert

import (
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"time"
)

func RenewCertificateWithRetry(certPath string) error {
	operation := func() error {
		// Attempt renewal (shell out to certbot or step-ca)
		err := renewCertificate(certPath)
		if err != nil {
			// Check if error is permanent (e.g., ACME account invalid)
			if isPermanentError(err) {
				return backoff.Permanent(err) // Don't retry
			}
			return err // Transient error, will retry
		}
		return nil
	}

	// Exponential backoff with 3 retries
	bo := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
	err := backoff.Retry(operation, bo)
	if err != nil {
		return fmt.Errorf("certificate renewal failed after retries: %w", err)
	}

	return nil
}

func isPermanentError(err error) bool {
	// Check for errors that won't be fixed by retrying
	// e.g., ACME account issues, DNS validation failures
	return false // Implement based on certbot/step-ca error types
}
```

### Pattern 5: Systemd Path Unit for Certificate Change Detection

**What:** Use systemd path units to monitor certificate files for changes (via PathChanged=). When certbot renews a certificate, systemd automatically triggers a reload service.

**When to use:** Certificate rotation automation on cloud relay and home device. Avoids polling, integrates with existing systemd setup.

**Example:**

```systemd
# Source: Systemd certificate reload automation (HIGH confidence)
# https://ibug.io/blog/2024/03/reload-ssl-cert-with-systemd/

# /etc/systemd/system/cert-reload.path
[Unit]
Description=Monitor certificate file changes
After=network.target

[Path]
PathChanged=/etc/letsencrypt/live/relay.example.com/fullchain.pem
PathChanged=/etc/letsencrypt/live/relay.example.com/privkey.pem

[Install]
WantedBy=multi-user.target

# /etc/systemd/system/cert-reload.service
[Unit]
Description=Reload services after certificate change

[Service]
Type=oneshot
ExecStart=/usr/bin/systemctl reload postfix.service
ExecStart=/usr/bin/systemctl reload caddy.service
```

### Pattern 6: Alert Rate Limiting (Sliding Window Deduplication)

**What:** Implement per-alert-type rate limiting with a 1-hour sliding window. Same alert type (e.g., "certificate expiry") can only fire once per hour to prevent notification storms during extended outages.

**When to use:** All alert notifications. Reuses existing notification system from Phase 02-02.

**Example:**

```go
// Source: Sliding window rate limiting pattern (MEDIUM confidence)
// https://github.com/RussellLuo/slidingwindow

package alert

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	lastSent    map[string]time.Time // alert type -> last sent timestamp
	dedupWindow time.Duration
}

func NewRateLimiter(dedupWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		lastSent:    make(map[string]time.Time),
		dedupWindow: dedupWindow,
	}
}

func (r *RateLimiter) ShouldSend(alertType string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	last, exists := r.lastSent[alertType]
	if !exists {
		r.lastSent[alertType] = time.Now()
		return true
	}

	if time.Since(last) >= r.dedupWindow {
		r.lastSent[alertType] = time.Now()
		return true
	}

	return false // Deduplicated
}
```

### Pattern 7: Push-Based External Monitoring (Healthchecks.io)

**What:** Send outbound HTTP pings to external uptime services (Healthchecks.io, UptimeRobot) on a schedule. Uses Dead Man's Switch pattern: if pings stop, service is down.

**When to use:** External uptime monitoring without exposing inbound ports (MON-03). Security-first approach.

**Example:**

```go
// Source: Healthchecks.io push monitoring pattern (HIGH confidence)
// https://healthchecks.io/docs/

package monitoring

import (
	"context"
	"net/http"
	"time"
)

type HealthchecksPinger struct {
	checkURL string
	client   *http.Client
}

func NewHealthchecksPinger(checkURL string) *HealthchecksPinger {
	return &HealthchecksPinger{
		checkURL: checkURL,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *HealthchecksPinger) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.checkURL, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Run pings every 5 minutes (configurable)
func (p *HealthchecksPinger) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.Ping(ctx) // Ignore errors, Healthchecks.io alerts on missed pings
		}
	}
}
```

### Anti-Patterns to Avoid

- **Polling certificate files in a loop:** Use systemd path units (PathChanged=) instead. Avoids busy-waiting and integrates with existing systemd setup.
- **Parsing human-readable mailq output:** Use `postqueue -j` (JSON output) instead. Fragile to parse, breaks on Postfix output changes.
- **Database for delivery history:** Use in-memory ring buffer with configurable retention (default 1000 recent messages). Avoids SQLite dependency for simple use case.
- **Restarting services for certificate rotation:** Use `postfix reload` (SIGHUP) and Caddy's automatic reload. Avoids brief service interruption.
- **Pull-based external monitoring:** Use push-based pings (Healthchecks.io) instead. Avoids exposing inbound HTTP endpoints.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Exponential backoff | Custom retry loop with sleep() | github.com/cenkalti/backoff/v4 | Already a project dependency, handles jitter, max retries, permanent errors. Custom implementations miss edge cases (e.g., retry budget exhaustion). |
| Certificate expiry parsing | Manual ASN.1 parsing | Go stdlib crypto/x509 | Certificate parsing is complex (multiple formats, extensions, validity chains). Stdlib handles all edge cases. |
| Rate limiting / deduplication | Custom timestamp map | Sliding window algorithm (RussellLuo/slidingwindow or simple map) | Distributed rate limiting requires careful window management. Simple map sufficient for single-process use case. |
| IMAP connection test | Custom socket code | net.Dial with timeout | IMAP protocol has STARTTLS, CAPABILITY negotiation. Simple TCP connection test is sufficient for readiness check. |
| Postfix queue parsing | Regex on mailq output | postqueue -j (native JSON) | Postfix 3.1+ provides structured output. Parsing human-readable output is fragile and breaks on version changes. |

**Key insight:** Mail server monitoring has well-established patterns (queue JSON output, TLS certificate inspection, service reload signals). Leveraging existing tools (Postfix JSON, systemd path units, stdlib crypto/x509) avoids reinventing complex parsing and reduces maintenance burden.

---

## Common Pitfalls

### Pitfall 1: False Positive Health Checks

**What goes wrong:** Health check marks service unhealthy during transient dependency hiccups (e.g., network blip to IMAP server), causing unnecessary container restarts.

**Why it happens:** Liveness and readiness checks are conflated. Single health check endpoint does both "is process alive?" and "are dependencies reachable?"

**How to avoid:** Separate `/health/live` (cheap, process alive) and `/health/ready` (deep, dependencies reachable). Docker uses liveness for restart decisions, readiness for load balancer routing. Set appropriate timeoutSeconds (2-5s) and failureThreshold (3 consecutive failures).

**Warning signs:** Container restart logs show health check failures during normal operation. Monitoring dashboards show brief "unhealthy" states during transient errors.

### Pitfall 2: Certificate Rotation Causes Brief Service Interruption

**What goes wrong:** Certificate renewal triggers full service restart instead of reload, causing brief email delivery interruption (TCP connections dropped).

**Why it happens:** Assuming services require restart to pick up new certificates. Many services (Postfix, Caddy) support hot reload via SIGHUP.

**How to avoid:** Use `postfix reload` (sends SIGHUP) instead of `systemctl restart postfix`. Caddy automatically reloads certificates when files change. Test that reload actually picks up new certs (some services cache cert paths).

**Warning signs:** Brief SMTP connection failures during certbot renewal. Postfix logs show service restart instead of reload.

### Pitfall 3: Let's Encrypt Rate Limits Hit During Testing

**What goes wrong:** Testing certificate renewal automation exhausts Let's Encrypt rate limits (50 certificates per registered domain per week). Production domains become unable to renew.

**Why it happens:** Testing renewal logic against production ACME endpoint instead of staging.

**How to avoid:** Use Let's Encrypt staging environment during development (`--staging` flag for certbot, `ca-url=https://acme-staging-v02.api.letsencrypt.org/directory` for step-ca). Staging has higher rate limits and issues untrusted certificates (safe for testing).

**Warning signs:** certbot renewal fails with "too many certificates" error. Let's Encrypt API returns HTTP 429 (rate limited).

### Pitfall 4: Alert Notification Storms During Outages

**What goes wrong:** Same alert fires repeatedly during extended outage (e.g., every 5 minutes for 6 hours = 72 emails). User email inbox flooded, real alerts lost in noise.

**Why it happens:** No rate limiting or deduplication. Each health check failure triggers immediate notification.

**How to avoid:** Implement per-alert-type deduplication with 1-hour window (user-specified requirement). Alert fires once, then suppressed for 1 hour even if condition persists. Log suppressed alerts for visibility.

**Warning signs:** Multiple identical alert emails within short time window. Alert notification count >> actual distinct issues.

### Pitfall 5: Postfix Queue JSON Parsing Fails on Older Versions

**What goes wrong:** `postqueue -j` returns "invalid option" error. Monitoring breaks on Postfix versions < 3.1.

**Why it happens:** JSON output added in Postfix 3.1 (released 2016). Older Postfix installations don't support `-j` flag.

**How to avoid:** Check Postfix version before using JSON output. Fall back to parsing `postqueue -p` (human-readable) if < 3.1. Docker images should use Postfix 3.1+ (2026 standard).

**Warning signs:** `postqueue -j` command fails with "unknown option". Error logs show JSON parsing failures.

### Pitfall 6: DKIM Key Rotation Breaks Email Authentication

**What goes wrong:** New DKIM key generated and DNS updated, but old emails in queue still signed with old key. Recipients see DKIM verification failures after rotation.

**Why it happens:** Immediate DNS record replacement. Old key removed before queued messages are delivered.

**How to avoid:** Dual-key overlap during rotation: (1) Generate new key with new selector (e.g., 2026q1 -> 2026q2), (2) Publish both keys in DNS, (3) Update signing to use new key, (4) Wait 7 days for old queue to drain, (5) Remove old key from DNS.

**Warning signs:** DMARC reports show DKIM failures after rotation. Recipients mark email as spam due to authentication failure.

### Pitfall 7: Certificate Expiry Alerts Too Late

**What goes wrong:** Alert fires 7 days before expiry, but renewal fails. Not enough time to debug before certificate expires.

**Why it happens:** Single alert threshold. No early warning for manual intervention.

**How to avoid:** Two-tier alerting: 14 days (early warning, "renewal should happen soon") and 7 days (urgent, "renewal hasn't happened, manual intervention needed"). User requirement specifies both thresholds.

**Warning signs:** Certificate expires before manual intervention possible. No advance warning of renewal failures.

---

## Code Examples

Verified patterns from official sources:

### Health Check HTTP Endpoint with Liveness and Readiness

```go
// Source: Docker health check best practices (HIGH confidence)
// https://oneuptime.com/blog/post/2026-01-30-docker-health-check-best-practices/view

package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthStatus struct {
	Status    string            `json:"status"` // "up" or "down"
	Checks    map[string]string `json:"checks,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

func healthLiveHandler(w http.ResponseWriter, r *http.Request) {
	// Liveness: is process alive?
	status := HealthStatus{
		Status:    "up",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/health+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func healthReadyHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Readiness: can we serve traffic?
		status := HealthStatus{
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		}

		// Check Postfix
		if err := checker.CheckPostfix(r.Context()); err != nil {
			status.Status = "down"
			status.Checks["postfix"] = err.Error()
		} else {
			status.Checks["postfix"] = "ok"
		}

		// Check IMAP
		if err := checker.CheckIMAP(r.Context()); err != nil {
			status.Status = "down"
			status.Checks["imap"] = err.Error()
		} else {
			status.Checks["imap"] = "ok"
		}

		// Check tunnel
		if err := checker.CheckTunnel(r.Context()); err != nil {
			status.Status = "down"
			status.Checks["tunnel"] = err.Error()
		} else {
			status.Checks["tunnel"] = "ok"
		}

		if status.Status == "" {
			status.Status = "up"
		}

		code := http.StatusOK
		if status.Status == "down" {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/health+json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(status)
	}
}

func main() {
	checker := NewChecker()

	http.HandleFunc("/health/live", healthLiveHandler)
	http.HandleFunc("/health/ready", healthReadyHandler(checker))

	http.ListenAndServe(":8080", nil)
}
```

### Postfix Queue Monitoring via JSON

```go
// Source: Postfix postqueue JSON output (HIGH confidence)
// https://www.postfix.org/postqueue.1.html

package queue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type QueueStats struct {
	Depth     int       `json:"depth"`      // Total messages in queue
	Deferred  int       `json:"deferred"`   // Messages in deferred queue
	Stuck     int       `json:"stuck"`      // Messages older than 24h
	Timestamp time.Time `json:"timestamp"`
}

func GetQueueStats() (*QueueStats, error) {
	cmd := exec.Command("postqueue", "-j")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	stats := &QueueStats{Timestamp: time.Now()}
	scanner := bufio.NewScanner(stdout)
	now := time.Now()
	stuckThreshold := 24 * time.Hour

	for scanner.Scan() {
		var msg struct {
			QueueName   string    `json:"queue_name"`
			ArrivalTime time.Time `json:"arrival_time"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue // Skip malformed lines
		}

		stats.Depth++

		if msg.QueueName == "deferred" {
			stats.Deferred++
		}

		if now.Sub(msg.ArrivalTime) > stuckThreshold {
			stats.Stuck++
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	return stats, nil
}
```

### Certificate Expiry Check

```go
// Source: Go crypto/x509 certificate parsing (HIGH confidence)
// https://pkg.go.dev/crypto/x509

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

type CertificateStatus struct {
	Path       string    `json:"path"`
	NotBefore  time.Time `json:"not_before"`
	NotAfter   time.Time `json:"not_after"`
	DaysLeft   int       `json:"days_left"`
	ShouldWarn bool      `json:"should_warn"`
	Subject    string    `json:"subject"`
}

func CheckCertificate(certPath string, warnDays int) (*CertificateStatus, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read cert: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", certPath)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert: %w", err)
	}

	timeLeft := time.Until(cert.NotAfter)
	daysLeft := int(timeLeft.Hours() / 24)

	status := &CertificateStatus{
		Path:       certPath,
		NotBefore:  cert.NotBefore,
		NotAfter:   cert.NotAfter,
		DaysLeft:   daysLeft,
		ShouldWarn: daysLeft <= warnDays,
		Subject:    cert.Subject.String(),
	}

	return status, nil
}
```

### Certificate Renewal with Exponential Backoff

```go
// Source: cenkalti/backoff Retry function (HIGH confidence)
// https://github.com/cenkalti/backoff

package cert

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func RenewCertificate(certName string, dryRun bool) error {
	operation := func() error {
		args := []string{"renew", "--cert-name", certName}
		if dryRun {
			args = append(args, "--dry-run")
		}

		cmd := exec.Command("certbot", args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			// Check if error is permanent
			if isPermanentError(output) {
				return backoff.Permanent(fmt.Errorf("certbot: %w\nOutput: %s", err, output))
			}
			return fmt.Errorf("certbot: %w\nOutput: %s", err, output)
		}

		return nil
	}

	// Exponential backoff with 3 retries
	bo := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)

	err := backoff.Retry(operation, bo)
	if err != nil {
		return fmt.Errorf("renewal failed after retries: %w", err)
	}

	return nil
}

func isPermanentError(output []byte) bool {
	// Certbot returns specific errors for permanent failures
	// e.g., "Account not found", "DNS validation failed"
	// Implement based on certbot error message patterns
	return false
}
```

### DKIM Key Rotation with Quarterly Selector

```go
// Source: DKIM key rotation best practices (MEDIUM confidence)
// https://www.m3aawg.org/DKIMKeyRotation

package cert

import (
	"fmt"
	"time"
)

// GenerateDKIMSelector returns selector in format: prefix-YYYYqQ
// Example: darkpipe-2026q1
func GenerateDKIMSelector(prefix string, t time.Time) string {
	year := t.Year()
	quarter := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%s-%dq%d", prefix, year, quarter)
}

// ShouldRotateDKIM checks if quarterly rotation is due
func ShouldRotateDKIM(lastRotation time.Time) bool {
	now := time.Now()

	// Calculate current quarter
	nowYear := now.Year()
	nowQuarter := (int(now.Month())-1)/3 + 1

	// Calculate last rotation quarter
	lastYear := lastRotation.Year()
	lastQuarter := (int(lastRotation.Month())-1)/3 + 1

	// Rotate if quarter has changed
	return nowYear != lastYear || nowQuarter != lastQuarter
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 90-day LE certs | 45-day LE certs (opt-in May 2026, default Feb 2028) | 2025-2028 transition | Renewal automation must handle shorter lifetimes. User decision to renew at 2/3 lifetime is future-proof. |
| Parsing mailq text output | postqueue -j (JSON) | Postfix 3.1 (2016) | Structured queue data eliminates fragile regex parsing. All modern Postfix installations support JSON. |
| Pull-based health checks (external service polls endpoint) | Push-based health checks (service pings external) | 2020s security shift | Avoids exposing inbound HTTP endpoints. Dead Man's Switch pattern (Healthchecks.io) detects service downtime without firewall rules. |
| Service restart for cert rotation | Hot reload via SIGHUP | Varies by service | Postfix supports reload since early versions. Caddy auto-reloads. Avoids brief service interruption. |
| Custom health check libraries | Stdlib net/http with JSON response | Go 1.x+ | Modern Go projects prefer minimal dependencies. Stdlib sufficient for simple health endpoints. |

**Deprecated/outdated:**

- **Polling certificate files in application code:** Systemd path units (PathChanged=) provide native file change monitoring. Avoids busy-waiting, integrates with existing systemd setup.
- **Human-readable mailq parsing:** Postfix 3.1+ JSON output is canonical. Parsing text output is fragile and breaks on version changes.
- **Single health check endpoint for liveness and readiness:** Docker and Kubernetes best practices recommend separate endpoints. Prevents false positives from transient dependency issues.

---

## Open Questions

### 1. Web Dashboard Location: Profile Server vs Separate Container?

**What we know:**
- Phase 08 profile server already runs HTTP server on home device for .mobileconfig downloads and QR codes
- User requirement: web dashboard for household members (less technical users)
- Dashboard shows 4 metrics: queue depth, delivery status, cert expiry, tunnel health
- Alternative: separate dashboard container adds footprint (conflicts with UX-02: minimal size)

**What's unclear:**
- Is profile server intended to be long-running (always available) or on-demand (only during device setup)?
- Does adding dashboard endpoints to profile server violate separation of concerns?

**Recommendation:**
- Add dashboard endpoints to profile server (`/status` route). Profile server is already HTTP server on home device, minimal additional code for status aggregation. Separate container adds ~10-20MB for lightweight Go dashboard, conflicts with minimal footprint goal.
- If profile server is on-demand only, add dashboard to existing home mail server container instead. Stalwart/Maddy/Postfix already run HTTP management interfaces, natural fit.

### 2. Delivery History Retention: How Many Recent Messages?

**What we know:**
- User requirement: check delivery status of "recent outbound messages"
- Status types: delivered, deferred, bounced
- No database dependency preferred (minimal footprint)

**What's unclear:**
- Is 100 messages sufficient? 1000? 10000?
- Should retention be time-based (last 7 days) or count-based (last 1000)?
- Is per-user tracking needed or system-wide sufficient?

**Recommendation:**
- Start with 1000 recent messages, system-wide, in-memory ring buffer. Configurable via environment variable. Time-based retention (7 days) adds complexity (TTL checks, purging). Count-based is simpler and aligns with "recent" (last N messages).
- If user needs per-user tracking, add later. System-wide sufficient for household use case (< 10 users).

### 3. Alert Severity Levels: Warn vs Critical, or Single Level?

**What we know:**
- User requirement: certificate expiry alerts at 14 days (early warning) and 7 days (urgent)
- Four trigger conditions: cert expiry, queue backup, delivery spike, tunnel down
- Multi-channel delivery: email, webhook, CLI warning

**What's unclear:**
- Should different alert types have different severity levels? (e.g., cert expiry at 14 days = warn, at 7 days = critical)
- Does severity affect delivery channel? (e.g., warn = email only, critical = email + webhook)

**Recommendation:**
- Two severity levels: `warn` and `critical`. Cert expiry at 14 days = warn, 7 days = critical. Queue backup threshold crossed = warn, stuck messages + tunnel down = critical. Both severities use all delivery channels (email + webhook + CLI). Severity is metadata for user filtering (e.g., Home Assistant automations can ignore warnings, act on critical).

---

## Sources

### Primary (HIGH confidence)

- [Postfix postqueue(1) manual](https://www.postfix.org/postqueue.1.html) - JSON output (-j flag) for queue monitoring
- [Go crypto/x509 package](https://pkg.go.dev/crypto/x509) - Certificate parsing and expiry inspection
- [cenkalti/backoff GitHub](https://github.com/cenkalti/backoff) - Exponential backoff retry pattern
- [Let's Encrypt: Decreasing Certificate Lifetimes to 45 Days](https://letsencrypt.org/2025/12/02/from-90-to-45) - Certificate lifetime transition timeline
- [Docker Health Check Best Practices (OneUptime, 2026)](https://oneuptime.com/blog/post/2026-01-30-docker-health-check-best-practices/view) - Liveness vs readiness separation
- [Systemd service(5) manual](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html) - ExecReload and SIGHUP handling

### Secondary (MEDIUM confidence)

- [Postfix TLS README](http://www.postfix.org/TLS_README.html) - TLS certificate configuration and reload behavior
- [M3AAWG DKIM Key Rotation Best Practices](https://www.m3aawg.org/DKIMKeyRotation) - Quarterly rotation recommendations
- [Healthchecks.io Documentation](https://healthchecks.io/docs/) - Push-based monitoring pattern
- [Reload SSL certificates with systemd (iBug, 2024)](https://ibug.io/blog/2024/03/reload-ssl-cert-with-systemd/) - PathChanged= automation pattern
- [Go Health Check Libraries](https://pkg.go.dev/github.com/alexliesenfeld/health) - Community health check patterns
- [Sliding Window Rate Limiting (RussellLuo)](https://github.com/RussellLuo/slidingwindow) - Alert deduplication algorithm

### Tertiary (LOW confidence)

- [Postfix queue management tutorials](https://easyengine.io/tutorials/mail/postfix-queue/) - Queue monitoring patterns (needs verification)
- [IMAP telnet testing guides](https://www.codetwo.com/kb/how-to-test-imap-connection-with-server/) - Manual IMAP health check approach

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All recommendations use Go stdlib, existing project dependencies (cenkalti/backoff), or native Postfix/systemd features
- Architecture: HIGH - Patterns verified against official docs (Postfix JSON output, Docker health checks, systemd path units)
- Pitfalls: HIGH - Common production issues documented in Let's Encrypt docs (rate limits), Postfix version differences (JSON support), Docker health check false positives

**Research date:** 2026-02-14
**Valid until:** 2026-03-16 (30 days - stable domain, Go stdlib and Postfix/systemd features change slowly)