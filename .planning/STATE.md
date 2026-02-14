# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-08)

**Core value:** Your email lives on your hardware, encrypted in transit, never stored on someone else's server -- and it still works like normal email from the outside.
**Current focus:** Phase 5 - Queue & Offline Handling (in progress)

## Current Position

Phase: 5 of 9 (Queue & Offline Handling)
Plan: 2 of 2 in current phase (Complete)
Status: Complete
Last activity: 2026-02-14 -- Completed 05-02 (S3 overflow storage + Phase 5 integration tests)

Progress: [████████░░] 52%

## Performance Metrics

**Velocity:**
- Total plans completed: 14
- Average duration: 6.5 minutes
- Total execution time: 1.57 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 (Transport Layer) | 3 | 889s | 296s |
| 02 (Cloud Relay) | 3 | 1142s | 381s |
| 03 (Home Mail Server) | 3 | 881s | 294s |
| 04 (DNS & Email Auth) | 3 | 1488s | 496s |
| 05 (Queue & Offline) | 2 | 1335s | 668s |

**Recent Trend:**
- Last 5 plans: 559s, 582s, 567s, 768s (05-02), avg: 619s
- Trend: Phase 05 complete - Queue & Offline Handling (12m 48s avg)

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
- [Phase 03-03]: Rspamd and Redis as shared services (NOT profiled, run with all mail server options)
- [Phase 03-03]: Greylisting 5-minute delay with score threshold >= 4.0 to avoid greylisting clean mail
- [Phase 03-03]: Private network whitelist (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) prevents greylisting cloud relay traffic
- [Phase 03-03]: Authenticated submission (port 587) bypasses Rspamd for all mail servers
- [Phase 03-03]: Phase test suite (test-mail-flow.sh + test-spam-filter.sh) validates all Phase 03 success criteria
- [04-01]: Time-based DKIM selector format ({prefix}-{YYYY}q{Q}) for quarterly rotation
- [04-01]: Explicit ip4: mechanism in SPF (not include:) to minimize DNS lookup count
- [04-01]: DMARC sp= tag with default sp=quarantine for subdomain protection
- [04-01]: DMARC p=none default for monitoring-first approach (progression: none -> quarantine -> reject)
- [04-01]: Single-line DKIM TXT records for DNS compatibility
- [04-01]: 0600 permissions for DKIM private keys (matches Phase 1 secrets pattern)
- [04-02]: Provider registration via init() to avoid import cycles
- [04-02]: Dry-run by default (--apply required for actual changes)
- [04-02]: Auto-detection from NS records (no manual provider selection)
- [04-02]: Propagation polling across 3 public DNS servers (Google, Cloudflare, OpenDNS)
- [04-02]: SPF duplicate prevention (update existing SPF instead of creating second)
- [04-02]: TXT record quoting for Route53 compatibility
- [04-03]: Use miekg/dns for validation (not stdlib) for controlled DNS server selection
- [04-03]: PTR verification uses stdlib net.LookupAddr (handles in-addr.arpa correctly)
- [04-03]: Detect multiple SPF records as RFC 7208 violation (pitfall #8)
- [04-03]: Parse Authentication-Results with emersion/go-msgauth/authres (RFC 8601 compliant)
- [04-03]: CLI uses flag package with environment variable overrides (12-factor pattern)
- [04-03]: Default DNS servers: 8.8.8.8:53 and 1.1.1.1:53 (public, reliable, fast)
- [05-01]: filippo.io/age for message encryption (industry-standard, simple API)
- [05-01]: CRC32 checksum before decrypt for fast rejection of corrupted data
- [05-01]: Message-ID deduplication to prevent duplicate queuing
- [05-01]: Queue enabled by default (RELAY_QUEUE_ENABLED=true)
- [05-01]: 200MB RAM limit default (leaves headroom in 256MB container)
- [05-01]: Rate limit to 10 messages/tick to prevent thundering herd on reconnection
- [Phase 05-02]: Hash-based S3 key generation (SHA-256 of Message-ID) to avoid special character issues
- [Phase 05-02]: Overflow disabled by default (requires user-provided S3 credentials)

### Pending Todos

None yet.

### Blockers/Concerns

- VPS port 25 restrictions are absolute blockers -- must validate provider before any relay work (Phase 1 prerequisite)
- IP warmup requires 4-6 weeks after Phase 4 (DNS/auth) completes -- time-based, not development
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) -- schema may change

## Session Continuity

Last session: 2026-02-14
Stopped at: Completed 05-02-PLAN.md (S3 overflow storage + Phase 5 integration tests)
Resume file: .planning/phases/05-queue-offline-handling/05-02-SUMMARY.md
Next plan: Phase 5 complete. Ready for Phase 6 or phase test suite.
