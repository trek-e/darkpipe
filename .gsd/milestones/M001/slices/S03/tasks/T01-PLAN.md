# T01: 03-home-mail-server 01

**Slice:** S03 — **Milestone:** M001

## Description

Set up the home device mail server foundation with all three selectable options (Stalwart, Maddy, Postfix+Dovecot), each providing inbound SMTP (port 25), IMAP (port 993), and SMTP submission (port 587).

Purpose: Establishes the receiving end of the mail pipeline -- the cloud relay (Phase 2) forwards mail via SMTP to port 25 on the home device, and this plan ensures a working mail server is listening there, storing messages, and serving them to IMAP clients. Docker compose profiles enable runtime mail server selection per MAIL-01.

Output: Complete home-device/ directory with Dockerfiles, config files, and docker-compose.yml for all three mail server options.

## Must-Haves

- [ ] "Each mail server option (Stalwart, Maddy, Postfix+Dovecot) accepts inbound SMTP on port 25 from the cloud relay"
- [ ] "Each mail server option provides IMAP on port 993 (implicit TLS) for mail client access"
- [ ] "Each mail server option provides SMTP submission on port 587 (STARTTLS) for sending from mail clients"
- [ ] "Docker compose profiles allow selecting exactly one mail server at deploy time"
- [ ] "A single test user can connect an IMAP client and see received messages"

## Files

- `home-device/docker-compose.yml`
- `home-device/stalwart/Dockerfile`
- `home-device/stalwart/config.toml`
- `home-device/maddy/Dockerfile`
- `home-device/maddy/maddy.conf`
- `home-device/postfix-dovecot/Dockerfile`
- `home-device/postfix-dovecot/postfix/main.cf`
- `home-device/postfix-dovecot/postfix/master.cf`
- `home-device/postfix-dovecot/dovecot/dovecot.conf`
- `home-device/postfix-dovecot/entrypoint.sh`
