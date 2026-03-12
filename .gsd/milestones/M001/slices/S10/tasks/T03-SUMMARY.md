---
id: T03
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
# T03: 10-mail-migration 03

**# Phase 10 Plan 03: Provider Integrations Summary**

## What Happened

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
