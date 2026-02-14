# Requirements: DarkPipe

**Defined:** 2026-02-08
**Core Value:** Your email lives on your hardware, encrypted in transit, never stored on someone else's server — and it still works like normal email from the outside.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Core Email Transport

- [x] **RELAY-01**: Cloud relay receives inbound SMTP with TLS (Let's Encrypt)
- [x] **RELAY-02**: Cloud relay forwards messages to home device without persistent storage (in-flight only)
- [x] **RELAY-03**: Cloud relay sends outbound SMTP via direct MTA delivery
- [x] **RELAY-04**: TLS enforced on all SMTP connections (SSL/STARTTLS)
- [x] **RELAY-05**: Optional strict mode to refuse mail from plaintext-only peers
- [x] **RELAY-06**: User notified when a remote mail server does not support a secure endpoint
- [x] **RELAY-07**: IMAP server on home device for mail client access
- [x] **RELAY-08**: SMTP submission (port 587) on home device for sending from clients

### Transport Security

- [x] **TRNS-01**: Encrypted WireGuard tunnel from cloud relay to home device
- [x] **TRNS-02**: Alternative mTLS transport option (user-selectable)
- [x] **TRNS-03**: Transport auto-reconnects after home internet interruption
- [x] **TRNS-04**: NAT traversal without port forwarding on home network

### Certificate Management

- [x] **CERT-01**: Let's Encrypt certificates via Certbot for public-facing TLS
- [x] **CERT-02**: Internal CA (step-ca) for relay↔home transport certificates
- [ ] **CERT-03**: Configurable certificate rotation (30/60/90 days)
- [ ] **CERT-04**: Certificate expiry monitoring with alerts

### Email Authentication & Deliverability

- [x] **AUTH-01**: Automated SPF record generation
- [x] **AUTH-02**: Automated DKIM key generation and signing (2048-bit minimum)
- [x] **AUTH-03**: Automated DMARC policy generation
- [x] **AUTH-04**: DNS validation checker (verify SPF/DKIM/DMARC/MX/PTR setup)
- [x] **AUTH-05**: DNS API integration for supported providers (Cloudflare, Route53, etc.)
- [x] **AUTH-06**: Manual DNS setup guide with copy-paste templates for unsupported providers
- [x] **AUTH-07**: Reverse DNS (PTR) verification and setup documentation

### Home Mail Server

- [x] **MAIL-01**: User-selectable mail server (Postfix+Dovecot, Stalwart, or Maddy)
- [x] **MAIL-02**: Multi-user mailbox support
- [x] **MAIL-03**: Multi-domain support
- [x] **MAIL-04**: Mail aliases and catch-all addresses
- [x] **MAIL-05**: Spam filtering via Rspamd
- [x] **MAIL-06**: Greylisting for spam reduction

### Webmail

- [ ] **WEB-01**: Web-based email client (Roundcube or SnappyMail, user-selectable)
- [ ] **WEB-02**: Mobile-responsive webmail UI

### Calendar & Contacts

- [ ] **CAL-01**: CalDAV server for calendar sync (Radicale, Baikal, or Stalwart built-in)
- [ ] **CAL-02**: CardDAV server for contacts sync

### Build System

- [ ] **BUILD-01**: GitHub Actions build pipeline — users select components via workflow config
- [ ] **BUILD-02**: Multi-architecture Docker images (arm64 + amd64)
- [ ] **BUILD-03**: Pre-built full-featured Docker image as alternative to custom build

### Queue & Offline Handling

- [x] **QUEUE-01**: Optional encrypted message queue on cloud relay when home device is offline
- [x] **QUEUE-02**: Optional queue overflow to Storj or S3-compatible object storage (encrypted at rest)
- [x] **QUEUE-03**: User can disable queuing entirely — mail bounces if home device unreachable

### Device Profiles & Client Setup

- [ ] **PROF-01**: Auto-generated Apple .mobileconfig profiles for iOS/macOS
- [ ] **PROF-02**: Auto-generated Android autoconfig profiles
- [ ] **PROF-03**: QR code generation for quick device setup
- [ ] **PROF-04**: Desktop mail client autodiscovery (Thunderbird autoconfig, Outlook autodiscover)
- [ ] **PROF-05**: App-generated passwords — users never create or manage mail passwords directly

### Monitoring

- [ ] **MON-01**: Mail queue health monitoring
- [ ] **MON-02**: Mail delivery status visibility
- [ ] **MON-03**: Cloud relay container health checks

### Deployment & UX

- [ ] **UX-01**: Tiered experience — simple defaults for non-technical users, full control for power users
- [x] **UX-02**: Cloud relay runs in smallest possible container footprint
- [ ] **UX-03**: Runs on Raspberry Pi 4+ (arm64), x64/arm64 Docker/Podman, TrueNAS Scale, Unraid

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Security

- **SEC-01**: MTA-STS policy publication for TLS enforcement
- **SEC-02**: DANE TLSA records for DNSSEC-based certificate verification
- **SEC-03**: TLS-RPT (TLS reporting) for visibility into TLS failures
- **SEC-04**: PGP/WKD support for end-to-end encryption
- **SEC-05**: 2FA (TOTP) for webmail and admin access

### Advanced Mail Features

- **ADV-01**: Sieve filtering rules with ManageSieve protocol
- **ADV-02**: Vacation/auto-reply via Sieve
- **ADV-03**: Full-text search with attachment indexing
- **ADV-04**: Rate limiting for outbound mail (anti-abuse)

### Operations

- **OPS-01**: Prometheus metrics export for homelab monitoring
- **OPS-02**: Backup/restore automation
- **OPS-03**: Audit logging for admin actions
- **OPS-04**: Deliverability scoring integration (mail-tester.com or similar)
- **OPS-05**: Automated blacklist monitoring with alerts

### Groupware

- **GRP-01**: Calendar web UI (view/edit without native client)
- **GRP-02**: Contacts web UI (manage without native client)
- **GRP-03**: Shared calendar support

### Protocol Support

- **PROTO-01**: ActiveSync (Exchange ActiveSync) for push email/contacts/calendar on mobile (evaluate Z-Push or Stalwart built-in)

### AI & Automation

- **AI-01**: Optional AI agent integration — users connect their own AI to act on messages (triage, reply drafts, summarization)
- **AI-02**: Scripting/rules engine for Gmail-like automation (auto-label, auto-forward, auto-archive based on conditions)
- **AI-03**: Webhook/API hooks for external automation (n8n, Home Assistant, custom scripts)

### Organization

- **ORG-01**: Advanced mail organization features (labels, tags, smart folders beyond basic IMAP folders)
- **ORG-02**: Priority inbox or importance scoring
- **ORG-03**: Conversation threading and grouping improvements

### Community & Governance

- **COM-01**: GitHub feature request section (Discussions category) for community-driven post-v1 prioritization
- **COM-02**: Public roadmap visibility — community sees what's planned and can vote on priorities

### Post-v1 Service

- **SVC-01**: Managed SMTP relay hosting as paid service option
- **SVC-02**: Smart host routing (research viability — privacy concerns around message inspection)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| POP3 support | Security liability; IMAP is the 2026 baseline. POP3 provides no sync, no folders, and encourages insecure plaintext auth patterns |
| Native mobile apps | Web-first via webmail + standard IMAP clients (Apple Mail, K-9, FairEmail) |
| Real-time chat/messaging | Different problem domain — email is asynchronous |
| Built-in file sharing (Nextcloud-style) | Massive scope creep; users install Nextcloud separately if needed |
| AI-powered spam/categorization | Privacy concern — adds dependencies, complexity; Rspamd Bayesian learning sufficient |
| Built-in VPN/Tor routing | Separate concern; document how to run behind VPN if desired |
| OAuth/social login | Email+password + app-generated passwords sufficient for v1 |
| Automatic Gmail/Outlook migration | High complexity, API changes break it; provide manual migration guide |
| Self-hosted DNS server | Increases attack surface; use existing DNS providers with API integration |
| Multi-tenant SaaS mode | DarkPipe is self-hosted for single user/family; multi-tenancy adds vast complexity |
| Windows/macOS native server | Linux is email server standard; production = Linux containers |
| Proprietary hardware | Helm failed this way — software-only, commodity hardware |
| Blockchain-based identity | Immature, complex, confusing — standard DNS + TLS certificates |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| TRNS-01 | Phase 1: Transport Layer | ✓ Complete |
| TRNS-02 | Phase 1: Transport Layer | ✓ Complete |
| TRNS-03 | Phase 1: Transport Layer | ✓ Complete |
| TRNS-04 | Phase 1: Transport Layer | ✓ Complete |
| CERT-02 | Phase 1: Transport Layer | ✓ Complete |
| RELAY-01 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-02 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-03 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-04 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-05 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-06 | Phase 2: Cloud Relay | ✓ Complete |
| CERT-01 | Phase 2: Cloud Relay | ✓ Complete |
| UX-02 | Phase 2: Cloud Relay | ✓ Complete |
| RELAY-07 | Phase 3: Home Mail Server | ✓ Complete |
| RELAY-08 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-01 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-02 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-03 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-04 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-05 | Phase 3: Home Mail Server | ✓ Complete |
| MAIL-06 | Phase 3: Home Mail Server | ✓ Complete |
| AUTH-01 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-02 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-03 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-04 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-05 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-06 | Phase 4: DNS & Email Authentication | ✓ Complete |
| AUTH-07 | Phase 4: DNS & Email Authentication | ✓ Complete |
| QUEUE-01 | Phase 5: Queue & Offline Handling | ✓ Complete |
| QUEUE-02 | Phase 5: Queue & Offline Handling | ✓ Complete |
| QUEUE-03 | Phase 5: Queue & Offline Handling | ✓ Complete |
| WEB-01 | Phase 6: Webmail & Groupware | ✓ Complete |
| WEB-02 | Phase 6: Webmail & Groupware | ✓ Complete |
| CAL-01 | Phase 6: Webmail & Groupware | ✓ Complete |
| CAL-02 | Phase 6: Webmail & Groupware | ✓ Complete |
| BUILD-01 | Phase 7: Build System & Deployment | ✓ Complete |
| BUILD-02 | Phase 7: Build System & Deployment | ✓ Complete |
| BUILD-03 | Phase 7: Build System & Deployment | ✓ Complete |
| UX-01 | Phase 7: Build System & Deployment | ✓ Complete |
| UX-03 | Phase 7: Build System & Deployment | ✓ Complete |
| PROF-01 | Phase 8: Device Profiles & Client Setup | Pending |
| PROF-02 | Phase 8: Device Profiles & Client Setup | Pending |
| PROF-03 | Phase 8: Device Profiles & Client Setup | Pending |
| PROF-04 | Phase 8: Device Profiles & Client Setup | Pending |
| PROF-05 | Phase 8: Device Profiles & Client Setup | Pending |
| MON-01 | Phase 9: Monitoring & Observability | Pending |
| MON-02 | Phase 9: Monitoring & Observability | Pending |
| MON-03 | Phase 9: Monitoring & Observability | Pending |
| CERT-03 | Phase 9: Monitoring & Observability | Pending |
| CERT-04 | Phase 9: Monitoring & Observability | Pending |

**Coverage:**
- v1 requirements: 50 total (12 categories)
- Mapped to phases: 50/50
- Unmapped: 0

---
*Requirements defined: 2026-02-08*
*Last updated: 2026-02-08 after roadmap creation — traceability complete*
