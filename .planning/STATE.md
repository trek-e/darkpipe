# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-08)

**Core value:** Your email lives on your hardware, encrypted in transit, never stored on someone else's server -- and it still works like normal email from the outside.
**Current focus:** Phase 2 - Cloud Relay

## Current Position

Phase: 2 of 9 (Cloud Relay)
Plan: 1 of 3 in current phase
Status: In Progress
Last activity: 2026-02-09 -- Completed 02-01-PLAN.md (cloud relay core)

Progress: [████░░░░░░] 40%

## Performance Metrics

**Velocity:**
- Total plans completed: 4
- Average duration: 4.7 minutes
- Total execution time: 0.31 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 (Transport Layer) | 3 | 889s | 296s |
| 02 (Cloud Relay) | 1 | 234s | 234s |

**Recent Trend:**
- Last 5 plans: 422s, 197s, 270s, 234s
- Trend: Consistent execution speed (3-7 minutes per plan)

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: 9 phases derived from 50 requirements across 12 categories
- [Roadmap]: Transport layer first (WireGuard + mTLS) -- both relay and home device depend on it
- [Roadmap]: VPS provider validation (port 25) folded into Phase 1 as prerequisite activity
- [Roadmap]: Certificate management split -- CERT-01 (public) with cloud relay, CERT-02 (internal) with transport, CERT-03/04 (lifecycle) with monitoring
- [01-01]: Use stdlib text/template for config generation (zero external dependencies)
- [01-01]: Wrap wg CLI rather than implement crypto (leverage official tools)
- [01-01]: Default PersistentKeepalive=25 for NAT traversal without port forwarding
- [01-01]: 0600 permissions for all config files to protect private keys
- [01-01]: Systemd auto-restart with 30s delay to prevent rapid failure loops
- [01-02]: cenkalti/backoff/v4 as only external Go dep for mTLS reconnection
- [01-02]: Go TLS defaults for cipher suites (TLS 1.3 + post-quantum in Go 1.24+)
- [01-02]: Shared testutil package for cert generation across mTLS tests
- [01-02]: Systemd timer renewal with ExecCondition needs-renewal + RandomizedDelaySec jitter
- [01-03]: golang.zx2c4.com/wireguard/wgctrl for kernel-level WireGuard control
- [01-03]: 5-minute health check threshold (PersistentKeepalive=25 refreshes ~2min)
- [01-03]: Unified transport health checker for consistent WireGuard/mTLS interface
- [01-03]: VPS provider guide prioritizes port 25 SMTP compatibility over price
- [Phase 02-01]: Use emersion/go-smtp for both server and client sides
- [Phase 02-01]: LMDB format for Postfix maps (BerkleyDB deprecated in Alpine)
- [Phase 02-01]: Transport abstraction via Forwarder interface for WireGuard/mTLS flexibility

### Pending Todos

None yet.

### Blockers/Concerns

- VPS port 25 restrictions are absolute blockers -- must validate provider before any relay work (Phase 1 prerequisite)
- IP warmup requires 4-6 weeks after Phase 4 (DNS/auth) completes -- time-based, not development
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) -- schema may change

## Session Continuity

Last session: 2026-02-09
Stopped at: Completed 02-01-PLAN.md (cloud relay core)
Resume file: .planning/phases/02-cloud-relay/02-01-SUMMARY.md
Next plan: 02-02-PLAN.md (TLS/SSL certificates)
