# S09: Monitoring Observability

**Goal:** Build the core monitoring data collection packages: health check framework with deep readiness probes, Postfix mail queue parser, and delivery status tracker with log parsing.
**Demo:** Build the core monitoring data collection packages: health check framework with deep readiness probes, Postfix mail queue parser, and delivery status tracker with log parsing.

## Must-Haves


## Tasks

- [x] **T01: 09-monitoring-observability 01**
  - Build the core monitoring data collection packages: health check framework with deep readiness probes, Postfix mail queue parser, and delivery status tracker with log parsing.

Purpose: These three packages form the data layer that all higher-level monitoring features (CLI, dashboard, alerts) consume. Health checks satisfy MON-03 (container health), queue monitoring satisfies MON-01 (mail queue health), and delivery tracking satisfies MON-02 (delivery status visibility).

Output: Three Go packages (monitoring/health, monitoring/queue, monitoring/delivery) with full test coverage.
- [x] **T02: 09-monitoring-observability 02**
  - Build the alert notification system with rate limiting and the certificate lifecycle management packages including automated renewal, DKIM rotation, and service reload orchestration.

Purpose: The alert system provides multi-channel notifications for all monitoring events (cert expiry, queue backup, delivery failures, tunnel down) with per-type rate limiting. Certificate lifecycle management automates renewal (CERT-03) and expiry monitoring (CERT-04). DKIM rotation integrates with the Phase 4 quarterly selector format.

Output: Two Go packages (monitoring/alert, monitoring/cert) with full test coverage.
- [x] **T03: 09-monitoring-observability 03**
  - Build the status aggregation layer, CLI command, web dashboard, push-based external monitoring, Docker health checks, and phase integration test suite. This plan wires all monitoring packages together into user-facing interfaces.

Purpose: This is the integration plan that makes monitoring visible to users via CLI (power users) and web dashboard (household members). It also adds Docker health check integration, Caddy health endpoint with Basic Auth, push-based external monitoring, and the comprehensive phase test suite.

Output: Status aggregator package, CLI command, web dashboard on profile server, Docker health checks, Caddy route, push pinger, and integration test suite.

## Files Likely Touched

- `monitoring/health/checker.go`
- `monitoring/health/postfix.go`
- `monitoring/health/imap.go`
- `monitoring/health/tunnel.go`
- `monitoring/health/server.go`
- `monitoring/health/checker_test.go`
- `monitoring/queue/mailq.go`
- `monitoring/queue/stats.go`
- `monitoring/queue/mailq_test.go`
- `monitoring/delivery/parser.go`
- `monitoring/delivery/tracker.go`
- `monitoring/delivery/status.go`
- `monitoring/delivery/parser_test.go`
- `monitoring/delivery/tracker_test.go`
- `monitoring/alert/notifier.go`
- `monitoring/alert/ratelimit.go`
- `monitoring/alert/triggers.go`
- `monitoring/alert/notifier_test.go`
- `monitoring/alert/ratelimit_test.go`
- `monitoring/cert/watcher.go`
- `monitoring/cert/rotator.go`
- `monitoring/cert/dkim.go`
- `monitoring/cert/reload.go`
- `monitoring/cert/watcher_test.go`
- `monitoring/cert/rotator_test.go`
- `monitoring/cert/dkim_test.go`
- `monitoring/status/aggregator.go`
- `monitoring/status/cli.go`
- `monitoring/status/dashboard.go`
- `monitoring/status/push.go`
- `monitoring/status/aggregator_test.go`
- `monitoring/status/cli_test.go`
- `monitoring/status/push_test.go`
- `home-device/profiles/cmd/profile-server/main.go`
- `home-device/profiles/cmd/profile-server/templates/status.html`
- `home-device/docker-compose.yml`
- `cloud-relay/docker-compose.yml`
- `cloud-relay/caddy/Caddyfile`
- `tests/test-monitoring.sh`
