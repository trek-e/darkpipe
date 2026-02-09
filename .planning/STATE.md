# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-08)

**Core value:** Your email lives on your hardware, encrypted in transit, never stored on someone else's server -- and it still works like normal email from the outside.
**Current focus:** Phase 2 complete, next: Phase 3 - Home Mail Server

## Current Position

Phase: 3 of 9 (Home Mail Server)
Plan: 2 of 3 in current phase
Status: In Progress
Last activity: 2026-02-09 -- Completed 03-02-PLAN.md (User and domain management)

Progress: [██████░░░░] 62%

## Performance Metrics

**Velocity:**
- Total plans completed: 8
- Average duration: 5.9 minutes
- Total execution time: 0.76 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 (Transport Layer) | 3 | 889s | 296s |
| 02 (Cloud Relay) | 3 | 1142s | 381s |
| 03 (Home Mail Server) | 2 | 606s | 303s |

**Recent Trend:**
- Last 5 plans: 366s, 542s, 300s, 306s
- Trend: Configuration plans consistent (~5 min), test suite plans slower (9 min)

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
- [02-01]: Use emersion/go-smtp for both server and client sides
- [02-01]: LMDB format for Postfix maps (BerkleyDB deprecated in Alpine)
- [02-01]: Transport abstraction via Forwarder interface for WireGuard/mTLS flexibility
- [02-02]: Webhook notifications rate-limited per domain (1-hour dedup window) to prevent spam
- [02-02]: Certificate watcher uses mtime-based change detection every 5 minutes
- [02-02]: Postfix TLS disabled on first boot until certificates are available
- [02-02]: Strict mode uses postconf for dynamic configuration without editing main.cf
- [02-02]: HTTP-01 challenge for initial cert obtain; DNS-01 documented as alternative
- [02-02]: TLS 1.2+ only with server cipher preference for modern security
- [02-03]: Ephemeral verification scans 5 Postfix queue dirs, ignores control files
- [02-03]: MockForwarder exported in forward/mock.go for cross-package testing
- [02-03]: Docker image optimized: Alpine 3.21, stripped binary, .dockerignore, target ~35MB
- [02-03]: Docker compose 256MB memory limit enforced via deploy.resources.limits
- [02-03]: All tests use stdlib testing only (no external frameworks) for zero dependencies
- [Phase 03-01]: Docker compose profiles for mail server selection (stalwart, maddy, postfix-dovecot)
- [Phase 03-01]: LMDB format for Postfix maps (BerkleyDB deprecated in Alpine 3.13+)
- [Phase 03-01]: Self-signed TLS for IMAP/submission within WireGuard tunnel
- [Phase 03-01]: Virtual mailbox domains only (vmail UID/GID 5000, no system users)
- [Phase 03-02]: Email address (user@domain) as username for uniform interface across all mail servers
- [Phase 03-02]: Multi-domain via virtual_mailbox_domains (Postfix), local_domains (Maddy), REST API (Stalwart)
- [Phase 03-02]: Aliases resolved before mailbox delivery (admin@example.com -> alice@example.com)
- [Phase 03-02]: Catch-all requires spam filtering (@example.org -> bob@example.org with Rspamd prerequisite)
- [Phase 03-02]: Postfix domains only in virtual_mailbox_domains (NOT virtual_alias_domains, avoids anti-pattern)

### Pending Todos

None yet.

### Blockers/Concerns

- VPS port 25 restrictions are absolute blockers -- must validate provider before any relay work (Phase 1 prerequisite)
- IP warmup requires 4-6 weeks after Phase 4 (DNS/auth) completes -- time-based, not development
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) -- schema may change

## Session Continuity

Last session: 2026-02-09
Stopped at: Completed 03-02-PLAN.md (User and domain management with multi-user, multi-domain, aliases, catch-all)
Resume file: .planning/phases/03-home-mail-server/03-02-SUMMARY.md
Next plan: 03-03-PLAN.md (Spam filtering)
