# T02: 10-mail-migration 02

**Slice:** S10 — **Milestone:** M001

## Description

Build CalDAV/CardDAV live sync and .vcf/.ics file import with contact merge logic and resume support.

Purpose: Enables migration of calendars and contacts alongside email. Supports both live sync from CalDAV/CardDAV servers (for providers with server access) and file import from .vcf/.ics exports (for providers like iCloud or Google Takeout). Contact merge follows locked decision: match by email, fill empty fields, don't overwrite existing values.

Output: Three files in deploy/setup/pkg/mailmigrate/ (caldav.go, carddav.go, fileimport.go) with tests.

## Must-Haves

- [ ] "User can sync calendars from a source CalDAV server to destination CalDAV server"
- [ ] "User can sync contacts from a source CardDAV server to destination CardDAV server"
- [ ] "User can import contacts from .vcf files and calendars from .ics files"
- [ ] "Contact duplicate handling merges fields (fill empty, don't overwrite) matching by email"
- [ ] "Calendar events and contacts track migration state for resume support"

## Files

- `deploy/setup/pkg/mailmigrate/caldav.go`
- `deploy/setup/pkg/mailmigrate/caldav_test.go`
- `deploy/setup/pkg/mailmigrate/carddav.go`
- `deploy/setup/pkg/mailmigrate/carddav_test.go`
- `deploy/setup/pkg/mailmigrate/fileimport.go`
- `deploy/setup/pkg/mailmigrate/fileimport_test.go`
