---
phase: 10-mail-migration
plan: 04
subsystem: mail-migration
tags: [cli, wizard, progress, testing, documentation]
dependencies:
  requires: ["10-01", "10-02", "10-03"]
  provides: ["darkpipe-setup migrate command", "migration wizard", "phase test suite"]
  affects: ["setup CLI", "user onboarding", "migration UX"]
tech-stack:
  added: ["survey/v2 prompts", "pterm progress bars", "cobra CLI framework"]
  patterns: ["CLI wizard pattern", "OAuth2 device flow UI", "progress callbacks"]
key-files:
  created:
    - deploy/setup/cmd/darkpipe-setup/migrate_cmd.go
    - deploy/setup/pkg/wizard/migrate_wizard.go
    - deploy/setup/pkg/wizard/progress.go
    - deploy/setup/pkg/wizard/oauth_flow.go
    - tests/test-mail-migration.sh
  modified:
    - deploy/setup/cmd/darkpipe-setup/main.go
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
decisions:
  - Dry-run by default (--apply required for execution) for safety
  - Progress bars with terminal fallback for non-TTY environments
  - OAuth2 client credentials via environment variables (GMAIL_CLIENT_ID, OUTLOOK_CLIENT_ID)
  - Provider-agnostic wizard using Provider.WizardPrompts() for extensibility
  - Destination defaults: localhost:993 (IMAP), localhost:5232 (CalDAV/CardDAV)
metrics:
  duration: 155
  completed: 2026-02-15T02:23:42Z
  tasks: 2
  files: 8
---

# Phase 10 Plan 04: CLI Wizard & Test Suite Summary

**One-liner:** Interactive `darkpipe-setup migrate` CLI with OAuth2 device flow, dry-run preview, per-folder progress bars, and comprehensive phase test suite validating all 4 success criteria.

## What Was Built

### CLI Migrate Command (migrate_cmd.go)
- Cobra subcommand `darkpipe-setup migrate` with 13 flags for all migration options
- Provider listing when --from not provided (table with 7 supported providers)
- Dry-run by default with --apply flag for safe execution
- Folder mapping, labels-as-folders, contacts merge mode, batch size configuration
- VCF/ICS file import support via --vcf-file and --ics-file flags

### Interactive Migration Wizard (migrate_wizard.go)
- Provider-specific authentication flows via Provider.WizardPrompts()
- OAuth2 device flow for Gmail/Outlook with RunOAuthDeviceFlow
- Text input prompts for generic providers, info displays for iCloud
- Source and destination connection management with cleanup
- **Dry-run phase** showing folder/message/calendar/contact counts with pterm tables
- Confirmation prompt before executing migration (--apply)
- Migration execution with progress callbacks wired to MigrationProgress
- Final summary table showing migrated/skipped/error counts
- State file location display for resume capability

### Progress Display (progress.go)
- MigrationProgress struct with per-folder and overall progress bars
- Terminal detection with graceful fallback to log-style output
- SetOverall, StartFolder, UpdateFolder, CompleteFolder methods
- Fallback mode prints "[3/12] Migrating INBOX: 142/1203 messages"
- Thread-safe progress tracking with sync.Mutex

### OAuth2 Device Flow UI (oauth_flow.go)
- RunOAuthDeviceFlow with pterm UI components
- Verification URL in highlighted box with centered title
- User code displayed in large pterm.BigText for easy reading
- Spinner while waiting for browser authorization
- Success message with checkmark on completion

### Phase Test Suite (test-mail-migration.sh)
- **SC1 tests:** Binary compilation, provider listing, flag recognition, IMAP/mapping/state unit tests
- **SC2 tests:** CalDAV/CardDAV/VCF/ICS import tests, contact merge tests
- **SC3 tests:** MailCow provider tests, provider registry tests
- **SC4 tests:** CLI flag validation, full test suite execution
- All 34 tests pass in 8 seconds (0 failures)
- Color-coded output with pass/fail counts and timing

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] GenericProvider field name mismatch**
- **Found during:** Task 1, migrate_wizard.go destination connection
- **Issue:** migrate_wizard.go used nonexistent fields IMAPUser, IMAPPass on GenericProvider
- **Fix:** Changed to Username, Password fields (actual GenericProvider API)
- **Files modified:** deploy/setup/pkg/wizard/migrate_wizard.go
- **Commit:** 7882fb5

**2. [Rule 1 - Bug] Redundant condition in destination endpoint defaults**
- **Found during:** Full test suite execution in Task 2
- **Issue:** `if cfg.DestCalDAV == "" && (cfg.DestCalDAV == "")` redundant logic
- **Fix:** Simplified to single condition check
- **Files modified:** deploy/setup/pkg/wizard/migrate_wizard.go
- **Commit:** afb11e1

**3. [Rule 3 - Blocking] Progress bar signature mismatch**
- **Found during:** Initial build attempt
- **Issue:** UpdateFolder(folder, count) didn't match IMAPSync.OnProgress signature (folder, current, total)
- **Fix:** Changed UpdateFolder signature to match, updated bar.Current assignment
- **Files modified:** deploy/setup/pkg/wizard/progress.go
- **Commit:** 7882fb5

