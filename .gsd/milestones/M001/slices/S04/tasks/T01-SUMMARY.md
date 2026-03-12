---
id: T01
parent: S04
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
# T01: 04-dns-email-auth 01

**# Phase 04 Plan 01: DKIM Keys and DNS Records Summary**

## What Happened

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
