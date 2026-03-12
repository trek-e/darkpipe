# T02: 03-home-mail-server 02

**Slice:** S03 — **Milestone:** M001

## Description

Configure multi-user mailboxes, multi-domain support, aliases, and catch-all addresses for all three mail server options.

Purpose: A home device serves a household -- multiple family members, each with their own mailbox and password, potentially across multiple custom domains. This plan adds the user/domain/alias layer on top of the working mail servers from Plan 01. Addresses MAIL-02 (multi-user), MAIL-03 (multi-domain), MAIL-04 (aliases and catch-all).

Output: Updated configurations for all three mail server options with multi-user examples, multi-domain support, alias mappings, and catch-all configuration. Setup scripts for initial user provisioning.

## Must-Haves

- [ ] "Multiple users each have separate mailboxes with independent credentials on the same home device"
- [ ] "Multiple domains are configured and mail for each domain delivers to the correct mailboxes"
- [ ] "Email sent to an alias delivers to the configured real mailbox"
- [ ] "Email sent to any undefined address at a catch-all domain delivers to the designated catch-all mailbox"
- [ ] "Users of different domains on the same server are fully isolated (user@domain1 cannot read user@domain2 mail)"

## Files

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
