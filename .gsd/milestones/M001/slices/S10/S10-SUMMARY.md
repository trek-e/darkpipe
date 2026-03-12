---
id: S10
parent: M001
milestone: M001
provides: []
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 
verification_result: passed
completed_at: 
blocker_discovered: false
---
# S10: Mail Migration

**# Phase 10 Plan 01: IMAP Migration Core Engine Summary**

## What Happened

# Phase 10 Plan 01: IMAP Migration Core Engine Summary

**Built the foundation for all provider integrations:** Thread-safe state tracking with JSON persistence, provider-specific folder mapping with user overrides, and IMAP sync engine with date/flag preservation using go-imap v2.

## What Was Built

### State Tracker (state.go)
- Thread-safe migration state with sync.RWMutex for concurrent access
- JSON persistence with 0600 permissions for security
- Auto-save every 100 operations to prevent data loss
- Tracks messages, calendar events, contacts, and folders by hash/ID
- Stats() method for migration summary display

**Deviations:** Linter enhanced the implementation with thread-safety (sync.RWMutex), auto-save logic, and folder tracking. These improvements were accepted as Rule 3 deviations (blocking issues/improvements).

### Folder Mapper (mapping.go)
- Provider-specific defaults for Gmail (skip All Mail/Important/Starred), Outlook (skip Clutter)
- User override support via map[string]string (empty value = skip)
- Label-to-keyword sanitization for IMAP atoms (replace spaces, remove special chars, handle digit prefixes)
- LabelsAsFolders flag for Dovecot/Maddy fallback (when keywords not supported)

### IMAP Sync Engine (imap.go)
- ListSourceFolders with STATUS command for fast message counts
- DryRun mode for migration preview without transferring
- SyncFolder with FETCH envelope/flags/body and APPEND with preserved INTERNALDATE
- SyncAll for sequential folder migration
- Progress callbacks (OnFolderStart, OnProgress, OnFolderDone) for UX
- Batch processing with configurable BatchSize (default 50)
- Skip-and-report error handling (errors logged, migration continues)

**Deviations:** go-imap v2 API differs from research examples. Fixed SeqSetRange() → SeqSet.AddRange(), msg.Envelope → msg.Collect().Envelope, statusData.NumMessages pointer dereference. These were Rule 3 deviations (blocking compilation errors).

## Commits

| Hash | Type | Description | Files |
|------|------|-------------|-------|
| 3bac915 | feat | State tracker and folder mapping | state.go, state_test.go, mapping.go, mapping_test.go, go.mod, go.sum |
| 788ae44 | feat | IMAP sync engine with date/flag preservation | imap.go, imap_test.go, go.mod, go.sum |

## Deviations from Plan

### Auto-fixed Issues (Rule 3 - Blocking)

**1. [Rule 3] Linter improved state.go with thread-safety and auto-save**
- **Found during:** Task 1
- **Issue:** Linter rewrote state.go with sync.RWMutex, auto-save every N operations, and constructor pattern
- **Fix:** Accepted linter improvements, updated tests to match new API (NewMigrationState(), error returns on Mark methods)
- **Files modified:** state.go, state_test.go
- **Impact:** Better implementation than planned (thread-safe, auto-save prevents data loss)

**2. [Rule 3] go-imap v2 API mismatches**
- **Found during:** Task 2
- **Issue:** Research examples used v1 API patterns (SeqSetRange, direct Envelope access). v2 requires SeqSet.AddRange(), Collect() for buffering, pointer dereference for NumMessages.
- **Fix:** Updated to v2 API: SeqSet.AddRange(1, totalMessages), msgData.Collect(), *statusData.NumMessages
- **Files modified:** imap.go
- **Impact:** Code now works with go-imap v2.0.0-beta.8

**3. [Rule 3] Linter created caldav.go and carddav.go with compilation errors**
- **Found during:** Task 1 and Task 2 test runs
- **Issue:** Linter auto-generated files for future tasks with API errors (PutCalendarObject return value mismatch)
- **Fix:** Removed caldav.go, carddav.go, caldav_test.go, carddav_test.go (out of scope for this plan)
- **Files removed:** 4 auto-generated files
- **Impact:** Clean build and test for core engine; files will be properly created in future plans

## Test Coverage

**34 tests, all passing:**

