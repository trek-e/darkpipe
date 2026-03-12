# T02: 06-webmail-groupware 02

**Slice:** S06 — **Milestone:** M001

## Description

Deploy CalDAV/CardDAV server (Radicale for Maddy/Postfix+Dovecot, Stalwart built-in for Stalwart profile) with well-known auto-discovery URLs, shared family calendar/address book, and phase integration test suite.

Purpose: Household members sync calendars and contacts bidirectionally with iOS/macOS/Android devices using the same credentials as their mail account. Shared family calendar and address book keep household information accessible to everyone.

Output: Radicale CalDAV/CardDAV container with profiled deployment, well-known URL auto-discovery via Caddy, user sync script, shared collections setup, and Phase 6 end-to-end test suite.

## Must-Haves

- [ ] "User can add CalDAV account on iOS/macOS/Android using mail.example.com and it auto-discovers calendars"
- [ ] "User can add CardDAV account on iOS/macOS/Android using mail.example.com and it auto-discovers address books"
- [ ] "CalDAV/CardDAV uses same credentials as mail account (user@domain + mail password)"
- [ ] "Default calendar and address book are auto-created for each user"
- [ ] "Shared family calendar and address book are accessible by all household members"
- [ ] "Calendar sharing works between household members (read-only or read-write via rights file)"
- [ ] "Well-known URLs redirect correctly for iOS/macOS auto-discovery"
- [ ] "CalDAV/CardDAV works remotely through cloud relay tunnel"

## Files

- `home-device/docker-compose.yml`
- `home-device/caldav-carddav/radicale/config/config`
- `home-device/caldav-carddav/radicale/rights`
- `home-device/caldav-carddav/radicale/users`
- `home-device/caldav-carddav/setup-collections.sh`
- `home-device/caldav-carddav/sync-users.sh`
- `cloud-relay/caddy/Caddyfile`
- `home-device/tests/test-webmail-groupware.sh`