**4. [Rule 3 - Blocking] OAuth endpoint helper functions missing**
- **Found during:** migrate_wizard.go implementation
- **Issue:** Plan referenced providers.GoogleEndpoint() and providers.MicrosoftEndpoint() which don't exist
- **Fix:** Inline oauth2.Endpoint structs in getOAuthConfig() with correct URLs
- **Files modified:** deploy/setup/pkg/wizard/migrate_wizard.go
- **Commit:** 7882fb5

**5. [Rule 3 - Blocking] Test script grep flag interpretation**
- **Found during:** test-mail-migration.sh execution
- **Issue:** grep interpreted --from as grep option instead of search pattern
- **Fix:** Added `--` separator before pattern to stop option parsing
- **Files modified:** tests/test-mail-migration.sh
- **Commit:** afb11e1

## Integration Points

### With Plan 10-01 (IMAP Core)
- IMAPSync.DryRun() called for folder preview
- IMAPSync.SyncAll() executed with progress callbacks
- State file path passed through to IMAPSync

### With Plan 10-02 (CalDAV/CardDAV)
- CalDAVSync.DryRun() and CardDAVSync.DryRun() for preview
- CalDAVSync.SyncAll() and CardDAVSync.SyncAll() for migration
- ImportVCF() and ImportICS() for file imports

### With Plan 10-03 (Providers)
- Provider.WizardPrompts() for auth flows
- Provider.ConnectIMAP/CalDAV/CardDAV() for connections
- GenericProvider used for destination DarkPipe server

## Verification Results

All verification steps passed:

1. ✓ `darkpipe-setup migrate --help` shows all providers and flags
2. ✓ `darkpipe-setup migrate` (no --from) lists 7 supported providers in table
3. ✓ `darkpipe-setup migrate --from nonexistent` returns "not found" error
4. ✓ `go build ./cmd/darkpipe-setup` compiles successfully
5. ✓ `bash tests/test-mail-migration.sh` — all 34 tests pass

## Documentation Updates

- **REQUIREMENTS.md:** Added MIG-01 requirement, removed from out-of-scope, marked UX-01/UX-03/MON-01/MON-02/MON-03/CERT-03/CERT-04 complete, updated to 51 v1 requirements
- **ROADMAP.md:** Marked Phase 10 complete (4/4 plans), updated progress table, added completion date 2026-02-15

## Key Decisions

1. **Dry-run by default:** --apply required for actual migration to prevent accidental data operations
2. **OAuth2 credentials via environment:** GMAIL_CLIENT_ID and OUTLOOK_CLIENT_ID from env vars (per OAuth2 best practices)
3. **Terminal detection:** Automatic fallback from pterm progress bars to log-style output for CI/scripts
4. **Provider-agnostic wizard:** Used Provider.WizardPrompts() pattern for extensibility (easy to add new providers)
5. **Destination defaults:** localhost:993 (IMAP), localhost:5232 (Radicale CalDAV/CardDAV) for typical DarkPipe setup

## Files Created

- `deploy/setup/cmd/darkpipe-setup/migrate_cmd.go` (141 lines) — Cobra migrate command with 13 flags
- `deploy/setup/pkg/wizard/migrate_wizard.go` (564 lines) — Interactive wizard with dry-run and migration phases
- `deploy/setup/pkg/wizard/progress.go` (153 lines) — pterm progress bars with terminal fallback
- `deploy/setup/pkg/wizard/oauth_flow.go` (49 lines) — OAuth2 device flow UI with pterm
- `tests/test-mail-migration.sh` (236 lines) — Phase 10 integration test suite

## Files Modified

- `deploy/setup/cmd/darkpipe-setup/main.go` — Registered migrateCmd
- `.planning/REQUIREMENTS.md` — Added MIG-01, updated traceability
- `.planning/ROADMAP.md` — Marked Phase 10 complete

## Commits

- `7882fb5` — feat(10-04): add CLI migrate command with interactive wizard and progress display
- `afb11e1` — feat(10-04): add phase test suite and update project documentation

## Self-Check: PASSED

### Created files exist:
```
FOUND: deploy/setup/cmd/darkpipe-setup/migrate_cmd.go
FOUND: deploy/setup/pkg/wizard/migrate_wizard.go
FOUND: deploy/setup/pkg/wizard/progress.go
FOUND: deploy/setup/pkg/wizard/oauth_flow.go
FOUND: tests/test-mail-migration.sh
```

### Commits exist:
```
FOUND: 7882fb5 (feat(10-04): add CLI migrate command with interactive wizard and progress display)
FOUND: afb11e1 (feat(10-04): add phase test suite and update project documentation)
```

### Test suite passes:
```
Phase 10 test suite: ALL TESTS PASSED
Tests Passed: 34
Tests Failed: 0
Total Time: 8s
```

## Next Steps

Phase 10 is complete. All 4 plans executed successfully:
- Plan 01: IMAP sync engine ✓
- Plan 02: CalDAV/CardDAV sync ✓
- Plan 03: Provider integrations ✓
- Plan 04: CLI wizard & test suite ✓

Ready for Phase 10 final verification and milestone completion.
