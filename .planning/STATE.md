# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-08)

**Core value:** Your email lives on your hardware, encrypted in transit, never stored on someone else's server -- and it still works like normal email from the outside.
**Current focus:** Phase 10 - Mail Migration (added before milestone completion)

## Current Position

Phase: 10 of 10 (Mail Migration)
Plan: 3 of 4 (Provider integrations complete)
Status: Completed 10-03 Provider integrations with OAuth2 and API clients
Last activity: 2026-02-15 -- Completed Phase 10 Plan 03

Progress: [█████████░] 92%

## Performance Metrics

**Velocity:**
- Total plans completed: 29
- Average duration: 6.0 minutes
- Total execution time: 2.9 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 (Transport Layer) | 3 | 889s | 296s |
| 02 (Cloud Relay) | 3 | 1142s | 381s |
| 03 (Home Mail Server) | 3 | 881s | 294s |
| 04 (DNS & Email Auth) | 3 | 1488s | 496s |
| 05 (Queue & Offline) | 2 | 1335s | 668s |
| 06 (Webmail & Groupware) | 2 | 353s | 177s |
| 07 (Build System & Deployment) | 3 | 1245s | 415s |
| 08 (Device Profiles & Client Setup) | 3 | 1274s | 425s |
| 09 (Monitoring & Observability) | 3 | 1294s | 431s |
| 10 (Mail Migration) | 3 | 1611s | 537s |

**Recent Trend:**
- Last 5 plans: 456s (09-03), 422s (10-01), 792s (10-02), 397s (10-03), avg: 517s
- Trend: Phase 10 in progress — Provider integrations added in 6.6 min

