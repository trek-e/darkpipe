---
phase: 10-mail-migration
plan: 02
subsystem: mail-migration
tags: [caldav, carddav, file-import, contact-merge, state-tracking]
dependency_graph:
  requires: [10-01 (migration state)]
  provides: [CalDAV sync, CardDAV sync, VCF import, ICS import]
  affects: []
tech_stack:
  added: [github.com/emersion/go-webdav, github.com/emersion/go-vcard, github.com/emersion/go-ical]
  patterns: [server sync, file import, contact merge, state resume]
key_files:
  created:
    - deploy/setup/pkg/mailmigrate/caldav.go
    - deploy/setup/pkg/mailmigrate/caldav_test.go
    - deploy/setup/pkg/mailmigrate/carddav.go
    - deploy/setup/pkg/mailmigrate/carddav_test.go
    - deploy/setup/pkg/mailmigrate/fileimport.go
    - deploy/setup/pkg/mailmigrate/fileimport_test.go
  modified:
    - deploy/setup/go.mod
    - deploy/setup/go.sum
decisions:
  - Contact merge uses "fill empty fields, don't overwrite" strategy per locked decision from 10-CONTEXT.md
  - CalDAV/CardDAV sync uses skip-and-report error handling (log and continue)
  - File import supports merge modes: append (default), overwrite, skip
  - VCALENDAR envelope wrapping for individual ICS events uses ical.NewCalendar() API
metrics:
  duration: 792s
  tasks_completed: 2
  files_created: 6
  completed_date: 2026-02-14
---

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
