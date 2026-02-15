# Milestones

## v1.0 MVP (Shipped: 2026-02-15)

**Phases completed:** 10 phases, 29 plans | 129 commits | 356 files | 38K LOC (Go)
**Timeline:** 7 days (2026-02-08 → 2026-02-15)
**Archive:** [Roadmap](milestones/v1.0-ROADMAP.md) | [Requirements](milestones/v1.0-REQUIREMENTS.md)

**Delivered:** Complete self-hosted email sovereignty stack — cloud relay, home mail server, DNS authentication, offline queuing, webmail, calendar/contacts, build system, device onboarding, monitoring, and mail migration.

**Key accomplishments:**
- Encrypted WireGuard + mTLS transport with NAT traversal and auto-reconnection
- Cloud relay SMTP gateway with TLS enforcement, ephemeral storage, and Let's Encrypt automation
- Home mail server (Stalwart/Maddy/Postfix+Dovecot) with multi-user, multi-domain, Rspamd spam filtering
- Automated SPF/DKIM/DMARC with Cloudflare/Route53 API integration and DNS validation CLI
- Encrypted offline queue with age encryption, S3-compatible overflow, configurable queue-or-bounce
- Webmail (Roundcube/SnappyMail) and CalDAV/CardDAV with shared family calendar/contacts
- GitHub Actions multi-arch build pipeline with interactive setup CLI for tiered UX
- Device onboarding: Apple .mobileconfig, QR codes, Thunderbird/Outlook autodiscovery, app passwords
- Health monitoring with delivery tracking, certificate lifecycle management, and web dashboard
- Mail migration from 7 providers (Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, generic) with CLI wizard

---