**Recent Plans:**
| Phase | Plan | Duration | Tasks | Files |
|-------|------|----------|-------|-------|
| Phase 09 | P03 | 456s | 2 | 12 |
| Phase 10 | P01 | 422s | 2 | 6 |
| Phase 10 | P02 | 792s | 2 | 6 |
| Phase 10 | P03 | 397s | 2 | 12 |

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
- [06-01]: Caddy reverse proxy on cloud relay (auto-HTTPS, lightweight, Go-based aligns with stack)
- [06-01]: IMAP passthrough authentication (no separate webmail user database)
- [06-01]: Roundcube Elastic skin for mobile responsiveness (WEB-02 requirement)
- [06-01]: extra_hosts mail-server:host-gateway pattern for mail server discovery
- [06-01]: 60-minute session timeout with auto-refresh to avoid UX frustration
- [06-02]: Radicale for Maddy/Postfix+Dovecot (Stalwart uses built-in CalDAV/CardDAV)
- [06-02]: Rights file ACLs for shared family calendar and address book
- [06-02]: Well-known URL redirects for iOS/macOS/Android CalDAV/CardDAV auto-discovery
- [06-02]: sync-users.sh syncs mail server users to Radicale htpasswd (same credentials)
- [07-01]: Multi-arch support via TARGETARCH build arg (buildx sets automatically)
- [07-01]: OCI labels on all images (source, version, description, licenses)
- [07-01]: Netcat health checks (more reliable than process-based checks)
- [07-01]: Docker secrets via _FILE suffix convention for all sensitive values
- [07-01]: Setup detection at /config/.darkpipe-configured
- [07-01]: GHCR as sole registry (no Docker Hub)
- [07-01]: Pre-built images for two stacks: default (Stalwart+SnappyMail) and conservative (Postfix+Dovecot+Roundcube+Radicale)
- [07-01]: GitHub Actions cache (type=gha) for layer optimization
- [Phase 07-02]: Quick vs Advanced setup modes for tiered UX - Quick asks 3 questions with opinionated defaults (Stalwart + SnappyMail), Advanced allows full customization
- [Phase 07-02]: Type-safe YAML generation using Go structs (not string templates) for compile-time safety and conditional service inclusion
- [Phase 07-02]: Separate Go module for setup tool to isolate dependencies (survey, pterm, cobra) from core mail services
- [Phase 07-03]: TrueNAS Scale 24.10+ as minimum version for Docker Compose support
- [Phase 07-03]: Raspberry Pi 4GB RAM recommended (2GB possible with optimization)
- [08-01]: App passwords use crypto/rand with charset excluding confusing characters (0/O/1/I)
- [08-01]: Bcrypt cost 12 for password hashing (balance of security and performance)
- [08-01]: Stalwart backend uses $app$<device-name>$<bcrypt-hash> format
- [08-01]: Dovecot and Maddy backends use JSON file storage with flock for concurrency
- [08-01]: Apple profiles are UNSIGNED for v1 (per research recommendation)
- [08-01]: .mobileconfig includes Email+CalDAV+CardDAV in ONE profile (per user decision)
- [08-01]: CalDAV/CardDAV payloads conditionally included based on config
- [08-01]: Autoconfig and autodiscover endpoints are public (no auth) for maximum client compatibility
- [08-01]: Used micromdm/plist (renamed from groob/plist) for Apple plist serialization
- [08-02]: QR codes encode single-use URLs (not inline settings) for revocability and auditability
- [08-02]: Token expiry 15 minutes (sufficient for mobile onboarding, short enough to limit exposure)
- [08-02]: Single-use enforcement: token marked as used IMMEDIATELY on validation (prevents race conditions)
- [08-02]: Profile server on port 8090 (separate from webmail on 8080 for service isolation)
- [08-02]: SRV records include _imap with target '.' (unavailable) per RFC 2782
- [08-02]: Caddy handle directives placed BEFORE default webmail reverse_proxy (first match wins)
- [08-03]: Profile server runs WITHOUT Docker Compose profile (always available like rspamd/redis)
- [08-03]: 64MB memory limit for profile server (sufficient for Go HTTP server with templates)
- [08-03]: Web UI uses Basic Auth with admin credentials (v1 simplification)
- [08-03]: Platform-specific instructions: iOS/macOS get QR+download, Android gets QR+manual, Thunderbird/Outlook get autodiscovery
- [08-03]: CLI QR command supports both terminal ASCII art and PNG file export
- [08-03]: Templates and static assets embedded via embed.FS (no runtime file dependencies)
- [09-01]: Liveness vs Readiness - Liveness always "up" (process alive), readiness performs deep checks (Postfix port 25, IMAP port 993, tunnel)
- [09-01]: Health check handlers return application/health+json (Kubernetes compatibility)
- [09-01]: Queue monitor uses postqueue -j JSON output for reliability over text parsing
- [09-01]: Stuck message threshold 24 hours (configurable via QueueStats.StuckThreshold)
- [09-01]: Delivery tracker uses ring buffer (not database) for 1000 recent entries to avoid I/O overhead
- [09-01]: Ring buffer capacity configurable via MONITOR_DELIVERY_HISTORY env var
- [09-01]: Thread-safe tracker with RWMutex for concurrent log parsing and query access
- [09-01]: Postfix log parser extracts both inbound and outbound deliveries
- [09-01]: Status normalization: sent->delivered for clearer user-facing language
- [09-02]: Alert rate limiting uses simple map-based implementation (not external library) for single-process use case
- [09-02]: Email channel uses sendmail command (assumes Postfix available on mail server)
- [09-02]: CLI alerts written to NDJSON file for CLI consumption (/data/monitoring/cli-alerts.json)
- [09-02]: Certificate renewal uses 2/3-lifetime rule (auto-handles Let's Encrypt 90→45 day transition)
- [09-02]: Exponential backoff with 3 retries for transient failures, permanent error detection for ACME issues
- [09-02]: DKIM rotation reuses Phase 4 selector format ({prefix}-{YYYY}q{Q}) with quarterly rotation
- [09-02]: Service reload uses hot reload (postfix reload, caddy reload) to avoid interruption
- [09-03]: Status aggregator uses interface-based design (HealthChecker, CertWatcher, DeliveryTracker) for testability
- [09-03]: Web dashboard added to profile server at /status (64MB memory sufficient)
- [09-03]: Push pinger uses Dead Man's Switch pattern: external service alerts when pings stop
- [09-03]: Docker HEALTHCHECK updated to /health/live (liveness check, always up if process alive)
- [10-01]: go-imap v2.0.0-beta.8 for IMAP sync (v2 API redesign with better concurrency)
- [10-01]: Thread-safe state tracking with sync.RWMutex and auto-save every 100 operations
- [10-01]: State file at /data/migration-state.json (aligns with container volume pattern)
- [10-01]: Gmail folder mappings: skip All Mail/Important/Starred, map Sent Mail -> Sent
- [10-01]: Outlook folder mappings: skip Clutter, map Deleted Items -> Trash
- [10-01]: IMAP APPEND preserves INTERNALDATE and flags for chronological mailbox order
- [10-01]: Folder-level migration tracking enables folder resume in future plans
- [10-03]: Custom XOAUTH2 SASL client (go-sasl only has OAUTHBEARER, not XOAUTH2)
- [10-03]: Provider registry pattern with init() registration to avoid import cycles
- [10-03]: OAuth2 device flow (RFC 8628) for Gmail and Outlook per locked decision
- [10-03]: iCloud requires app-specific passwords (2FA prerequisite documented)
- [10-03]: MailCow API uses X-API-Key header authentication
- [10-03]: Mailu API uses Bearer token authentication with fallback to IMAP-only
- [10-03]: Generic provider supports flexible IMAP/CalDAV/CardDAV endpoint configuration

### Pending Todos

None — migration tool promoted to Phase 10.

### Blockers/Concerns

- VPS port 25 restrictions are absolute blockers -- must validate provider before any relay work (Phase 1 prerequisite)
- IP warmup requires 4-6 weeks after Phase 4 (DNS/auth) completes -- time-based, not development
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) -- schema may change
- go-imap v2 is beta (v2.0.0-beta.8) -- API may change before stable release, monitor for breaking changes

## Session Continuity

Last session: 2026-02-15
Stopped at: Completed 10-03 Provider integrations
Resume file: .planning/phases/10-mail-migration/10-03-SUMMARY.md
Next plan: Phase 10 Plan 04 (CLI wizard with dry-run and progress bars)
