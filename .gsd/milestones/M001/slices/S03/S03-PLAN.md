# S03: Home Mail Server

**Goal:** Set up the home device mail server foundation with all three selectable options (Stalwart, Maddy, Postfix+Dovecot), each providing inbound SMTP (port 25), IMAP (port 993), and SMTP submission (port 587).
**Demo:** Set up the home device mail server foundation with all three selectable options (Stalwart, Maddy, Postfix+Dovecot), each providing inbound SMTP (port 25), IMAP (port 993), and SMTP submission (port 587).

## Must-Haves


## Tasks

- [x] **T01: 03-home-mail-server 01** `est:5min`
  - Set up the home device mail server foundation with all three selectable options (Stalwart, Maddy, Postfix+Dovecot), each providing inbound SMTP (port 25), IMAP (port 993), and SMTP submission (port 587).

Purpose: Establishes the receiving end of the mail pipeline -- the cloud relay (Phase 2) forwards mail via SMTP to port 25 on the home device, and this plan ensures a working mail server is listening there, storing messages, and serving them to IMAP clients. Docker compose profiles enable runtime mail server selection per MAIL-01.

Output: Complete home-device/ directory with Dockerfiles, config files, and docker-compose.yml for all three mail server options.
- [x] **T02: 03-home-mail-server 02** `est:5min`
  - Configure multi-user mailboxes, multi-domain support, aliases, and catch-all addresses for all three mail server options.

Purpose: A home device serves a household -- multiple family members, each with their own mailbox and password, potentially across multiple custom domains. This plan adds the user/domain/alias layer on top of the working mail servers from Plan 01. Addresses MAIL-02 (multi-user), MAIL-03 (multi-domain), MAIL-04 (aliases and catch-all).

Output: Updated configurations for all three mail server options with multi-user examples, multi-domain support, alias mappings, and catch-all configuration. Setup scripts for initial user provisioning.
- [x] **T03: 03-home-mail-server 03** `est:4min 35s`
  - Deploy Rspamd spam filtering with greylisting backed by Redis, integrate with all three mail server options via milter protocol, and create integration test scripts for the complete home device mail stack.

Purpose: Spam filtering is essential before enabling catch-all addresses or exposing the mail server to the internet. Rspamd provides industry-standard spam scoring and greylisting reduces unsolicited messages by temporarily rejecting first-time senders (legitimate servers retry, spammers typically do not). The integration test scripts verify the entire mail pipeline works end-to-end and serve as the phase test suite per project memory rules.

Output: Rspamd + Redis containers, milter integration for all mail servers, greylisting configuration, and test scripts covering mail flow and spam filtering.

## Files Likely Touched

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
- `home-device/stalwart/config.toml`
- `home-device/maddy/maddy.conf`
- `home-device/maddy/aliases`
- `home-device/postfix-dovecot/postfix/main.cf`
- `home-device/postfix-dovecot/postfix/vmailbox`
- `home-device/postfix-dovecot/postfix/virtual`
- `home-device/postfix-dovecot/dovecot/dovecot.conf`
- `home-device/postfix-dovecot/dovecot/users`
- `home-device/postfix-dovecot/entrypoint.sh`
- `home-device/stalwart/setup-users.sh`
- `home-device/maddy/setup-users.sh`
- `home-device/docker-compose.yml`
- `home-device/spam-filter/rspamd/local.d/greylist.conf`
- `home-device/spam-filter/rspamd/local.d/worker-proxy.conf`
- `home-device/spam-filter/rspamd/local.d/actions.conf`
- `home-device/spam-filter/rspamd/local.d/logging.inc`
- `home-device/spam-filter/rspamd/override.d/worker-controller.inc`
- `home-device/spam-filter/redis/redis.conf`
- `home-device/stalwart/config.toml`
- `home-device/maddy/maddy.conf`
- `home-device/postfix-dovecot/postfix/main.cf`
- `home-device/docker-compose.yml`
- `home-device/tests/test-mail-flow.sh`
- `home-device/tests/test-spam-filter.sh`
