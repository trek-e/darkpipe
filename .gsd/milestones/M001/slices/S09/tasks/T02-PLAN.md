# T02: 09-monitoring-observability 02

**Slice:** S09 — **Milestone:** M001

## Description

Build the alert notification system with rate limiting and the certificate lifecycle management packages including automated renewal, DKIM rotation, and service reload orchestration.

Purpose: The alert system provides multi-channel notifications for all monitoring events (cert expiry, queue backup, delivery failures, tunnel down) with per-type rate limiting. Certificate lifecycle management automates renewal (CERT-03) and expiry monitoring (CERT-04). DKIM rotation integrates with the Phase 4 quarterly selector format.

Output: Two Go packages (monitoring/alert, monitoring/cert) with full test coverage.

## Must-Haves

- [ ] "Alerts fire via email, webhook, and CLI warning for all four trigger conditions"
- [ ] "Same alert type suppressed for 1 hour after first notification (rate limiting)"
- [ ] "Certificate expiry alerts fire at 14 days and 7 days before expiry"
- [ ] "Certificate rotation renews at 2/3 of certificate lifetime with exponential backoff retry"
- [ ] "DKIM key rotation runs quarterly matching Phase 4 selector format"

## Files

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
