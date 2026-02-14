---
phase: 04-dns-email-auth
plan: 01
subsystem: dns
tags: [dkim, spf, dmarc, mx, dns-records, email-authentication]
dependency-graph:
  requires: []
  provides:
    - dkim-key-generation
    - dkim-signing
    - dns-record-generation
    - dns-records-guide
  affects:
    - email-authentication
    - deliverability
tech-stack:
  added:
    - emersion/go-msgauth (v0.7.0) - DKIM signing and verification
    - fatih/color (v1.18.0) - Terminal color output
  patterns:
    - stdlib-first (crypto/rsa, crypto/x509, encoding/pem for key management)
    - environment-based configuration (12-factor pattern)
    - time-based selector naming for automated rotation
key-files:
  created:
    - dns/dkim/keygen.go - 2048-bit RSA key generation and PEM encoding
    - dns/dkim/selector.go - Time-based DKIM selector naming ({prefix}-{YYYY}q{Q})
    - dns/dkim/signer.go - DKIM message signing via emersion/go-msgauth
    - dns/records/spf.go - SPF record generation with ip4: mechanism
    - dns/records/dkim.go - DKIM TXT record generation
    - dns/records/dmarc.go - DMARC record with subdomain protection (sp= tag)
    - dns/records/mx.go - MX record generation
    - dns/records/guide.go - DNS-RECORDS.md guide generation and terminal output
    - dns/config/config.go - Environment-based DNS configuration
  modified: []
