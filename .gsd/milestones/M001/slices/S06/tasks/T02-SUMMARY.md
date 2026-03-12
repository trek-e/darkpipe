---
id: T02
parent: S06
milestone: M001
provides:
  - Radicale CalDAV/CardDAV server with htpasswd authentication and file-based storage
  - Shared family calendar and address book accessible by all household members
  - Well-known URL redirects for iOS/macOS/Android auto-discovery
  - Setup script for default calendar/address book creation per user
  - Sync script for mail server user synchronization to Radicale
  - Phase 6 integration test suite covering all success criteria
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 3min 44sec
verification_result: passed
completed_at: 2026-02-14
blocker_discovered: false
---
# T02: 06-webmail-groupware 02

**# Phase 6 Plan 2: CalDAV/CardDAV Integration Summary**

## What Happened

# Phase 6 Plan 2: CalDAV/CardDAV Integration Summary

**Radicale CalDAV/CardDAV server with htpasswd authentication, shared family collections, and well-known auto-discovery for iOS/macOS/Android devices**

## Performance

- **Duration:** 3 min 44 sec
- **Started:** 2026-02-14T07:01:01Z
- **Completed:** 2026-02-14T07:04:45Z
- **Tasks:** 3
- **Files created:** 6
- **Files modified:** 2

## Accomplishments
- Radicale CalDAV/CardDAV server deployed as Docker compose profile for Maddy and Postfix+Dovecot setups
- Stalwart built-in CalDAV/CardDAV documented (no additional container needed)
- Same credentials work for mail, CalDAV, and CardDAV (htpasswd synchronized from mail server)
- Default calendar and address book auto-created per user via setup script
- Shared family calendar and address book accessible by all household members via rights file
- Well-known URL redirects enable iOS/macOS/Android auto-discovery without manual URL entry
- CalDAV/CardDAV remotely accessible through cloud relay WireGuard tunnel
- Phase 6 integration test suite validates all success criteria

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Radicale CalDAV/CardDAV server to home device** - `18e1cf2` (feat)
   - Radicale service with profile "radicale" (for Maddy/Postfix+Dovecot, not Stalwart)
   - Htpasswd auth with bcrypt encryption
   - Rights file allows owner access + shared family calendar/contacts
   - Setup script creates default calendar and address book per user
   - Sync script synchronizes mail server users to Radicale htpasswd

