# Roadmap: DarkPipe

## Overview

DarkPipe delivers email sovereignty by splitting the mail stack across a minimal cloud relay (internet-facing SMTP gateway) and a user-owned home device (persistent mail storage). The roadmap builds from the encrypted transport foundation upward through each component layer: cloud relay, home mail server, DNS authentication, offline queuing, webmail/groupware, build system, client device setup, and operational monitoring. Each phase delivers a coherent, testable capability that unlocks the next.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Transport Layer** - Encrypted tunnel between cloud relay and home device with NAT traversal (completed 2026-02-09)
- [x] **Phase 2: Cloud Relay** - Minimal internet-facing SMTP gateway that receives and forwards mail (completed 2026-02-09)
- [x] **Phase 3: Home Mail Server** - Full-featured mail server with IMAP access on user-owned hardware (completed 2026-02-09)
- [x] **Phase 4: DNS & Email Authentication** - Automated SPF/DKIM/DMARC setup and DNS provider integration (completed 2026-02-14)
- [ ] **Phase 5: Queue & Offline Handling** - Encrypted message queuing when home device is unreachable
- [ ] **Phase 6: Webmail & Groupware** - Web-based email access with calendar and contacts sync
- [ ] **Phase 7: Build System & Deployment** - GitHub Actions pipeline for user-customized multi-arch Docker images
- [ ] **Phase 8: Device Profiles & Client Setup** - Auto-generated device configuration for seamless mail client onboarding
- [ ] **Phase 9: Monitoring & Observability** - Mail health visibility, delivery tracking, and certificate lifecycle management

## Phase Details

### Phase 1: Transport Layer
**Goal**: A secure, resilient, NAT-traversing encrypted connection exists between a cloud VPS and a home device, surviving internet interruptions without user intervention
**Depends on**: Nothing (first phase)
**Requirements**: TRNS-01, TRNS-02, TRNS-03, TRNS-04, CERT-02
**Success Criteria** (what must be TRUE):
  1. Cloud relay and home device communicate over an encrypted WireGuard tunnel without any port forwarding configured on the home network
  2. User can select mTLS as an alternative transport mechanism and traffic flows encrypted between cloud and home
  3. After simulating a home internet outage (unplug for 60 seconds), the tunnel automatically re-establishes and data flows resume without manual intervention
  4. Internal transport certificates are issued by a private CA (step-ca) and are distinct from public-facing TLS certificates
**Plans:** 3 plans

Plans:
- [x] 01-01-PLAN.md -- WireGuard tunnel foundation: Go module, config generation, deployment scripts, systemd auto-restart
- [x] 01-02-PLAN.md -- mTLS alternative transport and internal CA: step-ca PKI, mTLS server/client, cert renewal automation
- [x] 01-03-PLAN.md -- Auto-reconnection hardening: WireGuard health monitoring, outage simulation, VPS provider guide

### Phase 2: Cloud Relay
**Goal**: An internet-facing SMTP gateway receives inbound mail and sends outbound mail with TLS encryption, forwarding everything to the home device without storing messages persistently
**Depends on**: Phase 1
**Requirements**: RELAY-01, RELAY-02, RELAY-03, RELAY-04, RELAY-05, RELAY-06, CERT-01, UX-02
**Success Criteria** (what must be TRUE):
  1. External mail servers can deliver email to the cloud relay on port 25 with STARTTLS, and the message arrives on the home device within seconds
  2. Cloud relay sends outbound SMTP from the home device to destination mail servers via direct MTA delivery (no smart host)
  3. Cloud relay container image is under 50MB and runs on a $5/month VPS with less than 256MB RAM usage
  4. When a remote mail server connects without TLS support, the user receives a notification, and if strict mode is enabled, the connection is refused
  5. No mail content persists on the cloud relay filesystem after successful forwarding to the home device
**Plans:** 3 plans

Plans:
- [x] 02-01-PLAN.md -- Postfix relay-only container with Go SMTP relay daemon bridging to WireGuard/mTLS transport
- [x] 02-02-PLAN.md -- TLS enforcement, strict mode, notification system, and Let's Encrypt certbot automation
- [x] 02-03-PLAN.md -- Ephemeral storage verification, container optimization, and comprehensive test suite

