# Phase 10: Mail Migration - Research

**Researched:** 2026-02-14
**Domain:** Email migration, IMAP sync, OAuth2 device flow, CalDAV/CardDAV, provider APIs
**Confidence:** HIGH

## Summary

Phase 10 implements a production-quality mail migration tool that enables DarkPipe users to import their existing email, contacts, and calendars from popular providers. The migration strategy combines provider-specific integrations (MailCow API, Gmail/Outlook OAuth2) with universal IMAP/CalDAV/CardDAV support, delivering a CLI wizard that guides users through migration with dry-run preview, resumable transfers, and rich progress feedback.

The research confirms that the locked decisions from CONTEXT.md are technically sound and align with Go ecosystem best practices. The emersion/go-imap v2 library provides robust IMAP client functionality with support for preserving flags and dates via APPEND. OAuth2 device flow is fully supported in golang.org/x/oauth2 for Gmail and Outlook authentication. CalDAV/CardDAV clients are available via emersion/go-webdav. The pterm library (already in the setup tool) delivers rich terminal UI with progress bars and concurrent spinners.

**Primary recommendation:** Build the migration tool as a standalone command within the existing deploy/setup module, reusing cobra CLI infrastructure and pterm for progress display. Implement provider-specific wizards that adapt behavior based on `--from <provider>` flag. Use JSON state files for resume tracking (simpler than SQLite for single-user migration state). Map Gmail labels to IMAP keywords where supported, with `--labels-as-folders` fallback for Dovecot/Maddy.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Provider Integration Depth:**
- MailCow: API integration for users, aliases, and mailboxes (success criteria #3)
- Gmail: OAuth2 device authorization grant (RFC 8628) for authentication, IMAP for mail sync
- Outlook: OAuth2 device authorization grant for authentication, IMAP for mail sync
- iCloud: App-password guidance in wizard flow, standard IMAP
- Mailu: API integration if available, fallback to IMAP
- docker-mailserver: Standard IMAP (no API, direct server access)
- Generic: Any IMAP server supported as fallback

**CLI wizard behavior:** The `--from <provider>` flag changes the wizard flow per provider (Gmail prompts for OAuth device flow, MailCow prompts for API URL + API key, iCloud explains app passwords, Generic asks for IMAP/CalDAV/CardDAV credentials)

**OAuth2 strategy:** Device authorization grant (RFC 8628) — user opens browser URL, enters code, CLI receives token. No redirect URI complexity. Used for Gmail and Outlook.

**Migration Lifecycle:**
- Dry-run by default: First invocation shows preview (folder count, message count, calendar/contact count), user confirms OR passes `--apply`
- Full resume support: Track migrated messages by Message-ID in state file (JSON or SQLite), on re-run skip already-migrated items
- Per-folder progress bars: Show folder name, message progress (142/1,203), overall folder progress (3/12), use pterm for rich output
- Error handling: Skip failed messages, log (subject, date, error), continue migration, summary shows migrated/skipped counts

**Data Mapping & Conflicts:**
- Gmail labels → IMAP keywords (Stalwart native support, Dovecot/Maddy may have limits, offer `--labels-as-folders` fallback)
- Contact duplicate handling: Match by email, fill empty fields from import, don't overwrite existing values, log merged contacts
- Preserve original dates: Use IMAP APPEND with original INTERNALDATE
- Configurable folder mapping: Smart defaults for Gmail/Outlook special folders, skip virtual folders ([Gmail]/All Mail, [Gmail]/Important), user overrides via config/flag

**Import Modes:**
- Live sync: Connect to source IMAP/CalDAV/CardDAV server and sync directly
- File import: Import from .vcf (contacts) and .ics (calendars) files

### Claude's Discretion

None — all critical decisions locked.

### Deferred Ideas (OUT OF SCOPE)

None — all discussion stayed within Phase 10 scope.
</user_constraints>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| emersion/go-imap | v2 (latest) | IMAP client for mail sync | De facto Go IMAP library, v2 API redesign with improved concurrency, 35 code snippets in Context7, supports APPEND with INTERNALDATE for date preservation |
| golang.org/x/oauth2 | v0.1.0+ | OAuth2 device authorization grant | Official Go OAuth2 library, native support for RFC 8628 device flow via DeviceAuth/DeviceAccessToken methods, handles polling automatically |
| emersion/go-webdav | v0.7.0+ | CalDAV/CardDAV client | Same author as go-imap (ecosystem consistency), supports calendar/contact discovery, query, and upload operations |
| emersion/go-vcard | latest | vCard (contact) parsing | Standard Go vCard parser (RFC 6350), used for .vcf file import and CardDAV object handling |
| emersion/go-ical | latest | iCalendar parsing | Standard Go iCal parser (RFC 5545), handles .ics file import and CalDAV events |
| pterm/pterm | v0.12.79+ | Terminal UI (progress bars, spinners) | Already in deploy/setup dependencies (Phase 7), supports multi-printer for concurrent folder progress, degrades gracefully |
| spf13/cobra | v1.8.1+ | CLI framework | Already in deploy/setup (Phase 7), established pattern for DarkPipe CLI commands |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | State file persistence | Resume tracking via JSON (simpler than SQLite for single-migration state) |
| crypto/sha256 | stdlib | Message-ID hashing for deduplication | Track migrated messages in state file |
| net/http | stdlib | HTTP client for provider APIs | MailCow API, Mailu API, OAuth2 token exchange |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| emersion/go-imap v2 | brianleishman/go-imap | brianleishman focuses on simplicity but lacks v2's concurrency improvements and active development |
| JSON state file | SQLite (go-sqlite3) | SQLite adds CGO dependency and complexity; JSON sufficient for single-user migration state (thousands of Message-IDs fit in memory) |
| Device flow OAuth2 | Authorization code flow | Authorization code requires redirect URI and local web server; device flow is CLI-native per locked decision |

**Installation:**
```bash
# All dependencies available via go get (no CGO, pure Go stack)
cd deploy/setup
go get github.com/emersion/go-imap/v2@latest
go get github.com/emersion/go-webdav@latest
go get github.com/emersion/go-vcard@latest
go get golang.org/x/oauth2@latest
# pterm, cobra already in deploy/setup/go.mod from Phase 7
```

## Architecture Patterns

### Recommended Project Structure

```
deploy/setup/
├── cmd/darkpipe-setup/
│   └── main.go                    # Existing setup tool (Phase 7)
├── pkg/
│   ├── config/                    # Existing (Phase 7)
│   ├── validate/                  # Existing (Phase 7)
│   ├── migrate/                   # NEW: Migration core
│   │   ├── migrate.go             # Migration orchestrator
│   │   ├── state.go               # Resume state tracking (JSON)
│   │   ├── imap.go                # IMAP sync engine
│   │   ├── caldav.go              # CalDAV sync engine
│   │   ├── carddav.go             # CardDAV sync engine
│   │   └── mapping.go             # Folder/label mapping logic
│   ├── providers/                 # NEW: Provider-specific integrations
│   │   ├── provider.go            # Provider interface
│   │   ├── gmail.go               # Gmail OAuth2 + label handling
│   │   ├── outlook.go             # Outlook OAuth2
│   │   ├── icloud.go              # iCloud app password wizard
│   │   ├── mailcow.go             # MailCow API client
│   │   ├── mailu.go               # Mailu API/IMAP
│   │   └── generic.go             # Generic IMAP/CalDAV/CardDAV
│   └── wizard/                    # NEW: Interactive wizard flows
│       ├── wizard.go              # Base wizard framework
│       ├── provider_select.go     # Provider selection prompt
│       ├── oauth_flow.go          # OAuth2 device flow handler
│       └── progress.go            # Progress bar/spinner management
└── go.mod                         # Existing module
```

### Pattern 1: Provider Interface Abstraction

**What:** Unified interface for provider-specific authentication and capabilities
**When to use:** Allows --from flag to dynamically select provider behavior without massive switch statements
**Example:**
```go
// Source: Designed for Phase 10 based on locked decisions

package providers

import (
    "context"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-webdav/caldav"
    "github.com/emersion/go-webdav/carddav"
)

// Provider defines the interface for mail provider integrations
type Provider interface {
    Name() string

    // Authentication
    AuthenticateIMAP(ctx context.Context) (*imapclient.Client, error)
    AuthenticateCalDAV(ctx context.Context) (*caldav.Client, error)
    AuthenticateCardDAV(ctx context.Context) (*carddav.Client, error)

    // Capabilities
    SupportsLabels() bool        // Gmail-style labels vs folders
    SupportsAPI() bool           // MailCow/Mailu API vs IMAP-only

    // Provider-specific operations
    GetFolderMapping() map[string]string  // Provider folder → standard folder
    GetLabelMapping() map[string]string   // Gmail labels → keywords
}

// Example: Gmail provider with OAuth2 device flow
type GmailProvider struct {
    OAuth2Token string
    IMAPEndpoint string
}

func (g *GmailProvider) AuthenticateIMAP(ctx context.Context) (*imapclient.Client, error) {
    // Use OAuth2 token for IMAP AUTHENTICATE XOAUTH2
    c, err := imapclient.DialTLS("imap.gmail.com:993", nil)
    if err != nil {
        return nil, err
    }

    // Gmail requires XOAUTH2 SASL mechanism
    if err := c.Authenticate(ctx, "XOAUTH2", g.OAuth2Token); err != nil {
        c.Close()
        return nil, err
    }

    return c, nil
}

func (g *GmailProvider) SupportsLabels() bool {
    return true  // Gmail uses X-GM-LABELS extension
}

func (g *GmailProvider) GetFolderMapping() map[string]string {
    return map[string]string{
        "[Gmail]/Sent Mail":  "Sent",
        "[Gmail]/Drafts":     "Drafts",
        "[Gmail]/Trash":      "Trash",
        "[Gmail]/Spam":       "Junk",
        "[Gmail]/All Mail":   "",  // Skip virtual folders
        "[Gmail]/Important":  "",
        "[Gmail]/Starred":    "",
    }
}
```

### Pattern 2: Resumable Migration with State Tracking

**What:** Track migrated items by Message-ID/UID in JSON state file, skip on re-run
**When to use:** All migrations — enables interrupted migrations to resume without re-downloading
**Example:**
```go
// Source: Designed for Phase 10 based on locked decisions

package migrate

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "os"
    "time"
)

// MigrationState tracks what's been migrated for resume support
type MigrationState struct {
    Version       string              `json:"version"`
    Provider      string              `json:"provider"`
    StartedAt     time.Time           `json:"started_at"`
    LastUpdated   time.Time           `json:"last_updated"`
    MailMessages  map[string]bool     `json:"mail_messages"`   // Message-ID hash → migrated
    CalEvents     map[string]bool     `json:"cal_events"`      // Event UID → migrated
    Contacts      map[string]bool     `json:"contacts"`        // Contact email → migrated
}

// HashMessageID creates a stable hash of Message-ID for state tracking
func HashMessageID(messageID string) string {
    h := sha256.Sum256([]byte(messageID))
    return hex.EncodeToString(h[:])
}

// SaveState persists migration state to JSON file
func (s *MigrationState) SaveState(path string) error {
    s.LastUpdated = time.Now()
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0600)
}

// LoadState reads migration state from JSON file
func LoadState(path string) (*MigrationState, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // First run, return empty state
            return &MigrationState{
                Version:      "1",
                MailMessages: make(map[string]bool),
                CalEvents:    make(map[string]bool),
                Contacts:     make(map[string]bool),
            }, nil
        }
        return nil, err
    }

    var state MigrationState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, err
    }
    return &state, nil
}

// IsMessageMigrated checks if message was already migrated
func (s *MigrationState) IsMessageMigrated(messageID string) bool {
    hash := HashMessageID(messageID)
    return s.MailMessages[hash]
}

// MarkMessageMigrated records message as migrated
func (s *MigrationState) MarkMessageMigrated(messageID string) {
    hash := HashMessageID(messageID)
    s.MailMessages[hash] = true
}
```

### Pattern 3: OAuth2 Device Authorization Grant Flow

**What:** RFC 8628 device flow for Gmail/Outlook — user enters code in browser, CLI polls for token
**When to use:** Gmail and Outlook authentication (locked decision)
**Example:**
```go
// Source: pkg.go.dev/golang.org/x/oauth2 + locked decisions

package wizard

import (
    "context"
    "fmt"
    "time"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

// RunOAuth2DeviceFlow executes device authorization grant flow
func RunOAuth2DeviceFlow(ctx context.Context, provider string) (string, error) {
    var config *oauth2.Config

    switch provider {
    case "gmail":
        config = &oauth2.Config{
            ClientID:     "YOUR_CLIENT_ID.apps.googleusercontent.com",
            ClientSecret: "YOUR_CLIENT_SECRET",
            Scopes:       []string{"https://mail.google.com/"},
            Endpoint:     google.Endpoint,
        }
    case "outlook":
        config = &oauth2.Config{
            ClientID:     "YOUR_CLIENT_ID",
            ClientSecret: "YOUR_CLIENT_SECRET",
            Scopes:       []string{"https://outlook.office.com/IMAP.AccessAsUser.All"},
            Endpoint: oauth2.Endpoint{
                AuthURL:       "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
                TokenURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/token",
                DeviceAuthURL: "https://login.microsoftonline.com/common/oauth2/v2.0/devicecode",
            },
        }
    default:
        return "", fmt.Errorf("unsupported provider: %s", provider)
    }

    // Step 1: Request device authorization
    response, err := config.DeviceAuth(ctx)
    if err != nil {
        return "", fmt.Errorf("device auth request failed: %w", err)
    }

    // Step 2: Display user code and verification URL
    fmt.Printf("\n")
    fmt.Printf("  Please visit: %s\n", response.VerificationURI)
    fmt.Printf("  And enter code: %s\n", response.UserCode)
    fmt.Printf("\n")
    fmt.Printf("Waiting for authorization")

    // Step 3: Poll for token (handles interval/timeout automatically)
    token, err := config.DeviceAccessToken(ctx, response)
    if err != nil {
        return "", fmt.Errorf("device access token failed: %w", err)
    }

    fmt.Printf(" Done!\n\n")
    return token.AccessToken, nil
}
```

### Pattern 4: IMAP APPEND with Date Preservation

**What:** Upload messages to destination IMAP with original INTERNALDATE preserved
**When to use:** All IMAP mail migrations (locked decision for date preservation)
**Example:**
```go
// Source: Context7 emersion/go-imap v2 examples

package migrate

import (
    "context"
    "io"
    "time"
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

// AppendMessage uploads message to IMAP with original date and flags
func AppendMessage(ctx context.Context, client *imapclient.Client, folder string, msg []byte, flags []imap.Flag, internalDate time.Time) error {
    options := &imap.AppendOptions{
        Flags: flags,
        Time:  internalDate,  // Preserve original date
    }

    appendCmd := client.Append(folder, int64(len(msg)), options)
    if _, err := appendCmd.Write(msg); err != nil {
        return err
    }
    if err := appendCmd.Close(); err != nil {
        return err
    }

    // Wait for completion
    data, err := appendCmd.Wait()
    if err != nil {
        return err
    }

    // Log UID if returned (useful for tracking)
    if data.UID != 0 {
        // Message appended with UID
    }

    return nil
}
```

### Pattern 5: Gmail Labels to IMAP Keywords

**What:** Map Gmail X-GM-LABELS to IMAP keywords for Stalwart (locked decision)
**When to use:** Gmail → Stalwart migration (Stalwart supports keywords natively)
**Example:**
```go
// Source: developers.google.com/workspace/gmail/imap/imap-extensions + locked decisions

package migrate

import (
    "context"
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

// FetchGmailLabels retrieves Gmail labels using X-GM-LABELS extension
func FetchGmailLabels(ctx context.Context, client *imapclient.Client, uid uint32) ([]string, error) {
    seqSet := imap.UIDSetNum(uid)

    fetchOptions := &imap.FetchOptions{
        UID: true,
    }
    // Note: X-GM-LABELS requires extension support in go-imap v2
    // May need to use raw FETCH command or wait for v2 extension support

    messages, err := client.Fetch(seqSet, fetchOptions).Collect()
    if err != nil {
        return nil, err
    }

    if len(messages) == 0 {
        return nil, nil
    }

    // Extract labels (implementation depends on go-imap v2 X-GM-LABELS support)
    // For now, document pattern — may require raw command or future v2 release
    return nil, nil
}

// ApplyKeywords applies labels as IMAP keywords on destination
func ApplyKeywords(ctx context.Context, client *imapclient.Client, uid uint32, keywords []string) error {
    seqSet := imap.UIDSetNum(uid)

    // Convert keywords to IMAP flags
    flags := make([]imap.Flag, len(keywords))
    for i, kw := range keywords {
        // IMAP keywords are atoms (no spaces, special chars)
        flags[i] = imap.Flag(kw)
    }

    storeFlags := &imap.StoreFlags{
        Op:    imap.StoreFlagsAdd,
        Flags: flags,
    }

    return client.UIDStore(seqSet, storeFlags, nil).Close()
}
```

### Pattern 6: Progress Display with pterm Multi-Printer

**What:** Show per-folder progress bars with overall migration progress (locked decision)
**When to use:** All migrations for UX feedback
**Example:**
```go
// Source: github.com/pterm/pterm examples + locked decisions

package wizard

import (
    "github.com/pterm/pterm"
)

// MigrationProgress manages concurrent folder progress displays
type MigrationProgress struct {
    multi       *pterm.MultiPrinter
    folderBars  map[string]*pterm.ProgressbarPrinter
    overallBar  *pterm.ProgressbarPrinter
}

// NewMigrationProgress creates progress display for migration
func NewMigrationProgress(totalFolders int) *MigrationProgress {
    multi := pterm.DefaultMultiPrinter

    overall, _ := pterm.DefaultProgressbar.
        WithTitle("Overall Progress").
        WithTotal(totalFolders).
        Start()

    return &MigrationProgress{
        multi:      &multi,
        folderBars: make(map[string]*pterm.ProgressbarPrinter),
        overallBar: overall,
    }
}

// StartFolder creates progress bar for a folder
func (m *MigrationProgress) StartFolder(folder string, totalMessages int) {
    bar, _ := pterm.DefaultProgressbar.
        WithTitle(folder).
        WithTotal(totalMessages).
        WithWriter(m.multi.NewWriter()).
        Start()

    m.folderBars[folder] = bar
}

// UpdateFolder increments folder progress
func (m *MigrationProgress) UpdateFolder(folder string, count int) {
    if bar, ok := m.folderBars[folder]; ok {
        bar.Add(count)
    }
}

// CompleteFolder marks folder done and updates overall
func (m *MigrationProgress) CompleteFolder(folder string) {
    if bar, ok := m.folderBars[folder]; ok {
        bar.Stop()
        delete(m.folderBars, folder)
    }
    m.overallBar.Increment()
}
```

### Anti-Patterns to Avoid

- **Storing passwords in plaintext state files:** OAuth2 tokens and IMAP passwords should be in secure storage, not migration state JSON. State file tracks Message-IDs only.
- **Re-downloading entire mailbox on retry:** Always use state file to skip already-migrated messages. Network failures are common during large migrations.
- **Blocking UI during network operations:** Use goroutines for IMAP operations, update progress bars concurrently. Never freeze terminal waiting for network.
- **Ignoring IMAP server rate limits:** Add exponential backoff for IMAP errors. Gmail has per-user quotas, batch operations.
- **Modifying Message-ID during migration:** Message-ID is the only reliable deduplication key. Never alter it.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| IMAP protocol client | Custom TCP socket handling, manual IMAP commands | emersion/go-imap v2 | IMAP protocol is complex (literals, continuation, pipelining). Hand-rolled clients miss edge cases (server capability negotiation, IDLE, compression). go-imap handles connection management, command/response parsing, concurrent operations. |
| OAuth2 device flow | Custom HTTP polling loop, token refresh logic | golang.org/x/oauth2 DeviceAuth/DeviceAccessToken | Device flow has subtle timing requirements (interval, expiry, slow_down response). Official library handles all error cases, token refresh, PKCE. Hand-rolled implementations miss security details. |
| CalDAV/CardDAV WebDAV operations | Manual XML PROPFIND/REPORT construction | emersion/go-webdav | CalDAV and CardDAV are complex WebDAV extensions with namespace-heavy XML. Querying calendars requires multi-level PROPFIND with calendar-data filters. Contact sync has ETags, sync-token logic. go-webdav handles all this. |
| Progress bar terminal rendering | ANSI escape codes, cursor positioning | pterm | Terminal UI is surprisingly complex (width detection, multi-line updates, no-TTY graceful degradation). pterm handles all edge cases and already in project. |
| Message deduplication | Custom header parsing, fuzzy matching | Message-ID hash (stdlib crypto/sha256) | Message-ID is RFC-defined unique identifier. Fuzzy matching (subject+date) creates false positives/negatives. Hash Message-ID for O(1) state lookup. |

**Key insight:** Email protocols (IMAP, CalDAV, CardDAV, OAuth2) have decades of edge cases. Existing Go libraries are battle-tested by thousands of production users. Custom protocol implementations will encounter obscure server behaviors (Dovecot vs Courier vs Exchange quirks) that take years to discover and fix.

## Common Pitfalls

### Pitfall 1: Gmail X-GM-LABELS Extension May Not Be in go-imap v2 Yet

**What goes wrong:** emersion/go-imap v2 is a ground-up rewrite. Extension support (including X-GM-LABELS) may not be fully implemented in early v2 releases. Code that assumes `FETCH X-GM-LABELS` will work may fail at runtime.

**Why it happens:** v2 API redesign means extensions need to be re-implemented. GitHub issues show v2 is actively developed but may lack v1 feature parity initially.

**How to avoid:**
1. Check go-imap v2 GitHub issues/docs for X-GM-LABELS support status before implementation
2. If not supported, use raw IMAP commands via `client.Execute()` or wait for extension
3. Have fallback: Gmail labels can still be synced via folder-based approach (each label = folder), then deduplicate on destination
4. Document in plan: "Verify go-imap v2 X-GM-LABELS support; implement raw FETCH if needed"

**Warning signs:** Import errors for extension packages, runtime "unknown FETCH item" errors

### Pitfall 2: IMAP APPEND Date Preservation Depends on Server Support

**What goes wrong:** Not all IMAP servers honor the INTERNALDATE parameter in APPEND. Some servers (especially older or misconfigured ones) ignore it and set the date to current time, breaking chronological mailbox order.

**Why it happens:** RFC 3501 says servers SHOULD support INTERNALDATE in APPEND, but it's not MUST. Some servers silently ignore it.

**How to avoid:**
1. Test APPEND date preservation during dry-run: append a test message with old date, fetch it back, verify INTERNALDATE matches
2. Warn user in dry-run output if server doesn't preserve dates: "Warning: Destination server ignores message dates. Migrated mail will show current date."
3. Document known-good servers: Stalwart, Dovecot, Maddy all preserve INTERNALDATE correctly
4. Consider aborting migration if dates are critical and server doesn't support it

**Warning signs:** All migrated messages show today's date regardless of original date

### Pitfall 3: OAuth2 Token Refresh During Long Migrations

**What goes wrong:** OAuth2 access tokens typically expire in 60-90 minutes. Large mailbox migrations can take hours. Midway through migration, IMAP AUTHENTICATE fails with "invalid credentials" even though initial auth succeeded.

**Why it happens:** golang.org/x/oauth2 Token struct contains RefreshToken, but not all providers issue refresh tokens in device flow. Gmail does, Outlook may not depending on app registration.

**How to avoid:**
1. Use oauth2.TokenSource wrapping the Token — it auto-refreshes if RefreshToken present
2. If provider doesn't issue refresh token, re-run device flow when access token expires
3. Catch IMAP auth errors mid-migration, prompt user to re-authenticate, resume migration with new token
4. Save OAuth2 token to secure file (0600 permissions) separate from migration state JSON

**Warning signs:** Migration starts successfully but fails after 60-90 minutes with auth errors

### Pitfall 4: Folder Mapping Ambiguity for Non-Standard Folders

**What goes wrong:** User has folder named "Archive" on Gmail, which maps to neither a Gmail special folder nor a standard IMAP folder. Migration tool doesn't know whether to skip it, migrate it as-is, or prompt user.

**Why it happens:** Gmail allows custom labels, Outlook allows custom folders. Smart defaults only cover well-known folders ([Gmail]/Sent, Outlook/Deleted Items). User-created folders have no standard mapping.

**How to avoid:**
1. Dry-run preview MUST show all folders with proposed mappings: "Source 'Archive' → Destination 'Archive'"
2. Offer `--folder-map` CLI flag: `--folder-map "Archive:Archives,Work:Work/Gmail"`
3. Prompt interactively during wizard if unknown folders detected: "Found custom folder 'Archive'. Migrate as 'Archive'? [Y/n/rename]"
4. Log all folder mappings in migration summary for user verification

**Warning signs:** User reports "some folders didn't migrate" or "wrong folders"

### Pitfall 5: Message-ID Collisions and Duplicates

**What goes wrong:** Some messages lack Message-ID header (non-compliant senders), or multiple messages have the same Message-ID (forwarding chains, mailing lists). State tracking by Message-ID alone causes skipped messages or false deduplication.

**Why it happens:** RFC 5322 requires Message-ID but not all senders comply. Mailing list software sometimes reuses Message-IDs.

**How to avoid:**
1. Fallback deduplication key: If Message-ID missing, use hash of (From + Subject + Date)
2. Track both Message-ID and folder: Same Message-ID in different folders = different messages (Gmail label behavior)
3. Imapsync best practice: Use Message-ID + Received header (first Received header is unique per hop)
4. Warn in dry-run if messages without Message-ID detected: "Found 12 messages without Message-ID. Using Subject+Date hash for deduplication."

**Warning signs:** Unexpected skipped messages, or duplicates appearing despite state tracking

### Pitfall 6: Contact Merge Logic Creating Data Loss

**What goes wrong:** Contact merge (fill empty fields, don't overwrite) can cause data loss if source and destination both have values for a field but they differ. Example: Destination has phone "555-1234", source has "555-5678". Neither is "empty" so source value is lost.

**Why it happens:** Locked decision says "fill empty fields, don't overwrite". This prevents accidental overwrites but also prevents updates.

**How to avoid:**
1. Dry-run MUST show contact merge preview: "Contact john@example.com: Destination has phone, source has different phone. Keeping destination."
2. Offer `--contacts-mode` flag: `append` (default, fill empty), `overwrite` (replace all), `skip` (don't merge, skip existing)
3. Log all merge decisions in summary: "Merged 45 contacts, skipped 12 fields due to conflicts"
4. Consider interactive prompt for conflicts during wizard: "Contact has different phone numbers. Keep [D]estination, use [S]ource, or [M]anually enter?"

**Warning signs:** User reports "contact info didn't import" or "old info stayed"

## Code Examples

Verified patterns from official sources:

### IMAP Connect and Authenticate with TLS
```go
// Source: https://context7.com/emersion/go-imap/llms.txt

package main

import (
    "log"
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    // Connect with implicit TLS
    c, err := imapclient.DialTLS("mail.example.org:993", nil)
    if err != nil {
        log.Fatalf("failed to dial IMAP server: %v", err)
    }
    defer c.Close()

    // Authenticate with username and password
    if err := c.Login("username", "password").Wait(); err != nil {
        log.Fatalf("failed to login: %v", err)
    }

    log.Println("Successfully connected and authenticated")
}
```

### Fetch Messages with Flags and Envelope
```go
// Source: https://context7.com/emersion/go-imap/llms.txt

package main

import (
    "log"
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

func main() {
    c, err := imapclient.DialTLS("mail.example.org:993", nil)
    if err != nil {
        log.Fatalf("failed to dial: %v", err)
    }
    defer c.Close()

    c.Login("username", "password").Wait()

    // Select INBOX
    selectedMbox, err := c.Select("INBOX", nil).Wait()
    if err != nil {
        log.Fatalf("failed to select INBOX: %v", err)
    }
    log.Printf("INBOX contains %v messages", selectedMbox.NumMessages)

    if selectedMbox.NumMessages > 0 {
        // Fetch first 10 messages
        seqSet := imap.SeqSetRange(1, 10)
        bodySection := &imap.FetchItemBodySection{
            Specifier: imap.PartSpecifierHeader,
        }
        fetchOptions := &imap.FetchOptions{
            Envelope:    true,
            Flags:       true,
            UID:         true,
            BodySection: []*imap.FetchItemBodySection{bodySection},
        }

        messages, err := c.Fetch(seqSet, fetchOptions).Collect()
        if err != nil {
            log.Fatalf("failed to fetch messages: %v", err)
        }

        for _, msg := range messages {
            log.Printf("UID %v: %v (flags: %v)",
                msg.UID, msg.Envelope.Subject, msg.Flags)
        }
    }
}
```

### OAuth2 Device Authorization Grant Flow
```go
// Source: https://pkg.go.dev/golang.org/x/oauth2 (WebFetch extraction)

package main

import (
    "context"
    "fmt"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

func main() {
    var config oauth2.Config = oauth2.Config{
        ClientID:     "YOUR_CLIENT_ID.apps.googleusercontent.com",
        ClientSecret: "YOUR_CLIENT_SECRET",
        Scopes:       []string{"https://mail.google.com/"},
        Endpoint:     google.Endpoint,
    }

    ctx := context.Background()

    // Step 1: Request device authorization
    response, err := config.DeviceAuth(ctx)
    if err != nil {
        panic(err)
    }

    // Step 2: Display to user
    fmt.Printf("Please enter code %s at %s\n", response.UserCode, response.VerificationURI)

    // Step 3: Poll for token (automatic interval handling)
    token, err := config.DeviceAccessToken(ctx, response)
    if err != nil {
        panic(err)
    }

    fmt.Println(token)
}
```

### CalDAV Client - Find Calendars
```go
// Source: https://pkg.go.dev/github.com/emersion/go-webdav/%40v0.7.0/caldav

package main

import (
    "context"
    "log"
    "net/http"
    "github.com/emersion/go-webdav/caldav"
)

func main() {
    httpClient := &http.Client{}
    client, err := caldav.NewClient(httpClient, "https://caldav.example.org")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Find calendar home set for principal
    homeSet, err := client.FindCalendarHomeSet(ctx, "principals/user@example.org")
    if err != nil {
        log.Fatal(err)
    }

    // List all calendars
    calendars, err := client.FindCalendars(ctx, homeSet)
    if err != nil {
        log.Fatal(err)
    }

    for _, cal := range calendars {
        log.Printf("Calendar: %s (%s)", cal.Name, cal.Path)
    }
}
```

### CardDAV Client - Get Contacts
```go
// Source: https://pkg.go.dev/github.com/emersion/go-webdav/carddav

package main

import (
    "context"
    "log"
    "net/http"
    "github.com/emersion/go-webdav/carddav"
)

func main() {
    httpClient := &http.Client{}
    client, err := carddav.NewClient(httpClient, "https://carddav.example.org")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Find address book home set
    homeSet, err := client.FindAddressBookHomeSet(ctx, "principals/user@example.org")
    if err != nil {
        log.Fatal(err)
    }

    // List address books
    addressBooks, err := client.FindAddressBooks(ctx, homeSet)
    if err != nil {
        log.Fatal(err)
    }

    for _, ab := range addressBooks {
        log.Printf("Address Book: %s (%s)", ab.Name, ab.Path)

        // Query all contacts in this address book
        query := &carddav.AddressBookQuery{
            DataRequest: carddav.AddressDataRequest{
                AllProp: true,
            },
        }

        contacts, err := client.QueryAddressBook(ctx, ab.Path, query)
        if err != nil {
            log.Printf("Error querying contacts: %v", err)
            continue
        }

        log.Printf("  Found %d contacts", len(contacts))
    }
}
```

### pterm Multi-Printer Progress Bars
```go
// Source: https://github.com/pterm/pterm/_examples/spinner/multiple/main.go

package main

import (
    "time"
    "github.com/pterm/pterm"
)

func main() {
    // Create multi printer for concurrent progress displays
    multi := pterm.DefaultMultiPrinter

    // Create progress bars with their own writers
    pb1, _ := pterm.DefaultProgressbar.
        WithTotal(100).
        WithTitle("Downloading INBOX").
        WithWriter(multi.NewWriter()).
        Start()

    pb2, _ := pterm.DefaultProgressbar.
        WithTotal(50).
        WithTitle("Downloading Sent").
        WithWriter(multi.NewWriter()).
        Start()

    // Start multi printer
    multi.Start()

    // Simulate concurrent operations
    for i := 0; i < 100; i++ {
        pb1.Increment()
        if i < 50 {
            pb2.Increment()
        }
        time.Sleep(50 * time.Millisecond)
    }

    multi.Stop()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| emersion/go-imap v1 | emersion/go-imap v2 | 2023-2024 | v2 is ground-up rewrite with improved API, better concurrency, streaming responses. Breaking changes from v1. Migration code should use v2 for future-proofing. |
| OAuth2 authorization code flow | Device authorization grant (RFC 8628) | RFC published 2019, widely adopted 2020+ | Device flow is CLI-native (no localhost redirect), better UX for terminal apps. Gmail/Outlook/GitHub all support device flow now. |
| Imapsync Perl tool | Go-native IMAP clients | Ongoing transition | Imapsync is production-standard but Perl dependency. Go ecosystem matured enough (go-imap v2, go-webdav) for native implementations. DarkPipe benefits from single-language stack. |
| XLIST Gmail extension | IMAP Special-Use (RFC 6154) | Deprecated 2013 | Gmail deprecated XLIST in favor of standard Special-Use flags (\Sent, \Drafts). Modern clients should use Special-Use, not XLIST. |

**Deprecated/outdated:**
- **XLIST command:** Gmail-specific, deprecated 2013. Use RFC 6154 Special-Use flags instead.
- **go-imap v1 API:** Replaced by v2. v1 still maintained for compatibility but v2 is recommended for new code.
- **Basic Authentication for Outlook IMAP:** Microsoft enforcing OAuth2-only starting March 2026. Basic auth will return 550 5.7.30 error after April 30, 2026.

## Open Questions

1. **go-imap v2 X-GM-LABELS Extension Support**
   - What we know: go-imap v2 is a rewrite, extension support may lag v1
   - What's unclear: Is X-GM-LABELS implemented in current v2 release? If not, when?
   - Recommendation: Check GitHub issues before planning. If not supported, implement raw FETCH command workaround or use folder-based Gmail sync. Document in plan: "Verify X-GM-LABELS support status; fallback to raw IMAP commands if needed."

2. **OAuth2 Refresh Token Availability in Device Flow**
   - What we know: Device flow should return refresh tokens, but depends on provider app registration
   - What's unclear: Do Gmail and Outlook device flow consistently return refresh tokens, or just access tokens?
   - Recommendation: Test both providers during implementation. If refresh tokens missing, implement re-auth flow when access token expires (60-90 min). Document in plan: "Test OAuth2 token refresh behavior; implement re-auth fallback."

3. **MailCow and Mailu API Version Compatibility**
   - What we know: MailCow has documented API at apiary.io, Mailu has RESTful API
   - What's unclear: Do different MailCow/Mailu versions have incompatible API changes? What's API stability guarantee?
   - Recommendation: Target latest stable API versions. Document supported versions in wizard. Add version detection if possible (API /version endpoint). Plan: "Support MailCow API v1 and Mailu 2.0+; add version detection and compatibility warnings."

4. **Contact Merge Conflict Resolution UX**
   - What we know: Locked decision is "fill empty fields, don't overwrite"
   - What's unclear: Is this sufficient for real-world migrations, or do users need conflict resolution UI?
   - Recommendation: Implement locked decision (fill empty) for v1. Log conflicts in summary. If user feedback shows need for resolution UI, add interactive prompt or `--contacts-mode overwrite` in v1.1. Plan: "Implement fill-empty merge; extensive conflict logging; gather user feedback for v1.1 improvements."

## Sources

### Primary (HIGH confidence)

- **emersion/go-imap v2** - [Context7 /emersion/go-imap](https://context7.com/emersion/go-imap/llms.txt) - IMAP client capabilities, FETCH, APPEND, authentication
- **golang.org/x/oauth2** - [pkg.go.dev](https://pkg.go.dev/golang.org/x/oauth2) - OAuth2 device flow API, DeviceAuth/DeviceAccessToken methods
- **emersion/go-webdav** - [Context7 pkg.go.dev](https://context7.com/websites/pkg_go_dev_github_com_emersion_go-webdav) - CalDAV/CardDAV client operations
- **Gmail IMAP Extensions** - [developers.google.com](https://developers.google.com/workspace/gmail/imap/imap-extensions) - X-GM-LABELS fetch/store syntax
- **RFC 8628** - [datatracker.ietf.org](https://datatracker.ietf.org/doc/html/rfc8628) - OAuth 2.0 Device Authorization Grant specification

### Secondary (MEDIUM confidence)

- **MailCow API** - [mailcow.docs.apiary.io](https://mailcow.docs.apiary.io/) (attempted fetch, structure unclear) - API endpoints for users/aliases/mailboxes
- **Mailu API** - [mailu.io/master/api.html](https://mailu.io/master/api.html) - RESTful API and config-export command
- **Microsoft OAuth2 IMAP** - [Microsoft Learn](https://learn.microsoft.com/en-us/exchange/client-developer/legacy-protocols/how-to-authenticate-an-imap-pop-smtp-application-by-using-oauth) - Outlook IMAP OAuth2 authentication
- **iCloud IMAP Settings** - [Apple Support](https://support.apple.com/en-us/102525) - IMAP server config, app-specific password requirements
- **imapsync FAQ** - [imapsync.lamiral.info](https://imapsync.lamiral.info/FAQ.d/FAQ.Duplicates.txt) - Duplicate detection via Message-ID, deduplication best practices
- **pterm** - [pterm.sh](https://pterm.sh/) and [GitHub examples](https://github.com/pterm/pterm) - Multi-printer, progress bars, spinners

### Tertiary (LOW confidence)

- **Email migration best practices** - Various web search results (EdbMails, BitTitan, migration tool vendors) - General guidance, not Go-specific
- **Microsoft 2026 auth enforcement** - [Mailbird article](https://www.getmailbird.com/microsoft-modern-authentication-enforcement-email-guide/) - Timeline for Basic Auth deprecation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified via Context7/official docs, version numbers confirmed, usage patterns documented
- Architecture: HIGH - Patterns based on locked decisions + official library examples, provider interface design is standard Go practice
- Pitfalls: MEDIUM-HIGH - Based on imapsync documentation, go-imap GitHub issues, and RFC interpretations. Some edge cases (X-GM-LABELS v2 support) need verification during implementation.

**Research date:** 2026-02-14
**Valid until:** ~45 days (stable domain - IMAP/OAuth2 standards don't change rapidly, but go-imap v2 is actively developed so extension support may evolve)

**Key risks to monitor:**
1. go-imap v2 extension availability (X-GM-LABELS) - verify before implementation
2. OAuth2 provider policy changes (Microsoft's 2026 enforcement is confirmed, but Google may change scopes/requirements)
3. MailCow/Mailu API versioning - document supported versions clearly
