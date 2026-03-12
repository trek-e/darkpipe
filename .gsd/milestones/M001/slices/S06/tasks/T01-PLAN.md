# T01: 06-webmail-groupware 01

**Slice:** S06 — **Milestone:** M001

## Description

Deploy Caddy reverse proxy on cloud relay and user-selectable webmail (Roundcube or SnappyMail) on home device, enabling browser-based email access remotely via HTTPS tunnel.

Purpose: Non-technical household members can read, compose, and send email through a web browser at mail.example.com without configuring any mail client. Remote access works through the cloud relay tunnel without VPN.

Output: Caddy reverse proxy on cloud relay forwarding HTTPS to home device webmail, Roundcube and SnappyMail as Docker compose profile options on home device with IMAP passthrough authentication and mobile-responsive UI.

## Must-Haves

- [ ] "User can access webmail at mail.example.com via HTTPS and the page loads"
- [ ] "User can log in to webmail using their IMAP credentials (user@domain + password)"
- [ ] "User can read, compose, and send email through the webmail interface"
- [ ] "Webmail interface is mobile-responsive (Elastic skin for Roundcube, default for SnappyMail)"
- [ ] "User selects Roundcube or SnappyMail via Docker compose profile (matches Phase 3 pattern)"

## Files

- `cloud-relay/docker-compose.yml`
- `cloud-relay/caddy/Caddyfile`
- `home-device/docker-compose.yml`
- `home-device/webmail/roundcube/config.inc.php`
- `home-device/webmail/snappymail/domains/default.json`