### Phase 3: Home Mail Server
**Goal**: Users access their email through standard IMAP clients and send mail via SMTP submission, with all messages stored on their own hardware with spam filtering, multi-user support, and multi-domain capability
**Depends on**: Phase 1, Phase 2
**Requirements**: RELAY-07, RELAY-08, MAIL-01, MAIL-02, MAIL-03, MAIL-04, MAIL-05, MAIL-06
**Success Criteria** (what must be TRUE):
  1. User can connect Apple Mail, Thunderbird, or K-9 Mail to the home device via IMAP (port 993) and see received messages
  2. User can send email from a mail client via SMTP submission (port 587) and the message routes through the cloud relay to its destination
  3. Multiple users each have separate mailboxes with independent credentials on the same home device
  4. Mail aliases and catch-all addresses work (email to alias@domain delivers to the configured real mailbox)
  5. Spam is filtered by Rspamd before delivery, with greylisting reducing unsolicited messages
**Plans:** 3 plans

Plans:
- [x] 03-01-PLAN.md -- Mail server foundation: Stalwart, Maddy, and Postfix+Dovecot configs with Docker compose profiles for IMAP/SMTP submission
- [x] 03-02-PLAN.md -- Multi-user mailboxes, multi-domain support, aliases, and catch-all configuration for all three options
- [x] 03-03-PLAN.md -- Rspamd spam filtering with greylisting, milter integration, and phase integration test suite

### Phase 4: DNS & Email Authentication
**Goal**: Email sent from DarkPipe passes SPF, DKIM, and DMARC authentication at all major providers (Gmail, Outlook, Yahoo), with DNS records automated or clearly documented for manual setup
**Depends on**: Phase 2, Phase 3
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07
**Success Criteria** (what must be TRUE):
  1. Running the DNS validation checker reports all-green for SPF, DKIM, DMARC, MX, and PTR records
  2. An email sent to a Gmail account shows "Passed" for SPF, DKIM, and DMARC when viewing the original message headers
  3. For supported DNS providers (Cloudflare, Route53), records are created automatically via API with a single command
  4. For unsupported DNS providers, the tool generates copy-paste DNS record templates with clear instructions
  5. Reverse DNS (PTR) verification confirms the cloud relay IP resolves to the mail server FQDN and vice versa
**Plans:** 3 plans

Plans:
- [x] 04-01-PLAN.md -- DKIM key generation (2048-bit RSA), SPF/DKIM/DMARC/MX record generation, DKIM signing via emersion/go-msgauth, DNS-RECORDS.md manual guide output
- [x] 04-02-PLAN.md -- DNSProvider interface, Cloudflare (cloudflare-go v6) and Route53 (aws-sdk-go-v2) API integration, NS-based auto-detection, dry-run safety, propagation polling
- [x] 04-03-PLAN.md -- DNS validation checker (miekg/dns), PTR verification, Authentication-Results parser, unified `darkpipe dns-setup` CLI with --validate-only, --json, --rotate-dkim modes

### Phase 5: Queue & Offline Handling
**Goal**: Users choose how mail is handled when their home device is offline -- queue it encrypted on the cloud relay, overflow to S3-compatible storage, or bounce it immediately
**Depends on**: Phase 2
**Requirements**: QUEUE-01, QUEUE-02, QUEUE-03
**Success Criteria** (what must be TRUE):
  1. With queuing enabled and home device offline, inbound mail queues encrypted on the cloud relay and delivers automatically when the home device reconnects
  2. When the cloud relay queue exceeds its threshold, overflow messages store encrypted in Storj/S3-compatible storage and deliver when home device is available
  3. With queuing disabled, the cloud relay returns a 4xx temporary failure to sending servers when the home device is unreachable, causing the sender's server to retry later (or bounce after its own timeout)
**Plans:** 2 plans

Plans:
- [ ] 05-01-PLAN.md -- Encrypted RAM queue with age encryption, QueuedForwarder wrapper, background processor, configurable queue-or-bounce behavior
- [ ] 05-02-PLAN.md -- S3-compatible overflow storage via minio-go (Storj/AWS S3/MinIO), queue integration, and phase integration test suite

### Phase 6: Webmail & Groupware
**Goal**: Non-technical household members access email through a web browser and sync calendars/contacts with their phones and computers
**Depends on**: Phase 3
**Requirements**: WEB-01, WEB-02, CAL-01, CAL-02
**Success Criteria** (what must be TRUE):
  1. User can read, compose, and send email through a web browser without configuring any mail client
  2. Webmail interface is usable on a mobile phone screen without horizontal scrolling or broken layouts
  3. User can add a CalDAV account to iOS/macOS Calendar or Android calendar and sync events bidirectionally
  4. User can add a CardDAV account to their phone's contacts app and sync contacts bidirectionally
