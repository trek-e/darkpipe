# T01: 09-monitoring-observability 01

**Slice:** S09 — **Milestone:** M001

## Description

Build the core monitoring data collection packages: health check framework with deep readiness probes, Postfix mail queue parser, and delivery status tracker with log parsing.

Purpose: These three packages form the data layer that all higher-level monitoring features (CLI, dashboard, alerts) consume. Health checks satisfy MON-03 (container health), queue monitoring satisfies MON-01 (mail queue health), and delivery tracking satisfies MON-02 (delivery status visibility).

Output: Three Go packages (monitoring/health, monitoring/queue, monitoring/delivery) with full test coverage.

## Must-Haves

- [ ] "Health check endpoints return pass/fail status for Postfix, IMAP, and tunnel services"
- [ ] "Queue monitoring reports depth, deferred count, and stuck message count from Postfix"
- [ ] "Delivery tracker records recent message delivery outcomes (delivered, deferred, bounced)"

## Files

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
