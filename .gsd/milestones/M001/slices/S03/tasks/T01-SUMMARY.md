---
id: T01
parent: S03
milestone: M001
provides:
  - Docker compose profiles for three mail server options (Stalwart, Maddy, Postfix+Dovecot)
  - SMTP inbound on port 25 (accepts mail from cloud relay via WireGuard)
  - IMAP on port 993 (serves mail to clients)
  - SMTP submission on port 587 (accepts outbound mail from clients)
  - Virtual mailbox support (no system users required)
  - Self-signed TLS certificates for IMAP/submission
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 5min
verification_result: passed
completed_at: 2026-02-09
blocker_discovered: false
---
# T01: 03-home-mail-server 01

**# Phase 03 Plan 01: Home Mail Server Foundation Summary**

## What Happened

# Phase 03 Plan 01: Home Mail Server Foundation Summary

**Three selectable mail server options (Stalwart, Maddy, Postfix+Dovecot) with SMTP inbound (25), IMAP (993), and submission (587) over self-signed TLS within WireGuard tunnel**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-09T13:38:15Z
- **Completed:** 2026-02-09T13:43:15Z
- **Tasks:** 2
- **Files created:** 10

## Accomplishments
- Docker compose profiles enable runtime mail server selection (`docker compose --profile stalwart up -d`)
- Stalwart 0.15.4 all-in-one mail server with IMAP4rev2, JMAP, CalDAV/CardDAV support
- Maddy 0.8.2 minimal Go-based mail server (15MB binary, ~15MB RAM)
- Postfix+Dovecot traditional split MTA+IMAP configuration
- All three options accept SMTP on port 25 from cloud relay (10.8.0.0/24 WireGuard subnet)
- All three options provide IMAP on port 993 and SMTP submission on port 587 for mail clients
- Virtual mailbox domains (no system users, all mail owned by vmail UID/GID 5000)
- LMDB format for Postfix maps (BerkleyDB deprecated in Alpine)
- Self-signed TLS certificates (appropriate for WireGuard-encrypted traffic)
- No open relay configuration (all servers restrict to local domains only)

## Task Commits

Each task was committed atomically:

1. **Task 1: Stalwart and Maddy mail server configurations with Docker compose profiles** - `9bcb5c6` (feat)
   - Docker compose with profiles for stalwart, maddy, postfix-dovecot
   - Stalwart config.toml with listeners for ports 25, 587, 993
   - Maddy maddy.conf with endpoints for ports 25, 587, 993
   - Environment variable support for MAIL_DOMAIN, MAIL_HOSTNAME, credentials
   - 512MB memory limit per container

2. **Task 2: Postfix+Dovecot mail server configuration** - `0c8f4fc` (feat)
   - Alpine 3.21 Dockerfile with postfix, dovecot, dovecot-lmtpd, dovecot-pigeonhole
   - Postfix main.cf with virtual mailbox domains and LMTP delivery to Dovecot
   - Postfix master.cf with SMTP inbound (25) and submission (587) with authentication
   - Dovecot dovecot.conf with IMAP (993), LMTP, and SASL auth for Postfix
   - Entrypoint script for cert generation, map initialization, service startup

## Files Created/Modified

### Created
- `home-device/docker-compose.yml` - Orchestration with profiles for mail server selection
- `home-device/stalwart/Dockerfile` - Stalwart official image with config
- `home-device/stalwart/config.toml` - Stalwart server configuration (listeners, storage, auth)
- `home-device/maddy/Dockerfile` - Maddy official image with config
- `home-device/maddy/maddy.conf` - Maddy server configuration (endpoints, storage, auth)
- `home-device/postfix-dovecot/Dockerfile` - Alpine-based Postfix+Dovecot build
- `home-device/postfix-dovecot/postfix/main.cf` - Postfix MTA configuration
- `home-device/postfix-dovecot/postfix/master.cf` - Postfix master process configuration
- `home-device/postfix-dovecot/dovecot/dovecot.conf` - Dovecot IMAP/LMTP configuration
- `home-device/postfix-dovecot/entrypoint.sh` - Service initialization and startup script

### Modified
None (all new files)

## Decisions Made

**1. Docker compose profiles for mail server selection**
- Enables runtime selection without rebuilding images
- Users choose with `--profile stalwart|maddy|postfix-dovecot` flag
- All three options provide identical port layout (25, 587, 993)

**2. Self-signed TLS certificates**
- IMAP and submission traffic is already encrypted via WireGuard tunnel
- Self-signed certs reduce operational complexity (no cert renewal)
- Production deployments with public DNS can replace with Let's Encrypt

**3. LMDB format for Postfix maps**
- BerkleyDB deprecated in Alpine 3.13+ (license changed to AGPL-3.0)
- LMDB is API-compatible replacement, now default in Alpine
- All map files use `lmdb:` prefix explicitly

**4. Virtual mailbox domains (no system users)**
- Postfix virtual_mailbox_domains with virtual_uid_maps = static:5000
- Dovecot static userdb with uid=5000, gid=5000
- All mail owned by vmail user, no system users required
- Simplifies multi-domain, multi-user configuration