- **State tests (6):** NewMigrationState, save/load roundtrip, mark/check methods, stats, auto-save trigger
- **Mapping tests (8):** Gmail/Outlook/generic defaults, user overrides, AllMappings, LabelToKeyword sanitization
- **IMAP tests (11):** NewIMAPSync, folder mapping, DryRun calculation, SyncResult aggregation, state integration, batch size, progress callbacks, date preservation, skip logic
- **Integration tests (3):** State persistence, folder skip logic, callback invocation

**Code quality:**
- `go build ./pkg/mailmigrate/` ✓
- `go test ./pkg/mailmigrate/... -v` ✓ (34/34 passed)
- `go vet ./pkg/mailmigrate/...` ✓ (no issues)

## Key Decisions Made

1. **State file default path:** `/data/migration-state.json` aligns with container volume pattern (consistent with `/data/monitoring/` from Phase 9)
2. **Auto-save interval:** 100 operations balances data safety and I/O overhead
3. **go-imap version:** v2.0.0-beta.8 (latest beta, v2 API redesign with better concurrency) over v1 (deprecated)
4. **Folder tracking in state:** Added IsFolderMigrated/MarkFolderMigrated (not in original plan) enables folder-level resume in future
5. **Defer fetchCmd.Close():** Linter added defer to ensure cleanup on error paths (good practice)

## Dependencies Added

- `github.com/emersion/go-imap/v2 v2.0.0-beta.8` - IMAP client library
- `github.com/emersion/go-message v0.18.2` - Message parsing (transitive)
- `github.com/emersion/go-sasl` - SASL authentication (transitive)

## Success Criteria

All success criteria from plan met:

- [x] mailmigrate package builds cleanly with go-imap v2 dependency
- [x] State tracker handles load/save/hash/check/mark operations with tests
- [x] Folder mapper supports Gmail, Outlook, and generic providers with user overrides
- [x] IMAP sync engine has DryRun and SyncFolder/SyncAll methods
- [x] Skip-and-report error handling logs failures and continues
- [x] Resume support verified via state integration tests

**Additional achievements:**
- Thread-safe state tracking (bonus: enables concurrent operations in future)
- Auto-save every 100 operations (bonus: prevents data loss on interruption)
- Folder-level migration tracking (bonus: enables folder resume in future plans)

## What's Next

**Plan 02:** Provider integrations (Gmail OAuth2, MailCow API, generic IMAP) that consume this core engine.

**Plan 03:** CLI wizard with pterm progress bars, dry-run preview, and interactive provider selection.

**Key links ready:**
- state.IsMessageMigrated/MarkMessageMigrated → used by IMAP sync
- mapper.Map → used by IMAP folder listing
- IMAPSync.DryRun → will be called by CLI wizard
- IMAPSync.SyncAll → will be called by CLI migration command

## Self-Check: PASSED

**Files verified:**
- [x] deploy/setup/pkg/mailmigrate/state.go exists
- [x] deploy/setup/pkg/mailmigrate/state_test.go exists
- [x] deploy/setup/pkg/mailmigrate/mapping.go exists
- [x] deploy/setup/pkg/mailmigrate/mapping_test.go exists
- [x] deploy/setup/pkg/mailmigrate/imap.go exists
- [x] deploy/setup/pkg/mailmigrate/imap_test.go exists

**Commits verified:**
- [x] 3bac915 exists in git log
- [x] 788ae44 exists in git log

**Build/test verified:**
- [x] Package builds without errors
- [x] All 34 tests pass
- [x] go vet reports no issues

# Phase 10 Plan 02: CalDAV/CardDAV Sync and File Import Summary

