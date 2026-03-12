# Project Research Summary

**Project:** DarkPipe
**Domain:** Privacy-First Self-Hosted Email System
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Executive Summary

DarkPipe is a privacy-focused self-hosted email system that solves the fundamental problem preventing home-based email: residential ISPs block port 25 and blacklist dynamic IPs. The solution is a cloud relay + home device split architecture where a minimal cloud VPS (25-30MB container) receives internet mail and forwards it through an encrypted WireGuard tunnel to a home device (Raspberry Pi 4 or similar) that stores all email data. This architecture ensures email data never persists in the cloud while maintaining internet-standard deliverability.

The recommended stack balances privacy, performance, and resource constraints: Postfix for the proven minimal cloud relay, Stalwart or Maddy for the modern home mail server (single-binary deployment), WireGuard kernel module for transport (500% faster than userspace on ARM64), and Go for orchestration glue code. The key differentiator is GitHub Actions-driven customizable builds - users select their stack components (mail server, webmail, calendar) via workflow configuration, and the system generates multi-architecture Docker images. This eliminates the "one size fits all" limitation of competitors like Mail-in-a-Box and Mailcow.

Critical risks center on email deliverability (IP reputation, DNS authentication, VPS provider port 25 restrictions) and operational complexity (self-hosted email requires ongoing maintenance that drives 80%+ user abandonment). Mitigation requires careful VPS provider selection (Linode/OVH have port 25 open by default), mandatory 4-6 week IP warmup period, automated DNS validation, and exceptional UX to reduce setup/maintenance burden below competitor levels.

## Key Findings

### Recommended Stack

DarkPipe requires a split-stack architecture optimized for different deployment targets: cloud relay (minimal container on amd64 VPS) and home device (full-featured on ARM64 or amd64).

**Core technologies:**
- **Postfix (Alpine)** for cloud relay - Battle-tested SMTP relay in 15-30MB container, proven minimal footprint, decades of production hardening
- **Stalwart or Maddy** for home mail server - Single-binary deployment (Stalwart 70MB with built-in CalDAV/CardDAV, Maddy 15MB for minimal footprint)
- **WireGuard kernel module** for transport - 500% faster than userspace on ARM64 Ethernet per Nord Security benchmarks, kernel module standard since Linux 5.6
- **Go** for orchestration/glue code - 2-5MB static binaries, excellent ARM64 cross-compilation, mature SMTP libraries (emersion/go-smtp)
- **step-ca** for internal CA - Modern ACME server with short-lived certificates and auto-renewal
- **Alpine Linux** base image - 5MB base trades 3MB for operational benefits (shell, package manager, debugging tools)

**Critical version notes:**
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026 with stable schema)
- Postfix 3.7+ requires LMDB (BerkleyDB deprecated in Alpine)
- WireGuard kernel module requires Linux 5.6+ (standard in modern distributions)

### Expected Features

Email is uniquely complex among self-hosted services - requires perfect DNS, reputation management, anti-spam, and ongoing monitoring.

**Must have (table stakes):**
- Core SMTP/IMAP/TLS transport with port 587 submission
- SPF/DKIM/DMARC authentication (Gmail/Yahoo/Microsoft mandate since 2024-2025)
- Reverse DNS (PTR) record configuration (missing = instant spam folder)
- Basic webmail (Roundcube/SnappyMail) for non-technical users
- Multi-user/multi-domain support
- Spam filtering (Rspamd standard)
- Let's Encrypt auto-renewal (90-day certificates)

**Should have (competitive advantage):**
- **Cloud-fronted architecture** - DarkPipe's core value proposition, solves port 25 blocking
- **GitHub Actions customization** - User-selectable stack via templated workflows (unique vs competitors)
- **First-class ARM64 support** - Raspberry Pi as primary deployment target
- **DNS API integration** - Automates record creation for Cloudflare/Route53/etc.
- CalDAV/CardDAV server (required for Gmail replacement)
- MTA-STS + DANE for TLS enforcement
- Prometheus metrics export for homelab users

**Defer (v2+):**
- Queue encryption at rest (conflicts with spam scanning, high complexity)
- PGP/WKD support (niche user base, key management complexity)
- Full-text search with attachment indexing (resource-intensive on Pi)
- Audit logging for admin actions (enterprise feature)

**Key differentiator vs competitors:** Mail-in-a-Box, Mailu, docker-mailserver, and Mailcow all assume public IP and full VPS deployment. DarkPipe is the only solution architected for home device storage with cloud relay for internet-facing SMTP. Additionally, competitors offer opinionated stacks while DarkPipe provides user customization via GitHub Actions.

### Architecture Approach

Cloud relay is minimal, stateless, and ephemeral. Home device is full-featured with persistent storage. Transport layer (WireGuard or mTLS) provides secure, NAT-traversing connection between them.

**Major components:**

1. **Cloud Relay (SMTP Gateway)** - Minimal Postfix in relay-only mode (25-30MB container), ephemeral RAM queue with optional encrypted S3 overflow, Let's Encrypt certificates, no persistent mail storage
2. **Transport Layer** - WireGuard hub-and-spoke tunnel (primary) with persistent keepalive for NAT traversal, or mTLS persistent connection (fallback for restricted networks)
3. **Home Device Mail Stack** - Full mail server (Postfix+Dovecot or Stalwart/Maddy), optional CalDAV/CardDAV (Radicale/Baikal or built into Stalwart), optional webmail (Roundcube/SnappyMail), all storage in Docker volumes
4. **Build System** - GitHub Actions with matrix builds for component selection and multi-arch images (arm64/amd64)
5. **DNS Automation** - Record generation, provider API integration, pre-deployment validation

**Critical architectural patterns:**
- Ephemeral cloud relay with persistent home storage (privacy requirement)
- WireGuard hub-and-spoke with home device as spoke (NAT-friendly)
- Single container vs Docker Compose stack options (resource flexibility)
- Build-time configuration via GitHub Actions (eliminates runtime complexity)

**Data flow:**
- Inbound: Internet SMTP → Cloud Relay → WireGuard → Home Device → User IMAP client
- Outbound: User SMTP → Home Device → WireGuard → Cloud Relay → Internet SMTP

### Critical Pitfalls

Research identified 11 critical pitfalls with HIGH confidence. Top 5 that shape roadmap:

1. **VPS provider port 25 restrictions** - DigitalOcean/Hetzner/Vultr block SMTP by default. Solution: Choose Linode/OVH for v1 (open by default), document provider policies, budget time for unblocking requests. **Must address in Phase 0 (Infrastructure Selection).**

2. **New VPS IPs start blacklisted or zero reputation** - Fresh IPs may be on RBL blacklists from previous tenants, or have no sending history causing Gmail/Outlook rejection. Solution: Check IP reputation before launch (MXToolbox/Spamhaus), mandatory 4-6 week warmup schedule (start 2-5 emails/day, gradually increase), continuous blacklist monitoring. **Extends MVP timeline by 4-6 weeks.**

3. **Missing/misconfigured SPF/DKIM/DMARC breaks deliverability** - Even one-character DNS typo causes authentication failures. Solution: Use 2048-bit DKIM keys (1024-bit weak), start DMARC with p=none for monitoring, automated testing with mail-tester.com (must score 9+/10). **Must address in Phase 1 (MVP).**

4. **Missing PTR record triggers instant spam filtering** - Reverse DNS controlled by VPS provider, not DNS registrar. Solution: Request PTR from provider before sending any email, verify forward-confirmed reverse DNS, include in deployment checklist. **Must address in Phase 1 (MVP).**

5. **WireGuard tunnel fails after home internet drop** - Tunnel doesn't auto-reconnect after ISP outages or dynamic IP changes. Solution: PersistentKeepalive=25 setting, systemd Restart=on-failure, use IP addresses or public DNS (not local DNS), monitor handshake timestamp. **Must address in Phase 1 (MVP).**