**Plans**: TBD

Plans:
- [ ] 06-01: Webmail integration (Roundcube/SnappyMail selection and setup)
- [ ] 06-02: CalDAV and CardDAV server deployment

### Phase 7: Build System & Deployment
**Goal**: Users produce custom Docker images tailored to their chosen stack components via GitHub Actions, with pre-built images available as an alternative, running on all target platforms
**Depends on**: Phase 2, Phase 3, Phase 6
**Requirements**: BUILD-01, BUILD-02, BUILD-03, UX-01, UX-03
**Success Criteria** (what must be TRUE):
  1. User forks the repository, selects mail server + webmail + calendar options in the GitHub Actions workflow, and receives working multi-arch Docker images published to GHCR
  2. Pre-built full-featured Docker images are available for users who want to skip customization
  3. Images build and run correctly on Raspberry Pi 4 (arm64), x64 Docker, TrueNAS Scale, and Unraid
  4. A non-technical user can deploy using simple defaults, while a power user can override every configuration option
  5. Multi-architecture images (arm64 + amd64) are produced from a single workflow run
**Plans**: TBD

Plans:
- [ ] 07-01: GitHub Actions build pipeline with component selection
- [ ] 07-02: Multi-architecture builds and pre-built image publishing
- [ ] 07-03: Platform validation (RPi4, TrueNAS Scale, Unraid) and tiered UX

### Phase 8: Device Profiles & Client Setup
**Goal**: Users onboard new devices (phones, tablets, desktops) to their DarkPipe mail server in under 2 minutes without manually entering server addresses, ports, or security settings
**Depends on**: Phase 3, Phase 6
**Requirements**: PROF-01, PROF-02, PROF-03, PROF-04, PROF-05
**Success Criteria** (what must be TRUE):
  1. iOS/macOS user installs a .mobileconfig profile and their device is immediately configured for email, calendar, and contacts
  2. Android user uses autoconfig and their email app auto-discovers all server settings
  3. User scans a QR code on their phone and the mail account is configured without typing any server details
  4. Thunderbird and Outlook auto-discover server settings when the user enters only their email address and password
  5. Users authenticate with app-generated passwords and never create or manage mail passwords directly
**Plans**: TBD

Plans:
- [ ] 08-01: Apple .mobileconfig and Android autoconfig generation
- [ ] 08-02: QR code generation and desktop autodiscovery (Thunderbird, Outlook)
- [ ] 08-03: App-generated password system

### Phase 9: Monitoring & Observability
**Goal**: Users have clear visibility into whether their email system is healthy -- mail is flowing, queues are clear, certificates are valid, and the cloud relay container is running properly
**Depends on**: Phase 2, Phase 3, Phase 5
**Requirements**: MON-01, MON-02, MON-03, CERT-03, CERT-04
**Success Criteria** (what must be TRUE):
  1. User can see mail queue depth and stuck/deferred message count at a glance
  2. User can check delivery status of recent outbound messages (delivered, deferred, bounced)
  3. Cloud relay container exposes health check endpoints that return pass/fail status
  4. Certificate rotation is configurable (30/60/90 day intervals) and rotations happen automatically without service interruption
  5. User receives an alert at least 14 days before any certificate expires, and again at 7 days if not renewed
**Plans**: TBD

Plans:
- [ ] 09-01: Mail queue and delivery status monitoring
- [ ] 09-02: Container health checks and certificate lifecycle management

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8 -> 9

Note: Phases 5 and 6 can execute in parallel after their dependencies are met. Phase 8 can begin once Phase 3 and Phase 6 complete, independent of Phases 4, 5, and 7.

| Phase | Plans Complete | Status | Completed |
|-------|---------------|--------|-----------|
| 1. Transport Layer | 3/3 | Complete | 2026-02-09 |
| 2. Cloud Relay | 3/3 | Complete | 2026-02-09 |
| 3. Home Mail Server | 3/3 | Complete | 2026-02-09 |
| 4. DNS & Email Authentication | 3/3 | Complete | 2026-02-14 |
| 5. Queue & Offline Handling | 0/2 | Not started | - |
| 6. Webmail & Groupware | 0/2 | Not started | - |
| 7. Build System & Deployment | 0/3 | Not started | - |
| 8. Device Profiles & Client Setup | 0/3 | Not started | - |
| 9. Monitoring & Observability | 0/2 | Not started | - |

---
*Roadmap created: 2026-02-08*
*Depth: Comprehensive (9 phases, 50 requirements mapped)*
