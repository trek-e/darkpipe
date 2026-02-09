# DarkPipe

## What This Is

DarkPipe is a cloud-fronted, personal-device-backed email service built entirely on open source technology. A minimal cloud relay handles internet-facing SMTP (receiving and sending), then securely transports messages — without storing them — to a mail server running on hardware the user controls at home. Users choose their own mail stack, calendar, contacts, and webmail components through a modular build system. It's email sovereignty without giving up deliverability.

## Core Value

Your email lives on your hardware, encrypted in transit, never stored on someone else's server — and it still works like normal email from the outside.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Cloud relay receives inbound SMTP with TLS (certbot/Let's Encrypt)
- [ ] Cloud relay sends outbound SMTP via direct MTA delivery
- [ ] TLS enforced on all connections (SSL/STARTTLS), with optional strict mode to refuse plaintext peers
- [ ] Notification when a remote mail server does not support a secure endpoint
- [ ] Encrypted transport from cloud relay to home device (WireGuard or mTLS, user-selectable)
- [ ] Certificate management with configurable rotation (30/60/90 days)
- [ ] Home device runs user-selected mail server (Postfix/Dovecot, Stalwart, Maddy, etc.)
- [ ] Home device supports optional calendar/contacts (Radicale, Baikal)
- [ ] Home device supports optional webmail (Roundcube, Snappymail)
- [ ] GitHub Actions build pipeline — users select components, outputs custom Docker images
- [ ] Pre-built full-featured Docker image available as alternative to custom build
- [ ] Runs on Raspberry Pi 4+ (arm64), x64/arm64 Docker/Podman, TrueNAS Scale, Unraid
- [ ] DKIM key generation, SPF/DMARC record generation — automated
- [ ] DNS API integration for supported providers (Cloudflare, Route53, etc.) with manual guide fallback
- [ ] Optional encrypted message queue on cloud relay when home device is offline
- [ ] Optional queue overflow to Storj or S3-compatible object storage (encrypted at rest)
- [ ] User can disable queueing entirely — mail bounces if home device is unreachable
- [ ] Tiered UX — simple defaults for non-technical users, full control for power users
- [ ] Cloud relay runs in smallest possible container footprint

### Out of Scope

- Smart host routing (post-v1 research — concerns about message inspection by third parties)
- Managed "SMTP as a service" relay hosting (post-v1 paid tier)
- Pre-configured hardware sales (post-v1 — learn from Helm's failure)
- Mobile apps (web-first via webmail component)

## Context

**Privacy motivation:** Existing email services require trusting a third party with message storage. Even "privacy-focused" providers hold your keys or your mail. DarkPipe eliminates that by keeping storage on user-owned hardware while maintaining a cloud presence for deliverability and DNS compliance.

**Helm precedent:** Helm (thehelm.com) attempted a similar vision with proprietary hardware. They failed — hardware dependency created a single point of failure and unsustainable economics. DarkPipe learns from this: software-only, open source, runs on commodity hardware the user already owns.

**SMTP port 25 restrictions:** Many VPS providers (notably DigitalOcean) block port 25 by default. Research required to identify reputable providers that allow SMTP without friction. This is a critical deployment constraint.

**Sustainability model:** AGPLv3 licensed, open source, community-contributed. Funded via donations (GitHub Sponsors, Open Collective, Liberapay/Ko-fi). Post-v1, explore managed relay hosting as a paid service to fund ongoing development.

**Community:** GitHub Discussions for all community interaction — issues, support, feature requests.

**Target platforms:**
- Raspberry Pi 4 and up (arm64)
- Docker or Podman on x64 or arm64 (including Mac)
- TrueNAS Scale applications
- Unraid
- Any Docker-capable device

## Constraints

- **License**: AGPLv3 — protects against closed forks, especially important if managed hosting launches post-v1
- **Container size**: Cloud relay must be as small as possible — minimal base image, no unnecessary dependencies
- **No cloud storage**: Messages must not be stored unencrypted on cloud relay; in-flight only unless user explicitly enables encrypted queue
- **Port 25**: Cloud relay hosting must be on providers that allow inbound/outbound SMTP port 25
- **ARM64 support**: All components must build and run on arm64 (Raspberry Pi is a primary target)
- **Open source only**: All mail stack components (mail server, calendar, contacts, webmail) must be open source

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| AGPLv3 license | Prevents closed forks, protects community investment, critical for future managed service | — Pending |
| Direct MTA for v1 (no smart host) | Smart hosts can inspect messages, defeating privacy promise | — Pending |
| WireGuard + mTLS as transport options | WireGuard for simplicity, mTLS for minimal footprint — user choice | — Pending |
| Certbot for public TLS, separate internal CA for relay↔home | Let's Encrypt for internet-facing, dedicated certs with configurable rotation for internal transport | — Pending |
| GitHub Actions for custom image builds | Users configure components via workflow, get tailored images without local build tooling | — Pending |
| Storj for overflow queue storage | S3-compatible, decentralized, aligns with privacy ethos better than AWS S3 | — Pending |
| No proprietary hardware | Learn from Helm's failure — software-only, commodity hardware | — Pending |
| Multiple donation channels | GitHub Sponsors + Open Collective + Liberapay/Ko-fi for maximum reach | — Pending |
| GitHub Discussions for community | Keep community close to code, single platform, low maintenance overhead | — Pending |

---
*Last updated: 2026-02-08 after initialization*
