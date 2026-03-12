# T04: 10-mail-migration 04

**Slice:** S10 — **Milestone:** M001

## Description

Build the CLI wizard command, progress display, and phase test suite that ties all migration components together into the user-facing `darkpipe-setup migrate` command.

Purpose: This is the user-facing integration layer. The `darkpipe-setup migrate --from <provider>` command guides users through authentication, shows a dry-run preview, and executes migration with rich progress feedback. The phase test suite validates all Phase 10 success criteria.

Output: Cobra migrate command, interactive wizard, progress display, and integration test suite.

## Must-Haves

- [ ] "User can run darkpipe-setup migrate --from gmail and be guided through OAuth2 device flow, preview, and migration"
- [ ] "User can run darkpipe-setup migrate --from mailcow and see MailCow users/aliases/mailboxes exported via API"
- [ ] "User can run darkpipe-setup migrate --from generic to migrate from any IMAP server"
- [ ] "Dry-run by default shows folder count, message count, calendar/contact count before migrating"
- [ ] "Per-folder progress bars display during migration with current/total counts"
- [ ] "Migration summary at end shows migrated count, skipped count, and error details"
- [ ] "Phase test suite validates all Phase 10 success criteria"

## Files

- `deploy/setup/cmd/darkpipe-setup/migrate_cmd.go`
- `deploy/setup/cmd/darkpipe-setup/main.go`
- `deploy/setup/pkg/wizard/migrate_wizard.go`
- `deploy/setup/pkg/wizard/progress.go`
- `deploy/setup/pkg/wizard/oauth_flow.go`
- `tests/test-mail-migration.sh`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
