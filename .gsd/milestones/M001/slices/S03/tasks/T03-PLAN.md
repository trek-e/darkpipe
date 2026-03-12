# T03: 03-home-mail-server 03

**Slice:** S03 — **Milestone:** M001

## Description

Deploy Rspamd spam filtering with greylisting backed by Redis, integrate with all three mail server options via milter protocol, and create integration test scripts for the complete home device mail stack.

Purpose: Spam filtering is essential before enabling catch-all addresses or exposing the mail server to the internet. Rspamd provides industry-standard spam scoring and greylisting reduces unsolicited messages by temporarily rejecting first-time senders (legitimate servers retry, spammers typically do not). The integration test scripts verify the entire mail pipeline works end-to-end and serve as the phase test suite per project memory rules.

Output: Rspamd + Redis containers, milter integration for all mail servers, greylisting configuration, and test scripts covering mail flow and spam filtering.

## Must-Haves

- [ ] "Rspamd scans all inbound mail and assigns spam scores before delivery to mailbox"
- [ ] "Greylisting temporarily rejects first-time sender/recipient/IP combinations, reducing unsolicited messages"
- [ ] "Rspamd integrates with all three mail server options via milter protocol (port 11332)"
- [ ] "Redis backend stores greylisting state and Rspamd statistics persistently"
- [ ] "Authenticated users (SMTP submission) bypass spam filtering and greylisting"
- [ ] "Integration test scripts verify end-to-end mail flow through the complete home device stack"

## Files

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