2. **Task 2: Configure Caddy well-known URLs and CalDAV/CardDAV routing** - `55e5e58` (feat)
   - Well-known CalDAV/CardDAV redirects (RFC 6764) for iOS/macOS auto-discovery
   - Radicale reverse proxy on /radicale/* path with header forwarding
   - Stalwart alternative documented (built-in CalDAV/CardDAV on port 8080)
   - CalDAV/CardDAV remotely accessible through cloud relay tunnel

3. **Task 3: Phase 6 integration test suite** - `bc90494` (test)
   - Tests webmail access (WEB-01)
   - Tests mobile responsiveness (WEB-02)
   - Tests CalDAV auto-discovery (CAL-01)
   - Tests CardDAV auto-discovery (CAL-02)
   - Tests Docker compose profile combinations
   - Color-coded output (green PASS, red FAIL)

## Files Created/Modified

### Created
- `home-device/caldav-carddav/radicale/config/config` - Radicale server config with htpasswd auth, file-based storage, rights file
- `home-device/caldav-carddav/radicale/rights` - Calendar/contact sharing permissions (owner access + shared family collections)
- `home-device/caldav-carddav/radicale/users` - Htpasswd users file (bcrypt encryption, managed by sync-users.sh)
- `home-device/caldav-carddav/setup-collections.sh` - Script to create default calendar + address book per user, plus shared family collections
- `home-device/caldav-carddav/sync-users.sh` - Script to sync mail server users to Radicale htpasswd
- `home-device/tests/test-webmail-groupware.sh` - Phase 6 integration test suite

### Modified
- `home-device/docker-compose.yml` - Added Radicale service with profile "radicale" and radicale-data volume, updated usage comments for CalDAV/CardDAV
- `cloud-relay/caddy/Caddyfile` - Added well-known CalDAV/CardDAV redirects and /radicale/* reverse proxy with Stalwart alternatives documented

## Decisions Made

1. **Radicale for Maddy/Postfix+Dovecot** - Stalwart has built-in CalDAV/CardDAV; standalone Radicale for other mail servers
2. **Manual user sync (interim)** - sync-users.sh provides manual sync from mail server to Radicale htpasswd; automation deferred to Phase 8
3. **File-based storage** - Radicale multifilesystem storage for simplicity, easy backup, and manual inspection
4. **Rights file for sharing** - Admin-configured sharing via rights file (no web UI); simpler than database-backed ACLs
5. **Well-known URLs on webmail domain** - CalDAV/CardDAV auto-discovery on mail.example.com (same domain as webmail)
6. **Shared family collections** - Default shared calendar and address book accessible by all household members

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - implementation proceeded smoothly following the plan specification.

## User Setup Required

**Radicale user synchronization (manual for Phase 6):**

When adding a new mail user, sync to Radicale:

```bash
# Sync all users from mail server
cd home-device
./caldav-carddav/sync-users.sh maddy  # or postfix-dovecot

# Or sync a single user
./caldav-carddav/sync-users.sh maddy --user alice@example.com --password secret

# Create shared family collections (run once on initial setup)
./caldav-carddav/setup-collections.sh alice@example.com --shared
```

**iOS/macOS Calendar/Contacts setup:**
1. Settings → Calendar → Accounts → Add Account → Other
2. Add CalDAV Account
3. Server: mail.example.com
4. Username: user@domain
5. Password: (mail password)
6. Description: DarkPipe Calendar
7. Calendar app will auto-discover personal and shared calendars

**Android DAVx5 setup:**
1. Install DAVx5 from F-Droid or Play Store
2. Add account → Login with URL and credentials
3. Server: https://mail.example.com/radicale/
4. Username: user@domain
5. Password: (mail password)
6. DAVx5 discovers calendars and address books
7. Select which to sync (personal + family shared)

**Environment variables (cloud-relay/.env):**
- `WEBMAIL_DOMAINS` - Already configured in Plan 06-01 (e.g., mail.example.com)

**DNS records:**
- Already configured in Plan 06-01 (mail.example.com A record pointing to cloud relay IP)

**Docker compose profile combinations:**
```bash
# Stalwart (built-in CalDAV/CardDAV) + Roundcube
docker compose --profile stalwart --profile roundcube up -d

# Maddy + SnappyMail + Radicale
docker compose --profile maddy --profile snappymail --profile radicale up -d

# Postfix+Dovecot + Roundcube + Radicale
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

## Next Phase Readiness

**Phase 6 complete** - All webmail and groupware success criteria met:
- WEB-01: Webmail accessible remotely at mail.example.com (Plan 06-01)
- WEB-02: Mobile-responsive webmail (Roundcube Elastic, SnappyMail viewport) (Plan 06-01)
- CAL-01: CalDAV auto-discovery with well-known URLs (Plan 06-02)
- CAL-02: CardDAV auto-discovery with well-known URLs (Plan 06-02)

**Ready for Phase 7 (Monitoring & Alerts):**
- Webmail and CalDAV/CardDAV services running
- Integration test suite provides baseline for monitoring
- Remote access pattern established for monitoring endpoints

**Blockers:** None

**Testing notes:**
- Test suite requires services running (webmail on port 8080, Radicale on port 5232)
- CalDAV/CardDAV authentication tests require users synced to Radicale
- Well-known URL redirects testable via curl or actual iOS/macOS/Android devices

## Self-Check: PASSED

All claimed files and commits verified:
- FOUND: home-device/caldav-carddav/radicale/config/config
- FOUND: home-device/caldav-carddav/radicale/rights
- FOUND: home-device/caldav-carddav/radicale/users
- FOUND: home-device/caldav-carddav/setup-collections.sh
- FOUND: home-device/caldav-carddav/sync-users.sh
- FOUND: home-device/tests/test-webmail-groupware.sh
- FOUND: 18e1cf2 (Task 1 commit)
- FOUND: 55e5e58 (Task 2 commit)
- FOUND: bc90494 (Task 3 commit)

---
*Phase: 06-webmail-groupware*
*Completed: 2026-02-14*
