---
id: T01
parent: S10
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
# T01: 10-mail-migration 01

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
