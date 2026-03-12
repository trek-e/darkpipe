# S06: Webmail Groupware

**Goal:** Deploy Caddy reverse proxy on cloud relay and user-selectable webmail (Roundcube or SnappyMail) on home device, enabling browser-based email access remotely via HTTPS tunnel.
**Demo:** Deploy Caddy reverse proxy on cloud relay and user-selectable webmail (Roundcube or SnappyMail) on home device, enabling browser-based email access remotely via HTTPS tunnel.

## Must-Haves


## Tasks

- [x] **T01: 06-webmail-groupware 01** `est:2min 9sec`
  - Deploy Caddy reverse proxy on cloud relay and user-selectable webmail (Roundcube or SnappyMail) on home device, enabling browser-based email access remotely via HTTPS tunnel.

Purpose: Non-technical household members can read, compose, and send email through a web browser at mail.example.com without configuring any mail client. Remote access works through the cloud relay tunnel without VPN.

Output: Caddy reverse proxy on cloud relay forwarding HTTPS to home device webmail, Roundcube and SnappyMail as Docker compose profile options on home device with IMAP passthrough authentication and mobile-responsive UI.
- [x] **T02: 06-webmail-groupware 02** `est:3min 44sec`
  - Deploy CalDAV/CardDAV server (Radicale for Maddy/Postfix+Dovecot, Stalwart built-in for Stalwart profile) with well-known auto-discovery URLs, shared family calendar/address book, and phase integration test suite.

Purpose: Household members sync calendars and contacts bidirectionally with iOS/macOS/Android devices using the same credentials as their mail account. Shared family calendar and address book keep household information accessible to everyone.

Output: Radicale CalDAV/CardDAV container with profiled deployment, well-known URL auto-discovery via Caddy, user sync script, shared collections setup, and Phase 6 end-to-end test suite.

## Files Likely Touched

- `cloud-relay/docker-compose.yml`
- `cloud-relay/caddy/Caddyfile`
- `home-device/docker-compose.yml`
- `home-device/webmail/roundcube/config.inc.php`
- `home-device/webmail/snappymail/domains/default.json`
- `home-device/docker-compose.yml`
- `home-device/caldav-carddav/radicale/config/config`
- `home-device/caldav-carddav/radicale/rights`
- `home-device/caldav-carddav/radicale/users`
- `home-device/caldav-carddav/setup-collections.sh`
- `home-device/caldav-carddav/sync-users.sh`
- `cloud-relay/caddy/Caddyfile`
- `home-device/tests/test-webmail-groupware.sh`
