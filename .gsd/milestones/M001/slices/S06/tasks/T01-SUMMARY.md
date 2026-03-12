---
id: T01
parent: S06
milestone: M001
provides:
  - Caddy reverse proxy on cloud relay forwarding HTTPS to home device webmail
  - Roundcube webmail with IMAP passthrough authentication and Elastic responsive skin
  - SnappyMail webmail as lightweight alternative
  - Docker compose profile pattern for webmail selection (roundcube, snappymail)
  - HTTPS termination with automatic Let's Encrypt certificates
  - Webmail accessible remotely at mail.example.com without VPN
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 2min 9sec
verification_result: passed
completed_at: 2026-02-14
blocker_discovered: false
---
# T01: 06-webmail-groupware 01

**# Phase 6 Plan 1: Webmail Access Summary**

## What Happened

# Phase 6 Plan 1: Webmail Access Summary

**Caddy reverse proxy on cloud relay with Roundcube/SnappyMail webmail on home device using IMAP passthrough authentication over WireGuard tunnel**

## Performance

- **Duration:** 2 min 9 sec
- **Started:** 2026-02-14T06:56:05Z
- **Completed:** 2026-02-14T06:58:14Z
- **Tasks:** 2
- **Files modified:** 2
- **Files created:** 3

## Accomplishments
- Caddy reverse proxy on cloud relay terminates HTTPS and forwards to home device webmail over WireGuard tunnel
- Roundcube and SnappyMail webmail available as Docker compose profiles (matches Phase 3 mail server profile pattern)
- IMAP passthrough authentication (webmail uses mail server credentials, no separate user database)
- Roundcube Elastic skin provides mobile-responsive layout (WEB-02 requirement)
- Remote webmail access at mail.example.com without VPN for non-technical household members

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Caddy reverse proxy to cloud relay** - `7a77145` (feat)
   - Caddy service with ports 80/443/443-udp for HTTP/HTTPS/HTTP3
   - Reverse proxy to 10.0.0.2:8080 with header forwarding
   - Auto-obtains Let's Encrypt certificates via ACME
   - Placeholder comments for CalDAV/CardDAV routes (Plan 06-02)

2. **Task 2: Add Roundcube and SnappyMail webmail to home device** - `10a0c86` (feat)
   - Roundcube (profile: roundcube) with Elastic responsive skin
   - SnappyMail (profile: snappymail) as lightweight alternative
   - IMAP passthrough authentication to mail server
   - extra_hosts mail-server:host-gateway for cross-profile compatibility

## Files Created/Modified

### Created
- `cloud-relay/caddy/Caddyfile` - Reverse proxy config forwarding HTTPS to home device webmail
- `home-device/webmail/roundcube/config.inc.php` - Roundcube config with IMAP passthrough, Elastic skin, 60-min session timeout
- `home-device/webmail/snappymail/domains/default.json` - SnappyMail domain config for IMAP/SMTP

### Modified
- `cloud-relay/docker-compose.yml` - Added Caddy service with ports 80/443/443-udp and named volumes
- `home-device/docker-compose.yml` - Added Roundcube and SnappyMail profiled services with extra_hosts

## Decisions Made

1. **Caddy over Nginx/Traefik** - Automatic HTTPS with Let's Encrypt, lightweight (128M limit), Go-based aligns with DarkPipe stack
2. **IMAP passthrough authentication** - Mail server is single source of truth, no sync issues, one set of credentials per user
3. **extra_hosts pattern** - `mail-server:host-gateway` allows webmail to connect to any mail server profile via host port binding
4. **Roundcube Elastic skin** - Mobile-responsive (WEB-02), modern UI, default in recent versions
5. **Session timeout 60 minutes with auto-refresh** - Avoids session expiry frustration (addresses webmail Pitfall 9 from research)
6. **Port 8080 for webmail** - Standard non-privileged port, Caddy forwards here from public 443

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - implementation proceeded smoothly following the plan specification.

## User Setup Required

**Environment variables (cloud-relay/.env):**
- `RELAY_DOMAIN` - Primary domain (e.g., example.com) - Caddy uses this for `mail.${RELAY_DOMAIN}`

**DNS records required:**
- `mail.example.com` A record pointing to cloud relay public IP
- Caddy will automatically obtain Let's Encrypt certificate via HTTP-01 challenge

**Usage examples:**
```bash
# Stalwart + Roundcube
cd home-device && docker compose --profile stalwart --profile roundcube up -d

# Maddy + SnappyMail
cd home-device && docker compose --profile maddy --profile snappymail up -d

# Postfix+Dovecot + Roundcube
cd home-device && docker compose --profile postfix-dovecot --profile roundcube up -d
```

**Access webmail:**
1. Visit https://mail.example.com
2. Log in with email address (user@domain) and mail password
3. Webmail connects to mail server via IMAP passthrough

## Next Phase Readiness

**Ready for Plan 06-02 (CalDAV/CardDAV):**
- Caddy reverse proxy infrastructure in place
- Placeholder comments for CalDAV/CardDAV routes already in Caddyfile
- Docker compose pattern established for adding groupware services
- Remote access tunnel working (same pattern for CalDAV/CardDAV)

**Blockers:** None

**Testing notes:**
- Webmail requires valid DNS record pointing to cloud relay IP
- Let's Encrypt certificate requires ports 80/443 accessible from internet
- User must have created mail account via Phase 3 user management

## Self-Check: PASSED

All claimed files and commits verified:
- FOUND: cloud-relay/caddy/Caddyfile
- FOUND: home-device/webmail/roundcube/config.inc.php
- FOUND: home-device/webmail/snappymail/domains/default.json
- FOUND: 7a77145 (Task 1 commit)
- FOUND: 10a0c86 (Task 2 commit)

---
*Phase: 06-webmail-groupware*
*Completed: 2026-02-14*