**5. SMTP delivery from cloud relay to home mail server**
- Cloud relay uses SMTP to port 25 (not LMTP)
- Preserves flexibility for all three mail server options
- Matches standard MX relay pattern
- LMTP used only for internal Postfix-to-Dovecot delivery

**6. Stalwart uses official multi-arch image**
- No custom build needed (stalwartlabs/stalwart:0.15.4)
- Official image supports arm64 and amd64
- Simplifies deployment on Raspberry Pi or x86_64

## Deviations from Plan

None - plan executed exactly as written.

All configuration details matched plan specifications:
- Three mail server options with Docker compose profiles
- Identical port layout (25, 587, 993) for all options
- Self-signed TLS for IMAP/submission
- Virtual mailbox domains
- LMDB format for Postfix maps
- No open relay configuration
- Environment variable support for customization

## Issues Encountered

None. All tasks completed without blocking issues.

## Verification Results

**Docker compose validation:**
- Three profiles defined: stalwart, maddy, postfix-dovecot
- All three services expose ports 25, 587, 993
- Shared darkpipe network configured
- Named volumes for mail-data and mail-config

**Stalwart configuration:**
- Listeners configured for ports 25, 587, 993
- Internal directory with SQLite backend
- Relay restricted to local domains (in-list check)
- Self-signed TLS enabled
- Storage: SQLite for data, filesystem for blobs

**Maddy configuration:**
- Endpoints configured for ports 25, 587, 993
- SQLite backend for mailbox storage and accounts
- Local domain restriction (destination check)
- Self-signed TLS configured
- Source IP filtering (10.8.0.0/24 + localhost only)

**Postfix+Dovecot configuration:**
- Postfix main.cf: virtual_transport = lmtp:unix:private/dovecot-lmtp
- Postfix main.cf: smtpd_relay_restrictions with reject_unauth_destination
- Postfix main.cf: All map files use lmdb: prefix (no hash: or btree:)
- Dovecot dovecot.conf: protocols = imap lmtp
- Dovecot dovecot.conf: ssl = required
- Entrypoint.sh: bash syntax valid, generates cert, creates maps, starts services

**Security verification:**
- No open relay configuration in any mail server option
- All servers restrict relay to local domains only
- Submission ports require authentication
- SMTP inbound restricted to WireGuard subnet (10.8.0.0/24)

## Integration Points

**Cloud relay → Home mail server:**
- Cloud relay forwards mail via SMTP to home device WireGuard IP (10.8.0.2:25)
- Home mail server accepts from mynetworks = 10.8.0.0/24 (Postfix) or source IP filter (Maddy)
- Flow: Cloud relay SMTP client → WireGuard tunnel → Home mail server port 25

**Home mail server → Mail clients:**
- IMAP on port 993 (implicit TLS) for retrieving mail
- SMTP submission on port 587 (STARTTLS) for sending mail
- Authentication required for both IMAP and submission

**Postfix+Dovecot internal delivery:**
- Flow: Postfix SMTP (25) → Postfix virtual_transport → Dovecot LMTP (Unix socket) → Maildir
- Dovecot LMTP socket: `/var/spool/postfix/private/dovecot-lmtp`
- Dovecot SASL socket: `/var/spool/postfix/private/auth`

## Next Phase Readiness

**Ready for Phase 03 Plan 02 (User/Domain Management):**
- Mail server foundation complete
- Virtual mailbox infrastructure in place
- User database files created (empty, ready for population)
- Map files initialized (vmailbox, virtual for aliases)

**Ready for Phase 03 Plan 03 (Spam Filtering):**
- Mail server ports (25, 587, 993) exposed
- Rspamd milter integration points documented in config comments
- Postfix supports milter protocol (smtpd_milters setting ready)
- Stalwart supports milter protocol (session.data.milter config ready)

**Blockers/Concerns:**
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026) - schema may change during upgrade
- Maddy is beta status - developers recommend caution for production use
- Self-signed TLS requires mail clients to accept certificate (manual step for users)

**Deployment Prerequisites:**
- VPS and home device must have WireGuard tunnel established (Phase 01 requirement)
- Cloud relay must be configured with home device WireGuard IP (Phase 02 requirement)
- Users must choose mail server profile before deployment: `--profile stalwart|maddy|postfix-dovecot`
- Environment variables must be set: MAIL_DOMAIN, MAIL_HOSTNAME, ADMIN_EMAIL, ADMIN_PASSWORD

---
*Phase: 03-home-mail-server*
*Plan: 01*
*Completed: 2026-02-09*

## Self-Check: PASSED

All files created:
- home-device/docker-compose.yml
- home-device/stalwart/Dockerfile
- home-device/stalwart/config.toml
- home-device/maddy/Dockerfile
- home-device/maddy/maddy.conf
- home-device/postfix-dovecot/Dockerfile
- home-device/postfix-dovecot/postfix/main.cf
- home-device/postfix-dovecot/postfix/master.cf
- home-device/postfix-dovecot/dovecot/dovecot.conf
- home-device/postfix-dovecot/entrypoint.sh

All commits verified:
- 9bcb5c6: feat(03-01): add Stalwart and Maddy mail server configs with Docker compose profiles
- 0c8f4fc: feat(03-01): add Postfix+Dovecot mail server configuration
