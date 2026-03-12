# T01: 10-mail-migration 01

**Slice:** S10 — **Milestone:** M001

## Description

Build the IMAP mail migration core engine: resumable state tracking, IMAP sync with date/flag preservation, and configurable folder mapping.

Purpose: This is the foundation that all provider integrations and the CLI wizard depend on. The IMAP sync engine handles the actual message transfer (fetch from source, append to destination), the state tracker enables resume after interruption, and the folder mapper translates provider-specific folders to standard IMAP folders.

Output: Three packages in deploy/setup/pkg/mailmigrate/ (state.go, imap.go, mapping.go) with comprehensive tests.

## Must-Haves

- [ ] "Migration state tracks migrated messages by Message-ID hash and survives process restart"
- [ ] "IMAP sync fetches messages from source and appends to destination preserving flags and INTERNALDATE"
- [ ] "Folder mapping translates provider-specific folders to standard IMAP folders with skip support"
- [ ] "Messages without Message-ID are tracked via fallback hash (From+Subject+Date)"
- [ ] "Already-migrated messages are skipped on re-run without re-downloading"

## Files

- `deploy/setup/pkg/mailmigrate/state.go`
- `deploy/setup/pkg/mailmigrate/state_test.go`
- `deploy/setup/pkg/mailmigrate/imap.go`
- `deploy/setup/pkg/mailmigrate/imap_test.go`
- `deploy/setup/pkg/mailmigrate/mapping.go`
- `deploy/setup/pkg/mailmigrate/mapping_test.go`
