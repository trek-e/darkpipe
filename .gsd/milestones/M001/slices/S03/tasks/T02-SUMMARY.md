---
id: T02
parent: S03
milestone: M001
provides:
  - Multi-user mailbox configuration for all three mail server options
  - Multi-domain support (example.com, example.org) for all three options
  - Alias mappings (admin@example.com -> alice@example.com)
  - Catch-all configuration (@example.org -> bob@example.org)
  - User provisioning scripts (setup-users.sh) for Stalwart and Maddy
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
# T02: 03-home-mail-server 02

**# Phase 03 Plan 02: User and Domain Management Summary**

## What Happened

# Phase 03 Plan 02: User and Domain Management Summary

**Multi-user mailboxes, multi-domain support, aliases, and catch-all addresses configured for Stalwart, Maddy, and Postfix+Dovecot**

## Performance

- **Duration:** 5 min 6 sec
- **Started:** 2026-02-09T13:46:46Z
- **Completed:** 2026-02-09T13:51:52Z
- **Tasks:** 2
- **Files created:** 6
- **Files modified:** 6

## Accomplishments

- Multi-user mailbox configuration for all three mail server options
- Stalwart: REST API-based user/domain/alias management with setup-users.sh automation
- Maddy: CLI-based user management (maddy creds, maddy imap-acct) with setup-users.sh automation
- Postfix+Dovecot: File-based user management with vmailbox and users files
- Multi-domain support: example.com and example.org configured in all three options
- User isolation: alice@example.com and bob@example.org have separate maildirs, cannot access each other's mail
- Alias resolution: admin@example.com and support@example.com deliver to alice@example.com
- Catch-all configuration: @example.org delivers to bob@example.org
- Spam warnings documented: catch-all requires Rspamd (Plan 03) to avoid spam overload
- Critical anti-pattern avoided: Postfix domains only in virtual_mailbox_domains, NOT virtual_alias_domains

## Task Commits

Each task was committed atomically:

1. **Task 1: Multi-user and multi-domain configuration for all three mail servers** - `834ded1` (feat)
   - Stalwart: Updated config.toml with domain/user management documentation, REST API examples
   - Maddy: Updated maddy.conf with multiple domains (example.com, example.org), added user creation docs
   - Postfix+Dovecot: Updated main.cf for multiple virtual_mailbox_domains, created vmailbox and users files
   - Created setup-users.sh scripts for Stalwart (REST API) and Maddy (CLI)
   - Docker compose: Added setup script volumes and config file volumes

2. **Task 2: Aliases and catch-all configuration for all three mail servers** - `be15066` (feat)
   - Stalwart: Added catch-all documentation to config.toml, updated setup-users.sh with alias/catch-all API calls
   - Maddy: Created aliases file, updated maddy.conf with alias resolution pipeline, updated Dockerfile
   - Postfix: Created virtual alias file with alias mappings and catch-all, updated entrypoint.sh
   - Spam warnings documented in all configs
   - Anti-pattern verification: no virtual_alias_domains/virtual_mailbox_domains overlap

## Files Created/Modified

### Created
- `home-device/stalwart/setup-users.sh` - REST API-based user provisioning script
- `home-device/maddy/setup-users.sh` - CLI-based user provisioning script
- `home-device/maddy/aliases` - Alias mapping file (admin@, support@, @example.org catch-all)
- `home-device/postfix-dovecot/postfix/vmailbox` - Virtual mailbox mappings (alice@example.com, bob@example.org)
- `home-device/postfix-dovecot/postfix/virtual` - Virtual alias mappings (admin@, support@, @example.org catch-all)
- `home-device/postfix-dovecot/dovecot/users` - User credentials file (passwd-file format)

### Modified
- `home-device/stalwart/config.toml` - Multi-domain/user documentation, catch-all config
- `home-device/maddy/maddy.conf` - Multiple domains, alias table reference, alias resolution pipeline
- `home-device/maddy/Dockerfile` - Copy aliases file
- `home-device/postfix-dovecot/postfix/main.cf` - Multiple virtual_mailbox_domains
- `home-device/postfix-dovecot/entrypoint.sh` - Verify config files, create maildirs for all users
- `home-device/docker-compose.yml` - Added setup script and config file volumes

## Decisions Made

**1. Email address as username for all three options**
- Uniform interface: user@domain is the username for authentication
- Simplifies Phase 7 (Build System) and Phase 8 (Device Profiles)
- All three mail servers use same user model

**2. Multi-domain support varies by implementation**
- Stalwart: Domains managed via REST API or Web UI
- Maddy: Domains listed in $(local_domains) variable (space-separated)
- Postfix: Domains listed in virtual_mailbox_domains (comma-separated)

**3. Alias resolution before mailbox delivery**
- Stalwart: Aliases are account type="alias" with memberOf field (REST API)
- Maddy: Aliases in table.file, resolved via replace_rcpt modifier
- Postfix: Aliases in virtual_alias_maps, resolved before virtual_mailbox_maps

**4. Catch-all requires spam filtering**
- All three configs include warnings about catch-all spam load
- Plan 03 (Rspamd) should be deployed before catch-all is active in production
- Catch-all documented but users must enable consciously

