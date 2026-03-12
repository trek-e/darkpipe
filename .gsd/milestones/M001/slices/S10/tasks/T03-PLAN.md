# T03: 10-mail-migration 03

**Slice:** S10 — **Milestone:** M001

## Description

Implement provider-specific integrations with authentication, capability detection, and folder mapping defaults for all supported providers.

Purpose: Each provider has unique authentication requirements (OAuth2 for Gmail/Outlook, API key for MailCow, app passwords for iCloud, plain credentials for generic) and folder mapping needs. The provider abstraction lets the CLI wizard and sync engine work uniformly regardless of source.

Output: Provider interface and implementations for Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, and generic IMAP in deploy/setup/pkg/providers/.

## Must-Haves

- [ ] "Gmail provider authenticates via OAuth2 device flow and returns IMAP client"
- [ ] "Outlook provider authenticates via OAuth2 device flow and returns IMAP client"
- [ ] "MailCow provider connects via API to export users, aliases, and mailboxes"
- [ ] "iCloud provider guides user to create app passwords then uses standard IMAP"
- [ ] "Generic provider accepts IMAP/CalDAV/CardDAV credentials directly"
- [ ] "All providers expose folder mapping defaults and capability flags"

## Files

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