**Additional critical pitfalls:**
- Residential/dynamic IP from home device gets blacklisted (architecture prevents this by design - home never sends direct SMTP)
- TLS certificate expiration breaks email silently (automated renewal + monitoring required)
- Becoming an open relay leads to immediate blacklisting (require SASL auth, test with MXToolbox)
- Docker volume management loses mail data on updates (use named volumes, test restore)
- Raspberry Pi ARM64 runs out of memory under load (don't run ClamAV/SpamAssassin on Pi, filter on cloud relay)
- Users give up due to complex setup (UX must exceed competitors, automate DNS/cert management)

## Implications for Roadmap

Based on dependency analysis and pitfall research, suggested phase structure:

### Phase 0: Infrastructure Selection & Validation
**Rationale:** VPS provider port 25 restrictions are absolute blockers. Must research and validate provider before any development. IP reputation baseline must be established.

**Delivers:**
- VPS provider compatibility list (port 25 open: Linode, OVH; unblocking available: BuyVM, Vultr)
- IP reputation validation checklist (MXToolbox, Spamhaus, multi-RBL)
- Provider selection guide for users

**Addresses:** Pitfall #1 (port 25 restrictions), Pitfall #2 (IP blacklisting)

**Research needed:** None - provider policies change frequently, maintain living document

### Phase 1: Foundation (Configuration & Transport)
**Rationale:** Configuration schema drives all components. Transport layer (WireGuard) is architectural requirement that both cloud relay and home device depend on.

**Delivers:**
- Configuration schema (darkpipe.yaml) and validation
- WireGuard tunnel setup with NAT traversal and auto-reconnection
- DNS automation library (record generation, validation, provider APIs)

**Uses:** Go for config/DNS tools, WireGuard kernel module

**Addresses:** Architecture foundation, Pitfall #8 (tunnel reconnection)

**Research needed:** LOW - WireGuard well-documented, Go stdlib sufficient

### Phase 2: Cloud Relay (Minimal Gateway)
**Rationale:** Cloud relay is simpler than home device (relay-only vs full mail server). Depends on transport layer from Phase 1.

**Delivers:**
- Postfix relay container (Alpine-based, 25-30MB)
- Ephemeral RAM queue with configurable overflow
- Let's Encrypt/Certbot automation
- Health checks and basic monitoring

**Uses:** Postfix 3.7+, Alpine Linux, Certbot, WireGuard client

**Addresses:** Pitfall #6 (certificate expiration), Pitfall #7 (open relay prevention)

**Research needed:** MEDIUM - S3 overflow queue pattern needs validation

### Phase 3: Home Device (Mail Storage)
**Rationale:** Home device provides persistent storage and full mail server functionality. Depends on transport layer and benefits from cloud relay for testing.

**Delivers:**
- Single-container option (docker-mailserver extended or Stalwart/Maddy)
- Docker Compose alternative (Postfix + Dovecot + Radicale)
- Volume management and persistence
- Basic webmail integration (Roundcube or SnappyMail)

**Uses:** Stalwart/Maddy/Postfix+Dovecot (user-selectable), Dovecot, Radicale or Baikal

**Addresses:** Pitfall #9 (Docker volume data loss), Pitfall #10 (Pi memory limits)

**Research needed:** MEDIUM - CalDAV/CardDAV integration patterns, ARM64 performance testing

### Phase 4: DNS & Authentication (Deliverability)
**Rationale:** Email authentication (SPF/DKIM/DMARC) and DNS automation are critical for deliverability but depend on functional mail stack for testing.

**Delivers:**
- Automated SPF/DKIM/DMARC record generation
- DNS provider API integrations (Cloudflare, Route53)
- Pre-deployment validation (mail-tester.com integration)
- PTR record verification and documentation

**Uses:** DNS automation library from Phase 1, OpenDKIM or built-in signing

**Addresses:** Pitfall #3 (SPF/DKIM/DMARC), Pitfall #4 (PTR records)

**Research needed:** LOW - Standards well-documented, provider APIs mature

### Phase 5: Build System (User Customization)
**Rationale:** GitHub Actions builds enable DarkPipe's key differentiator (user-selectable components) but require complete stack to be implemented first.

**Delivers:**
- Multi-arch build workflows (arm64, amd64)
- Component selection inputs (mail server, webmail, calendar)
- Build matrix for parallel builds
- Image publishing to GHCR

**Uses:** GitHub Actions, Docker buildx, QEMU for emulation

**Addresses:** Core differentiator vs competitors

**Research needed:** LOW - Multi-arch Docker builds well-documented

### Phase 6: Deployment & UX (Launch Readiness)
**Rationale:** Exceptional UX required to overcome "self-hosted email is too hard" barrier. All components must be functional for end-to-end testing.

**Delivers:**
- CLI wizard for initial setup
- Pre-flight checks (DNS, ports, IP reputation)
- Status dashboard (authentication, reputation, queue health)
- Plain-language error messages
- Zero-downtime update process

**Uses:** Go CLI (cobra/viper), web framework for dashboard

**Addresses:** Pitfall #11 (complex setup UX)

**Research needed:** MEDIUM - Update strategies need validation, UX testing required

### Phase 7: IP Warmup & Production Validation (4-6 weeks)
**Rationale:** IP warmup cannot be skipped or accelerated. This is a time-based phase, not development.

**Delivers:**
- IP warmup schedule execution (2-5 emails/day → 25-50/day over 4-6 weeks)
- Deliverability monitoring (bounce rates, spam folder placement)
- Blacklist monitoring automation
- DMARC report collection and analysis

**Uses:** Existing mail stack, external monitoring tools

**Addresses:** Pitfall #2 (IP reputation)

**Research needed:** None - warmup schedules standard, execution required

### Phase 8: Scaling & Operations (Post-Launch)
**Rationale:** Operational features that reduce ongoing maintenance burden. Can be added after MVP validates core concept.

**Delivers:**
- Automated blacklist monitoring with alerts
- Queue management and cleanup
- Certificate expiry monitoring
- Backup/restore automation
- Rate limiting and abuse prevention
- Prometheus metrics export

**Addresses:** Reduces ongoing maintenance burden (Pitfall #11)

**Research needed:** LOW - Standard patterns, needs integration testing

### Phase Ordering Rationale

**Why this order:**
- **Phase 0 before anything:** Provider restrictions are absolute blockers
- **Configuration/transport before implementations:** Shared dependencies
- **Cloud relay before home device:** Simpler component validates transport layer
- **DNS after functional stack:** Needs working mail server for testing
- **Build system after components:** Requires complete stack to parameterize
- **UX after functionality:** Can't polish non-existent features
- **IP warmup is time-based:** 4-6 weeks regardless of development speed
- **Operations last:** Nice-to-have after core validation

**How this avoids pitfalls:**
- Provider selection (Phase 0) prevents port 25 issues
- Transport resilience (Phase 1) prevents tunnel failures
- Volume management (Phase 3) prevents data loss
- DNS automation (Phase 4) prevents authentication failures
- UX focus (Phase 6) reduces abandonment
- IP warmup (Phase 7) ensures deliverability

**Why grouping makes sense:**
- Phase 1 groups foundational dependencies (config, transport, DNS library)
- Phases 2-3 group deployment targets (cloud vs home)
- Phase 4 groups deliverability requirements (DNS, auth)
- Phase 6 groups user-facing features (UX, dashboard)
- Phase 8 groups operational maturity features

### Research Flags

Phases likely needing deeper research during planning:

- **Phase 3 (Home Device):** CalDAV/CardDAV integration patterns unclear - Stalwart has built-in but Postfix+Dovecot requires separate service. ARM64 performance benchmarks needed for spam filtering decision.
- **Phase 6 (UX/Updates):** Zero-downtime update strategies for mail servers need validation - queue draining and mail-in-flight handling during container swap.
- **Phase 8 (Operations):** S3-compatible queue overflow pattern proven by Maddy/Stalwart but implementation details sparse.

Phases with standard patterns (skip research-phase):

- **Phase 1 (Configuration):** YAML schemas and Go CLI tooling well-documented
- **Phase 2 (Cloud Relay):** Postfix relay configuration extremely well-documented
- **Phase 4 (DNS):** Email authentication standards (SPF/DKIM/DMARC) mature with clear specs
- **Phase 5 (Build System):** GitHub Actions multi-arch builds have official docs and many examples
- **Phase 7 (IP Warmup):** Warmup schedules standardized across industry

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Core technologies (Postfix, WireGuard, Docker) verified with official documentation. Stalwart/Maddy recommended based on feature comparison and current versions. |
| Features | MEDIUM-HIGH | Table stakes validated against competitors and standards (SPF/DKIM/DMARC mandates). Differentiators based on competitor analysis. Anti-features informed by Helm failure analysis. |
| Architecture | MEDIUM | Core patterns (relay, WireGuard tunnel, Docker deployment) HIGH confidence. Advanced patterns (S3 overflow, CalDAV integration) MEDIUM - proven by other projects but implementation details need validation. |
| Pitfalls | HIGH | Critical pitfalls verified with official sources (Spamhaus, provider docs, Let's Encrypt). Community pitfalls (UX complexity, IP warmup) validated by multiple 2026 blog posts and user abandonment reports. |

**Overall confidence:** MEDIUM-HIGH

### Gaps to Address

Areas where research was inconclusive or needs validation during implementation:

- **CalDAV/CardDAV without separate service:** Stalwart provides built-in CalDAV/CardDAV but research found no evidence of Dovecot-native implementation. Need to decide: recommend Stalwart for integrated approach, or Postfix+Dovecot+Radicale for flexibility. Phase 3 planning should resolve this.

- **Queue overflow threshold calculation:** No authoritative source on when to trigger S3 overflow from RAM queue. Needs simulation/testing based on average email size and delivery latency. Phase 2 planning should establish thresholds.

- **Certificate rotation without downtime:** How to rotate mTLS certificates (if mTLS transport used) without breaking persistent connection. WireGuard doesn't have this issue (uses pre-shared keys). Phase 1 planning for mTLS fallback should investigate.

- **Multi-user resource budgeting on Pi:** Architecture assumes single user or small family (5-10 users). What are actual memory/CPU requirements per user on RPi4? Phase 3 needs empirical testing.

- **ARM64 spam filtering performance:** Should ClamAV/SpamAssassin run on cloud relay (defeats privacy model by scanning plaintext in cloud) or home device (may exceed Pi4 resources)? Phase 3 needs benchmarking to inform default configuration.

- **DNS propagation delays on initial deployment:** How to handle 24-48 hour DNS propagation when mail may arrive immediately after MX record creation? Does cloud relay need extended queue retention for first deployment? Phase 4 planning should establish strategy.

- **Backup strategy ownership:** Should DarkPipe provide automated backup (additional complexity) or document user's responsibility? Where should backups be stored (user's cloud account, local USB drive)? Phase 8 planning should decide scope.

## Sources

### Primary (HIGH confidence)

Research drew from official documentation for core technologies:

- **STACK.md sources:** Stalwart/Maddy/Postfix official docs, Docker multi-platform build documentation, WireGuard quickstart, step-ca GitHub, emersion/go-smtp library documentation
- **FEATURES.md sources:** Mail-in-a-Box/Mailu/Mailcow official sites and GitHub repositories, Helm shutdown FAQ and reviews, email authentication standards (SPF/DKIM/DMARC RFCs), CalDAV/CardDAV specifications
- **ARCHITECTURE.md sources:** Postfix configuration README, WireGuard documentation, Docker Compose documentation, docker-mailserver GitHub repository
- **PITFALLS.md sources:** VPS provider official SMTP policies (DigitalOcean, Linode, OVH, Hetzner, Vultr), Spamhaus PBL documentation, Let's Encrypt challenge types, email deliverability guides from major providers

### Secondary (MEDIUM confidence)

Community resources and comparative analyses:

- Alpine vs Distroless vs Scratch container base image comparisons (Medium, OneUptime blog)
- WireGuard kernel vs userspace performance benchmarks (Nord Security blog with RPi4 testing)
- SnappyMail vs Roundcube webmail comparison (Forward Email blog)
- Multi-architecture Docker image build guides (Blacksmith, Red Hat developer articles)
- Self-hosting email sustainability discussions (2026 blog posts on email self-hosting challenges)
- IP warmup best practices (Mailwarm, Iterable, Mailtrap guides)

### Tertiary (LOW confidence, needs validation)

- S3-compatible queue overflow pattern (proven by Maddy/Stalwart but implementation details sparse)
- CalDAV/CardDAV integration approaches (Dovecot mailing list discussions, community examples)
- Raspberry Pi resource constraints under mail server load (forum discussions, community reports)
- Zero-downtime Docker update strategies for mail servers (generic Docker guides, needs mail-specific validation)

---
*Research completed: 2026-02-08*
*Ready for roadmap: yes*

# Architecture Research: DarkPipe

**Domain:** Cloud-Relay + Home-Device Email System
**Researched:** 2026-02-08
**Confidence:** MEDIUM

## Standard Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           INTERNET                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  Inbound SMTP (port 25)           Outbound SMTP (port 25/587)            │
│          ↓                                     ↑                          │
│  ┌───────────────────────────────────────────────────────────┐           │
│  │              CLOUD RELAY (Minimal Gateway)                 │           │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │           │
│  │  │ SMTP Receive │→ │ Ephemeral    │→ │ SMTP Forward │    │           │
│  │  │ (Postfix)    │  │ Queue (RAM)  │  │ (Postfix)    │    │           │
│  │  └──────────────┘  └──────────────┘  └──────────────┘    │           │
│  │         ↓                   ↓                                          │
│  │  ┌──────────────┐  ┌──────────────┐                                   │
│  │  │ TLS/Certbot  │  │ Overflow to  │                                   │
│  │  │ (Let's Enc)  │  │ S3 Storage   │                                   │
│  │  └──────────────┘  └──────────────┘                                   │
│  └───────────────────────────────────────────────────────────┘           │
│                                ↕                                          │
│                    TRANSPORT LAYER                                        │
│              (WireGuard Tunnel OR mTLS Connection)                        │
│                                ↕                                          │
│  ┌───────────────────────────────────────────────────────────┐           │
│  │         HOME DEVICE (Full Mail Stack on RPi4+)             │           │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │           │
│  │  │ Transport    │  │ Mail Server  │  │ Webmail      │    │           │
│  │  │ Endpoint     │  │ (Postfix/    │  │ (Roundcube/  │    │           │
│  │  │ (WG/mTLS)    │→ │  Dovecot)    │  │  others)     │    │           │
│  │  └──────────────┘  └──────┬───────┘  └──────────────┘    │           │
│  │  ┌──────────────┐         │                                           │
│  │  │ CalDAV/      │←────────┘                                           │
│  │  │ CardDAV      │                                                      │
│  │  │ (Radicale)   │                                                      │
│  │  └──────────────┘                                                      │
│  │         ↓                                                              │
│  │  ┌──────────────────────────────────────────┐                         │
│  │  │ Persistent Storage (Docker Volumes)      │                         │
│  │  └──────────────────────────────────────────┘                         │
│  └───────────────────────────────────────────────────────────┐           │
└──────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Cloud Relay (SMTP Gateway)** | Minimal SMTP gateway that receives mail and immediately forwards via transport layer | Postfix in relay-only mode, smallest possible container |
| **Ephemeral Queue** | In-memory queue for active forwarding; overflow to encrypted S3-compatible storage | RAM-based spool with configurable overflow threshold |
| **Transport Layer** | Secure, NAT-traversing persistent connection between cloud and home | WireGuard tunnel (primary) or mTLS persistent connection |
| **Home Device Mail Stack** | Full-featured mail server with IMAP, SMTP, CalDAV, CardDAV | Docker Mailserver or component stack (Postfix + Dovecot + Radicale + webmail) |
| **Certificate Management** | TLS certificates for cloud relay (public) and relay↔home (internal) | Certbot for public certs, internal CA for relay-to-home authentication |
| **Build System** | User-configurable CI/CD that produces custom multi-arch Docker images | GitHub Actions with matrix builds for arm64/amd64 |
| **DNS Management** | Auto-generate and validate required DNS records (MX, SPF, DKIM, DMARC) | DNS provider API integration with validation |
| **Configuration** | User-friendly config format that drives all components | YAML primary, env vars for secrets, CLI wizard for initial setup |

## Recommended Project Structure

```
darkpipe/
├── cloud-relay/              # Minimal SMTP gateway
│   ├── Dockerfile            # Multi-stage build for smallest image
│   ├── postfix/              # Relay-only Postfix config
│   │   ├── main.cf.template  # Template with variable substitution
│   │   └── master.cf         # Process configuration
│   ├── transport/            # WireGuard or mTLS endpoint
│   │   ├── wireguard/        # WireGuard config templates
│   │   └── mtls/             # mTLS connection handler
│   ├── queue/                # Ephemeral queue with overflow
│   │   ├── ram-spool/        # In-memory queue manager
│   │   └── s3-overflow/      # S3-compatible backup storage
│   └── certbot/              # Let's Encrypt automation
│       └── renewal-hooks/    # Certificate rotation scripts
├── home-device/              # Home mail stack
│   ├── single-container/     # All-in-one Docker Mailserver approach
│   │   └── Dockerfile        # Extended docker-mailserver image
│   ├── compose-stack/        # Multi-container alternative
│   │   ├── docker-compose.yml
│   │   ├── postfix/          # MTA configuration
│   │   ├── dovecot/          # IMAP/POP3 server
│   │   ├── radicale/         # CalDAV/CardDAV server
│   │   └── roundcube/        # Webmail interface
│   └── transport/            # WireGuard or mTLS client
│       ├── wireguard/        # WireGuard peer config
│       └── mtls/             # mTLS client with reconnection
├── build-system/             # GitHub Actions workflows
│   ├── .github/workflows/
│   │   ├── build-cloud.yml   # Cloud relay build (amd64)
│   │   ├── build-home.yml    # Home device build (arm64/amd64)
│   │   └── user-config.yml   # User-triggered custom build
│   └── config-templates/     # Component selection templates
├── dns-tools/                # DNS automation
│   ├── generator/            # DNS record generation from config
│   ├── validators/           # Pre-deployment validation
│   └── providers/            # DNS provider API integrations
│       ├── cloudflare.go
│       ├── route53.go
│       └── generic-api.go
├── config/                   # Configuration management
│   ├── darkpipe.yaml.example # Main configuration template
│   ├── schema.json           # JSON schema for validation
│   └── wizard/               # Interactive CLI setup
│       └── setup.go
└── docs/                     # User documentation
    ├── architecture.md       # This document
    ├── deployment.md         # Deployment guide
    └── troubleshooting.md    # Common issues
```

### Structure Rationale

- **cloud-relay/**: Separated by deployment target (cloud VPS). Single-purpose design for minimal attack surface and resource usage.
- **home-device/**: Provides both single-container and compose-stack options to support different use cases (resource-constrained RPi4 vs. more powerful home servers).
- **build-system/**: User-driven CI/CD allows users to select components without maintaining their own build infrastructure.
- **dns-tools/**: DNS automation is critical for email delivery; separating this allows pre-deployment validation and multi-provider support.
- **config/**: Single source of truth (YAML) that drives all other components, reducing configuration drift.

## Architectural Patterns

### Pattern 1: Ephemeral Cloud Relay with Persistent Home Storage

**What:** Cloud relay stores nothing persistently (except optional encrypted overflow queue). All permanent storage lives on home device.

**When to use:** When privacy is paramount and cloud infrastructure must be minimal and stateless.

**Trade-offs:**
- **Pro:** Zero persistent cloud storage means no data breach exposure; minimal cloud costs.
- **Pro:** Cloud relay container can be ultra-small (50-100MB) and horizontally scalable.
- **Con:** Home device must be highly available or mail delivery fails.
- **Con:** Transport layer becomes single point of failure; requires robust reconnection logic.

**Example:**
```yaml
# cloud-relay postfix main.cf
# Relay immediately, no local delivery
mydestination =
local_recipient_maps =
relay_domains = $mydomain
transport_maps = hash:/etc/postfix/transport

# Minimal queue
queue_run_delay = 30s
minimal_backoff_time = 60s
maximal_queue_lifetime = 4h

# Forward via WireGuard tunnel to home device
# /etc/postfix/transport
example.com    smtp:[10.8.0.2]:25
```

### Pattern 2: WireGuard Hub-and-Spoke with NAT Traversal

**What:** Cloud relay acts as WireGuard hub (static IP, always reachable). Home device acts as spoke (behind NAT, connects outbound to hub). Uses persistent keepalive for NAT hole punching.

**When to use:** When home device is behind residential NAT and cannot accept inbound connections directly.

**Trade-offs:**
- **Pro:** Works with any residential ISP, no port forwarding required.
- **Pro:** WireGuard's performance is excellent (minimal overhead, uses kernel cryptography).
- **Pro:** Simple configuration; WireGuard handles roaming IPs automatically.
- **Con:** Home device must maintain persistent connection (battery consideration for mobile devices).
- **Con:** Cellular NATs with strict UDP randomization may require fallback.

**Example:**
```ini
# Cloud relay WireGuard config
[Interface]
PrivateKey = <cloud-private-key>
Address = 10.8.0.1/24
ListenPort = 51820

[Peer]
PublicKey = <home-device-public-key>
AllowedIPs = 10.8.0.2/32

# Home device WireGuard config
[Interface]
PrivateKey = <home-private-key>
Address = 10.8.0.2/24

[Peer]
PublicKey = <cloud-public-key>
Endpoint = relay.example.com:51820
AllowedIPs = 10.8.0.1/32
PersistentKeepalive = 25  # NAT hole punch every 25s
```

### Pattern 3: mTLS Persistent Connection with Reconnection Handling

**What:** Alternative to WireGuard. Cloud relay and home device maintain persistent mTLS connection with mutual certificate authentication. Home device initiates connection (NAT-friendly). Connection pooling and automatic reconnection.

**When to use:** When WireGuard is unavailable (restrictive networks) or when HTTP/2 multiplexing over single connection is beneficial.

**Trade-offs:**
- **Pro:** Works through HTTP-aware proxies and restrictive firewalls.
- **Pro:** Certificate-based authentication eliminates pre-shared keys.
- **Pro:** Connection pooling allows multiple mail streams over single connection.
- **Con:** More complex to implement than WireGuard.
- **Con:** Higher CPU overhead for TLS handshakes (mitigated by session resumption).
- **Con:** Requires careful certificate expiration and renewal automation.

**Example:**
```go
// Home device mTLS client with reconnection
func maintainConnection(ctx context.Context) {
    backoff := 1 * time.Second
    for {
        select {
        case <-ctx.Done():
            return
        default:
            conn, err := dialMTLS()
            if err != nil {
                log.Printf("Connection failed: %v, retrying in %v", err, backoff)
                time.Sleep(backoff)
                backoff = min(backoff*2, 5*time.Minute)
                continue
            }
            backoff = 1 * time.Second
            handleConnection(conn) // Blocks until connection drops
        }
    }
}
```

### Pattern 4: Single Container vs. Docker Compose Stack

**What:** Two deployment approaches for home device:
1. **Single container**: Extends docker-mailserver with all components in one image.
2. **Compose stack**: Separate containers for Postfix, Dovecot, Radicale, webmail.

**When to use:**
- **Single container**: Resource-constrained devices (RPi4 with 4GB RAM), simplicity priority.
- **Compose stack**: More powerful home servers (8GB+ RAM), component flexibility priority.

**Trade-offs:**

**Single Container:**
- **Pro:** Lower memory overhead (shared processes, no duplicate libraries).
- **Pro:** Simpler deployment (single `docker run` command).
- **Pro:** Easier updates (single image pull).
- **Con:** All components must restart together (no isolated updates).
- **Con:** Harder to customize individual components.

**Compose Stack:**
- **Pro:** Independent scaling and updates per component.
- **Pro:** Easier to swap components (e.g., replace Roundcube with Rainloop).
- **Pro:** Resource limits per container for better control.
- **Con:** Higher memory overhead (duplicate base images, separate processes).
- **Con:** More complex networking (inter-container communication).

### Pattern 5: Build-Time Configuration via GitHub Actions

**What:** Users fork repository, configure YAML file, trigger GitHub Actions workflow that builds custom multi-arch Docker images with their configuration baked in.

**When to use:** When distributing a configurable product that users deploy to their own infrastructure.

**Trade-offs:**
- **Pro:** No runtime configuration complexity; images are ready-to-run.
- **Pro:** User-specific builds can optimize for their exact component selection.
- **Pro:** Multi-arch builds (amd64 for cloud, arm64 for RPi) handled automatically.
- **Con:** Re-build required for configuration changes (not runtime reconfigurable).
- **Con:** GitHub Actions minutes consumption (mitigated by caching).
- **Con:** Users need GitHub account and basic understanding of workflows.

**Example workflow:**
```yaml
# .github/workflows/build-home.yml
name: Build Home Device Image
on:
  workflow_dispatch:
    inputs:
      enable_caldav:
        description: 'Enable CalDAV/CardDAV'
        required: true
        type: boolean
      webmail_choice:
        description: 'Webmail interface'
        required: true
        type: choice
        options:
          - roundcube
          - rainloop
          - none

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        platform: [linux/amd64, linux/arm64]
    steps:
      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v5
        with:
          platforms: ${{ matrix.platform }}
          build-args: |
            ENABLE_CALDAV=${{ inputs.enable_caldav }}
            WEBMAIL=${{ inputs.webmail_choice }}
          tags: ghcr.io/${{ github.repository }}/darkpipe-home:latest
          push: true
```

### Pattern 6: DNS Automation with Pre-Deployment Validation

**What:** CLI tool generates required DNS records (MX, SPF, DKIM, DMARC) from configuration, integrates with DNS provider APIs to apply them, and validates before deploying mail services.

**When to use:** Always. Email delivery fails without correct DNS configuration.

**Trade-offs:**
- **Pro:** Eliminates most common deployment error (incorrect DNS).
- **Pro:** Provider-agnostic abstraction over various DNS APIs.
- **Con:** Requires DNS provider API credentials (security consideration).
- **Con:** Not all DNS providers have APIs (fallback to manual instructions).

**Example:**
```bash
# Generate DNS records from config
$ darkpipe dns generate --config darkpipe.yaml
Generated DNS records for example.com:
  MX     10 relay.example.com
  TXT    "v=spf1 ip4:203.0.113.10 -all"
  TXT    (DKIM key for selector 'default')
  TXT    "_dmarc" "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"

# Validate before deployment
$ darkpipe dns validate --config darkpipe.yaml
✓ MX record points to relay.example.com (203.0.113.10)
✓ SPF record authorizes relay IP
✗ DKIM record not found (deployment will fail)
✗ DMARC record missing

# Apply via DNS provider API
$ darkpipe dns apply --config darkpipe.yaml --provider cloudflare
Applied 4 DNS records to Cloudflare
Waiting for propagation (this may take up to 5 minutes)...
✓ All records validated
```

## Data Flow

### Inbound Mail Flow (Internet → Mailbox)

```
┌─────────────┐
│ Sender's    │
│ Mail Server │
└──────┬──────┘
       │ SMTP (port 25)
       ↓
┌─────────────────────────────────────────────┐
│ Cloud Relay                                  │
│ 1. Postfix receives on port 25              │
│ 2. TLS negotiation (STARTTLS)               │
│ 3. Anti-spam checks (basic)                 │
│ 4. Enqueue to RAM spool                     │
│ 5. Dequeue immediately (if transport up)    │
└──────┬──────────────────────────────────────┘
       │ SMTP over WireGuard tunnel (10.8.0.2:25)
       │ OR SMTP over mTLS persistent connection
       ↓
┌─────────────────────────────────────────────┐
│ Home Device                                  │
│ 6. Transport endpoint receives              │
│ 7. Forward to local Postfix (localhost:25)  │
│ 8. Postfix applies local rules              │
│ 9. MDA (Dovecot LMTP) delivers to mailbox   │
│ 10. Mail stored in Docker volume             │
└──────┬──────────────────────────────────────┘
       │ IMAP/POP3 (ports 993/995)
       ↓
┌─────────────┐
│ User's Mail │
│ Client      │
└─────────────┘
```

**Critical decision points:**
- **Step 4-5**: If transport is down, queue in RAM. If RAM threshold exceeded, overflow to encrypted S3 storage.
- **Step 7**: Authentication happens here via WireGuard tunnel (network-level) or mTLS (application-level).
- **Step 8**: Home device applies user-specific filtering, spam rules, etc. Cloud relay does minimal processing.

### Outbound Mail Flow (Mailbox → Internet)

```
┌─────────────┐
│ User's Mail │
│ Client      │
└──────┬──────┘
       │ SMTP Submission (port 587 with AUTH)
       ↓
┌─────────────────────────────────────────────┐
│ Home Device                                  │
│ 1. Postfix receives on port 587             │
│ 2. User authentication (SASL)               │
│ 3. DKIM signing                              │
│ 4. Route via transport to cloud relay       │
└──────┬──────────────────────────────────────┘
       │ SMTP over WireGuard tunnel (10.8.0.1:587)
       │ OR SMTP over mTLS persistent connection
       ↓
┌─────────────────────────────────────────────┐
│ Cloud Relay                                  │
│ 5. Receive from home via transport          │
│ 6. Verify source (WireGuard peer/mTLS cert) │
│ 7. Rewrite headers (hide home IP)           │
│ 8. Send to destination via SMTP (port 25)   │
└──────┬──────────────────────────────────────┘
       │ SMTP (port 25)
       ↓
┌─────────────┐
│ Recipient's │
│ Mail Server │
└─────────────┘
```

**Critical decision points:**
- **Step 4**: Home device trusts cloud relay to send on its behalf. Cloud relay IP must have good reputation.
- **Step 6**: Cloud relay MUST verify source via transport authentication to prevent open relay abuse.
- **Step 7**: Cloud relay rewrites `Received:` headers to show relay IP, not home IP (privacy).

### Calendar/Contact Sync Flow (CalDAV/CardDAV)

```
┌─────────────┐
│ User's      │
│ Calendar    │
│ Client      │
└──────┬──────┘
       │ HTTPS (CalDAV/CardDAV)
       ↓
┌─────────────────────────────────────────────┐
│ Home Device (via WireGuard or VPN)          │
│ 1. Radicale receives HTTPS request          │
│ 2. Authentication (HTTP Basic/Digest)       │
│ 3. Read/write iCal/vCard files              │
│ 4. Store in Docker volume                   │
└─────────────────────────────────────────────┘
```

**Note:** CalDAV/CardDAV do NOT route through cloud relay. Direct connection to home device required (WireGuard tunnel or VPN).

### Update/Upgrade Flow

```
┌─────────────────────────────────────────────┐
│ 1. User triggers update                     │
│    $ docker pull darkpipe/home:latest       │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 2. Docker pulls new image                   │
│    (configuration baked into image)         │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 3. Health check on new container            │
│    Wait until new container responds        │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 4. Route traffic to new container           │
│    (load balancer or Docker networking)     │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 5. Graceful shutdown of old container       │
│    Wait for in-flight mail to complete      │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 6. Remove old container                     │
│    Data persists in Docker volumes          │
└─────────────────────────────────────────────┘
```

**Critical considerations:**
- **Configuration persistence**: User config stored in Docker volumes (`.env` files or YAML), not baked into images. Alternative: rebuild custom image via GitHub Actions.
- **Zero-downtime strategy**: Use Docker Compose with `update_config` for rolling updates, or blue-green deployment with Nginx routing.
- **Database migrations**: Mail formats rarely change, but if schema updates required, include migration scripts in entrypoint.

## Scaling Considerations

| Concern | Single User (10 emails/day) | Power User (100 emails/day) | Small Organization (1000 emails/day) |
|---------|----------------------------|------------------------------|--------------------------------------|
| **Cloud Relay** | 512MB VPS ($3-5/mo) | 1GB VPS ($5-10/mo) | 2GB VPS ($10-20/mo) with horizontal scaling |
| **Home Device** | RPi4 4GB | RPi4 8GB or NUC | Dedicated server or TrueNAS Scale |
| **Transport Layer** | Single WireGuard tunnel | Single WireGuard tunnel | Multiple WireGuard tunnels or mTLS connection pool |
| **Storage** | 10GB Docker volume | 50GB Docker volume | 500GB+ with S3 archival |
| **Container Strategy** | Single container | Single container or Compose | Compose stack with resource limits |

### Scaling Priorities

1. **First bottleneck: Transport layer reconnection**
   - **What breaks first:** If home device loses connection frequently (unreliable network), cloud relay queue fills up.
   - **How to fix:** Implement robust reconnection logic with exponential backoff. Add monitoring/alerting for transport state. Consider overflow to S3-compatible storage (e.g., Storj) for temporary queue persistence.

2. **Second bottleneck: Home device CPU (spam filtering)**
   - **What breaks next:** SpamAssassin and ClamAV are CPU-intensive on RPi4.
   - **How to fix:** Move heavy spam filtering to cloud relay (breaks privacy model) OR disable ClamAV (rely on sender's scanning) OR upgrade to more powerful home device (NUC with i5+).

3. **Third bottleneck: Cloud relay reputation**
   - **What breaks next:** If cloud relay IP gets blacklisted (compromised account sending spam), all mail delivery fails.
   - **How to fix:** Implement rate limiting, strict authentication, DMARC monitoring. Use multiple cloud relays with DNS round-robin. Consider dedicated IP from VPS provider.

## Anti-Patterns

### Anti-Pattern 1: Storing Mail Persistently in Cloud Relay

**What people do:** Configure cloud relay with local mailboxes, thinking it's a backup.

**Why it's wrong:** Defeats the entire privacy model of DarkPipe. Cloud provider has access to plaintext mail. Increases attack surface. Regulatory compliance issues (GDPR, etc.).

**Do this instead:** Cloud relay should be 100% stateless. Only ephemeral queue in RAM with optional encrypted overflow to S3. If user wants cloud backup, they should use encrypted backup of home device volumes.

### Anti-Pattern 2: Using Let's Encrypt Certificates for Internal Transport

**What people do:** Request Let's Encrypt certs for both cloud relay (correct) and internal WireGuard IPs (wrong).

**Why it's wrong:** Let's Encrypt requires public DNS validation. Internal IPs (10.8.0.x) cannot be validated. Also, WireGuard doesn't use TLS (has its own encryption), and mTLS should use internal CA for client certs.

**Do this instead:**
- **Cloud relay**: Let's Encrypt certs via Certbot for public SMTP (STARTTLS).
- **WireGuard transport**: No TLS needed; WireGuard provides encryption.
- **mTLS transport**: Internal CA for mutual authentication certificates.

### Anti-Pattern 3: Running Certbot Inside Docker Container Without Volume Persistence

**What people do:** Include Certbot in Docker image, run cert renewal inside container, forget to persist `/etc/letsencrypt`.

**Why it's wrong:** Certificates are lost on container restart. Let's Encrypt has rate limits (5 certs per domain per week). Container restart triggers new cert request, hitting rate limit, breaking mail delivery.

**Do this instead:** Mount `/etc/letsencrypt` as Docker volume. Use Certbot renewal hooks to reload Postfix. Alternative: use external cert management (certbot on host, mount certs into container).

### Anti-Pattern 4: Baking Secrets Into Docker Images

**What people do:** Include DKIM private keys, WireGuard private keys, passwords in Dockerfile or config files committed to Git.

**Why it's wrong:** Secrets exposed in image layers (even if deleted in later layer). Anyone with image access has secrets. GitHub Actions logs may leak secrets.

**Do this instead:** Secrets passed via environment variables at runtime or Docker secrets (Swarm) or mounted from encrypted volumes. Build system generates keys at deployment time, never stores in Git.

### Anti-Pattern 5: No Health Checks in Docker Compose

**What people do:** Deploy containers without `healthcheck` configuration, assuming "if container runs, it's healthy."

**Why it's wrong:** Postfix/Dovecot may start but fail to bind to ports (port conflict). Service may be running but rejecting connections (config error). Load balancer routes traffic to unhealthy container.

**Do this instead:**
```yaml
services:
  mail:
    image: darkpipe/home:latest
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "25"]  # Check SMTP port
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s  # Allow time for Postfix to start
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **DNS Providers** | REST API (Cloudflare, Route53, etc.) | Required for automated record creation; fallback to manual instructions if no API |
| **S3-Compatible Storage** | AWS SDK (works with Storj, Backblaze B2, MinIO) | For encrypted queue overflow; optional feature |
| **Let's Encrypt** | ACME protocol via Certbot | Cloud relay only; automatic renewal with DNS-01 or HTTP-01 challenge |
| **Monitoring/Alerting** | Prometheus metrics + Grafana OR simple health endpoint | For transport layer status, queue depth, delivery rates |
| **VPS Providers** | Manual deployment or Terraform | User deploys cloud relay to their chosen provider |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| **Cloud Relay ↔ Home Device** | SMTP over WireGuard (encrypted tunnel) OR SMTP over mTLS (persistent connection) | Primary data path; must be highly reliable |
| **Postfix ↔ Dovecot (Home)** | LMTP (Local Mail Transfer Protocol) over localhost | Standard MTA-to-MDA handoff; no authentication needed (localhost trusted) |
| **Webmail ↔ Dovecot (Home)** | IMAP over localhost | Roundcube/Rainloop connects via IMAP to Dovecot on same host |
| **CalDAV/CardDAV ↔ Storage** | File I/O to Docker volume | Radicale stores iCal/vCard files directly; no database |
| **GitHub Actions ↔ User Config** | YAML file in user's forked repo | User commits config, triggers workflow, receives custom image |

## Build Order Implications

Based on dependency analysis, recommended build order for development:

### Phase 1: Foundation (No Dependencies)
1. **Configuration schema and validation** - Defines data structures for all other components.
2. **DNS automation library** - Standalone; no dependencies; critical for all deployments.
3. **Build system (GitHub Actions workflows)** - Enables user-driven builds from day one.

### Phase 2: Transport Layer (Depends on Config)
4. **WireGuard tunnel setup** - Simpler than mTLS; implement first.
5. **mTLS persistent connection** - Alternative transport; shares config schema with WireGuard.

### Phase 3: Cloud Relay (Depends on Transport)
6. **Minimal Postfix relay container** - Core cloud relay; depends on transport to forward mail.
7. **Ephemeral queue with RAM spool** - Cloud relay needs queue; start with RAM-only (simpler).
8. **S3-compatible overflow storage** - Optional feature; add after RAM queue works.
9. **Certbot automation** - Cloud relay needs public TLS; integrate Let's Encrypt.

### Phase 4: Home Device (Depends on Transport)
10. **Postfix + Dovecot single container** - Core home mail server; simplest deployment.
11. **Docker Compose stack alternative** - Multi-container option; shares config with single container.
12. **CalDAV/CardDAV integration (Radicale)** - Optional component; depends on home device foundation.
13. **Webmail interface (Roundcube)** - Optional component; depends on Dovecot IMAP.

### Phase 5: User Experience (Depends on All)
14. **CLI wizard for initial setup** - Guides user through config creation; needs understanding of all components.
15. **Update/upgrade automation** - Zero-downtime updates; needs full system understanding.
16. **Monitoring and health checks** - Observability across all components.

### Dependency Graph

```
Configuration Schema
  ├─→ DNS Automation
  ├─→ Build System
  ├─→ WireGuard Setup
  │     ├─→ Cloud Relay (Postfix)
  │     │     ├─→ Ephemeral Queue
  │     │     │     └─→ S3 Overflow
  │     │     └─→ Certbot
  │     └─→ Home Device (Postfix+Dovecot)
  │           ├─→ Compose Stack Alternative
  │           ├─→ CalDAV/CardDAV
  │           └─→ Webmail
  ├─→ mTLS Setup (Alternative to WireGuard)
  │     └─→ (same downstream as WireGuard)
  └─→ CLI Wizard (needs all components)
        └─→ Update/Upgrade System
              └─→ Monitoring
```

**Critical path:** Configuration → WireGuard → Cloud Relay → Home Device → CLI Wizard

**Parallel development opportunities:**
- DNS Automation and Build System can be developed in parallel.
- WireGuard and mTLS are alternatives; pick one for MVP, add the other later.
- CalDAV/CardDAV and Webmail are optional; can be developed after core mail works.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| **Cloud Relay (Postfix)** | HIGH | Well-documented standard practice. [Postfix docs](https://www.postfix.org/BASIC_CONFIGURATION_README.html) authoritative. |
| **WireGuard Transport** | HIGH | Mature protocol with excellent documentation. [Official WireGuard docs](https://www.wireguard.com/quickstart/) and [NAT traversal research](https://nordvpn.com/blog/achieving-nat-traversal-with-wireguard/) confirm patterns. |
| **mTLS Patterns** | MEDIUM | Well-established pattern but implementation details vary. [Medium article on connection handling](https://medium.com/@wolfroma/handling-mtls-connection-spikes-haproxy-tomcat-httpclient-ecab9f18707e) provides real-world guidance. |
| **Home Device (Docker Mailserver)** | HIGH | [docker-mailserver](https://github.com/docker-mailserver/docker-mailserver) is production-ready and well-documented for single-container deployment. |
| **CalDAV/CardDAV Integration** | MEDIUM-LOW | Separate applications (like Radicale) are the norm. No evidence of unified Postfix+Dovecot+CalDAV stack. [Dovecot mailing list discussions](https://dovecot.org/pipermail/dovecot/2022-October/125533.html) confirm this. |
| **S3-Compatible Queue** | MEDIUM | [Maddy mail server](https://maddy.email/reference/blob/s3/) demonstrates S3 blob storage for mail. Stalwart also supports S3. Pattern is proven but not mainstream. |
| **GitHub Actions Multi-Arch** | HIGH | [Official Docker docs](https://docs.docker.com/build/ci/github-actions/multi-platform/) and [multiple community examples](https://github.com/sredevopsorg/multi-arch-docker-github-workflow) provide clear patterns. |
| **DNS Automation** | MEDIUM | Provider APIs are well-documented (Cloudflare, Route53), but [DNS-PERSIST-01](https://www.certkit.io/blog/dns-persist-01) is new for 2026 and may not be widely supported yet. |
| **Zero-Downtime Updates** | MEDIUM | [Docker update patterns](https://oneuptime.com/blog/post/2026-01-06-docker-update-without-downtime/view) are documented, but mail-specific considerations (queue draining) need validation. |
| **RPi4 Resource Constraints** | MEDIUM | [Community reports](https://forums.raspberrypi.com/viewtopic.php?t=310308) confirm Docker Mailserver works on RPi4, but resource limits need empirical testing. |

## Open Questions Requiring Phase-Specific Research

1. **Queue overflow threshold**: What RAM threshold triggers S3 overflow? How to calculate based on average email size and delivery latency?

2. **Reconnection backoff strategy**: What backoff algorithm balances quick reconnection vs. not hammering cloud relay? Needs simulation/testing.

3. **CalDAV/CardDAV without separate service**: Is there a way to serve CalDAV/CardDAV from Dovecot directly, or is Radicale/Sabre always required?

4. **Certificate rotation without downtime**: How to rotate mTLS certificates (both cloud and home) without breaking persistent connection? Needs protocol-level investigation.

5. **Multi-user support on single home device**: Architecture assumes single user. What changes needed for family (5-10 users) on one RPi4?

6. **ARM64 performance**: Quantify SpamAssassin and ClamAV performance on RPi4 arm64 vs. amd64 NUC. May influence default component selection.

7. **DNS propagation delays**: How to handle initial deployment when DNS records take 24-48 hours to propagate? Does cloud relay need to queue mail for extended period?

8. **Backup strategy**: Where should automated backups fit in architecture? Cloud relay? Home device? User's responsibility?

## Sources

### High Confidence Sources (Official Documentation)
- [Postfix Basic Configuration](http://www.postfix.org/BASIC_CONFIGURATION_README.html)
- [Postfix Standard Configuration Examples](https://www.postfix.org/STANDARD_CONFIGURATION_README.html)
- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [Docker Multi-Platform Images with GitHub Actions](https://docs.docker.com/build/ci/github-actions/multi-platform/)
- [docker-mailserver GitHub Repository](https://github.com/docker-mailserver/docker-mailserver)
- [Certbot User Guide](https://eff-certbot.readthedocs.io/en/stable/using.html)
- [Let's Encrypt Challenge Types](https://letsencrypt.org/docs/challenge-types/)
- [Docker Compose Environment Variables](https://docs.docker.com/compose/how-tos/environment-variables/set-environment-variables/)
- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

### Medium Confidence Sources (Verified with Multiple Sources)
- [SMTP Relay Explained - Mailtrap 2026](https://mailtrap.io/blog/smtp-relay/)
- [How NAT Traversal Works - NordVPN](https://nordvpn.com/blog/achieving-nat-traversal-with-wireguard/)
- [WireGuard NAT Traversal - Nettica](https://nettica.com/nat-traversal-hole-punch/)
- [Email Forwarding FAQ - Forward Email](https://forwardemail.net/en/faq)
- [Docker Mailserver Basic Installation Tutorial](https://docker-mailserver.github.io/docker-mailserver/latest/examples/tutorials/basic-installation/)
- [Handling mTLS Connection Spikes - Medium](https://medium.com/@wolfroma/handling-mtls-connection-spikes-haproxy-tomcat-httpclient-ecab9f18707e)
- [Certificate Rotation Best Practices - Workik](https://workik.com/certificate-rotation-script-generator)
- [Multi-Arch Docker Images with GitHub Actions - Red Hat](https://developers.redhat.com/articles/2023/12/08/build-multi-architecture-container-images-github-actions)
- [DNS-PERSIST-01 Validation - CertKit](https://www.certkit.io/blog/dns-persist-01)
- [How to Update Docker Container with Zero Downtime - Atlantic.Net](https://www.atlantic.net/vps-hosting/how-to-update-docker-container-with-zero-downtime/)
- [How Email Works: MUA, MSA, MTA, MDA - Oxilor](https://oxilor.com/blog/how-does-email-work)
- [Docker Container Resource Limits - OneUpTime](https://oneuptime.com/blog/post/2026-01-30-docker-container-resource-limits/view)

### Community Sources (Lower Confidence, Needs Validation)
- [Raspberry Pi Email Server Options - Forward Email Blog](https://forwardemail.net/en/blog/open-source/raspberry-pi-email-server)
- [Raspberry Pi docker-mailserver Discussion - RPi Forums](https://forums.raspberrypi.com/viewtopic.php?t=310308)
- [Adding CalDAV/CardDAV Next to Dovecot - Dovecot Mailing List](https://dovecot.org/pipermail/dovecot/2022-October/125533.html)
- [S3-Compatible Storage Solutions 2026 - Cloudian](https://cloudian.com/guides/s3-storage/best-s3-compatible-storage-solutions-top-5-in-2026/)
- [Maddy Mail Server S3 Storage](https://maddy.email/reference/blob/s3/)

---

**Architecture research for:** Cloud-Relay + Home-Device Email System (DarkPipe)
**Researched:** 2026-02-08
**Overall Confidence:** MEDIUM - Core patterns (Postfix, WireGuard, Docker) are HIGH confidence from official sources. Advanced patterns (S3 queue overflow, CalDAV integration, zero-downtime updates) are MEDIUM confidence requiring validation during implementation.

# Technology Stack

**Project:** DarkPipe
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Executive Summary

DarkPipe requires a split-stack architecture: minimal cloud relay (15-50MB container) and full-featured home server. Stalwart or Maddy for Rust/Go single-binary deployments, Postfix for battle-tested minimal relay, WireGuard kernel module for transport, step-ca for internal PKI, and Go for orchestration glue code.

---

## Cloud Relay Stack (Minimal Container)

### SMTP Relay/MTA

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Postfix (Alpine)** | 3.7.4+ | Internet-facing SMTP relay | Battle-tested, 15-30MB container, stable API, excellent documentation, standard choice for minimal relays | HIGH |
| **Maddy** | 0.8.2 | All-in-one Go mail server | Single 15MB binary, Go-native, ~15MB memory footprint, built-in DKIM/SPF/DMARC, excellent for minimal cloud relay | MEDIUM-HIGH |
| **Haraka** | Latest | Node.js SMTP relay | Async/event-driven, plugin architecture, good for custom filtering, but Node.js runtime adds ~50MB vs Go/Rust | MEDIUM |

**Recommendation:** Use **Postfix** for cloud relay. Proven minimal footprint, decades of production hardening, perfect match for "receive and forward" relay pattern. Maddy is excellent backup if Go-native stack preferred.

**Why not Stalwart for relay:** Stalwart is a full mail server (150MB+ with dependencies). Overkill for cloud relay that only needs SMTP receive + forward.

### Container Base Image

| Technology | Size | Purpose | Why Recommended | Confidence |
|------------|------|---------|-----------------|------------|
| **Alpine Linux** | ~5MB | Minimal container base | Package manager available, musl libc, broad hardware support, proven for Postfix/WireGuard, debugging tools available | HIGH |
| **Distroless (Debian)** | ~2MB | Minimal security-focused base | No shell/package manager, better security posture, but harder to debug, requires static binaries | MEDIUM-HIGH |
| **Scratch** | 0MB | Empty base for static binaries | Smallest possible, but no CA certs, no DNS resolution, requires bundling everything | MEDIUM |

**Recommendation:** Use **Alpine** for cloud relay. Trade 3MB for massive operational benefits: shell for debugging, package manager for updates, standard troubleshooting tools. Security through minimal attack surface + regular updates > distroless complexity.

**For Go glue code containers:** Use distroless or scratch with static binary.

---

## Home Server Stack (Full-Featured)

### Mail Server (User-Selectable)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Stalwart** | 0.15.4 (v1.0.0 Q2 2026) | All-in-one Rust mail server | Single binary, JMAP/IMAP4rev2/POP3/SMTP, built-in CalDAV/CardDAV, SQLite/RocksDB storage, memory-safe, REST API, excellent ARM64 support | HIGH |
| **Maddy** | 0.8.2 | All-in-one Go mail server | Single binary, replaces Postfix+Dovecot+OpenDKIM, 15MB memory, built-in DKIM/SPF/DMARC, simple config | HIGH |
| **Postfix + Dovecot** | Postfix 3.7+, Dovecot 2.3+ | Traditional split MTA+MDA | Most battle-tested, maximum flexibility, best documentation, proven ARM64 support, but complex configuration | HIGH |

**Recommendation:** Default to **Stalwart** for modern deployments. Single binary, modern protocols (JMAP), built-in calendar/contacts server, excellent security (memory-safe Rust). **Maddy** for Go preference or tighter resource constraints. **Postfix+Dovecot** for maximum compatibility or existing expertise.

**Why Stalwart over Maddy:**
- Built-in CalDAV/CardDAV (eliminates separate Radicale/Baikal)
- JMAP support (modern protocol, better than IMAP for sync)
- v1.0.0 due Q2 2026 with stable schema/auto-upgrades
- REST API for automation

**Why Maddy over Stalwart:**
- Smaller memory footprint (~15MB vs Stalwart's ~50MB)
- Simpler if only need email (no calendar/contacts)
- Go codebase if team prefers Go

**Why Postfix+Dovecot:**
- 20+ years production hardening
- Most documentation/examples available
- Maximum flexibility for complex routing
- ISP/enterprise familiarity

### Calendar/Contacts Server (If Not Using Stalwart)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Radicale** | 3.x | Lightweight CalDAV/CardDAV | Python, minimal dependencies, file-system storage, 5-10MB memory, simple setup, GPLv3 | MEDIUM-HIGH |
| **Baikal** | Latest | PHP CalDAV/CardDAV with web UI | Admin interface, MySQL/SQLite, more polished UI, 5M+ Docker pulls, multi-arch support | MEDIUM-HIGH |

**Recommendation:** Skip separate calendar server if using **Stalwart** (built-in). Otherwise use **Baikal** for better admin UX and proven Docker support. **Radicale** if minimizing dependencies (Python vs PHP).

**Skip if:** Using Stalwart (built-in CalDAV/CardDAV).

### Webmail Client (Optional)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **SnappyMail** | Latest | Modern lightweight webmail | 138KB download (Brotli), 99% Lighthouse score, no database required, actively maintained RainLoop fork, significantly faster than Roundcube | MEDIUM-HIGH |
| **Roundcube** | Latest | Traditional PHP webmail | Most mature, extensive plugin ecosystem, fpm-alpine variant available, 58x more popular than SnappyMail, but heavier | HIGH |

**Recommendation:** Use **SnappyMail** for modern, minimal deployment. Dramatically faster, no database, minimal resource usage. Use **Roundcube** only if need specific plugins or organizational familiarity.

**Why SnappyMail over Roundcube:**
- 138KB vs multi-MB page loads
- No database required (simpler deployment)
- Better mobile experience
- Actively maintained (Roundcube slower development)

**Why Roundcube over SnappyMail:**
- Massive plugin ecosystem
- 20+ years of production use
- More organizational adoption
- Better documentation

---

## Transport Security Stack

### Encrypted Relay ↔ Home Transport

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **WireGuard (Kernel)** | Kernel module | VPN tunnel cloud ↔ home | Lowest latency, best throughput on ARM64, ~500% faster than userspace on Ethernet, kernel merged since Linux 5.6, standard WireGuard tools | HIGH |
| **mTLS over SMTP** | TLS 1.3 | Application-layer mutual auth | No separate tunnel, SMTP-native, works with Postfix relay_clientcerts, simpler architecture, but requires certificate distribution | MEDIUM-HIGH |

**Recommendation:** Use **WireGuard kernel module** for primary transport. Proven performance advantage on ARM64 (Nord Security study: kernel dramatically faster than userspace on RPi4). mTLS over SMTP as fallback/alternative for users unable to run kernel modules.

**Implementation:** WireGuard kernel module + custom Go relay daemon using emersion/go-smtp for SMTP protocol handling.

**Why WireGuard kernel over userspace:**
- 500% throughput improvement on ARM64 Ethernet (Nord Security testing)
- Lower power consumption (critical for RPi4)
- Better latency under load
- Kernel module standard since 5.6, widely available

**Why WireGuard over pure mTLS:**
- Simpler than certificate distribution to Postfix
- Tunnel isolates all relay ↔ home traffic (not just SMTP)
- Better for future multi-protocol support
- Easier NAT traversal

**Why mTLS as alternative:**
- No kernel dependency (works in restricted environments)
- SMTP-native (no separate tunnel)
- Standard Postfix configuration
- Better for environments blocking VPN ports

### Go WireGuard Libraries

| Library | Purpose | Why Recommended | Confidence |
|---------|---------|-----------------|------------|
| **golang.zx2c4.com/wireguard/wgctrl** | Kernel WireGuard control | Official Go bindings, standard for kernel module management | HIGH |
| **wireguard-go** | Userspace fallback | Official userspace implementation, use only when kernel unavailable | HIGH |

**Recommendation:** Use **wgctrl** for kernel control. Only use **wireguard-go** userspace as fallback for environments without kernel module access.

---

## Certificate Management Stack

### Public Certificates (Internet-Facing Relay)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Certbot** | Latest | Let's Encrypt ACME client | Official EFF tool, Alpine Docker image available, standard automation, supports DNS-01 and HTTP-01 challenges | HIGH |
| **acme.sh** | Latest | Alternative ACME client | Lightweight shell script, broader DNS provider support, smaller footprint than Certbot | MEDIUM-HIGH |

**Recommendation:** Use **Certbot** in sidecar container. Standard choice, excellent documentation, official Docker image (certbot/certbot), proven DNS-01 automation for wildcard certs.

**DNS-01 challenge pattern:** CNAME _acme-challenge subdomain to validation-specific server (security best practice per Let's Encrypt docs).

### Internal CA (Relay ↔ Home Transport)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **step-ca** | 0.29.0 | Private certificate authority | Modern ACME server, short-lived certs with auto-renewal, SSH CA support, OAuth/OIDC integration, multiple database backends, REST API, Apache 2.0 license | HIGH |
| **cfssl** | Latest | CloudFlare's PKI toolkit | Simple JSON API, proven in production, but less active development than step-ca | MEDIUM |

**Recommendation:** Use **step-ca** for internal CA. Modern design, built-in ACME server (works with Certbot/Caddy), short-lived certificate best practices, active development, excellent documentation.

**Why step-ca over cfssl:**
- Built-in ACME server (automated renewal)
- Short-lived certificates (security best practice)
- Active development (SmallStep Labs)
- Better OAuth/SSO integration options
- Badger/BoltDB/Postgres/MySQL backends

**Why step-ca over manual OpenSSL:**
- Automated renewal eliminates manual process
- ACME protocol standardization
- REST API for automation
- Short-lived certs by default (better security)

### Go mTLS Libraries

| Library | Purpose | Why Recommended | Confidence |
|---------|---------|-----------------|------------|
| **crypto/tls (stdlib)** | Standard library TLS | Built-in, zero dependencies, RequireAndVerifyClientCert support, sufficient for mTLS | HIGH |
| **github.com/stephen-fox/mtls** | Certificate generation helper | Simplifies cert/key pair generation for testing/development | MEDIUM |

**Recommendation:** Use **crypto/tls** from Go standard library. Zero dependencies, well-tested, sufficient for production mTLS. Use mtls library only for development/testing certificate generation.

---

## Orchestration & Glue Code Stack

### Primary Language

| Language | Binary Size | ARM64 Support | SMTP Ecosystem | Why Recommended | Confidence |
|----------|-------------|---------------|----------------|-----------------|------------|
| **Go** | 2-5MB static | Excellent | Excellent (emersion/go-smtp) | Simple deployment, fast compilation, excellent stdlib, standard for cloud-native tools, mature SMTP libraries, great ARM64 cross-compilation | HIGH |
| **Rust** | 2-5MB static | Excellent | Growing (minismtp, Stalwart SMTP) | Smaller binaries with optimization, memory safety, but slower compilation, steeper learning curve | MEDIUM-HIGH |
| **Python** | N/A (interpreted) | Excellent | Excellent (aiosmtpd) | Rapid development, but requires runtime (~50MB+), slower performance, not ideal for minimal containers | MEDIUM |

**Recommendation:** Use **Go** for all orchestration and glue code. Single static binary deployment, excellent ARM64 cross-compilation, mature SMTP ecosystem (emersion/go-smtp), fast compilation, standard for cloud-native tools (similar to Docker, Kubernetes).

**Why Go over Rust:**
- Faster compilation (critical for CI/CD iteration)
- Simpler syntax (easier contributor onboarding)
- Larger SMTP library ecosystem
- Standard choice for cloud-native infrastructure
- Better error handling for network operations
- Comparable binary size (2-5MB range)

**Why Go over Python:**
- Static binary (no runtime dependency)
- ~10x smaller container footprint
- Better performance for network I/O
- Easier cross-compilation for ARM64

### Go SMTP Libraries

| Library | Version | Purpose | Why Recommended | Confidence |
|---------|---------|---------|-----------------|------------|
| **github.com/emersion/go-smtp** | 0.24.0 | SMTP client & server | Active development (1.1K projects depend on it), RFC 5321 compliant, ESMTP extensions, UTF-8 support, LMTP, MIT license, 2K+ stars | HIGH |
| **github.com/mhale/smtpd** | Latest | Minimal SMTP server | Simple API, but inactive (no updates in 12+ months), not recommended for new projects | LOW |

**Recommendation:** Use **emersion/go-smtp** exclusively. Active maintenance, comprehensive feature set, production-proven (1.1K dependents), RFC compliant, excellent for building custom relay logic.

---

## Build & CI/CD Stack

### Multi-Architecture Docker Builds

| Technology | Purpose | Why Recommended | Confidence |
|------------|---------|-----------------|------------|
| **docker/setup-qemu-action** | ARM64 emulation on amd64 runners | Standard GitHub Actions approach, works but slower (~3-5x) | HIGH |
| **Native ARM64 runners** | Build ARM64 on ARM64 hardware | Fastest (no emulation overhead), but requires ARM64 runner access (GitHub-hosted ARM64 runners or self-hosted) | HIGH |
| **docker/build-push-action** | Multi-platform builds | Official Docker action, supports buildx, push to registries | HIGH |

**Recommendation:** Use **QEMU emulation** for simplicity (no cost, standard GitHub Actions). Optimize to **native ARM64 runners** if build time becomes bottleneck. Use matrix strategy to build each platform separately and merge manifests.

**Pattern:**
```yaml
strategy:
  matrix:
    platform: [linux/amd64, linux/arm64]
```

Build each platform on dedicated runner (or QEMU), push by digest, merge job creates manifest list with `docker buildx imagetools create`.

### User-Selectable Components

**Build Strategy:** Use GitHub Actions build matrix with boolean inputs for component selection:

```yaml
on:
  workflow_dispatch:
    inputs:
      mail_server:
        type: choice
        options: [stalwart, maddy, postfix-dovecot]
      calendar_server:
        type: choice
        options: [none, radicale, baikal]
      webmail:
        type: choice
        options: [none, snappymail, roundcube]
```

Use Docker multi-stage builds with build args to select components. Each component as separate build stage, final image only includes selected components.

---

## Supporting Libraries (Go)

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| **golang.zx2c4.com/wireguard/wgctrl** | Latest | WireGuard kernel control | Managing WireGuard interfaces from Go | HIGH |
| **github.com/emersion/go-smtp** | 0.24.0 | SMTP protocol implementation | Custom relay logic, SMTP server/client | HIGH |
| **crypto/tls** (stdlib) | stdlib | TLS/mTLS implementation | Secure connections, certificate handling | HIGH |
| **github.com/spf13/cobra** | Latest | CLI framework | Building user-facing CLI tools | HIGH |
| **github.com/spf13/viper** | Latest | Configuration management | Loading config from files/env/flags | HIGH |

---

## Installation & Quickstart

### Cloud Relay Container (Postfix + Go Glue)

**Dockerfile pattern:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay ./cmd/relay

FROM alpine:3.21
RUN apk add --no-cache postfix ca-certificates wireguard-tools
COPY --from=builder /build/relay /usr/local/bin/relay
COPY postfix-config/ /etc/postfix/
EXPOSE 25
CMD ["/usr/local/bin/relay"]
```

**Result:** ~25-30MB container (Alpine 5MB + Postfix 15MB + WireGuard tools 5MB + Go binary 2-5MB).

### Home Server Container (Stalwart)

**Dockerfile pattern:**
```dockerfile
FROM stalwartlabs/stalwart:0.15.4
# Stalwart single binary, ~50MB total
# Add WireGuard for relay connection
RUN apt-get update && apt-get install -y wireguard-tools && rm -rf /var/lib/apt/lists/*
```

**Result:** ~70MB container (Stalwart includes all mail + calendar/contacts).

### Development Dependencies

```bash
# Go development
go install github.com/emersion/go-smtp@latest
go install golang.zx2c4.com/wireguard/wgctrl@latest

# Container building
docker buildx create --use

# Certificate management (local testing)
brew install step  # or: wget https://dl.step.sm/gh-release/cli/docs-ca-install/v0.29.0/step_linux_0.29.0_amd64.tar.gz
```

---

## Alternatives Considered

| Category | Recommended | Alternative | When to Use Alternative | Confidence |
|----------|-------------|-------------|------------------------|------------|
| Cloud Relay | Postfix | Haraka | Need custom JavaScript plugins, async event processing | MEDIUM |
| Cloud Relay | Postfix | Maddy | Prefer Go-native stack, comfortable with less battle-tested option | MEDIUM-HIGH |
| Home Mail Server | Stalwart | Maddy | Tighter memory constraints (<50MB), don't need CalDAV/CardDAV | HIGH |
| Home Mail Server | Stalwart | Postfix+Dovecot | Maximum compatibility, existing expertise, complex routing rules | HIGH |
| Calendar/Contacts | Baikal (if not Stalwart) | Radicale | Minimizing dependencies (Python vs PHP), file-based storage preference | MEDIUM-HIGH |
| Webmail | SnappyMail | Roundcube | Need specific plugins, organizational standard, more conservative choice | HIGH |
| Transport | WireGuard Kernel | mTLS-only | Kernel module unavailable, restricted environment, VPN ports blocked | MEDIUM-HIGH |
| CA | step-ca | cfssl | Simpler JSON API preference, less feature-rich CA sufficient | MEDIUM |
| Language | Go | Rust | Team Rust expertise, willing to accept slower compile times for memory safety | MEDIUM-HIGH |
| Base Image | Alpine | Distroless | Maximum security posture, comfortable with no-shell debugging constraints | MEDIUM-HIGH |

---

## What NOT to Use

| Avoid | Why | Use Instead | Confidence |
|-------|-----|-------------|------------|
| **Exim** | Complex configuration, less modern, smaller community than Postfix | Postfix | HIGH |
| **Sendmail** | Ancient, notorious configuration complexity, security history | Postfix | HIGH |
| **RainLoop (original)** | Abandoned project (forked to SnappyMail), security concerns | SnappyMail | HIGH |
| **Mail-in-a-Box** | Opinionated all-in-one with Ubuntu dependency, not containerizable, forces specific stack choices | Stalwart or Maddy | MEDIUM-HIGH |
| **BoringTun (Rust WireGuard)** | Userspace implementation, slower than kernel on ARM64, use only if kernel unavailable | WireGuard kernel module | HIGH |
| **wireguard-go (userspace)** | 3-5x slower than kernel, higher CPU usage, acceptable only as fallback | WireGuard kernel module | HIGH |
| **BerkleyDB** | Deprecated in Alpine (Oracle license change to AGPL-3.0), replaced with LMDB | LMDB (default in Alpine Postfix) | HIGH |
| **Python for relay glue** | Requires 50MB+ runtime, slower, larger containers | Go | HIGH |
| **Node.js for relay glue** | Requires 50MB+ runtime, async complexity for simple relay logic | Go | HIGH |

---

## Stack Patterns by Deployment

### Pattern 1: Minimal (Best for RPi4, 2GB RAM)
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Maddy (15MB binary) + SnappyMail (no DB) = ~40MB total
**Transport:** WireGuard kernel module
**Total footprint:** ~70MB containers + ~15MB RAM (Maddy)

### Pattern 2: Modern (Recommended for RPi4 4GB+)
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Stalwart 0.15.4 (70MB) with built-in CalDAV/CardDAV
**Transport:** WireGuard kernel module
**Total footprint:** ~100MB containers + ~50MB RAM (Stalwart)

### Pattern 3: Maximum Compatibility
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Postfix + Dovecot (~80MB) + Baikal (~30MB) + Roundcube (~40MB) = 150MB
**Transport:** mTLS over SMTP (no WireGuard)
**Total footprint:** ~180MB containers + ~100MB RAM

---

## Version Compatibility Notes

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Postfix 3.7+ | Alpine 3.20+ | Requires LMDB (BerkleyDB deprecated) |
| Stalwart 0.15.4 | Pre-v1.0 schema | Breaking changes from 0.14.x, read upgrade docs |
| step-ca 0.29.0 | Certbot latest | Full ACME server compatibility |
| WireGuard kernel | Linux 5.6+ | Kernel module mainlined, standard in modern kernels |
| emersion/go-smtp 0.24.0 | Go 1.20+ | Requires generics support |

---

## Sources

### Official Documentation (HIGH Confidence)
- [Stalwart Mail Server](https://stalw.art/mail-server/) - Official docs and latest release (v0.15.4)
- [Stalwart GitHub Releases](https://github.com/stalwartlabs/stalwart/releases) - Version 0.15.4, 2026-01-19
- [Maddy Mail Server](https://maddy.email/) - Official documentation
- [Maddy GitHub](https://github.com/foxcpp/maddy) - Version 0.8.2, 2026-01-14
- [emersion/go-smtp GitHub](https://github.com/emersion/go-smtp) - Version 0.24.0, 2025-08-05
- [step-ca GitHub](https://github.com/smallstep/certificates) - Version 0.29.0, 2025-12-03
- [Docker Multi-Platform Builds Official Docs](https://docs.docker.com/build/ci/github-actions/multi-platform/)
- [Let's Encrypt Challenge Types](https://letsencrypt.org/docs/challenge-types/) - ACME DNS-01 documentation
- [WireGuard Official](https://www.wireguard.com/) - Kernel module documentation

### Docker Hub & Container Images (HIGH Confidence)
- [Postfix Alpine Containers](https://github.com/bokysan/docker-postfix) - Multi-arch Alpine/Debian/Ubuntu
- [Stalwart Docker Hub](https://hub.docker.com/r/stalwartlabs/stalwart) - Official multi-arch images
- [Maddy Docker Hub](https://hub.docker.com/r/foxcpp/maddy) - Official image
- [Certbot Docker Hub](https://hub.docker.com/r/certbot/certbot) - Official EFF image
- [step-ca Docker Hub](https://hub.docker.com/r/smallstep/step-ca/) - Official Smallstep image
- [Roundcube Docker Hub](https://hub.docker.com/r/roundcube/roundcubemail/) - Official fpm-alpine variant
- [Baikal Docker Hub](https://hub.docker.com/r/ckulka/baikal) - Multi-arch support

### Community & Comparisons (MEDIUM Confidence)
- [Alpine, Distroless, or Scratch Comparison (Medium)](https://medium.com/@cloudwithusama/alpine-distroless-or-scratch-choosing-the-right-lightweight-base-image-f5b12dc5d4f6) - 2026 container base image comparison
- [Docker Image Size Reduction (OneUpTime Blog)](https://oneuptime.com/blog/post/2026-01-16-docker-reduce-image-size/view) - Alpine vs Distroless vs Scratch
- [WireGuard Kernel vs Userspace Performance (Nord Security)](https://nordsecurity.com/blog/wireguard-kernel-module-vs-user-space) - ARM64 benchmarks on Raspberry Pi 4
- [SnappyMail vs Roundcube (Forward Email Blog)](https://forwardemail.net/en/blog/open-source/webmail-email-clients) - 2026 webmail comparison
- [Go vs Rust in 2026 (Bitfield Consulting)](https://bitfieldconsulting.com/posts/rust-vs-go)
- [Building Multi-Platform Docker Images (Blacksmith)](https://www.blacksmith.sh/blog/building-multi-platform-docker-images-for-arm64-in-github-actions)
- [GotaTun Rust WireGuard (Mullvad)](https://mullvad.net/en/blog/announcing-gotatun-the-future-of-wireguard-at-mullvad-vpn) - 2026 Rust WireGuard developments

### GitHub & Library Documentation (MEDIUM-HIGH Confidence)
- [Haraka SMTP Server GitHub](https://github.com/haraka/Haraka) - Node.js SMTP relay
- [Radicale GitHub](https://github.com/Kozea/Radicale) - CalDAV/CardDAV Python server
- [Baikal GitHub](https://github.com/sabre-io/Baikal) - CalDAV/CardDAV PHP server
- [minismtp Rust Library](https://github.com/saefstroem/minismtp) - Minimal Rust SMTP server
- [mhale/smtpd Go Library](https://github.com/mhale/smtpd) - Inactive, not recommended
- [acme-dns GitHub](https://github.com/joohoi/acme-dns) - DNS-01 challenge automation

---

*Stack research for: DarkPipe Privacy Email*
*Researched: 2026-02-08*
*Researcher: GSD Project Researcher*

# Feature Research

**Domain:** Privacy-First Self-Hosted Email
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

#### Core Email Transport (SMTP/IMAP)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Inbound SMTP relay | Fundamental email receiving | Low | Standard Postfix/mail server functionality |
| Outbound SMTP relay | Fundamental email sending | Low | Required for any email system |
| IMAP server | Standard protocol for modern email clients | Medium | Dovecot is standard; IMAP vastly better UX than POP3 |
| SMTP Submission (port 587) | Modern email sending standard | Low | Required by all clients; port 25 blocked by ISPs |
| TLS for SMTP/IMAP | Encryption in transit | Low | Let's Encrypt makes this trivial; users expect https:// everywhere |
| STARTTLS support | Opportunistic encryption | Low | Standard for SMTP; implicit in modern stacks |

#### Email Authentication & Deliverability

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| SPF record generation | Gmail/Yahoo/MS mandate for bulk senders | Low | DNS TXT record; trivial if automated |
| DKIM signing | Industry baseline for email authentication | Medium | Requires key generation, DNS publication, Postfix/OpenDKIM config |
| DMARC policy setup | Mandated by major providers in 2024-2025; baseline in 2026 | Low | DNS TXT record; policy must align with SPF/DKIM |
| Reverse DNS (PTR) setup | Deliverability requirement; missing = instant spam folder | Low | Cloud relay must handle this; user device cannot |
| MX record configuration | Email routing to your server | Low | DNS configuration; expected to be automated |

#### Basic Mail Server Features

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Multiple mailboxes/users | Multi-user support is assumed | Low | Postfix/Dovecot handle this natively |
| Multiple domains | Many users consolidate multiple identities | Medium | Postfix virtual domains; DNS delegation per-domain |
| Mail aliases | Forwarding/catch-all addresses | Low | Standard Postfix virtual alias maps |
| Folder management | IMAP folders/labels | Low | Dovecot handles this; client-driven |
| Basic search | Finding emails in mailbox | Low | Dovecot FTS (full-text search) with Xapian or Solr |
| Spam filtering | Inbound spam protection | Medium | SpamAssassin/Rspamd standard; auto-learning improves accuracy |
| Greylisting | Reduces spam via temporary deferrals | Low | Postgrey or Rspamd greylisting module |
| Virus scanning | ClamAV integration expected | Low | Performance overhead; some users disable for personal use |

#### Webmail Basics

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Web-based email access | Not everyone wants native clients | Medium | Roundcube/SOGo are standard options |
| Email composition | Writing/sending via web | Low | Built into webmail solutions |
| Attachment handling | Drag-drop upload, download | Low | Modern webmail (Roundcube 1.5+) supports drag-drop |
| Mobile-responsive UI | Mobile devices dominate email access | Medium | Roundcube/SOGo responsive by default; verify on actual devices |
| HTML email rendering | Modern emails are HTML-heavy | Low | Built into webmail clients |
| Basic contacts integration | Accessing address book from compose | Medium | Requires CardDAV integration or webmail-native contacts |

#### Security Fundamentals

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| TLS certificate auto-renewal | Let's Encrypt requires 90-day renewal | Low | Certbot handles this; failure = service outage |
| Fail2ban or equivalent | Brute force protection | Low | Standard hardening; watches auth logs and bans IPs |
| Firewall rules | Only expose required ports | Low | ufw/iptables; 25, 587, 993, 443, 22 (if SSH) |

#### Basic Administration

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Web admin panel | GUI for non-technical users | Medium | Mail-in-a-Box, Mailu, Mailcow all provide this |
| Add/remove mailboxes | User management | Low | Admin panel or CLI |
| Quota management | Disk space limits per user | Low | Dovecot quota plugin |
| Backup configuration | Automated backups expected | Medium | Mail-in-a-Box uses Duplicity to S3; critical for disaster recovery |

### Differentiators (Competitive Advantage)

Features that set product apart. Not expected, but valued.

#### Privacy & Control Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Cloud-fronted architecture | Internet-facing SMTP without exposing home IP | High | **DarkPipe's core value prop**: cloud relay + home device split |
| User-owned hardware for storage | Email stored on user's device, not cloud provider | Medium | Differentiates from hosted solutions; requires secure transport |
| No vendor lock-in for storage | Users control their own hardware/data | Low | Natural consequence of architecture |
| Transparent relay operation | Users see what cloud relay does (audit logs) | Medium | Builds trust; critical for privacy-focused users |
| Minimal cloud footprint | Cloud relay doesn't store mail (forward-only) | Medium | Reduces privacy exposure; requires reliable home device |
| Data sovereignty | Email never leaves user's jurisdiction | Low | Marketing/trust angle; practical for some use cases |

#### TLS Enforcement & Security Visibility

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| MTA-STS policy publication | Modern TLS enforcement standard | Low | DNS + HTTPS-hosted policy file; Mailu/Mailcow have this |
| DANE TLSA records | DNSSEC-based certificate verification | Medium | Requires DNSSEC-enabled DNS provider; high security users want this |
| TLS-RPT (TLS reporting) | Visibility into TLS failures with peers | Low | DNS TXT record + report receiver |
| Strict mode: refuse plaintext peers | Reject mail from servers without TLS | Medium | High privacy users want this; may break mail from legacy systems |
| Notification of insecure peers | Alert when peer doesn't support TLS | Medium | Transparency into security posture; rspamd can log this |
| Queue encryption at rest | Encrypt mail spool on disk | High | **Rarely implemented**; requires content filter or filesystem encryption |

#### DNS Automation & Deliverability

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Auto-generate DNS records | Eliminates manual DNS configuration errors | Medium | Mail-in-a-Box does this if it's your nameserver; DarkPipe likely needs instructions |
| DNS validation checker | Verify SPF/DKIM/DMARC/MX setup | Low | Mail-in-a-Box includes this; critical for deliverability troubleshooting |
| DNS API integration | Programmatic DNS updates (Cloudflare, Route53, etc.) | Medium | Enables true one-click setup for supported providers |
| DNS record templates | Copy-paste instructions for manual setup | Low | Fallback for unsupported DNS providers |

#### Calendar & Contacts (Groupware)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| CalDAV server | Standards-based calendar sync | Medium | Nextcloud, Baïkal, SOGo, Radicale; required for full Gmail replacement |
| CardDAV server | Standards-based contacts sync | Medium | Same implementations as CalDAV; often bundled |
| Shared calendars | Family/team calendar sharing | Medium | SOGo supports this well; Nextcloud too |
| Calendar web UI | View/edit calendars without client | Medium | SOGo provides this; Nextcloud Calendar app |
| Contacts web UI | Manage contacts without client | Low | SOGo, Nextcloud Contacts |

#### Advanced Mail Features

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Sieve filtering rules | Server-side mail rules (folder sorting, auto-reply) | Medium | ManageSieve protocol; Roundcube plugin for GUI management |
| Vacation/auto-reply | Out-of-office messages | Low | Sieve script or Dovecot pigeonhole plugin |
| Mail forwarding rules | Forward to external addresses | Low | Postfix aliases or Sieve |
| Catch-all addresses | domain@example.com forwards all unknown addresses | Low | Postfix virtual alias wildcard |
| Full-text search with attachments | Search email body + PDF/DOCX content | High | Dovecot FTS with Apache Tika for attachment indexing |

#### Build/Deployment Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| GitHub Actions build customization | Users choose their stack via workflow config | High | **DarkPipe's unique approach**: templated workflows, user customizes |
| Multi-architecture images | ARM64 (Raspberry Pi, Mac) + x86_64 support | Medium | Docker buildx with `--platform linux/amd64,linux/arm64` |
| One-click deploy templates | Deploy to cloud providers via marketplace | High | DigitalOcean App Platform, AWS Lightsail, etc.; high initial effort |
| Documented stack alternatives | Choose Postfix vs Stalwart, Roundcube vs SOGo | Medium | Flexibility vs "opinionated simplicity" tradeoff |
| Reproducible builds | GitOps for email server config | Medium | Natural with GitHub Actions; versioned infrastructure |

#### Monitoring & Observability

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Mail delivery status dashboard | See which emails sent/received, delivery state | Medium | Postfix logs + parsing; Mailcow has queue viewer |
| Queue health monitoring | Detect stuck/deferred mail | Low | Postfix queue monitoring (mailq); alert on buildup |
| Certificate expiry alerts | Proactive notification before Let's Encrypt expires | Low | Simple cron job + email; critical to avoid outages |
| SMTP/IMAP connection logs | Audit who's connecting, from where | Low | Standard Postfix/Dovecot logs; privacy users want visibility |
| Prometheus metrics export | Integration with monitoring stack | Medium | Postfix exporter + Dovecot exporter exist; homelab users want this |
| Deliverability scoring | Check reputation of your sending IP | Medium | Integration with mail-tester.com or similar APIs |

#### Advanced Security Features

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| PGP/WKD support | End-to-end encryption for power users | Medium | Stalwart has this; Web Key Directory for automatic key discovery |
| S/MIME support | Certificate-based email signing/encryption | Medium | Standards-based alternative to PGP; enterprise users may want this |
| 2FA for webmail/admin | Protect web interfaces | Low | TOTP via Google Authenticator; Mail-in-a-Box includes this |
| Audit logging | Detailed logs of admin actions | Medium | Mailcow has this; critical for multi-admin environments |
| Rate limiting outbound mail | Prevent account takeover abuse | Low | Postfix policyd-rate-limit or Rspamd ratelimit module |

### Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Built-in webmail from scratch | Massive undertaking; solved problem | Use Roundcube or SOGo; focus on integration, not reinvention |
| Custom email client | Out of scope; existing clients work | Support standard protocols (IMAP, SMTP, CalDAV, CardDAV) |
| AI-powered features (spam, categorization) | Adds complexity, dependencies, privacy concerns | Use proven tools (Rspamd has Bayes learning); let users opt into external AI |
| Built-in VPN/Tor routing | Scope creep; separate concern | Document how to run behind VPN if desired; don't bundle |
| Blockchain-based identity | Complexity, immaturity, user confusion | Use standard DNS + TLS certificates |
| Built-in file sharing (Nextcloud-style) | Massive scope; solved problem | CalDAV/CardDAV only; if users want files, they install Nextcloud separately |
| Real-time collaborative editing | Wrong product category | Email is asynchronous; leave collaboration to Nextcloud/office suites |
| Built-in chat/messaging | Different problem domain | Email server, not Signal/Matrix |
| Support for legacy protocols (POP3, insecure SMTP) | Security liability; 2026 baseline is IMAP+TLS | IMAP only; refuse plaintext auth |
| Automatic migration from Gmail/Outlook | High complexity, API changes break it | Provide clear manual migration guide; let users use Thunderbird tools |
| Native mobile apps | Requires ongoing maintenance for iOS/Android | Users use native Mail apps or K-9/FairEmail; support standard protocols |
| Self-hosted DNS server | Increases attack surface, complexity | Use existing DNS providers with API integration or manual setup |
| Catch-all relay (forward all mail to Gmail) | Defeats privacy purpose | If users want Gmail as backup, they can configure forwarding manually |
| Machine learning spam training UI | Users won't use it; Bayesian auto-learn works | Rspamd auto-learns from user actions (move to spam folder) |
| Multi-tenant SaaS mode | DarkPipe is self-hosted; multi-tenancy adds vast complexity | Single-user or family-scale; if users want SaaS, recommend hosted providers |
| Windows/macOS native server support | Linux is email server standard; porting is massive effort | Document Docker Desktop for local testing; production = Linux VPS + home Linux |

## Feature Dependencies

```
Email Fundamentals
├── SMTP Inbound/Outbound → Required for everything
├── IMAP → Required for webmail, clients
└── TLS Certificates → Required for SMTP/IMAP encryption, HTTPS

Email Authentication (Deliverability)
├── SPF → Required before DMARC
├── DKIM → Required before DMARC
├── DMARC → Requires SPF + DKIM
├── MTA-STS → Requires TLS certs + HTTPS hosting
└── DANE TLSA → Requires DNSSEC-enabled DNS

DNS Automation
├── DNS API Integration → Enables auto-setup of SPF/DKIM/DMARC/MTA-STS
└── DNS Validation → Requires readable DNS records (after setup)

Webmail
├── Web Server (Nginx/Apache) → Hosts webmail interface
├── IMAP Server → Webmail backend
├── TLS Certificates → HTTPS for webmail
└── Contacts Integration → Requires CardDAV server

Groupware (Calendar/Contacts)
├── CalDAV Server → Calendar sync
├── CardDAV Server → Contacts sync
└── Web Server → Optional web UI for calendar/contacts

Advanced Mail Features
├── Sieve Filtering → Requires Dovecot + Pigeonhole
├── ManageSieve → Requires Sieve + protocol server for remote management
└── Full-Text Search → Requires Dovecot FTS plugin + Xapian/Solr

Security Features
├── Queue Encryption → Requires content filter or filesystem encryption
├── Strict TLS Mode → Requires MTA-STS or custom Postfix policy
├── PGP/WKD → Requires key management + WKD hosting
└── 2FA → Requires TOTP library + session management

Monitoring
├── Prometheus Metrics → Requires exporters for Postfix/Dovecot
├── Delivery Status Dashboard → Requires log parsing + database
└── Certificate Expiry Alerts → Requires cron job + notification system

DarkPipe-Specific
├── Cloud Relay → Required for inbound SMTP (port 25)
├── Secure Transport (Cloud → Home) → Requires VPN or authenticated relay
└── GitHub Actions Build → Required for multi-stack support
```

### Dependency Notes

- **SPF/DKIM before DMARC:** DMARC policy is meaningless without SPF and DKIM configured. Must implement in order.
- **TLS Certificates before HTTPS-dependent features:** MTA-STS, webmail, admin panel all require valid TLS certificates. Let's Encrypt must be working first.
- **DNS API integration enhances but doesn't replace validation:** Even with automated DNS updates, validation checks are required to catch provider API issues.
- **Sieve requires Dovecot:** If you swap Dovecot for another IMAP server, Sieve support may not be available.
- **CalDAV/CardDAV are bundled:** Solutions like Nextcloud, Baïkal, Radicale provide both; rarely makes sense to split them.
- **Queue encryption conflicts with content filters:** If you encrypt queue at rest, Amavis/SpamAssassin/ClamAV can't read messages. Must decrypt for scanning or scan before encryption.
- **Cloud relay is DarkPipe's architectural constraint:** Home devices can't receive inbound SMTP (port 25 blocked, dynamic IP). Cloud relay is mandatory for receiving mail.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **Inbound SMTP relay (cloud)** — Receives mail from internet
- [x] **Outbound SMTP relay (home device)** — Sends mail to internet
- [x] **IMAP server (home device)** — Users access mail via clients
- [x] **SPF/DKIM/DMARC setup** — Deliverability baseline
- [x] **TLS encryption (STARTTLS)** — Required for modern email
- [x] **Let's Encrypt auto-renewal** — Avoid certificate expiry outages
- [x] **Basic webmail (Roundcube)** — Web access for non-technical users
- [x] **Spam filtering (Rspamd/SpamAssassin)** — Inbound spam protection
- [x] **DNS validation tool** — Check if setup is correct
- [x] **Multi-user support** — Family/small team use case
- [x] **GitHub Actions build system** — Core differentiator: user customizes stack
- [x] **Multi-arch Docker images (ARM64 + x86_64)** — Raspberry Pi + VPS support

**Rationale:** This is the minimum feature set to replace Gmail for a privacy-focused user. Without cloud relay + home device architecture, DarkPipe has no value prop. Without DNS validation, users will fail at setup. Without webmail, non-technical family members can't participate. Without multi-arch images, Raspberry Pi users (core audience) can't run it.

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **CalDAV/CardDAV server** — Trigger: users request calendar/contacts sync (likely immediate)
- [ ] **Sieve filtering with ManageSieve** — Trigger: power users want server-side rules
- [ ] **MTA-STS + DANE** — Trigger: privacy-focused users want TLS enforcement
- [ ] **DNS API integration (Cloudflare, Route53)** — Trigger: users struggle with manual DNS setup
- [ ] **TLS-RPT logging** — Trigger: users want visibility into TLS failures
- [ ] **Queue health monitoring** — Trigger: users experience stuck mail, don't know why
- [ ] **Certificate expiry alerts** — Trigger: user's cert expires, service goes down
- [ ] **Prometheus metrics export** — Trigger: homelab users want Grafana dashboards
- [ ] **Strict TLS mode (reject plaintext peers)** — Trigger: high-privacy users want this hardening
- [ ] **Multiple domain support** — Trigger: users want to consolidate identities

**Rationale:** These features are high-value but not launch-blockers. CalDAV/CardDAV will likely be requested immediately (Gmail replacement requires calendar/contacts). Monitoring features become critical as users scale up. TLS enforcement features serve the privacy-focused niche.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Queue encryption at rest** — Why defer: high complexity, niche demand, conflicts with spam scanning
- [ ] **PGP/WKD support** — Why defer: niche user base, requires key management UX
- [ ] **S/MIME support** — Why defer: enterprise feature, less relevant for personal email
- [ ] **Full-text search with attachment indexing** — Why defer: high resource usage, not critical for small mailboxes
- [ ] **Audit logging for admin actions** — Why defer: relevant for multi-admin orgs, not v1 audience
- [ ] **Rate limiting outbound mail** — Why defer: anti-abuse feature, less relevant for trusted family users
- [ ] **Deliverability scoring integration** — Why defer: nice-to-have diagnostics, not core functionality
- [ ] **One-click deploy to cloud marketplaces** — Why defer: high partnership/integration effort
- [ ] **SOGo webmail (alternative to Roundcube)** — Why defer: dual webmail support adds maintenance burden

**Rationale:** These are "nice-to-have" features that serve narrow use cases or require disproportionate effort. Queue encryption and PGP serve extreme privacy users (niche within niche). Deliverability scoring and audit logs are diagnostic tools that can wait. Full-text search with attachments is resource-intensive and most users search by sender/subject anyway.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Cloud relay architecture | HIGH | HIGH | P0 (MVP blocker) |
| SMTP/IMAP/TLS | HIGH | LOW | P0 (MVP blocker) |
| SPF/DKIM/DMARC | HIGH | LOW | P0 (MVP blocker) |
| DNS validation tool | HIGH | LOW | P0 (MVP blocker) |
| Webmail (Roundcube) | HIGH | MEDIUM | P0 (MVP blocker) |
| Multi-arch builds | HIGH | MEDIUM | P0 (MVP blocker) |
| GitHub Actions customization | HIGH | HIGH | P0 (core differentiator) |
| CalDAV/CardDAV | HIGH | MEDIUM | P1 (post-launch) |
| Sieve filtering | MEDIUM | MEDIUM | P1 (post-launch) |
| MTA-STS + DANE | MEDIUM | MEDIUM | P1 (post-launch) |
| DNS API integration | HIGH | MEDIUM | P1 (reduces support burden) |
| Certificate expiry alerts | HIGH | LOW | P1 (prevents outages) |
| Queue health monitoring | MEDIUM | LOW | P1 (prevents support tickets) |
| Prometheus metrics | MEDIUM | LOW | P1 (homelab users expect this) |
| Strict TLS mode | MEDIUM | MEDIUM | P2 (niche privacy feature) |
| TLS-RPT | LOW | LOW | P2 (diagnostics) |
| Multiple domain support | MEDIUM | MEDIUM | P2 (defer until requested) |
| Queue encryption at rest | LOW | HIGH | P3 (niche, high complexity) |
| PGP/WKD | LOW | HIGH | P3 (niche, key mgmt complexity) |
| Full-text search + attachments | MEDIUM | HIGH | P3 (resource-intensive) |
| Deliverability scoring | LOW | MEDIUM | P3 (nice-to-have diagnostics) |
| One-click cloud deploy | MEDIUM | HIGH | P3 (partnership effort) |
| Audit logging | LOW | MEDIUM | P3 (enterprise feature) |

**Priority key:**
- **P0:** Must have for launch (MVP blockers)
- **P1:** Should have, add in first 3 months post-launch
- **P2:** Could have, add based on user feedback
- **P3:** Won't have initially, consider for v2.0+

## Competitor Feature Analysis

| Feature | Mail-in-a-Box | Mailu | docker-mailserver | Mailcow | DarkPipe (Planned) |
|---------|---------------|-------|-------------------|---------|-------------------|
| **Architecture** | Monolithic Ubuntu script | Docker Compose | Single Docker image | Docker Compose | **Cloud relay + home device** |
| **SMTP/IMAP/TLS** | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ User-chosen via GitHub Actions |
| **SPF/DKIM/DMARC** | ✅ Auto-generated | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Auto-generated + validation |
| **MTA-STS** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **DANE** | ✅ | ✅ | ❌ | ❌ | ✅ (planned) |
| **Webmail** | ✅ Roundcube | ✅ Roundcube/Rainloop | ❌ (BYO) | ✅ SOGo/Roundcube | ✅ Roundcube (default) |
| **CalDAV/CardDAV** | ✅ Nextcloud | ❌ (separate install) | ❌ (BYO) | ✅ SOGo | ✅ (planned: Baïkal or SOGo) |
| **Admin Panel** | ✅ Custom | ✅ Custom | ❌ (CLI only) | ✅ Custom + API | ✅ (planned: minimal web UI) |
| **DNS Management** | ✅ Can host DNS | ❌ Manual | ❌ Manual | ❌ Manual | **✅ API integration (Cloudflare, Route53)** |
| **Spam Filtering** | ✅ SpamAssassin | ✅ Rspamd | ✅ Rspamd | ✅ Rspamd | ✅ Rspamd (default) |
| **Greylisting** | ✅ Postgrey | ✅ Rspamd | ✅ Postgrey | ✅ Rspamd | ✅ Rspamd |
| **Virus Scanning** | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV (optional) |
| **Backups** | ✅ Duplicity to S3 | ❌ (manual) | ❌ (manual) | ✅ Via admin panel | ✅ (planned: backup to user's cloud) |
| **Multi-Domain** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **Sieve Filtering** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **2FA** | ✅ TOTP | ❌ | ❌ | ✅ TOTP | ✅ (planned) |
| **Multi-Arch (ARM64)** | ❌ (x86_64 only) | ✅ | ✅ | ⚠️ (experimental) | **✅ First-class ARM64 support** |
| **User Customization** | ❌ Opinionated | ❌ Opinionated | ⚠️ Config files | ⚠️ Limited | **✅ GitHub Actions: choose your stack** |
| **Home Device Support** | ❌ Assumes public IP | ❌ Assumes public IP | ❌ Assumes public IP | ❌ Assumes public IP | **✅ Cloud relay for inbound** |
| **Privacy Focus** | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | **✅ Storage on user device** |

### Key Differentiators vs Competitors

1. **DarkPipe's unique value:** Cloud relay + home device split solves the "home ISP blocks port 25" problem that kills other self-hosted solutions.
2. **GitHub Actions customization:** No other solution lets users choose stack components (Postfix vs Stalwart, Roundcube vs SOGo) via templated builds.
3. **First-class ARM64 support:** Mail-in-a-Box doesn't support ARM. Mailcow's ARM support is experimental. DarkPipe treats Raspberry Pi as primary deployment target.
4. **DNS API integration:** Competitors require manual DNS setup or (Mail-in-a-Box) full DNS hosting. DarkPipe automates via provider APIs.
5. **Privacy architecture:** Competitors assume VPS = trusted. DarkPipe assumes cloud relay = untrusted, storage = home device only.

### What Competitors Do Well (Learn From)

- **Mail-in-a-Box:** Excellent DNS validation UI. Copy this. Comprehensive status checks. One-command install.
- **Mailu:** Clean Docker Compose structure. Role-based admin delegation. Good anti-spam (Rspamd).
- **docker-mailserver:** Configuration via files (GitOps-friendly). `setup.sh` utility for CLI admin tasks.
- **Mailcow:** Polished admin UI. SOGo integration (best webmail + groupware). Queue viewer for debugging.

### What Competitors Struggle With (Avoid)

- **Mail-in-a-Box:** Monolithic bash script. Can't customize components. No Docker (harder to deploy/update). x86_64 only.
- **Mailu:** No CalDAV/CardDAV (deal-breaker for Gmail replacement). Manual DNS setup (high failure rate).
- **docker-mailserver:** No admin panel (CLI-only intimidates users). No webmail (BYO). Steep learning curve.
- **Mailcow:** Heavy resource usage (runs many containers). Complex to customize. Assumes public IP (no home device support).

## Helm's Feature Set & Failure Points

### What Helm Offered

Helm was a **$499 hardware device** (personal email server) that launched in 2018 and shut down in December 2022. It targeted non-technical users wanting email privacy.

**Feature Set:**
- Email server (SMTP/IMAP) with custom domain support
- Contacts and calendar (CalDAV/CardDAV)
- Notes and file storage
- **120GB SSD storage** (expandable to 5TB)
- **Proximity-based security:** Bluetooth token for 2FA
- **Full disk encryption** with Secure Enclave-managed keys
- **Cloud backup:** Nightly encrypted backups
- DMARC/SPF/DKIM authentication
- TLS encryption (client-server, server-server)
- PGP/S/MIME support (at MUA level, not MTA)

### Why Helm Failed

**Critical Failure Points:**

1. **Supply chain issues:** Shifted manufacturing from Mexico to China in late 2019 to reduce costs. COVID-19 devastated supply chain. Unable to manufacture/ship devices.

2. **Hardware dependence:** Physical device = capital expenditure, inventory risk, shipping logistics. Software-only competitors (Proton Mail) scaled without these constraints.

3. **Subscription dependency:** Service required ongoing cloud backup subscription. When revenue dropped (couldn't ship new devices), cloud costs became unsustainable.

4. **No webmail:** Users had to configure IMAP clients. Non-technical users struggled. Helm's target audience expected "just works" web access.

5. **Home network complexity:** Users had to configure port forwarding, DDNS, or use Helm's relay. Many failed at network setup.

6. **Locked-in hardware:** $499 device became e-waste when service shut down. Users lost investment. (Helm did release Armbian conversion firmware, but most users didn't convert.)

7. **Single point of failure:** Helm the company = Helm the service. No open-source alternative to migrate to.

### Lessons for DarkPipe

**Do:**
- ✅ **Separate hardware and software:** DarkPipe runs on user's existing hardware (Raspberry Pi, old laptop). No inventory/shipping risk.
- ✅ **Open-source + self-hosted:** If DarkPipe project ends, users keep running their servers. No vendor lock-in.
- ✅ **Cloud relay for inbound:** Solves Helm's "home network complexity" problem. Users don't configure port forwarding.
- ✅ **Include webmail:** Roundcube/SOGo required for non-technical users.
- ✅ **Donation-funded, not subscription:** No recurring revenue pressure. Cloud relay costs covered by donations or minimal fees.

**Don't:**
- ❌ **Don't sell hardware:** Helm's $499 device = capital risk. DarkPipe = bring your own device.
- ❌ **Don't create service dependencies:** Helm's cloud backup was a recurring cost center. DarkPipe users manage their own backups.
- ❌ **Don't target "zero-configuration":** Helm promised simplicity, delivered complexity. DarkPipe targets "homelab-adjacent" users who tolerate config.
- ❌ **Don't hide relay operation:** Helm's relay was opaque. DarkPipe should show users exactly what cloud relay does (logs, audit trail).

**Helm's Legacy:**
- Proved market demand for privacy-focused email (1000s of units sold).
- Showed that hardware + subscription model is fragile.
- Validated need for "works behind home ISP" solution (port 25 blocking).
- Demonstrated that non-technical users will pay for privacy (if UX is good enough).

## Sources

### Competitor Analysis
- [Mail-in-a-Box](https://mailinabox.email/)
- [Mail-in-a-Box GitHub](https://github.com/mail-in-a-box/mailinabox)
- [Mailu](https://mailu.io/)
- [Mailu GitHub](https://github.com/Mailu/Mailu)
- [docker-mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/)
- [docker-mailserver GitHub](https://github.com/docker-mailserver/docker-mailserver)
- [Mailcow January 2026 Update](https://mailcow.email/posts/2026/release-2026-01/)
- [Mailcow GitHub](https://github.com/mailcow/mailcow-dockerized)
- [Mailcow features analysis](https://www.servercow.de/mailcow)

### Helm Analysis
- [Helm Shutdown FAQ](https://support.thehelm.com/hc/en-us/articles/10831596925203-Helm-Shutdown-FAQ)
- [Helm Email Server Review](https://blog.strom.com/wp/?p=6990)
- [Fortune: Helm's Private Email Server](https://fortune.com/2019/03/09/helm-server-gmail-privacy/)
- [SlashGear: Helm Personal Email Server](https://www.slashgear.com/helm-personal-email-server-promises-perfect-privacy-17550402/)
- [GeekWire: Seattle startup vets take on Google with Helm](https://www.geekwire.com/2018/seattle-startup-vets-take-tech-giants-helm-new-personal-email-server/)

### Technical Standards & Protocols
- [How to Host Your Own Email Server in 2026](https://elementor.com/blog/how-to-host-your-own-email-server/)
- [Self-Hosting Email in 2026: Is Running a Linux Mail Server Still Worth It?](https://securityboulevard.com/2025/12/self-hosting-email-in-2026-is-running-a-linux-mail-server-still-worth-it/)
- [SPF DKIM DMARC Explained 2026](https://skynethosting.net/blog/spf-dkim-dmarc-explained-2026/)
- [Email Security Market: 8 Things to Know for 2026 Planning](https://abnormal.ai/blog/email-security-market)
- [MTA-STS Guide](https://redsift.com/guides/email-security-guide/mta-sts)
- [DMARC, DANE, MTA-STS, TLS, and DKIM Explained](https://www.anubisnetworks.com/blog/dmarc_dane_explained)

### CalDAV/CardDAV
- [Calendar & Contacts - awesome-selfhosted](https://awesome-selfhosted.net/tags/calendar--contacts.html)
- [Baïkal](https://sabre.io/baikal/)
- [Best 11 Open-source CalDAV Self-hosted Servers](https://medevel.com/11-caldav-os-servers/)
- [Building a self-hosted CalDAV server](https://rfrancocantero.medium.com/building-a-self-hosted-caldav-server-the-technical-reality-behind-calendar-sharing-9a930af28ff0)

### Sieve Filtering
- [Sieve (mail filtering language) - Wikipedia](https://en.wikipedia.org/wiki/Sieve_(mail_filtering_language))
- [Advanced Email Filtering with Sieve - Docker Mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/config/advanced/mail-sieve/)
- [Proton Mail Sieve Filters](https://proton.me/support/sieve-advanced-custom-filters)

### Multi-Architecture & Deployment
- [How to Build Multi-Architecture Docker Images (ARM64 + AMD64)](https://oneuptime.com/blog/post/2026-01-06-docker-multi-architecture-images/view)
- [How to Build Multi Architecture Docker Images](https://devopscube.com/build-multi-arch-docker-image/)
- [GitHub Actions: Early February 2026 updates](https://github.blog/changelog/2026-02-05-github-actions-early-february-2026-updates/)

### Monitoring & Security
- [SSL Certificate Monitor - Dynatrace](https://www.dynatrace.com/hub/detail/ssl-certificate-monitor/)
- [TLS Certificate Monitoring with OpenTelemetry](https://www.elastic.co/observability-labs/blog/edot-certificate-monitoring)
- [NETSCOUT SSL certificate monitoring](https://www.helpnetsecurity.com/2026/01/27/netscout-ngeniusone-enhancements/)
- [Postfix Hardening Guide](https://linux-audit.com/postfix-hardening-guide-for-security-and-privacy/)
- [Postfix TLS Support](http://www.postfix.org/TLS_README.html)

### Privacy-Focused Email
- [Privacy Guides: Self-Hosting Email](https://www.privacyguides.org/en/self-hosting/email-servers/)
- [Self Hosted Email: Privacy, Security and Full Control](https://www.smartertools.com/blog/2025/01/self-hosting-email)
- [Forward Email - 100% open-source and privacy-focused](https://forwardemail.net/en/blog/best-private-email-service)

### DNS Automation
- [DNSimple TLSA Record Support](https://blog.dnsimple.com/2026/01/tlsa-record-support/)
- [Best Email API For Developers](https://mailtrap.io/blog/best-email-api/)

---
*Feature research for: DarkPipe (Privacy-First Self-Hosted Email)*
*Researched: 2026-02-08*
*Confidence: MEDIUM-HIGH (Web search verified with official docs; some gaps around queue encryption and bleeding-edge features)*

# Pitfalls Research: Self-Hosted Email Relay System

**Domain:** Cloud-fronted email relay with home device backend
**Researched:** 2026-02-08
**Confidence:** HIGH

---

## Critical Pitfalls

### Pitfall 1: VPS Provider Port 25 Restrictions Block Direct MTA

**What goes wrong:**
You deploy your cloud relay on a VPS only to discover port 25 (SMTP) is blocked at the network level, preventing your server from receiving or sending email directly. Your MTA cannot function as a mail server without port 25 access.

**Why it happens:**
VPS providers block SMTP ports (25, 465, 587) by default to prevent spam abuse on their infrastructure. DigitalOcean, Hetzner, Vultr, Scaleway, and many others implement these restrictions. Developers assume "VPS = full control" without researching provider-specific limitations.

**How to avoid:**
- **Choose port-25-friendly providers for v1:** Linode/Akamai and OVH have port 25 open by default
- **Budget time for unblocking:** BuyVM allows unblocking via support ticket; Vultr similar but less reliable
- **Avoid DigitalOcean for direct MTA:** Port 25 restrictions cannot be reliably removed
- **Hetzner requires 1-month wait:** New accounts must wait ~1 month before requesting port 25 access
- **Document provider policies:** Track current policies as they change frequently

**Warning signs:**
- SMTP connection timeouts on port 25 during testing
- "Connection refused" or "Network unreachable" errors
- Can telnet to port 587 but not port 25
- Support ticket responses mentioning "anti-spam policy"

**Phase to address:**
**Phase 0 (Infrastructure Selection)** - Provider selection must happen before any MTA development. Research and document current port 25 policies for target providers. Consider maintaining a fallback provider list.

**Confidence:** HIGH - Verified with official documentation from DigitalOcean, Linode, OVH, Hetzner, Vultr, BuyVM, and Scaleway

---

### Pitfall 2: New VPS IPs Start Blacklisted or with Zero Reputation

**What goes wrong:**
You launch your mail server with a fresh VPS IP address and discover it's already on RBL blacklists (previous tenant spam), or Gmail/Outlook reject 50-100% of your emails because the IP has no sending history. Deliverability is catastrophic from day one.

**Why it happens:**
VPS providers recycle IP addresses. Your "new" IP may have been used by spammers months ago. Even clean IPs start with zero reputation - major providers (Gmail, Microsoft) treat unknown IPs with suspicion. ISPs maintain both public blacklists (Spamhaus, Barracuda) and private reputation systems.

**How to avoid:**
- **Check IP reputation BEFORE launching:** Use MXToolbox, Spamhaus, and multi-RBL checkers during VPS provisioning
- **Budget 4-6 weeks for IP warmup:** Gradual volume increase is mandatory, not optional
- **Request IP replacement if blacklisted:** Most providers allow one IP change if you catch it early
- **Start with engaged recipients only:** First 2 weeks should be trusted contacts who will open/reply
- **Conservative warmup schedule:**
  - Days 1-3: 2-5 emails/day
  - Days 4-7: 5-10 emails/day
  - Week 2: 10-20 emails/day
  - Week 3: 15-30 emails/day
  - Week 4: 25-50 emails/day
- **Monitor blacklists continuously:** Automated checking every 3-6 hours with RBLTracker, MXToolbox, or ZeroBounce
- **Never send bulk from day one:** Fastest way to permanent blacklisting

**Warning signs:**
- High bounce rates (>5%) in first week
- 450/451 "Greylisting" responses that never clear
- Gmail sending to spam folder consistently
- Spamhaus or Barracuda listings appearing in logs
- Microsoft returning "550 5.7.1 Service unavailable" errors

**Phase to address:**
**Phase 1 (MVP)** - IP reputation verification must be part of deployment checklist. Warmup period extends MVP timeline by 4-6 weeks. **Phase 2** should add automated blacklist monitoring before scaling.

**Confidence:** HIGH - Multiple sources including Spamhaus, deliverability guides, and mail server provider documentation

---

### Pitfall 3: Missing or Misconfigured SPF/DKIM/DMARC Breaks Deliverability

**What goes wrong:**
Emails fail authentication checks and land in spam folders or get rejected outright. Gmail shows "via unknown.net" warnings. Microsoft flags messages as spoofed. Deliverability drops to <30% despite clean IP reputation.

**Why it happens:**
Modern email requires SPF (sender authorization), DKIM (cryptographic signature), and DMARC (policy enforcement). Even one-character DNS typos break authentication. Self-hosters often set up SPF but skip DKIM key rotation, use weak 1024-bit keys, or create DMARC records without testing SPF/DKIM alignment first.

**How to avoid:**
- **SPF common mistakes:**
  - Exceeding 10 DNS lookup limit (causes hard fail)
  - Including cloudflare.com or broad includes that balloon lookups
  - Forgetting "+a +mx" for your own server
  - Using "v=spf1 -all" (rejects everything) vs "v=spf1 ~all" (soft fail)
- **DKIM requirements:**
  - Use 2048-bit keys minimum (1024-bit considered weak in 2026)
  - Test DKIM signature with mail-tester.com before production
  - Avoid duplicate selectors if multiple systems sign
  - Set up key rotation policy (annual minimum)
- **DMARC essentials:**
  - Start with "p=none" for monitoring, not "p=reject"
  - Verify SPF and DKIM alignment BEFORE setting policy
  - Include "rua=mailto:dmarc-reports@yourdomain" for feedback
  - Subdomain policy: "sp=quarantine" or "sp=reject"
- **Testing protocol:**
  - Send to mail-tester.com (should score 9+/10)
  - Send to Gmail and check "Show original" → Passed SPF/DKIM/DMARC
  - Use dmarcian.com or similar to validate records
  - Monitor DMARC reports weekly

**Warning signs:**
- Mail-tester.com scores below 8/10
- "SPF fail" or "DKIM fail" in bounced message headers
- DMARC reports showing alignment failures
- Gmail "via" warnings on your own emails
- High spam folder placement (>10%)

**Phase to address:**
**Phase 1 (MVP)** - Authentication must be working before any production traffic. Include DNS setup in deployment automation. **Phase 2** should add DMARC report parsing and alerting on failures.

**Confidence:** HIGH - Verified with official email authentication standards, Gmail/Microsoft documentation, and multiple 2026 deliverability guides

---

### Pitfall 4: Missing PTR (Reverse DNS) Record Triggers Instant Spam Filtering

**What goes wrong:**
All major email providers reject or spam-folder your emails because your sending IP has no PTR record, or PTR doesn't match forward DNS. SpamRATS RATS-NoPtr blacklist flags your IP immediately. Deliverability drops below 50%.

**Why it happens:**
Reverse DNS (PTR record) proves your IP is legitimately configured for email. Without PTR → A record match (forward-confirmed reverse DNS / FCrDNS), providers assume botnet or compromised machine. VPS users forget PTR records are controlled by the hosting provider, not your DNS registrar.

**How to avoid:**
- **Contact VPS provider to set PTR:** You cannot set PTR in your own DNS, must request from provider
- **PTR must point to your mail server FQDN:** mail.yourdomain.com, not generic vps12345.provider.com
- **Verify forward-confirmed reverse DNS:**
  ```bash
  # Get PTR record
  dig -x YOUR.IP.ADDRESS +short
  # Verify A record matches
  dig mail.yourdomain.com +short
  # Should return YOUR.IP.ADDRESS
  ```
- **PTR and A must match exactly:** mail.example.com → 1.2.3.4 AND 1.2.3.4 → mail.example.com
- **Set PTR before sending ANY email:** Include in deployment checklist
- **Monitor PTR continuously:** Some providers reset PTR during maintenance

**Warning signs:**
- SpamRATS RATS-NoPtr blacklist listing
- Bounce messages mentioning "reverse DNS lookup failed"
- Mail-tester.com flagging PTR issues
- "dig -x" shows no PTR or generic hostname
- Deliverability suddenly drops after VPS migration

**Phase to address:**
**Phase 1 (MVP)** - PTR verification must be in pre-launch checklist. Automated testing should verify PTR/A match before deployment. Document provider-specific PTR request process.

**Confidence:** HIGH - Verified with Spamhaus, SpamRATS, and multiple deliverability guides. PTR requirement universal across major providers.

---

### Pitfall 5: Residential/Dynamic IP from Home Device Gets Blacklisted

**What goes wrong:**
The home device (Raspberry Pi, local server) attempts to send email directly from a residential ISP IP address. Spamhaus PBL (Policy Block List) and other services immediately blacklist it. Even with perfect configuration, major providers reject all mail.

**Why it happens:**
Residential IP ranges are flagged in DNS blacklists as "should never send email directly." ISPs assign dynamic IPs with PTR records indicating home/residential use (e.g., "c-123-45-67-89.hsd1.ca.comcast.net"). Email ecosystem assumes residential IPs = compromised machines in botnets.

**How to avoid:**
- **NEVER send directly from home device:** This is why DarkPipe uses cloud relay architecture
- **Cloud relay must handle ALL outbound SMTP:** Home device only stores mail, relay sends
- **WireGuard tunnel for relay → home:** Relay fetches from home, not home pushing to internet
- **If forced to use home IP:**
  - Request static IP from ISP (still likely blacklisted but allows PTR)
  - Use ISP's SMTP relay as smarthost (defeats purpose of self-hosting)
  - Accept emails will be rejected by Gmail/Outlook
- **PBL-specific issue:** Spamhaus PBL lists ALL residential IP ranges globally

**Warning signs:**
- Spamhaus PBL (Policy Block List) listings
- PTR record shows residential naming pattern
- ISP ToS prohibits running mail servers
- IP changes every few days/weeks (dynamic)
- "Cannot relay" errors from recipient servers

**Phase to address:**
**Phase 1 (MVP)** - Architecture prevents this by design. Cloud relay handles all MTA functions. Home device must NEVER have SMTP port 25 accessible from internet. Document this as architectural constraint.

**Confidence:** HIGH - Verified with Spamhaus PBL documentation, residential IP blacklist research, and ISP policies

---

### Pitfall 6: TLS Certificate Expiration Breaks Email Flow Silently

**What goes wrong:**
Your Let's Encrypt certificate expires after 90 days. Receiving mail servers reject connections with "certificate expired" errors. Email clients show warnings. Mail queues build up, deliverability drops to zero. Problem discovered only when users complain.

**Why it happens:**
Manual certificate renewal is missed due to lack of expiry tracking. Certbot auto-renewal cron jobs fail silently (disk full, DNS changes, permission errors). Email-specific certificate requirements differ from web hosting - must cover mail.example.com, smtp.example.com, potentially wildcards.

**How to avoid:**
- **Automate with Let's Encrypt/Certbot:** 90-day certs require automation, not manual renewal
- **Monitor renewal success:** Certbot runs renewals but doesn't alert on failures
- **Certificate coverage requirements:**
  - Include all mail hostnames: mail.example.com, smtp.example.com, imap.example.com
  - Consider wildcard cert (*.example.com) for flexibility
  - Verify SAN (Subject Alternative Names) includes all aliases
- **Test renewal process:**
  ```bash
  certbot renew --dry-run
  ```
- **Set up expiry monitoring:**
  - Monitor certificate expiry 30/14/7 days before expiration
  - Alert if certbot renewal fails
  - Use external monitoring (UptimeRobot, SSL Labs)
- **Common failure modes:**
  - DNS changes break DNS-01 validation
  - Firewall blocks HTTP-01 validation on port 80
  - Disk space full prevents cert writing
  - Permissions prevent certbot from restarting services
- **Docker-specific:** Certificates in containers need volume mounting; renewal must restart containers

**Warning signs:**
- "Certificate expired" in logs
- Bounce messages mentioning TLS/SSL errors
- External SSL checkers showing expired certs
- Certbot logs showing renewal failures
- Sudden drop in delivered mail with TLS errors

**Phase to address:**
**Phase 1 (MVP)** - Automated certificate management from day one. Include monitoring in deployment. **Phase 2** should add alerting for renewal failures and expiry warnings.

**Confidence:** HIGH - Verified with Let's Encrypt documentation, email server TLS guides, and certificate management best practices

---

### Pitfall 7: Becoming an Open Relay Leads to Immediate Blacklisting

**What goes wrong:**
Your mail server is misconfigured to relay email for anyone without authentication. Spammers discover it within hours (automated scanning), flood thousands of spam messages through your server. Your IP gets blacklisted on every major RBL within 24-48 hours. Permanent reputation damage.

**Why it happens:**
Default mail server configs often allow relaying for localhost/local networks. Developers test without authentication, then forget to restrict. Docker networking creates unexpected relay paths. IPv6 adds additional relay surface area. Automated scanners constantly probe port 25 for open relays.

**How to avoid:**
- **Require SMTP authentication for ALL relay:** No exceptions for "internal" networks
- **Restrict relay by:**
  - Authenticated users only (SASL authentication)
  - Specific IP addresses (if absolutely necessary)
  - Known domains only (reject unknown senders)
- **Docker-specific risks:**
  - Container bridge networks may allow relay
  - Check relay access from other containers
  - Don't expose port 25 to host network without auth
- **Test for open relay:**
  - Use MXToolbox Open Relay test
  - Test from external IP without authentication
  - Verify logs show "Relay access denied" for unauthorized attempts
- **Monitor for abuse:**
  - Alert on sudden volume spikes
  - Log all relay attempts (authorized and denied)
  - Check queue size daily - spam floods create huge queues
- **Rate limiting as backup:**
  - Limit emails per connection
  - Limit connections per IP
  - Queue delay for unknown senders

**Warning signs:**
- Queue suddenly fills with thousands of messages
- Logs show connections from unknown IPs sending mail
- Blacklist monitoring alerts (RBL listings)
- Outbound bandwidth spike
- Bounces from hundreds of domains you don't recognize

**Phase to address:**
**Phase 1 (MVP)** - Relay restrictions must be tested before launch. Include open relay testing in deployment checklist. **Phase 2** should add automated relay testing and queue monitoring.

**Confidence:** HIGH - Verified with SMTP relay security documentation, open relay prevention guides, and mail server configuration best practices

---

### Pitfall 8: WireGuard Tunnel Fails After Home Internet Drop, Mail Stalls

**What goes wrong:**
Home internet connection drops (ISP maintenance, power outage, router restart). WireGuard tunnel doesn't automatically reconnect. Cloud relay can't reach home device to fetch stored mail. Outbound emails queue indefinitely. Users report "emails not sending."

**Why it happens:**
WireGuard is designed as a simple tunnel - it doesn't include automatic reconnection logic. When home IP changes (dynamic IP) or connection drops, tunnel breaks. DNS resolution fails if local DNS was used. Keepalive settings too long or missing. Home device reboots don't restore tunnel.

**How to avoid:**
- **Configure PersistentKeepalive on home device:**
  ```
  [Peer]
  PersistentKeepalive = 25
  ```
  Sends keepalive packet every 25 seconds to maintain NAT mapping
- **Use systemd to restart tunnel on failure:**
  ```bash
  [Unit]
  After=network-online.target nss-lookup.target
  Wants=network-online.target nss-lookup.target

  [Service]
  Restart=on-failure
  RestartSec=30
  ```
- **DNS resolution issues:**
  - Use IP addresses in Endpoint, not hostnames (or use public DNS)
  - Set DNS = 1.1.1.1, 8.8.8.8 in [Interface] section
  - Don't rely on local DNS resolver
- **Dynamic IP handling:**
  - Cloud relay needs to handle home IP changes
  - Use Dynamic DNS (DDNS) for home endpoint
  - Implement endpoint update mechanism
- **Monitor tunnel health:**
  - Check "latest handshake" timestamp (should be <3 minutes with keepalive)
  - Alert if tunnel down >5 minutes
  - Test connectivity from cloud relay periodically
- **Fallback strategy:**
  - Queue mail on cloud relay if tunnel down
  - Retry delivery when tunnel restores
  - Alert operators if tunnel down >30 minutes

**Warning signs:**
- "Endpoint resolution failed" in WireGuard logs
- Handshake timestamp >5 minutes old
- Ping across tunnel fails
- Mail queue on cloud relay growing
- Users reporting "emails stuck"

**Phase to address:**
**Phase 1 (MVP)** - Tunnel resilience is critical for architecture. Test reconnection during development. **Phase 2** should add monitoring and automatic alerting for tunnel failures.

**Confidence:** MEDIUM - Based on WireGuard documentation and community forum discussions. Specific implementation details need testing.

---

### Pitfall 9: Docker Volume Management Loses Mail Data on Container Update

**What goes wrong:**
You update the mail server Docker container (security patch, version upgrade). Container recreates without persistent volumes properly configured. All stored emails, queue data, and configuration vanish. Data loss discovered only when users report missing emails.

**Why it happens:**
Anonymous volumes get orphaned on container removal. Bind mounts use relative paths that break. docker-compose.yml doesn't specify volumes correctly. Developers test with empty data, miss volume config. Container logs show data written, but to ephemeral layer not persistent volume.

**How to avoid:**
- **Use named volumes, never anonymous:**
  ```yaml
  volumes:
    mail-data:
      driver: local

  services:
    mail:
      volumes:
        - mail-data:/var/mail
        - mail-queue:/var/spool/postfix
        - mail-config:/etc/postfix
  ```
- **Critical mail server paths to persist:**
  - `/var/mail` - User mailboxes
  - `/var/spool/postfix` - Mail queue
  - `/etc/postfix` - Configuration
  - `/etc/letsencrypt` - TLS certificates
  - `/var/log` - Logs (for debugging)
- **Docker Compose best practices:**
  - Define volumes at top level
  - Use descriptive names (mail-data not data1)
  - Document what each volume contains
- **Backup volumes regularly:**
  - Automated daily backups to offsite location
  - Test restore process before production
  - "Volume without backup = single point of failure"
- **Permission issues:**
  - Mail UID/GID must match between container and volume
  - Check `ls -la` shows correct ownership
  - Use docker-compose user: directive if needed
- **Dangling volume cleanup:**
  - `docker volume ls -qf dangling=true` shows orphaned volumes
  - Don't auto-clean without verification (might contain data)
  - Document volume lifecycle in runbook

**Warning signs:**
- Container starts with empty mailboxes after update
- "Permission denied" errors for mail directories
- Queue shows as empty when it shouldn't be
- Config resets to defaults after container restart
- `docker volume ls` shows dangling volumes

**Phase to address:**
**Phase 1 (MVP)** - Volume configuration must be correct from first deployment. Test container update process during development. **Phase 2** should add automated backup and restore testing.

**Confidence:** HIGH - Verified with Docker documentation, Docker Mailserver guides, and container volume best practices

---

### Pitfall 10: Raspberry Pi ARM64 Runs Out of Memory Under Load

**What goes wrong:**
Home device (Raspberry Pi 4) handles mail storage and indexing. Under load (virus scanning, large attachments, multiple IMAP connections), memory usage spikes. OOM killer terminates mail processes. Mail delivery fails. System becomes unresponsive.

**Why it happens:**
ARM LPAE limits any single process to 3GB RAM (1GB reserved for kernel) even on 8GB Pi. Mail servers with SpamAssassin, ClamAV, and Dovecot can easily exceed this. Large mailboxes trigger memory-intensive indexing. Swap on SD card thrashes and kills performance.

**How to avoid:**
- **Memory budgeting for Pi 4 8GB:**
  - OS + base services: 1-2GB
  - Mail storage (Dovecot): 512MB-1GB
  - Mail indexing (FTS): 512MB-1GB
  - Remaining for buffers/cache: ~3-4GB
  - **Don't run:** ClamAV (500MB+), SpamAssassin (heavy), webmail
- **Run minimal services on home device:**
  - Mail storage only (Dovecot IMAP/LMTP)
  - Virus scanning on cloud relay (more resources)
  - Spam filtering on cloud relay
  - Webmail on cloud relay or separate service
- **Swap considerations:**
  - Don't use SD card for swap (wear + slow)
  - Use USB SSD if swap needed
  - Limit swappiness: `vm.swappiness=10`
  - Monitor swap usage - if heavy, system underpowered
- **Storage I/O limits:**
  - SD card random I/O kills mailbox performance
  - Use USB 3.0 SSD for mail storage
  - Monitor I/O wait time
- **Resource monitoring:**
  - Alert on memory usage >80%
  - Alert on swap usage >10%
  - Monitor OOM killer logs
  - Track process memory with `ps aux --sort=-%mem`

**Warning signs:**
- "Out of memory" in kernel logs
- OOM killer messages in `dmesg`
- System freezes under load
- IMAP connections timeout
- Swap usage climbing
- Load average >4.0 on Pi

**Phase to address:**
**Phase 1 (MVP)** - Minimize services on home device from start. Test under realistic load before production. **Phase 2** should add memory monitoring and alerts.

**Confidence:** MEDIUM - Based on Raspberry Pi specifications and mail server memory requirements. Specific limits need testing with actual workload.

---

### Pitfall 11: Users Give Up Due to Complex Setup and Ongoing Maintenance

**What goes wrong:**
Setup requires deep knowledge of DNS, Linux server administration, mail protocols, and debugging. Users struggle through installation, hit deliverability issues, spend hours troubleshooting. After launch, weekly maintenance (spam rules, blacklist monitoring, log review) becomes overwhelming. Users give up and return to Gmail.

**Why it happens:**
Email is the most complex self-hosted service. Unlike web servers (simple HTTP), email requires perfect DNS, reputation management, anti-spam, authentication, TLS, and ongoing monitoring. Documentation assumes Linux expertise. Error messages are cryptic. One misconfiguration breaks everything. Most users lack time for "week after week" maintenance.

**How to avoid:**
- **Acknowledge this is HARD:** "Self-hosted email is notoriously difficult" - set expectations
- **Provide automated setup:**
  - One-command installation script
  - Automated DNS record generation and validation
  - Pre-configured SPF/DKIM/DMARC templates
  - Automated certificate management
- **UX improvements:**
  - Pre-flight checks before deployment (DNS, ports, IP reputation)
  - Plain-language error messages ("Your DNS SPF record has a typo at position 45: 'inlcude' should be 'include'")
  - Step-by-step troubleshooting guides for common issues
  - Status dashboard showing what's working/broken
- **Reduce ongoing maintenance:**
  - Automated blacklist monitoring with alerts
  - Automatic spam rule updates
  - Self-healing for common issues (cert renewal, disk space)
  - Weekly health report via email
- **Clear escape hatches:**
  - Export all data easily
  - Migrate to hosted provider without data loss
  - Don't lock users in
- **Target realistic users:**
  - Technical users comfortable with Linux
  - Small environments (personal/family, not business)
  - Users who value privacy over convenience
  - Users with time for maintenance

**Warning signs:**
- Installation failing at DNS setup step
- Users asking "why isn't this working?" without details
- Forum posts: "I give up, moving to ProtonMail"
- High churn in first 30 days
- Support requests outnumber successful deployments

**Phase to address:**
**Phase 1 (MVP)** - UX must be better than competitors (Mail-in-a-Box, Mailcow) from day one. Focus on installation experience. **Phase 2** should reduce ongoing maintenance burden with automation.

**Confidence:** HIGH - Verified with user experience research, self-hosted email blog posts, and forum discussions about why users abandon self-hosting

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-Term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip IP warmup | Launch immediately | Blacklisted IP, permanent reputation damage | Never - warmup mandatory |
| Use anonymous Docker volumes | Faster setup | Data loss on container updates | Never for production mail |
| No PTR record | Avoid provider support ticket | Emails rejected/spam-foldered | Never - PTR required |
| Self-signed certificates | Skip Let's Encrypt | Mail clients show warnings, some servers reject | Development only |
| Single point of failure (no backup relay) | Simpler architecture | When primary down, all mail lost | MVP acceptable, must fix Phase 2 |
| Manual certificate renewal | Avoid automation complexity | Silent expiration breaks email | Never - 90-day certs need automation |
| No DMARC monitoring | Skip report parsing | Miss authentication failures | MVP acceptable, add Phase 2 |
| Spam filtering on home device | Consolidated architecture | Pi out of memory | Never - filter on cloud relay |

---

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| VPS Provider | Assume port 25 open | Research SMTP policy before purchasing; maintain provider compatibility list |
| Let's Encrypt | Use webroot for validation when web server not running | Use DNS-01 challenge for mail servers, or run minimal HTTP server for HTTP-01 |
| WireGuard | Use hostname in Endpoint with local DNS | Use IP addresses or public DNS (1.1.1.1, 8.8.8.8) to prevent resolution failures |
| Docker networks | Expose port 25 to bridge network | Bind only to host network or specific IPs with authentication required |
| Dynamic DNS | Update only when IP changes | Continuous verification with TTL monitoring and update retry logic |
| Cloudflare proxy | Proxy mail records (orange cloud) | DNS-only mode (grey cloud) for MX, SPF, mail A records |

---

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Mailbox storage on SD card | Slow IMAP, corruption | Use USB SSD for mail storage on Pi | >1GB mailbox or >100 messages/day |
| No queue limits | Disk fills with spam queue | Implement queue size limits, message age limits | First spam attack (could be day one) |
| Synchronous virus scanning | Mail delivery delays >30s | Scan asynchronously after delivery, or on cloud relay only | >50 messages/day with attachments |
| Full-text search on Pi | High memory usage, OOM kills | Disable FTS on Pi, use client-side search | >5000 messages indexed |
| Logging everything at DEBUG level | Disk fills quickly | INFO level for production, DEBUG only for troubleshooting | >1000 messages/day |
| No log rotation | Disk space exhaustion | Configure logrotate with compression and retention limits | Within 30 days for active servers |

---

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Open relay (no auth required) | Spammers abuse, immediate blacklisting, permanent reputation damage | Require SASL authentication for all relay; test with MXToolbox |
| Weak DKIM keys (1024-bit) | Signature forgery possible | Use 2048-bit minimum; rotate annually |
| No rate limiting | Spam floods, DoS attacks | Limit connections per IP, messages per connection, recipient rate |
| Plaintext passwords on wire | Credential theft | Require STARTTLS for SMTP submission; TLS for IMAP/POP3 |
| No fail2ban/similar | Brute force attacks succeed | Block IPs after failed auth attempts; monitor auth logs |
| Root certificates not updated | TLS validation failures | Keep ca-certificates package updated; monitor certificate trust store |
| Insecure cipher suites | Downgrade attacks possible | Disable SSLv3, TLS 1.0, weak ciphers; use Mozilla SSL Config Generator |
| World-readable mail files | Local users read all email | Correct permissions: mail files 600, dirs 700, owned by mail user |

---

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Cryptic DNS error messages | Users don't know what to fix | "SPF record missing - add this TXT record to your DNS: ..." |
| No pre-flight checks | Deployment fails halfway through | Validate DNS, ports, IP reputation BEFORE installation starts |
| Silent certificate expiration | Email stops working, users confused | Alert 30/14/7 days before expiration; auto-renew with verification |
| No status dashboard | Users don't know if system healthy | Show: IP reputation, authentication status, queue size, recent deliverability |
| Assuming Linux expertise | Installation fails for 80% of users | Provide one-command install; automate complex steps |
| No migration guide | Users locked in to DarkPipe | Document export process; provide scripts to move to other platforms |
| Overwhelming maintenance tasks | Users give up after 2 weeks | Automate: blacklist monitoring, spam rules, certificate renewal, disk cleanup |

---

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Email sending works:** Often missing - IP warmup period (4-6 weeks) - verify gradual volume ramp scheduled
- [ ] **DNS configured:** Often missing - PTR record from provider - verify reverse DNS resolves correctly
- [ ] **TLS certificates:** Often missing - Auto-renewal testing - verify `certbot renew --dry-run` succeeds
- [ ] **SPF/DKIM/DMARC:** Often missing - Alignment verification - verify mail-tester.com scores 9+/10
- [ ] **Docker volumes:** Often missing - Backup/restore testing - verify restore from backup actually works
- [ ] **WireGuard tunnel:** Often missing - Reconnection after ISP drop - verify tunnel recovers from home internet outage
- [ ] **Open relay testing:** Often missing - External relay test - verify MXToolbox shows "not an open relay"
- [ ] **Monitoring setup:** Often missing - Alerting verification - verify alerts actually fire and reach operators
- [ ] **IP reputation:** Often missing - Multi-RBL check - verify not listed on Spamhaus, Barracuda, SpamRATS, etc.
- [ ] **Rate limiting:** Often missing - Abuse testing - verify system blocks rapid-fire spam attempts

---

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Blacklisted IP | HIGH (4-8 weeks) | 1. Stop all mail. 2. Request removal from each RBL (some auto-delist after 48hrs). 3. If permanent, request new IP from provider. 4. Start IP warmup from scratch. |
| Open relay exploited | HIGH (permanent reputation damage possible) | 1. Immediately restrict relay (require auth). 2. Flush mail queue. 3. Request RBL removal. 4. Monitor for re-listing. 5. Consider new IP if damage severe. |
| Certificate expired | MEDIUM (4-8 hours) | 1. Renew immediately: `certbot renew --force-renewal`. 2. Restart mail services. 3. Test with `openssl s_client`. 4. Investigate why auto-renewal failed. 5. Set up monitoring. |
| Data loss (volumes) | HIGH (may be unrecoverable) | 1. Stop container immediately. 2. Check `docker volume ls` for dangling volumes. 3. Attempt data recovery from Docker layers. 4. Restore from backup (if exists). 5. If no backup: data lost. |
| WireGuard tunnel down | LOW (30 minutes) | 1. Restart WireGuard on both ends. 2. Check home IP hasn't changed (update DDNS if needed). 3. Verify keepalive settings. 4. Test connectivity. 5. Process queued mail. |
| SPF/DKIM misconfigured | MEDIUM (2-24 hours) | 1. Fix DNS records immediately. 2. Wait for DNS propagation (use low TTL). 3. Test with mail-tester.com. 4. Send to Gmail, check headers. 5. Monitor deliverability. |
| Out of disk space | MEDIUM (2-4 hours) | 1. Clear logs: `journalctl --vacuum-time=7d`. 2. Clean mail queue of spam. 3. Run `docker system prune`. 4. Add disk space monitoring. 5. Implement log rotation. |
| Memory exhaustion (Pi) | LOW (15 minutes) | 1. Restart mail services. 2. Identify memory hog with `ps aux`. 3. Disable heavy services (ClamAV, etc.). 4. Monitor with `htop`. 5. Consider offloading to cloud relay. |

---

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| VPS port 25 blocked | Phase 0 (Infrastructure) | Telnet to port 25 from external IP succeeds |
| New IP blacklisted | Phase 1 (MVP Deployment) | MXToolbox multi-RBL check shows clean |
| Missing SPF/DKIM/DMARC | Phase 1 (MVP Deployment) | Mail-tester.com scores 9+/10 |
| Missing PTR record | Phase 1 (MVP Deployment) | `dig -x IP` returns mail server FQDN |
| Residential IP blacklisted | Phase 1 (Architecture) | Home device never sends direct SMTP |
| TLS certificate expiration | Phase 1 (MVP Deployment) | Certbot auto-renewal tested and monitored |
| Open relay | Phase 1 (MVP Deployment) | MXToolbox open relay test shows "not open" |
| WireGuard tunnel failure | Phase 1 (MVP Deployment) | Tunnel recovers automatically from home internet drop |
| Docker volume data loss | Phase 1 (MVP Deployment) | Container update preserves data; restore from backup tested |
| Pi memory exhaustion | Phase 1 (MVP Deployment) | Memory usage <80% under realistic load |
| Complex setup UX | Phase 1 (MVP) | Non-expert user completes setup in <2 hours |
| IP reputation (no warmup) | Phase 1-2 (Warmup Period) | Deliverability >95% to Gmail/Outlook after 6 weeks |
| Blacklist monitoring | Phase 2 (Scaling) | Automated alerts fire within 6 hours of listing |
| DMARC report analysis | Phase 2 (Scaling) | Weekly reports parsed, failures alerted |
| Queue management | Phase 2 (Scaling) | Automatic cleanup of deferred/old messages |
| Log rotation | Phase 2 (Scaling) | Logs compressed and retained for 30 days max |
| Backup/disaster recovery | Phase 2 (Scaling) | Monthly restore test from backup succeeds |
| Rate limiting abuse | Phase 2 (Scaling) | System blocks spam floods automatically |

---

## Sources

### VPS Provider Port 25 Policies
- [Best Mail Server Providers - Guide 2026 - Forward Email](https://forwardemail.net/en/blog/docs/best-mail-server-providers)
- [GitHub - awesome-mail-server-providers](https://github.com/forwardemail/awesome-mail-server-providers)
- [Why Is SMTP Blocked?? | Vultr Docs](https://docs.vultr.com/support/products/compute/why-is-smtp-blocked)
- [Why is SMTP blocked? | DigitalOcean Documentation](https://docs.digitalocean.com/support/why-is-smtp-blocked/)
- [Send email on Akamai Cloud](https://techdocs.akamai.com/cloud-computing/docs/send-email)
- [BuyVM FAQ - Frantech/BuyVM Wiki](https://wiki.buyvm.net/doku.php/faq)
- [OVHcloud AntiSpam - Best practices](https://support.us.ovhcloud.com/hc/en-us/articles/16100926574995)
- [Setting up SMTP | Scaleway Documentation](https://www.scaleway.com/en/docs/transactional-email/reference-content/smtp-configuration/)

### IP Reputation and Blacklisting
- [Email Reputation Management for VPS Hosting - ServerSpan](https://www.serverspan.com/en/blog/email-reputation-management-for-vps-hosting-beyond-spf-dkim-and-dmarc)
- [The 2026 Playbook for Email Sender Reputation - Smartlead](https://www.smartlead.ai/blog/how-to-improve-email-sender-reputation)
- [Spamhaus Policy Blocklist (PBL)](https://brandergroup.net/whats-spamhaus-policy-blocklist-pbl/)
- [SpamRATS RATS Dyna Blacklist - Suped](https://www.suped.com/blocklists/spamrats-rats-dyna-blacklist)

### Helm Email Server Analysis
- [Helm Email Server: secure and stylish, but has issues - Web Informant](https://blog.strom.com/wp/?p=6990)
- [Helm Personal Email Server | Hacker News](https://news.ycombinator.com/item?id=18238581)
- [What happens to my email if power/Internet goes out? - Helm Support](https://support.thehelm.com/hc/en-us/articles/230119308)

### SPF/DKIM/DMARC Deliverability
- [How to Host Your Own Email Server in 2026 - Elementor](https://elementor.com/blog/how-to-host-your-own-email-server/)
- [Email Deliverability in 2026: SPF, DKIM, DMARC Checklist - EGen Consulting](https://www.egenconsulting.com/blog/email-deliverability-2026.html)
- [SPF, DKIM, DMARC: Common Setup Mistakes - InfraForge](https://www.infraforge.ai/blog/spf-dkim-dmarc-common-setup-mistakes)
- [The Ultimate SPF / DKIM / DMARC Best Practices 2026 - Uriports](https://www.uriports.com/blog/spf-dkim-dmarc-best-practices/)

### PTR Records and Reverse DNS
- [How Reverse DNS Impacts Email Deliverability - InfraForge](https://www.infraforge.ai/blog/reverse-dns-email-deliverability-impact)
- [PTR Records and Email Sending [2026 Update] - Mailtrap](https://mailtrap.io/blog/ptr-records/)
- [RATS-NoPtr Blacklist - Warmy](https://www.warmy.io/blog/rats-noptr-blacklist-how-to-get-delist/)

### Self-Hosted Email Sustainability
- [After self-hosting my email for twenty-three years I have thrown in the towel](https://cfenollosa.com/blog/after-self-hosting-my-email-for-twenty-three-years-i-have-thrown-in-the-towel-the-oligopoly-has-won.html)
- [Self-Hosting Email in 2026: Is Running a Linux Mail Server Still Worth It? - Security Boulevard](https://securityboulevard.com/2025/12/self-hosting-email-in-2026-is-running-a-linux-mail-server-still-worth-it/)
- [Self-Hosted email is the hardest it's ever been, but also the easiest - Vadosware](https://vadosware.io/post/its-never-been-easier-or-harder-to-self-host-email)

### TLS and Certificate Management
- [Understanding Common SSL Misconfigurations - Encryption Consulting](https://www.encryptionconsulting.com/understanding-common-ssl-misconfigurations-and-how-to-prevent-them/)
- [SSL & TLS Certificate Errors in Email Servers - Warmy](https://www.warmy.io/blog/ssl-and-tls-certificate-errors-in-email-servers-how-they-impact-deliverability/)
- [Top 10 SSL/TLS Misconfigurations, Risks and Solutions - CheapSSLWeb](https://cheapsslweb.com/blog/top-10-ssl-tls-misconfigurations-risks-and-its-solutions/)

### Open Relay Prevention
- [SMTP Open Relay Vulnerabilities - DuoCircle](https://www.duocircle.com/email-security/smtp-open-relay-vulnerabilities-how-to-prevent-security-breaches)
- [What Is An SMTP Relay Attack? - Twingate](https://www.twingate.com/blog/glossary/smtp%20relay%20attack)
- [About open relay on SMTP servers - Broadcom](https://techdocs.broadcom.com/us/en/symantec-security-software/email-security/email-security-cloud/1-0/about-email-anti-malware/about-open-relay-on-smtp-servers.html)

### Raspberry Pi Limitations
- [15 Notable Open Source Email Servers for Raspberry Pi 2026 - Forward Email](https://forwardemail.net/en/blog/open-source/raspberry-pi-email-server)
- [Raspberry Pi Email Server complete solution - Raspberry Pi Forums](https://forums.raspberrypi.com/viewtopic.php?t=210084)

### Docker Volume Management
- [How to Use Docker Volumes for Persistent Data - OneUptime](https://oneuptime.com/blog/post/2026-02-02-docker-volumes-persistent-data/view)
- [12 Best Practices for Docker Volume Management - DevOps Training Institute](https://www.devopstraininginstitute.com/blog/12-best-practices-for-docker-volume-management)
- [FAQ - Docker Mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/faq/)

### WireGuard Tunnel Management
- [Troubleshooting WireGuard DNS Issues - Pro Custodibus](https://www.procustodibus.com/blog/2023/09/troubleshooting-wireguard-dns-issues/)
- [WireGuard breaks with NetworkManager - Arch Linux Forums](https://bbs.archlinux.org/viewtopic.php?id=289926)

### Email Rate Limiting and Throttling
- [Email Throttling Strategies - Mailpool](https://www.mailpool.ai/blog/email-throttling-strategies-managing-send-limits-across-multiple-providers)
- [Mastering Email Throttling - Allegrow](https://www.allegrow.co/knowledge-base/email-throttling-deliverability)
- [Email Sending Limits by Provider: 2026 Complete Guide - GrowthList](https://growthlist.co/email-sending-limits-of-various-email-service-providers/)

### IP Warmup Best Practices
- [Master Email Warm Up in 2026 [Full Guide] - Mailwarm](https://www.mailwarm.com/blog/email-warm-up)
- [Building a Strong Email Reputation With IP Warm-Up - Iterable](https://iterable.com/blog/building-a-strong-email-reputation-with-ip-warm-up/)
- [Email IP Reputation Explained [2026] - Mailtrap](https://mailtrap.io/blog/email-ip-reputation/)

### GDPR and Legal Compliance
- [Email Privacy Laws & Regulations 2026: GDPR, CCPA Guide - Mailbird](https://www.getmailbird.com/email-privacy-laws-regulations-compliance/)
- [Complete GDPR Compliance Guide (2026-Ready) - SecurePrivacy](https://secureprivacy.ai/blog/gdpr-compliance-2026)
- [GDPR Compliance for U.S. Companies: The 2026 Definitive Guide - MeetERGO](https://meetergo.com/en/magazine/gdpr-compliance-for-us-companies)

### Disaster Recovery and Backup
- [Exchange Server disaster recovery - Microsoft Learn](https://learn.microsoft.com/en-us/exchange/high-availability/disaster-recovery/disaster-recovery)
- [What is email backup? - Barracuda Networks](https://www.barracuda.com/support/glossary/email-backup)

### Mail Queue Management
- [The Complete Guide to Postfix Mail Queue Management - TheLinuxCode](https://thelinuxcode.com/postfix_mail_queue_management/)
- [Queues and messages in queues in Exchange Server - Microsoft Learn](https://learn.microsoft.com/en-us/exchange/mail-flow/queues/queues)
- [Postfix deferred queue - Bobcares](https://bobcares.com/blog/postfix-deferred-queue/)

---

*Pitfalls research for: Self-Hosted Email Relay System (DarkPipe)*
*Researched: 2026-02-08*
*Confidence: HIGH - Based on official documentation, community research, and current 2026 standards*