**5. Postfix anti-pattern explicitly avoided**
- Domains only in virtual_mailbox_domains, NOT virtual_alias_domains
- Catch-all (@example.org) works within virtual_mailbox_domains
- Aliases (virtual file) and mailboxes (vmailbox file) are separate
- This matches research Pitfall 2

**6. User isolation by maildir paths**
- Postfix+Dovecot: /var/mail/vhosts/example.com/alice vs /var/mail/vhosts/example.org/bob
- Maddy: Account IDs include domain (alice@example.com vs bob@example.org)
- Stalwart: Internal directory stores users with full email addresses
- Different domains = different accounts = automatic isolation

## Deviations from Plan

None - plan executed exactly as written.

All configuration details matched plan specifications:
- Multi-user and multi-domain configuration for all three options
- Alias mappings (admin@, support@ -> alice@)
- Catch-all configuration (@example.org -> bob@)
- Setup scripts created for Stalwart and Maddy
- Anti-patterns from research explicitly avoided
- Spam warnings documented

## Issues Encountered

None. All tasks completed without blocking issues.

## Verification Results

**Multi-user configuration:**
- Stalwart: setup-users.sh creates alice@example.com and bob@example.org via REST API
- Maddy: setup-users.sh creates alice@example.com and bob@example.org via CLI (maddy creds, maddy imap-acct)
- Postfix+Dovecot: vmailbox has 2 users across 2 domains, users file has matching credentials

**Multi-domain support:**
- Stalwart: Domains managed via REST API (example.com, example.org)
- Maddy: $(local_domains) = example.com example.org
- Postfix: virtual_mailbox_domains = example.com, example.org

**Alias resolution:**
- Stalwart: setup-users.sh creates admin@example.com and support@example.com aliases via REST API
- Maddy: aliases file has admin@example.com: alice@example.com and support@example.com: alice@example.com
- Postfix: virtual file has admin@example.com alice@example.com and support@example.com alice@example.com

**Catch-all configuration:**
- Stalwart: setup-users.sh sets domain catch-all via REST API (PUT /api/v1/domain/example.org)
- Maddy: aliases file has @example.org: bob@example.org
- Postfix: virtual file has @example.org bob@example.org

**Anti-pattern verification:**
- Postfix main.cf does NOT have virtual_alias_domains defined
- Domains only appear in virtual_mailbox_domains
- No overlap between virtual_alias_domains and virtual_mailbox_domains (critical requirement met)

**Setup scripts:**
- Stalwart setup-users.sh: executable, syntax valid, creates domains/users/aliases via REST API
- Maddy setup-users.sh: executable, syntax valid, creates users via CLI tools
- Postfix+Dovecot entrypoint.sh: syntax valid, verifies config files, creates maildirs

## Integration Points

**User authentication flow:**
- IMAP login: user@domain + password
- SMTP submission login: user@domain + password
- All three options use same authentication pattern

**Alias resolution flow:**
- Inbound mail to admin@example.com -> resolved to alice@example.com -> delivered to alice's mailbox
- Inbound mail to support@example.com -> resolved to alice@example.com -> delivered to alice's mailbox

**Catch-all flow:**
- Inbound mail to anything@example.org -> resolved to bob@example.org -> delivered to bob's mailbox
- Only works for undefined addresses (does not override defined mailboxes)

**Multi-domain isolation:**
- alice@example.com and bob@example.org are separate accounts
- Different maildirs (Postfix+Dovecot), different account IDs (Maddy, Stalwart)
- No cross-domain access possible

## Next Phase Readiness

**Ready for Phase 03 Plan 03 (Spam Filtering):**
- Multi-user mailboxes configured
- Catch-all documented (awaiting Rspamd before activation)
- All three mail servers accept inbound mail on port 25

**Ready for Phase 07 (Build System):**
- Uniform user model (user@domain) across all three mail server options
- Setup scripts demonstrate automated user provisioning
- File-based configuration ready for templating

**Ready for Phase 08 (Device Profiles):**
- Multi-domain support demonstrated
- User isolation verified
- Alias and catch-all mechanisms documented

**Blockers/Concerns:**
- Catch-all should NOT be enabled in production until Rspamd (Plan 03) is deployed
- Self-signed TLS certificates require mail clients to accept certificate (manual step)
- Stalwart 0.15.4 is pre-v1.0 (schema may change during upgrade)

**Deployment Prerequisites:**
- Users must run setup scripts after first deployment:
  - Stalwart: `docker compose exec stalwart /opt/stalwart-mail/setup-users.sh`
  - Maddy: `docker compose exec maddy /data/setup-users.sh`
  - Postfix+Dovecot: Users already created from vmailbox/users files
- Default passwords must be changed after deployment
- Catch-all should be disabled until Rspamd is deployed

---
*Phase: 03-home-mail-server*
*Plan: 02*
*Completed: 2026-02-09*

## Self-Check: PASSED

All files created:
- home-device/stalwart/setup-users.sh
- home-device/maddy/setup-users.sh
- home-device/maddy/aliases
- home-device/postfix-dovecot/postfix/vmailbox
- home-device/postfix-dovecot/postfix/virtual
- home-device/postfix-dovecot/dovecot/users

All commits verified:
- 834ded1: feat(03-02): add multi-user and multi-domain configuration for all mail servers
- be15066: feat(03-02): add alias and catch-all configuration for all mail servers
