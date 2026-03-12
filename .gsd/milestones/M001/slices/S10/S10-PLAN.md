# S10: Mail Migration

**Goal:** Build the IMAP mail migration core engine: resumable state tracking, IMAP sync with date/flag preservation, and configurable folder mapping.
**Demo:** Build the IMAP mail migration core engine: resumable state tracking, IMAP sync with date/flag preservation, and configurable folder mapping.

## Must-Haves


## Tasks

- [x] **T01: 10-mail-migration 01**
  - Build the IMAP mail migration core engine: resumable state tracking, IMAP sync with date/flag preservation, and configurable folder mapping.

Purpose: This is the foundation that all provider integrations and the CLI wizard depend on. The IMAP sync engine handles the actual message transfer (fetch from source, append to destination), the state tracker enables resume after interruption, and the folder mapper translates provider-specific folders to standard IMAP folders.

Output: Three packages in deploy/setup/pkg/mailmigrate/ (state.go, imap.go, mapping.go) with comprehensive tests.
- [x] **T02: 10-mail-migration 02**
  - Build CalDAV/CardDAV live sync and .vcf/.ics file import with contact merge logic and resume support.

Purpose: Enables migration of calendars and contacts alongside email. Supports both live sync from CalDAV/CardDAV servers (for providers with server access) and file import from .vcf/.ics exports (for providers like iCloud or Google Takeout). Contact merge follows locked decision: match by email, fill empty fields, don't overwrite existing values.

Output: Three files in deploy/setup/pkg/mailmigrate/ (caldav.go, carddav.go, fileimport.go) with tests.
- [x] **T03: 10-mail-migration 03**
  - Implement provider-specific integrations with authentication, capability detection, and folder mapping defaults for all supported providers.

Purpose: Each provider has unique authentication requirements (OAuth2 for Gmail/Outlook, API key for MailCow, app passwords for iCloud, plain credentials for generic) and folder mapping needs. The provider abstraction lets the CLI wizard and sync engine work uniformly regardless of source.

Output: Provider interface and implementations for Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, and generic IMAP in deploy/setup/pkg/providers/.
- [x] **T04: 10-mail-migration 04**
  - Build the CLI wizard command, progress display, and phase test suite that ties all migration components together into the user-facing `darkpipe-setup migrate` command.

Purpose: This is the user-facing integration layer. The `darkpipe-setup migrate --from <provider>` command guides users through authentication, shows a dry-run preview, and executes migration with rich progress feedback. The phase test suite validates all Phase 10 success criteria.

Output: Cobra migrate command, interactive wizard, progress display, and integration test suite.

## Files Likely Touched

- `deploy/setup/pkg/mailmigrate/state.go`
- `deploy/setup/pkg/mailmigrate/state_test.go`
- `deploy/setup/pkg/mailmigrate/imap.go`
- `deploy/setup/pkg/mailmigrate/imap_test.go`
- `deploy/setup/pkg/mailmigrate/mapping.go`
- `deploy/setup/pkg/mailmigrate/mapping_test.go`
- `deploy/setup/pkg/mailmigrate/caldav.go`
- `deploy/setup/pkg/mailmigrate/caldav_test.go`
- `deploy/setup/pkg/mailmigrate/carddav.go`
- `deploy/setup/pkg/mailmigrate/carddav_test.go`
- `deploy/setup/pkg/mailmigrate/fileimport.go`
- `deploy/setup/pkg/mailmigrate/fileimport_test.go`
- `deploy/setup/pkg/providers/provider.go`
- `deploy/setup/pkg/providers/gmail.go`
- `deploy/setup/pkg/providers/outlook.go`
- `deploy/setup/pkg/providers/icloud.go`
- `deploy/setup/pkg/providers/mailcow.go`
- `deploy/setup/pkg/providers/mailu.go`
- `deploy/setup/pkg/providers/dockermailserver.go`
- `deploy/setup/pkg/providers/generic.go`
- `deploy/setup/pkg/providers/oauth.go`
- `deploy/setup/pkg/providers/provider_test.go`
- `deploy/setup/cmd/darkpipe-setup/migrate_cmd.go`
- `deploy/setup/cmd/darkpipe-setup/main.go`
- `deploy/setup/pkg/wizard/migrate_wizard.go`
- `deploy/setup/pkg/wizard/progress.go`
- `deploy/setup/pkg/wizard/oauth_flow.go`
- `tests/test-mail-migration.sh`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
