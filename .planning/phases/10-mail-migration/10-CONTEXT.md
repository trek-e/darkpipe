# Phase 10: Mail Migration — Context

**Created:** 2026-02-14
**Source:** Discussion with user about implementation approach

## Provider Integration Depth

**Decision:** Provider-specific integrations where available, not just generic IMAP.

- **MailCow**: API integration for users, aliases, and mailboxes (success criteria #3)
- **Gmail**: OAuth2 device authorization grant (RFC 8628) for authentication, IMAP for mail sync
- **Outlook**: OAuth2 device authorization grant for authentication, IMAP for mail sync
- **iCloud**: App-password guidance in wizard flow, standard IMAP
- **Mailu**: API integration if available, fallback to IMAP
- **docker-mailserver**: Standard IMAP (no API, direct server access)
- **Generic**: Any IMAP server supported as fallback

**CLI wizard behavior:** The `--from <provider>` flag changes the wizard flow per provider:
- Gmail prompts for OAuth device flow
- MailCow prompts for API URL + API key
- iCloud explains how to create app passwords
- Generic asks for IMAP/CalDAV/CardDAV credentials directly

**OAuth2 strategy:** Device authorization grant (RFC 8628) — user opens browser URL, enters code, CLI receives token. No redirect URI complexity. Used for Gmail and Outlook.

## Migration Lifecycle

**Decision:** Production-quality UX with resumability, dry-run, and rich progress.

### Dry-run by default
- First invocation shows a preview: folder count, message count, calendar/contact count
- User must confirm interactively OR pass `--apply` to skip preview
- Aligns with Phase 4 DNS pattern (dry-run by default, `--apply` to execute)

### Full resume support
- Track migrated messages by Message-ID in a state file (JSON or SQLite)
- On re-run, skip already-migrated items
- State file stored alongside migration config (e.g., `/data/migration-state.json`)
- Allows interrupted migrations to pick up where they left off

### Per-folder progress bars
- Show current folder name, message count progress (142/1,203)
- Show overall folder progress (3/12 folders)
- Use pterm for rich terminal output (already in setup tool dependencies)
- Degrade gracefully to log-style output if terminal doesn't support it

### Error handling: skip and report
- Failed messages are logged (subject, date, error reason) and skipped
- Migration continues past individual failures
- Summary at end shows: migrated count, skipped count, skipped details
- No retry on individual messages (resume handles re-attempts on next run)

## Data Mapping & Conflicts

**Decision:** Smart defaults with maximum data preservation.

### Gmail labels → IMAP keywords
- Map Gmail labels to IMAP keywords/flags (Stalwart supports keywords natively)
- Preserves multi-label semantics without storage duplication
- Standard IMAP folders (Inbox, Sent, Drafts, Trash, Spam) mapped normally
- For Dovecot/Maddy: document that keywords may have limited support, offer `--labels-as-folders` fallback

### Contact duplicate handling: merge fields
- Match existing contacts by email address
- Fill in empty fields from import (phone, address, birthday, etc.)
- Do NOT overwrite fields that already have values
- Log merged contacts in summary

### Preserve original dates
- Use IMAP APPEND with original INTERNALDATE
- Messages appear at correct chronological position
- Essential for usability — migrated mailbox looks identical to source

### Configurable folder mapping
- **Smart defaults:** Map well-known provider folders:
  - `[Gmail]/Sent Mail` → `Sent`
  - `[Gmail]/Drafts` → `Drafts`
  - `[Gmail]/Trash` → `Trash`
  - `[Gmail]/Spam` → `Junk`
  - Skip virtual folders: `[Gmail]/All Mail`, `[Gmail]/Important`, `[Gmail]/Starred`
  - Outlook: `Clutter` → skip, `Deleted Items` → `Trash`
- **User overrides:** Config file or `--folder-map` CLI flag for custom mappings
- **Unknown folders:** Migrate with original name by default

## Import Modes

**Decision:** Support both file import AND live sync.

- **Live sync:** Connect to source IMAP/CalDAV/CardDAV server and sync directly
- **File import:** Import from .vcf (contacts) and .ics (calendars) files
- File import covers providers with export-only (e.g., iCloud Contacts export, Google Takeout)
- File import is simpler fallback when live server access is unavailable

## Deferred Ideas

None — all discussion stayed within Phase 10 scope.

---
*Context captured: 2026-02-14*