**One-liner:** CalDAV/CardDAV live sync and VCF/ICS file import with contact merge (fill empty, don't overwrite) and state-based resume support

## What Was Built

### Task 1: CalDAV and CardDAV Sync Engines
- **CalDAV sync** discovers calendars via FindCalendarHomeSet, syncs events with UID-based resume tracking
- **CardDAV sync** discovers address books, syncs contacts with email-based matching and merge logic
- **Contact merge logic** implements locked decision: match by email, fill empty fields from source, preserve existing destination values
- **Merge modes** support append (default), overwrite, and skip strategies
- **Progress callbacks** provide OnProgress, OnCalendarStart, OnCalendarDone for UI integration
- **Dry-run support** for both CalDAV and CardDAV to preview without syncing
- **State integration** uses IsCalEventMigrated/MarkCalEventMigrated and IsContactMigrated/MarkContactMigrated

### Task 2: VCF and ICS File Import
- **VCF import** parses multi-contact .vcf files, applies same merge logic as CardDAV sync
- **ICS import** parses multi-event .ics files, wraps events in VCALENDAR envelope for CalDAV PUT
- **Dry-run for files** counts contacts (with/without email) and events without importing
- **State tracking** prevents duplicate imports on resume
- **Helper functions** parseVCFFile and parseICSFile handle multi-object file formats
- **Edge cases handled**: empty files, single contact, contacts without email

## Deviations from Plan

None - plan executed exactly as written.

## Dependencies Added

- `github.com/emersion/go-webdav v0.7.0` - CalDAV/CardDAV client library
- `github.com/emersion/go-vcard v0.0.0-20241024213814-c9703dde27ff` - vCard parsing
- `github.com/emersion/go-ical v0.0.0-20250609112844-439c63cef608` - iCalendar parsing

## Test Coverage

**CalDAV tests:**
- extractEventUID from iCal data (with UID, without UID)
- State integration (IsCalEventMigrated, MarkCalEventMigrated)

**CardDAV tests:**
- extractPrimaryEmail from vCard (single, multiple, none)
- mergeContact logic (fill empty fields, preserve existing, all merge modes)
- State integration (IsContactMigrated, MarkContactMigrated)

**File import tests:**
- parseVCFFile (multi-contact, single contact, empty file)
- parseICSFile (multi-event)
- DryRunVCF (count with/without email)
- DryRunICS (count events)

All tests pass. Code compiles without warnings.

## Integration Points

- **State tracking:** Integrates with MigrationState from 10-01
- **Contact merge:** Uses extractPrimaryEmail and mergeContact shared functions
- **Error handling:** Skip-and-report pattern consistent with IMAP sync
- **Progress reporting:** Callback pattern matches IMAP OnProgress style

## Next Steps

- Plan 10-03: OAuth2 provider integration (Gmail, Outlook device authorization grant)
- Plan 10-04: CLI wizard with dry-run, provider selection, progress bars

## Self-Check: PASSED

All claimed files verified to exist.
All claimed commits verified to exist in git history.

# Phase 10 Plan 03: Provider Integrations Summary

**Provider abstraction with OAuth2 device flow, MailCow API client, and 7 provider implementations covering Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, and generic IMAP/CalDAV/CardDAV servers.**

## What Was Built

### Task 1: Provider Interface and OAuth2 Providers

**Provider Interface (provider.go)**
- `Provider` interface with 15 methods covering authentication, capabilities, and configuration
- Capability flags: SupportsLabels, SupportsAPI, SupportsCalDAV, SupportsCardDAV
- Connection methods: ConnectIMAP, ConnectCalDAV, ConnectCardDAV
- Folder management: GetFolderMapping, GetSkipFolders
- Wizard integration: WizardPrompts returns provider-specific setup flows
- Registry pattern with Register/Get/List for provider lookup by slug
- Global `DefaultRegistry` populated via init() functions

**OAuth2 Device Flow (oauth.go)**
- `RunDeviceFlow` implements RFC 8628 device authorization grant
- `TokenSourceFromToken` creates auto-refreshing OAuth2 token source
- Custom XOAUTH2 SASL client (go-sasl only has OAUTHBEARER, not XOAUTH2)
- `xoauth2Client` implements sasl.Client interface for go-imap v2 compatibility
- XOAUTH2Token helper formats authentication string: `user=<email>\x01auth=Bearer <token>\x01\x01`

**Gmail Provider (gmail.go)**
- OAuth2 authentication via device flow
- XOAUTH2 SASL for IMAP (imap.gmail.com:993)
- CalDAV client pointing to https://www.googleapis.com/caldav/v2/
- CardDAV client pointing to https://www.googleapis.com/.well-known/carddav
- Folder mappings: [Gmail]/Sent Mail -> Sent, [Gmail]/Spam -> Junk
- Skip folders: [Gmail]/All Mail, [Gmail]/Important, [Gmail]/Starred
- SupportsLabels = true (Gmail uses labels instead of folders)
- OAuth2 scopes: mail.google.com, calendar, contacts
- Requires GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables

**Outlook Provider (outlook.go)**
- OAuth2 authentication via device flow (Microsoft requires OAuth2 after April 2026)
- XOAUTH2 SASL for IMAP (outlook.office365.com:993)
- CalDAV/CardDAV not supported (users should export .ics/.vcf from Outlook web)
- Folder mappings: Deleted Items -> Trash, Sent Items -> Sent, Junk Email -> Junk
- Skip folders: Clutter
- OAuth2 scopes: IMAP.AccessAsUser.All, offline_access
- Device auth URL: https://login.microsoftonline.com/common/oauth2/v2.0/devicecode
- Requires OUTLOOK_CLIENT_ID environment variable

**iCloud Provider (icloud.go)**
- App-specific password authentication (2FA prerequisite)
- Standard IMAP login (imap.mail.me.com:993)
- CalDAV client pointing to https://caldav.icloud.com/
- CardDAV client pointing to https://contacts.icloud.com/
- basicAuthTransport for CalDAV/CardDAV HTTP authentication
- Wizard prompts include detailed instructions for creating app-specific password at appleid.apple.com
- No folder mappings (iCloud uses standard names)

### Task 2: MailCow, Mailu, docker-mailserver, and Generic Providers

**MailCow Provider (mailcow.go)**
- SupportsAPI = true (API key authentication via X-API-Key header)
- API methods:
  - GetMailboxes: GET /api/v1/get/mailbox/all
  - GetAliases: GET /api/v1/get/alias/all
  - GetDomains: GET /api/v1/get/domain/all
  - ExportConfig: Aggregates all API data for migration preview
- MailCowMailbox struct: Username, Name, QuotaUsed, Messages, Active
- MailCowAlias struct: Address, GoTo, Active
- MailCowDomain struct: DomainName, Active
- Standard IMAP authentication for mailbox sync
- No CalDAV/CardDAV (users should use file import or separate server)

**Mailu Provider (mailu.go)**
- SupportsAPI = true with fallback to IMAP-only mode
- API methods:
  - GetUsers: GET /api/v1/user (Bearer token authentication)
  - GetAliases: GET /api/v1/alias
  - GetDomains: GET /api/v1/domain
- API errors trigger fallback warnings (older Mailu versions may not have API)
- Wizard prompts include optional API URL/key for IMAP-only operation

**docker-mailserver Provider (dockermailserver.go)**
- Standard IMAP-only provider
- No API, CalDAV, or CardDAV support
- Minimal wizard prompts: IMAP host, email, password
- Uses standard folder names

**Generic Provider (generic.go)**
- Flexible IMAP/CalDAV/CardDAV endpoint configuration
- Configurable IMAPPort with defaults: 993 (TLS), 143 (STARTTLS)
- SupportsCalDAV/SupportsCardDAV based on URL presence
- basicAuthTransport for CalDAV/CardDAV HTTP authentication
- No provider-specific folder mappings (user can override with --folder-map)
- Wizard prompts for all configurable fields (IMAP host/port/TLS, CalDAV URL, CardDAV URL)

**Comprehensive Test Suite (provider_test.go)**
- Registry tests: Register, Get, List, DefaultRegistry validation
- Capability tests: All 7 providers verify correct flags
- Folder mapping tests: Gmail and Outlook mappings verified
- XOAUTH2Token formatting test
- MailCow API tests: GetMailboxes, GetAliases with httptest mock server
- Mailu API tests: GetUsers with Bearer token authentication
- Generic provider tests: IMAPEndpoint with various port configurations
- WizardPrompts tests: All providers return valid prompt structures
- 19 tests, all passing

## Commits

| Hash    | Type | Description                                                          | Files |
|---------|------|----------------------------------------------------------------------|-------|
| 7720aa6 | feat | Provider interface with OAuth2 device flow and Gmail/Outlook/iCloud | 7     |
| 37e92c3 | feat | MailCow, Mailu, docker-mailserver, generic providers with tests     | 5     |

## Deviations from Plan

None - plan executed exactly as written.

**Note:** Custom XOAUTH2 SASL client was necessary because go-sasl only provides OAUTHBEARER (RFC 7628), not XOAUTH2 (Gmail/Outlook legacy mechanism). This is a correct implementation, not a deviation.

## Dependencies Added

- `golang.org/x/oauth2 v0.35.0` - OAuth2 client library
- `cloud.google.com/go/compute/metadata v0.3.0` - Google OAuth2 metadata (transitive)

## Test Coverage

**19 tests, all passing:**

- **Registry tests (3):** Register/Get, List, DefaultRegistry with all 7 providers
- **Gmail tests (3):** FolderMapping, SkipFolders, Capabilities
- **Outlook tests (2):** FolderMapping, Capabilities
- **iCloud tests (1):** Capabilities
- **MailCow tests (2):** Capabilities, API GetMailboxes/GetAliases
- **Mailu tests (2):** Capabilities, API GetUsers
- **docker-mailserver tests (1):** Capabilities
- **Generic provider tests (2):** Capabilities, IMAPEndpoint
- **OAuth2 tests (1):** XOAUTH2Token formatting
- **Wizard tests (1):** All providers return valid prompts
- **Integration test (1):** httptest-based API mocking for MailCow and Mailu

**Code quality:**
- `go build ./pkg/providers/` ✓
- `go test ./pkg/providers/... -v` ✓ (19/19 passed)
- `go vet ./pkg/providers/...` ✓ (no issues)

## Success Criteria

All success criteria from plan met:

- [x] Provider interface cleanly abstracts authentication and capabilities
- [x] Gmail and Outlook use OAuth2 device flow (RFC 8628) per locked decision
- [x] MailCow exports users, aliases, and mailboxes via API per success criteria #3
- [x] iCloud guides user through app password creation
- [x] Generic provider works with any IMAP server
- [x] All providers expose correct folder mappings and capability flags
- [x] Registry pattern allows provider lookup by slug

## Integration Points

**With mailmigrate package (10-01, 10-02):**
- Provider.ConnectIMAP returns imapclient.Client used by IMAPSync
- Provider.ConnectCalDAV returns caldav.Client used by CalDAVSync
- Provider.ConnectCardDAV returns carddav.Client used by CardDAVSync
- Provider.GetFolderMapping feeds into FolderMapper

**With CLI wizard (10-04, future):**
- DefaultRegistry.List() provides provider selection menu
- Provider.WizardPrompts() defines provider-specific setup flows
- Provider capability flags guide wizard flow (skip CalDAV if not supported)

## Key Decisions Made

1. **Custom XOAUTH2 SASL client:** go-sasl only has OAUTHBEARER, not XOAUTH2 (Gmail/Outlook requirement)
2. **Provider registry init():** Avoids import cycles by registering providers in package init functions
3. **Environment variables for OAuth2:** GMAIL_CLIENT_ID/GMAIL_CLIENT_SECRET and OUTLOOK_CLIENT_ID required (documented in wizard prompts)
4. **iCloud app-specific passwords:** Wizard prompts include full setup instructions with appleid.apple.com link
5. **Mailu API fallback:** Optional API with graceful degradation to IMAP-only mode for older versions
6. **Generic provider flexibility:** Supports any combination of IMAP/CalDAV/CardDAV based on URL configuration

## What's Next

**Plan 04:** CLI wizard with pterm progress bars, dry-run preview, interactive provider selection using DefaultRegistry and WizardPrompts.

**Provider integration flow:**
1. CLI wizard calls DefaultRegistry.List() for provider menu
2. User selects provider slug
3. CLI wizard calls provider.WizardPrompts() to get setup questions
4. For OAuth2 providers, CLI calls RunDeviceFlow with display callback
5. CLI creates provider instance with credentials
6. CLI passes provider.ConnectIMAP/CalDAV/CardDAV to sync engines from 10-01 and 10-02

## Self-Check: PASSED

**Files verified:**
- [x] deploy/setup/pkg/providers/provider.go exists
- [x] deploy/setup/pkg/providers/oauth.go exists
- [x] deploy/setup/pkg/providers/gmail.go exists
- [x] deploy/setup/pkg/providers/outlook.go exists
- [x] deploy/setup/pkg/providers/icloud.go exists
- [x] deploy/setup/pkg/providers/mailcow.go exists
- [x] deploy/setup/pkg/providers/mailu.go exists
- [x] deploy/setup/pkg/providers/dockermailserver.go exists
- [x] deploy/setup/pkg/providers/generic.go exists
- [x] deploy/setup/pkg/providers/provider_test.go exists

**Commits verified:**
- [x] 7720aa6 exists in git log
- [x] 37e92c3 exists in git log

**Build/test verified:**
- [x] Package builds without errors
- [x] All 19 tests pass
- [x] go vet reports no issues

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