decisions:
  - decision: Use time-based DKIM selector format ({prefix}-{YYYY}q{Q})
    rationale: Enables automated quarterly key rotation, makes rotation window obvious, allows multiple selectors to coexist during transition
    alternatives: [date-based, incremental numbering, random strings]
  - decision: Use explicit ip4: mechanism in SPF records (not include:)
    rationale: Minimizes DNS lookup count (RFC 7208 enforces 10 lookup limit), ip4: doesn't count as a lookup
    alternatives: [include: mechanisms for third-party services]
  - decision: Include sp= tag in DMARC records with default sp=quarantine
    rationale: Protects subdomains from spoofing (pitfall #5 from research), more restrictive than primary domain policy
    alternatives: [omit sp= tag, use sp=none]
  - decision: Default DMARC policy to p=none for initial setup
    rationale: Monitoring-first approach allows verification before enforcement, aligns with best practices progression (none -> quarantine -> reject)
    alternatives: [start with p=quarantine, start with p=reject]
  - decision: Single-line DKIM TXT record output
    rationale: Multi-line TXT records can cause DNS compatibility issues, single-line is universally supported
    alternatives: [multi-line with proper formatting]
  - decision: 0600 permissions for DKIM private keys
    rationale: Matches Phase 1 pattern for secrets, prevents unauthorized access to signing keys
    alternatives: [0400 (read-only), 0640 (group-readable)]
  - decision: Print DNS records to terminal AND save to DNS-RECORDS.md
    rationale: Terminal output for immediate viewing, file for reference and offline access, aligns with user decision from context
    alternatives: [terminal-only, file-only]
metrics:
  duration: 347s
  tasks: 2
  files: 17
  commits: 2
  completed: 2026-02-14T05:13:44Z
---

# Phase 04 Plan 01: DKIM Keys and DNS Records Summary

**One-liner:** Implemented 2048-bit DKIM key generation with quarterly rotation selectors, SPF/DKIM/DMARC/MX record generation following RFC best practices, and DNS-RECORDS.md guide with provider documentation links.

## What Was Built

Created the DNS foundation for email authentication with three core components:

### 1. DKIM Key Management (dns/dkim/)

**keygen.go** - RSA key pair generation and PEM encoding:
- `GenerateKeyPair(bits int)` - Generates 2048-bit RSA keys (enforces minimum 2048-bit per NIST 2025 recommendations)
- `SaveKeyPair()` - Saves private key with 0600 permissions, public key with 0644
- `LoadPrivateKey()` - Loads PEM-encoded private keys from disk
- `PublicKeyBase64()` - Returns base64-encoded DER public key for DNS TXT records

**selector.go** - Time-based selector naming for automated rotation:
- Format: `{prefix}-{YYYY}q{Q}` (e.g., "darkpipe-2026q1")
- `GenerateSelector()` - Creates selector for any time period
- `GetCurrentSelector()` - Returns selector for current quarter
- `GetNextSelector()` - Returns selector for next quarter (rotation planning)
- `ShouldRotate()` - Detects quarter boundaries for triggering rotation

**signer.go** - DKIM message signing wrapper:
- Uses emersion/go-msgauth with relaxed/relaxed canonicalization
- Signs critical headers: From, To, Subject, Date, Message-ID, MIME-Version, Content-Type
- Returns complete signed message with DKIM-Signature header

### 2. DNS Record Generation (dns/records/)

**spf.go** - SPF record builder:
- Uses explicit `ip4:` mechanism (not `include:`) to avoid DNS lookup limit
- Always terminates with `-all` (hard fail) for maximum protection
- Supports multiple IP addresses

**dkim.go** - DKIM TXT record builder:
- Format: `v=DKIM1; k=rsa; p={base64PublicKey}`
- Single-line output (no newlines) for DNS compatibility
- Subdomain format: `{selector}._domainkey.{domain}`

**dmarc.go** - DMARC record builder:
- Includes `sp=` tag for subdomain protection (pitfall #5)
- Configurable policy: none/quarantine/reject
- Optional aggregate (rua=) and forensic (ruf=) reporting
- Percentage-based rollout support (pct=)

**mx.go** - MX record builder:
- Points to cloud relay FQDN
- Default priority 10 (standard for primary mail server)

**guide.go** - DNS-RECORDS.md guide generation:
- Markdown output with all four record types
- Explanations of what each record does and why it matters
- Copy-paste values formatted for DNS providers
- Links to Cloudflare, Route53, GoDaddy, Namecheap, Google Domains documentation
- Verification section with built-in validator command and external tools (MXToolbox, mail-tester.com)
- DMARC policy progression guide (none -> quarantine -> reject)
- `PrintRecords()` - Colored terminal checklist output
- `PrintJSON()` - Machine-readable JSON format
- `SaveGuide()` - Writes guide to disk

### 3. Configuration (dns/config/)

**config.go** - Environment-based configuration following 12-factor pattern:
- Required: `DARKPIPE_DOMAIN`, `RELAY_HOSTNAME`, `RELAY_IP`
- DKIM settings: `DKIM_KEY_BITS`, `DKIM_KEY_DIR`, `DKIM_SELECTOR_PREFIX`
- DMARC settings: `DMARC_POLICY`, `DMARC_RUA`, `DMARC_RUF`
- Output settings: `OUTPUT_FORMAT` (text/json), `DNS_RECORDS_FILE`
- Validation settings: `DNS_PROPAGATION_TIMEOUT`, `DNS_SERVERS`

## Test Coverage

All packages have comprehensive test coverage:

- **keygen_test.go** - Key generation at 2048/4096 bits, <2048 rejection, save/load roundtrip, base64 encoding
- **selector_test.go** - Quarter calculation (Q1-Q4), rotation detection, next quarter logic
- **signer_test.go** - DKIM signature creation, header inclusion verification, minimal message handling
- **spf_test.go** - Single/multiple IP formatting, v=spf1 prefix, -all suffix
- **dkim_test.go** - Subdomain format, single-line output, v=DKIM1 structure
- **dmarc_test.go** - Policy variations, sp= tag presence, rua/ruf formatting
- **mx_test.go** - Default/custom priority, hostname formatting
- **guide_test.go** - All record types present, provider links, verification section, JSON validity

All tests pass: `go test ./dns/... -v` ✓
No vet issues: `go vet ./dns/...` ✓
All packages build: `go build ./dns/...` ✓

## Deviations from Plan

None - plan executed exactly as written. All must-haves delivered:

✓ 2048-bit RSA DKIM key pair can be generated and saved to disk in PEM format
✓ SPF record is generated with the cloud relay IP and correct -all qualifier
✓ DKIM TXT record is generated with the public key in base64 format under the correct selector._domainkey.domain name
✓ DMARC record is generated with configurable policy (none/quarantine/reject) and rua/ruf reporting addresses
✓ MX record is generated pointing to the cloud relay FQDN
✓ Time-based DKIM selector follows {prefix}-{YYYY}q{Q} format for automated rotation
✓ All DNS records are printed to terminal AND saved to a DNS-RECORDS.md file with explanations
✓ DKIM signing wraps emersion/go-msgauth and signs email messages with the generated private key

## Dependencies Added

- **emersion/go-msgauth v0.7.0** - RFC-compliant DKIM signing (RFC 6376), DMARC lookup (RFC 7489), Authentication-Results parsing (RFC 8601)
- **fatih/color v1.18.0** - Terminal color output for human-friendly checklists

## Next Steps

This plan provides the foundational DNS record generation. Upcoming plans will:

- **Plan 04-02**: DNS provider API integration (Cloudflare, Route53) with auto-detection via NS lookup and dry-run mode
- **Plan 04-03**: DNS validation tooling with propagation polling, PTR verification, and end-to-end email authentication test

## Technical Notes

### Key Generation Security
- Enforces 2048-bit minimum (rejects <2048 per NIST 2025 recommendations)
- Uses `crypto/rand.Reader` for secure random number generation
- Private keys saved with 0600 permissions (matches Phase 1 pattern)

### Selector Rotation Strategy
- Quarterly rotation window (Q1-Q4)
- Time-based naming makes rotation window obvious
- Multiple selectors can coexist during transition (7-day grace period recommended)
- Rotation workflow: publish new selector → wait for propagation → switch signing → wait 7 days → remove old selector

### SPF Best Practices
- Uses explicit `ip4:` mechanism to avoid DNS lookup limit
- `ip4:` doesn't count toward RFC 7208's 10-lookup limit
- Hard fail (`-all`) for maximum protection
- No `include:` mechanisms in default config (users can add if needed)

### DMARC Protection
- Includes `sp=` tag for subdomain protection (prevents subdomain spoofing)
- Defaults to `p=none` for monitoring (recommended progression: none → quarantine → reject)
- Supports percentage-based rollout (`pct=`) for gradual enforcement
- Optional aggregate (rua=) and forensic (ruf=) reporting

### DKIM Signature
- Relaxed/relaxed canonicalization (most compatible, RFC 6376 default)
- Signs critical headers: From, To, Subject, Date, Message-ID, MIME-Version, Content-Type
- Single-line TXT record output for DNS compatibility

## Self-Check: PASSED

**Files created:**
```
dns/dkim/keygen.go - FOUND
dns/dkim/keygen_test.go - FOUND
dns/dkim/selector.go - FOUND
dns/dkim/selector_test.go - FOUND
dns/dkim/signer.go - FOUND
dns/dkim/signer_test.go - FOUND
dns/records/spf.go - FOUND
dns/records/spf_test.go - FOUND
dns/records/dkim.go - FOUND
dns/records/dkim_test.go - FOUND
dns/records/dmarc.go - FOUND
dns/records/dmarc_test.go - FOUND
dns/records/mx.go - FOUND
dns/records/mx_test.go - FOUND
dns/records/guide.go - FOUND
dns/records/guide_test.go - FOUND
dns/config/config.go - FOUND
```

**Commits:**
```
ac6ac4d - feat(04-01): implement DKIM key generation, selector naming, and signing - FOUND
7cb96ea - feat(04-01): implement SPF/DKIM/DMARC/MX record generation and DNS guide - FOUND
```

All files and commits verified successfully.
