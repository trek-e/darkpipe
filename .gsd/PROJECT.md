# DarkPipe

## What This Is

DarkPipe is a complete self-hosted email sovereignty stack. A minimal cloud relay handles internet-facing SMTP (receiving and sending), then securely transports messages — without storing them — to a mail server running on hardware the user controls at home. Users choose their mail server (Stalwart, Maddy, or Postfix+Dovecot), webmail (Roundcube or SnappyMail), and calendar/contacts components through a modular build system. The system includes automated DNS authentication, encrypted offline queuing, device onboarding via QR codes and profiles, monitoring with a web dashboard, and migration tools for 7 popular providers.

## Core Value

Your email lives on your hardware, encrypted in transit, never stored on someone else's server — and it still works like normal email from the outside.

## Current State

**Shipped:** v1.0 MVP (2026-02-15)
**Codebase:** 38K LOC (Go), 356 files, 10 phases, 29 plans
**Tech stack:** Go 1.25, Postfix, Stalwart/Maddy/Dovecot, Rspamd, Redis, Caddy, Roundcube/SnappyMail, Radicale, WireGuard, step-ca, age encryption, Docker multi-arch

**Known issues:**
- go-imap v2 is beta (v2.0.0-beta.8) — monitor for breaking changes
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) — schema may change
- VPS port 25 restrictions are deployment-specific — provider validation required
- IP warmup requires 4-6 weeks after deployment — time-based, not development

## Requirements

### Validated

- ✓ Cloud relay receives inbound SMTP with TLS (Let's Encrypt) — v1.0
- ✓ Cloud relay sends outbound SMTP via direct MTA delivery — v1.0
- ✓ TLS enforced on all SMTP connections with optional strict mode — v1.0
- ✓ Notification when remote server lacks TLS support — v1.0
- ✓ Encrypted transport (WireGuard or mTLS, user-selectable) — v1.0
- ✓ Certificate management with configurable rotation (30/60/90 days) — v1.0
- ✓ Home device runs user-selected mail server (Stalwart, Maddy, Postfix+Dovecot) — v1.0
- ✓ CalDAV/CardDAV with shared family calendar/contacts — v1.0
- ✓ Webmail (Roundcube or SnappyMail, mobile-responsive) — v1.0
- ✓ GitHub Actions multi-arch build pipeline with component selection — v1.0
- ✓ Pre-built Docker images for two stack configurations — v1.0
- ✓ Runs on RPi4+, x64/arm64 Docker, TrueNAS Scale, Unraid — v1.0
- ✓ Automated SPF/DKIM/DMARC with DNS validation CLI — v1.0
- ✓ DNS API integration (Cloudflare, Route53) with manual guide fallback — v1.0
- ✓ Encrypted offline queue with S3-compatible overflow — v1.0
- ✓ Configurable queue-or-bounce behavior — v1.0
- ✓ Tiered UX — simple defaults for non-technical users, full control for power users — v1.0
- ✓ Minimal cloud relay container footprint — v1.0
- ✓ Multi-user, multi-domain with aliases and catch-all — v1.0
- ✓ Spam filtering (Rspamd) with greylisting — v1.0
- ✓ Device onboarding via Apple .mobileconfig, QR codes, Thunderbird/Outlook autodiscovery — v1.0
- ✓ App-generated passwords for mail clients — v1.0
- ✓ Mail queue health monitoring and delivery status tracking — v1.0
- ✓ Certificate expiry alerts and automatic renewal — v1.0
- ✓ Web-based monitoring dashboard — v1.0
- ✓ Mail migration from 7 providers with CLI wizard and dry-run — v1.0

### Active

(None — fresh requirements defined via `/gsd:new-milestone`)

### Out of Scope

- Smart host routing — concerns about message inspection by third parties
- Managed "SMTP as a service" relay hosting — post-v1 paid tier
- Pre-configured hardware sales — learn from Helm's failure
- Mobile apps — web-first via webmail + standard IMAP clients
- POP3 support — security liability; IMAP is the 2026 baseline

## Context

**Privacy motivation:** Existing email services require trusting a third party with message storage. Even "privacy-focused" providers hold your keys or your mail. DarkPipe eliminates that by keeping storage on user-owned hardware while maintaining a cloud presence for deliverability and DNS compliance.

**Helm precedent:** Helm (thehelm.com) attempted a similar vision with proprietary hardware. They failed — hardware dependency created a single point of failure and unsustainable economics. DarkPipe learns from this: software-only, open source, runs on commodity hardware the user already owns.

**SMTP port 25 restrictions:** Many VPS providers (notably DigitalOcean) block port 25. VPS provider guide included with SMTP compatibility matrix. This remains a deployment constraint.

**Sustainability model:** AGPLv3 licensed, open source, community-contributed. Funded via donations (GitHub Sponsors, Open Collective, Liberapay/Ko-fi). Post-v1, explore managed relay hosting as a paid service.

**Community:** GitHub Discussions for all community interaction.

**Target platforms:**
- Raspberry Pi 4+ (arm64, 4GB RAM recommended)
- Docker or Podman on x64 or arm64 (including Mac Silicon)
- TrueNAS Scale 24.10+ applications
- Unraid Community Applications
- Proxmox LXC containers
- Synology Container Manager

## Constraints

- **License**: AGPLv3
- **Container size**: Cloud relay minimal footprint (~35MB)
- **No cloud storage**: Messages not stored unencrypted on cloud relay
- **Port 25**: Cloud relay requires VPS providers that allow SMTP
- **ARM64 support**: All components build and run on arm64
- **Open source only**: All mail stack components must be open source

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| AGPLv3 license | Prevents closed forks, protects community investment | ✓ Good |
| Direct MTA for v1 (no smart host) | Smart hosts can inspect messages, defeating privacy promise | ✓ Good |
| WireGuard + mTLS as transport options | WireGuard for simplicity, mTLS for minimal footprint | ✓ Good |
| Certbot for public TLS, step-ca for internal | Let's Encrypt internet-facing, private CA for transport | ✓ Good |
| GitHub Actions for custom image builds | Users configure components via workflow dispatch | ✓ Good |
| S3-compatible overflow (Storj/AWS/MinIO) | Flexible encrypted overflow for offline queuing | ✓ Good |
| No proprietary hardware | Software-only, commodity hardware (Helm lesson) | ✓ Good |
| Docker Compose profiles for component selection | Single compose file, profile flags for stack selection | ✓ Good |
| Separate Go modules for setup/profiles tools | Isolates CLI dependencies (survey, pterm) from core services | ✓ Good |
| Dry-run by default for DNS and migration | Safe operations — require explicit --apply for changes | ✓ Good |
| emersion/go-imap v2 beta for migration | Only Go IMAP v2 client; beta risk accepted, monitor updates | ⚠️ Revisit |
| Stalwart as default mail server | Most feature-complete (built-in CalDAV/CardDAV); pre-v1.0 risk | ⚠️ Revisit |
| OAuth2 device flow for Gmail/Outlook migration | RFC 8628 — no browser redirect needed for CLI tool | ✓ Good |

## Milestones

| Milestone | Status | Completed |
|-----------|--------|-----------|
| M001: MVP (Phases 1-10) | ✓ Shipped | 2026-02-15 |
| M002: Post-Launch Hardening | Planning | — |

---
*Last updated: 2026-03-11 after M001 summary*
