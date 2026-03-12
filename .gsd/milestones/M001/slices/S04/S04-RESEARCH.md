# Phase 4: DNS & Email Authentication - Research

**Researched:** 2026-02-13
**Domain:** DNS management APIs, DKIM/DMARC/SPF email authentication, DNS validation
**Confidence:** HIGH

## Summary

Phase 4 implements automated email authentication (SPF, DKIM, DMARC) with DNS provider integration and validation tooling. The Go ecosystem provides mature, production-ready libraries for DNS API management (official Cloudflare and AWS SDKs), DKIM signing (emersion/go-msgauth), and DNS querying (miekg/dns). The standard library's crypto/rsa handles 2048-bit key generation without external dependencies. All major components have official Go support with recent 2026 updates.

The implementation pattern follows DarkPipe's existing architecture: stdlib-first (crypto/rsa, crypto/x509, net), minimal external dependencies (official provider SDKs only), environment variable configuration, and interface-based abstraction for provider extensibility. DNS validation uses miekg/dns for querying records and emersion/go-msgauth for parsing Authentication-Results headers from test emails.

**Primary recommendation:** Use official Cloudflare (cloudflare-go v6) and AWS SDK v2 for DNS APIs, emersion/go-msgauth for DKIM/DMARC operations, miekg/dns for DNS validation queries, and stdlib crypto/rsa for key generation. Implement DNSProvider interface pattern for community extensibility.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### DNS Provider Scope
- Cloudflare and Route53 only at launch. Add more providers post-v1 based on user demand.
- Go interface pattern (DNSProvider interface) so community contributors can add providers later
- Credentials via environment variables only (CLOUDFLARE_API_TOKEN, AWS_ACCESS_KEY_ID, etc.) — 12-factor, CI-friendly
- Auto-detect DNS provider via NS record lookup — no need for user to specify provider manually
- Dry-run by default — show what WOULD be created, require --apply flag to make changes
- PTR records: automate via VPS provider API where possible, verify-only for others

#### DKIM Key Management
- 2048-bit RSA keys as default
- Automated key rotation: tool rotates keys on schedule, publishes new selector, waits for propagation, switches signing key

#### CLI Tool Workflow
- Single `darkpipe dns-setup` command does everything: generates DKIM keys, creates/verifies DNS records
- Validation built into dns-setup (runs after apply). Add --validate-only flag for standalone checks
- Human-friendly colored checklist output by default, --json flag for machine-readable output
- DNS + headers check: after DNS validation, send test email and inspect Authentication-Results headers for end-to-end verification
- Built-in DNS propagation wait: after creating records, poll DNS until propagation confirmed (with timeout), then validate

#### Manual Setup Guide
- Print DNS records to terminal AND save to a DNS-RECORDS.md file for reference
- Include record values + explanation of what each record does and why
- Link to each provider's own "how to add DNS records" documentation (no screenshots — they go stale)
- Verification section points users to both the built-in validator AND external tools (MXToolbox, mail-tester.com)

### Claude's Discretion
- DKIM selector naming convention (choose what works best with automated rotation)
- DKIM private key storage location (align with Phase 3 mail server config patterns)
- Multi-domain handling (single domain per run vs batch — align with existing home mail server multi-domain code)
- PTR automation providers (pick based on API simplicity and Phase 1 VPS provider guide)
- CLI UX mode (wizard vs one-shot — align with tiered UX requirement: simple defaults for non-technical, full control for power users)
- Whether DKIM rotation gets its own command or stays in dns-setup (align with automated rotation design)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope

</user_constraints>

## Standard Stack

### Core Libraries

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| cloudflare/cloudflare-go | v6 (latest v4.5.1) | Cloudflare DNS API client | Official Cloudflare SDK, Stainless-generated, 73 code snippets in Context7, HIGH source reputation (82.2 benchmark) |
| aws/aws-sdk-go-v2 | v2 (latest v1_39_0) | AWS Route53 DNS API client | Official AWS SDK v2, replaces deprecated v1, 1324 code snippets in Context7, HIGH source reputation (87.4 benchmark) |
| emersion/go-msgauth | Latest (v0.6.4) | DKIM signing/verification, DMARC lookup, Authentication-Results parsing | Production-ready, RFC-compliant (RFC 6376 DKIM, RFC 7489 DMARC, RFC 8601 Authentication-Results), includes dkim-keygen tool, MIT license |
| miekg/dns | v2 (published 2026-01-22) | DNS queries (TXT, MX, NS, PTR), propagation checking | Complete DNS implementation, supports all RR types including DNSSEC, most recent 2026 update, industry standard for Go DNS |
| crypto/rsa | stdlib | 2048-bit RSA key generation | Go standard library, no external deps, PKCS#1 and RFC 8017 compliant, secure key generation via crypto/rand |
| crypto/x509 | stdlib | PEM encoding/decoding, certificate parsing | Go standard library, used in existing DarkPipe PKI code (transport/pki/ca) |
| net/mail | stdlib | Email header parsing | Go standard library, RFC 5322/6532 compliant |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| mjl-/mox/iprev | Latest | PTR forward/reverse DNS verification | Implements RFC standard PTR check: reverse lookup IP → hostname, forward lookup hostname → IP, verify match |
| fatih/color | Latest | Terminal color output | Human-friendly checklist output (already used in Go CLI tools, optional dependency) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| cloudflare-go v6 | cloudflare-go v1/v2 | v6 is Stainless-generated (better API design), v1/v2 are legacy |
| aws-sdk-go-v2 | aws-sdk-go (v1) | v1 deprecated, EOL July 31, 2025 — v2 is required for future support |
| emersion/go-msgauth | toorop/go-dkim | toorop/go-dkim is DKIM-only, go-msgauth covers DKIM + DMARC + Authentication-Results parsing |
| miekg/dns | Go stdlib net package | stdlib net.LookupTXT/LookupMX work for simple cases, but miekg/dns supports full RR types, raw DNS queries, and controlled DNS server selection (required for propagation checking) |
| emersion/go-msgauth | Manual DKIM implementation | DKIM canonicalization and signature verification are complex (RFC 6376 has many edge cases) — use battle-tested library |

**Installation:**
```bash
go get github.com/cloudflare/cloudflare-go/v6
go get github.com/aws/aws-sdk-go-v2/service/route53
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/emersion/go-msgauth
go get github.com/miekg/dns
go get github.com/mjl-/mox/iprev
# stdlib packages (crypto/rsa, crypto/x509, net/mail) are included automatically
```

## Architecture Patterns

### Recommended Project Structure

Following DarkPipe's existing patterns (transport/wireguard, transport/pki, cloud-relay/relay):

```
dns/
├── provider/              # DNS provider abstraction
│   ├── interface.go       # DNSProvider interface
│   ├── cloudflare/        # Cloudflare implementation
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── detect.go      # NS record detection
│   ├── route53/           # AWS Route53 implementation
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── detect.go
│   └── detector.go        # Auto-detect provider via NS lookup
├── dkim/                  # DKIM key management
│   ├── keygen.go          # 2048-bit RSA key generation
│   ├── keygen_test.go
│   ├── signer.go          # DKIM signing (wraps emersion/go-msgauth)
│   ├── signer_test.go
│   └── selector.go        # Selector naming and rotation
├── records/               # DNS record generation
│   ├── spf.go             # SPF record builder
│   ├── spf_test.go
│   ├── dkim.go            # DKIM TXT record builder
│   ├── dkim_test.go
│   ├── dmarc.go           # DMARC record builder
│   ├── dmarc_test.go
│   ├── mx.go              # MX record builder
│   └── mx_test.go
├── validator/             # DNS validation and propagation
│   ├── checker.go         # Query DNS records via miekg/dns
│   ├── checker_test.go
│   ├── propagation.go     # Poll DNS until propagated
│   ├── propagation_test.go
│   ├── ptr.go             # PTR verification via mjl-/mox/iprev
│   └── ptr_test.go
├── authtest/              # End-to-end email auth test
│   ├── sender.go          # Send test email via cloud relay
│   ├── parser.go          # Parse Authentication-Results headers
│   └── parser_test.go
└── cmd/
    └── dns-setup/
        └── main.go        # CLI entry point
```

### Pattern 1: DNSProvider Interface

**What:** Provider abstraction allows Cloudflare, Route53, and future providers to implement a common interface.

**When to use:** Required for provider auto-detection and extensibility (community can add providers post-v1).

**Example:**
```go
// Source: DarkPipe pattern aligned with transport/wireguard vs transport/mtls abstraction

package provider

import "context"

// DNSProvider abstracts DNS API operations across different providers.
type DNSProvider interface {
	// CreateRecord creates a DNS record (A, TXT, MX, etc.)
	CreateRecord(ctx context.Context, rec Record) error

	// UpdateRecord updates an existing DNS record by ID
	UpdateRecord(ctx context.Context, recordID string, rec Record) error

	// ListRecords lists DNS records, optionally filtered by type and name
	ListRecords(ctx context.Context, filter RecordFilter) ([]Record, error)

	// DeleteRecord deletes a DNS record by ID
	DeleteRecord(ctx context.Context, recordID string) error

	// GetZoneID returns the zone/hosted zone ID for a domain
	GetZoneID(ctx context.Context, domain string) (string, error)
}

// Record represents a DNS record (provider-agnostic).
type Record struct {
	Type    string // "A", "TXT", "MX", "CNAME"
	Name    string // "example.com" or "mail.example.com"
	Content string // IP, text value, hostname
	TTL     int    // seconds
	Priority *int  // MX/SRV priority (nil for other types)
}

// RecordFilter filters DNS record queries.
type RecordFilter struct {
	Type string // empty = all types
	Name string // empty = all names
}
```

### Pattern 2: Configuration via Environment Variables

**What:** Follow 12-factor app pattern (existing pattern in cloud-relay/relay/config).

**When to use:** All credential loading, all configurable parameters.

**Example:**
```go
// Source: Aligned with cloud-relay/relay/config/config.go pattern

package config

import (
	"fmt"
	"os"
	"time"
)

type DNSConfig struct {
	// Provider credentials (auto-detected from available env vars)
	CloudflareAPIToken string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string

	// DKIM settings
	DKIMKeyBits        int
	DKIMKeyDir         string
	DKIMSelectorPrefix string

	// Validation settings
	PropagationTimeout time.Duration
	DNSServers         []string // for validation queries

	// Output
	OutputFormat       string // "text" or "json"
	RecordsFile        string // path to DNS-RECORDS.md
}

func LoadFromEnv() (*DNSConfig, error) {
	cfg := &DNSConfig{
		CloudflareAPIToken: getEnv("CLOUDFLARE_API_TOKEN", ""),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		DKIMKeyBits:        getEnvInt("DKIM_KEY_BITS", 2048),
		DKIMKeyDir:         getEnv("DKIM_KEY_DIR", "/etc/darkpipe/dkim"),
		DKIMSelectorPrefix: getEnv("DKIM_SELECTOR_PREFIX", "darkpipe"),
		PropagationTimeout: getEnvDuration("DNS_PROPAGATION_TIMEOUT", 5*time.Minute),
		DNSServers:         getEnvSlice("DNS_SERVERS", []string{"8.8.8.8:53", "1.1.1.1:53"}),
		OutputFormat:       getEnv("OUTPUT_FORMAT", "text"),
		RecordsFile:        getEnv("DNS_RECORDS_FILE", "DNS-RECORDS.md"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ... getEnvInt, getEnvDuration, getEnvSlice helpers ...
```

### Pattern 3: Dry-Run by Default

**What:** Never mutate DNS without explicit --apply flag (safety requirement from CONTEXT.md).

**When to use:** All DNS write operations.

**Example:**
```go
// Source: Safety pattern for DNS mutations

package provider

import (
	"context"
	"fmt"
)

type DryRunProvider struct {
	underlying DNSProvider
	dryRun     bool
}

func NewDryRunProvider(underlying DNSProvider, dryRun bool) *DryRunProvider {
	return &DryRunProvider{
		underlying: underlying,
		dryRun:     dryRun,
	}
}

func (p *DryRunProvider) CreateRecord(ctx context.Context, rec Record) error {
	if p.dryRun {
		fmt.Printf("[DRY RUN] Would create %s record: %s -> %s\n", rec.Type, rec.Name, rec.Content)
		return nil
	}
	return p.underlying.CreateRecord(ctx, rec)
}

func (p *DryRunProvider) UpdateRecord(ctx context.Context, recordID string, rec Record) error {
	if p.dryRun {
		fmt.Printf("[DRY RUN] Would update record %s: %s -> %s\n", recordID, rec.Name, rec.Content)
		return nil
	}
	return p.underlying.UpdateRecord(ctx, recordID, rec)
}

// DeleteRecord, ListRecords (read-only, no dry-run check), GetZoneID (read-only)...
```

### Pattern 4: Propagation Polling with Timeout

**What:** After creating DNS records, poll multiple public DNS servers until record propagates or timeout.

**When to use:** After DNS record creation, before validation step.

**Example:**
```go
// Source: DNS propagation best practices 2026

package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// WaitForPropagation polls DNS servers until record propagates or timeout.
func WaitForPropagation(ctx context.Context, recordName, recordType, expectedValue string, timeout time.Duration) error {
	servers := []string{"8.8.8.8:53", "1.1.1.1:53", "208.67.222.222:53"} // Google, Cloudflare, OpenDNS
	deadline := time.Now().Add(timeout)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("propagation timeout after %s", timeout)
			}

			allPropagated := true
			for _, server := range servers {
				propagated, err := checkServer(recordName, recordType, expectedValue, server)
				if err != nil || !propagated {
					allPropagated = false
					break
				}
			}

			if allPropagated {
				return nil // Success
			}

			fmt.Printf("Waiting for DNS propagation...\n")
		}
	}
}

func checkServer(name, recordType, expectedValue, server string) (bool, error) {
	c := new(dns.Client)
	m := new(dns.Msg)

	// Convert recordType string to dns.Type constant
	var qtype uint16
	switch recordType {
	case "TXT":
		qtype = dns.TypeTXT
	case "MX":
		qtype = dns.TypeMX
	case "A":
		qtype = dns.TypeA
	default:
		return false, fmt.Errorf("unsupported record type: %s", recordType)
	}

	m.SetQuestion(dns.Fqdn(name), qtype)

	r, _, err := c.Exchange(m, server)
	if err != nil {
		return false, err
	}

	for _, ans := range r.Answer {
		switch recordType {
		case "TXT":
			if txt, ok := ans.(*dns.TXT); ok {
				for _, s := range txt.Txt {
					if s == expectedValue {
						return true, nil
					}
				}
			}
		case "MX":
			if mx, ok := ans.(*dns.MX); ok {
				if mx.Mx == dns.Fqdn(expectedValue) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
```

### Pattern 5: DKIM Selector Rotation

**What:** Time-based selector naming enables key rotation without downtime.

**When to use:** DKIM key generation and rotation scheduling.

**Recommended convention:** `{prefix}-{YYYYQQ}` format (e.g., `darkpipe-2026q1`).

**Why:** Based on 2026 best practices research:
- Quarterly rotation is recommended minimum
- Time-based selectors make rotation window obvious
- Multiple selectors can coexist during transition (old + new)
- Prefix allows per-domain or per-environment identification

**Example:**
```go
// Source: DKIM selector best practices 2026

package dkim

import (
	"fmt"
	"time"
)

// GenerateSelector creates a time-based DKIM selector name.
// Format: {prefix}-{YYYY}q{Q} (e.g., "darkpipe-2026q1")
func GenerateSelector(prefix string, t time.Time) string {
	year := t.Year()
	quarter := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%s-%dq%d", prefix, year, quarter)
}

// GetCurrentSelector returns the selector for the current quarter.
func GetCurrentSelector(prefix string) string {
	return GenerateSelector(prefix, time.Now())
}

// GetNextSelector returns the selector for the next quarter (for rotation).
func GetNextSelector(prefix string) string {
	nextQuarter := time.Now().AddDate(0, 3, 0)
	return GenerateSelector(prefix, nextQuarter)
}

// ShouldRotate returns true if current time is within rotation window
// (first month of quarter = rotation time).
func ShouldRotate() bool {
	month := time.Now().Month()
	return month == 1 || month == 4 || month == 7 || month == 10
}
```

### Anti-Patterns to Avoid

- **Don't parse DNS records with regex** — Use miekg/dns structured parsing (handles edge cases like multiline TXT records, quoted strings)
- **Don't implement custom DKIM canonicalization** — RFC 6376 has subtle edge cases (whitespace handling, header folding) that emersion/go-msgauth handles correctly
- **Don't skip propagation checks** — DNS changes aren't instant; polling prevents false validation failures
- **Don't hardcode DNS servers** — Allow environment variable override for private DNS or testing
- **Don't store credentials in code** — Use environment variables only (12-factor principle)
- **Don't create multiple SPF records** — Multiple SPF TXT records for same domain breaks authentication

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| DKIM signing | Custom RFC 6376 implementation | emersion/go-msgauth | DKIM has complex canonicalization rules (relaxed/simple headers/body), signature generation, and edge cases. Library is RFC-compliant and battle-tested. |
| DNS queries | Custom UDP DNS packets | miekg/dns | DNS protocol has many edge cases (compression, DNSSEC, EDNS0). miekg/dns is complete implementation used by production DNS tools. |
| SPF record validation | Custom SPF lookup limit checker | DNS query via miekg/dns + count includes | SPF has 10 DNS lookup limit (RFC 7208). Correctly counting "include:" mechanisms requires recursive resolution. Easy to miss nested includes. |
| Authentication-Results parsing | Regex on email headers | emersion/go-msgauth or net/mail | RFC 8601 header syntax has optional versions, method-specific tokens, and property escaping. Structured parser avoids regex fragility. |
| PTR verification | Manual reverse + forward DNS | mjl-/mox/iprev | RFC-standard PTR check is: reverse lookup IP → names, forward lookup each name → IPs, verify original IP in results. mjl-/mox/iprev implements this correctly. |
| RSA key generation | Custom crypto | crypto/rsa stdlib | Secure random number generation (crypto/rand), proper key size validation, and PKCS#1 encoding are critical. Stdlib is audited and maintained. |
| PEM encoding | Custom base64 + headers | crypto/x509 stdlib | PEM has specific line length (64 chars), header/footer format, and encoding rules. Stdlib handles correctly. |

**Key insight:** Email authentication is standards-heavy (RFC 6376 DKIM, RFC 7208 SPF, RFC 7489 DMARC, RFC 8601 Authentication-Results). Hand-rolling parsers/generators leads to subtle non-compliance that causes auth failures at major providers. Use libraries that explicitly state RFC compliance.

## Common Pitfalls

### Pitfall 1: SPF Record Exceeds 10 DNS Lookup Limit

**What goes wrong:** SPF record validation fails because "include:" mechanisms trigger too many DNS lookups (RFC 7208 enforces 10 lookup limit).

**Why it happens:** Each "include:" counts as one lookup, and nested includes count recursively. Easy to exceed limit when including multiple third-party services (Google Workspace, Mailchimp, etc.).

**How to avoid:**
- Generate SPF record with explicit IP addresses when possible (use `ip4:` mechanism instead of `include:`)
- Document current lookup count in SPF record comments
- Validate SPF record with miekg/dns to count lookups before publishing

**Warning signs:**
- SPF validation fails with "PermError" in Authentication-Results headers
- Third-party SPF validators report "too many DNS lookups"

### Pitfall 2: DKIM Key Length Less Than 2048 Bits

**What goes wrong:** 1024-bit DKIM keys are no longer secure (attackers can break them as of 2026). Major email providers may reject or flag emails.

**Why it happens:** Older tutorials and tools default to 1024-bit keys for "compatibility."

**How to avoid:**
- Always generate 2048-bit keys (crypto/rsa.GenerateKey(rand.Reader, 2048))
- Validate key length in tests
- NIST recommends 2048-bit minimum as of 2026

**Warning signs:**
- DKIM signature verification succeeds but email still goes to spam
- Security audits flag weak key length

### Pitfall 3: DKIM TXT Record Formatting Errors

**What goes wrong:** DNS TXT record contains extra spaces, line breaks, or syntax errors causing DKIM validation to fail.

**Why it happens:** Manual copy-paste from terminal output can introduce whitespace. Some DNS providers split long TXT records into multiple strings incorrectly.

**How to avoid:**
- Generate TXT record value as single-line base64 encoded public key
- Validate TXT record immediately after publishing via DNS query
- Include record format in DNS-RECORDS.md with exact copy-paste instructions

**Warning signs:**
- DKIM validation shows "neutral" or "fail" in Authentication-Results
- DNS query returns TXT record but DKIM verification fails

### Pitfall 4: DMARC Policy Stuck at p=none

**What goes wrong:** DMARC is published but never enforced (p=none means monitor-only). 75-80% of domains with DMARC records are stuck at p=none for months/years.

**Why it happens:** Lack of DMARC report monitoring or fear of breaking legitimate email.

**How to avoid:**
- Start with p=none for testing, but include migration timeline in documentation
- Provide clear instructions for monitoring DMARC aggregate reports (rua= tag)
- Recommend p=quarantine after verification, then p=reject for full enforcement

**Warning signs:**
- DMARC record exists but no enforcement action taken by receivers
- Spoofed emails from your domain still reach inboxes

### Pitfall 5: Ignoring Subdomain DMARC Protection

**What goes wrong:** DMARC policy applies only to primary domain, leaving subdomains vulnerable to spoofing.

**Why it happens:** DMARC "sp=" tag (subdomain policy) is optional and often omitted.

**How to avoid:**
- Include `sp=quarantine` or `sp=reject` in DMARC record generation
- Document subdomain policy in DNS-RECORDS.md

**Warning signs:**
- Spoofed emails from subdomain@example.com bypass DMARC checks

### Pitfall 6: DNS Propagation Timeout Too Short

**What goes wrong:** Validation fails because DNS records haven't propagated globally yet (TTL-dependent).

**Why it happens:** DNS propagation can take 5-10 minutes depending on TTL and global DNS server refresh cycles. Default timeout too aggressive.

**How to avoid:**
- Set propagation timeout to 5 minutes minimum (configurable via env var)
- Lower TTL to 300 seconds (5 min) before making changes, increase after
- Poll multiple public DNS servers (Google 8.8.8.8, Cloudflare 1.1.1.1, OpenDNS 208.67.222.222)

**Warning signs:**
- Validation fails immediately after DNS record creation
- Record visible in provider console but DNS queries return NXDOMAIN

### Pitfall 7: PTR Record Cannot Be Automated (VPS Provider Limitation)

**What goes wrong:** Tool tries to automate PTR record creation but VPS provider doesn't offer API.

**Why it happens:** PTR records are in reverse DNS zones controlled by IP address owner (VPS provider), not forward DNS zones. Not all providers expose reverse DNS API.

**How to avoid:**
- Implement PTR automation only for providers with API (DigitalOcean, Vultr, Linode)
- For others, provide verify-only mode: check PTR exists, document manual setup steps
- Phase 1 VPS provider guide should note PTR API availability

**Warning signs:**
- PTR automation fails with "access denied" or "zone not found"
- VPS provider documentation shows PTR setup via support ticket only

### Pitfall 8: Multiple SPF Records Break Authentication

**What goes wrong:** Domain has multiple TXT records starting with "v=spf1" causing all SPF checks to fail.

**Why it happens:** Adding new SPF record instead of updating existing one (especially when migrating providers).

**How to avoid:**
- Before creating SPF record, query existing TXT records for domain
- If SPF record exists, update it instead of creating new one
- Validate post-deployment: ensure only ONE SPF record exists

**Warning signs:**
- All emails fail SPF check suddenly after DNS change
- DNS query shows multiple "v=spf1" TXT records

### Pitfall 9: DKIM Rotation Without Overlap Period

**What goes wrong:** Old DKIM selector removed immediately when switching to new selector, causing in-flight emails to fail verification.

**Why it happens:** Eager cleanup after rotation without grace period.

**How to avoid:**
- Publish new DKIM public key (new selector) FIRST
- Wait for propagation (5+ minutes)
- Update mail server to sign with new selector
- Wait 7 days (email can be delayed in recipient queues)
- Only then remove old selector DNS record

**Warning signs:**
- DKIM verification failures spike immediately after rotation
- Delayed emails (queued for retry) fail DKIM check

## Code Examples

Verified patterns from official sources and research:

### Cloudflare DNS Record Creation

```go
// Source: https://context7.com/cloudflare/cloudflare-go/llms.txt

package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func main() {
	client := cloudflare.NewClient(
		option.WithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN")),
	)

	zoneID := "023e105f4ecef8ad9ca31a8372d0c353" // Get via client.Zones.List()

	// Create TXT record for DKIM
	record, err := client.DNS.Records.New(context.TODO(), dns.RecordNewParams{
		ZoneID:  cloudflare.F(zoneID),
		Type:    cloudflare.F(dns.RecordTypeTXT),
		Name:    cloudflare.F("darkpipe-2026q1._domainkey.example.com"),
		Content: cloudflare.F("v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A..."),
		TTL:     cloudflare.F(int64(3600)),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("DKIM TXT record created: %s\n", record.ID)

	// Create MX record
	mxRecord, err := client.DNS.Records.New(context.TODO(), dns.RecordNewParams{
		ZoneID:   cloudflare.F(zoneID),
		Type:     cloudflare.F(dns.RecordTypeMX),
		Name:     cloudflare.F("example.com"),
		Content:  cloudflare.F("mail.example.com"),
		TTL:      cloudflare.F(int64(3600)),
		Priority: cloudflare.F(10),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("MX record created: %s\n", mxRecord.ID)
}
```

### AWS Route53 DNS Record Creation

```go
// Source: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/

package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func main() {
	// Load credentials from environment (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	client := route53.NewFromConfig(cfg)
	hostedZoneID := "/hostedzone/Z1234567890ABC" // Get via ListHostedZones

	// Create TXT record for SPF
	_, err = client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionCreate,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String("example.com"),
						Type: types.RRTypeTxt,
						TTL:  aws.Int64(3600),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("\"v=spf1 ip4:203.0.113.10 -all\"")},
						},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("SPF TXT record created")
}
```

### DKIM Key Generation (2048-bit RSA)

```go
// Source: https://pkg.go.dev/crypto/rsa

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Generate 2048-bit RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Save private key as PEM
	privateKeyFile, err := os.Create("dkim-private.pem")
	if err != nil {
		panic(err)
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		panic(err)
	}

	// Save public key as PEM
	publicKeyFile, err := os.Create("dkim-public.pem")
	if err != nil {
		panic(err)
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		panic(err)
	}

	fmt.Println("DKIM keys generated: dkim-private.pem, dkim-public.pem")
}
```

### DKIM Signing with emersion/go-msgauth

```go
// Source: https://github.com/emersion/go-msgauth

package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-msgauth/dkim"
)

func main() {
	// Load private key
	privateKeyPEM, err := os.ReadFile("dkim-private.pem")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	// Email message to sign
	message := `From: sender@example.com
To: recipient@example.com
Subject: Test Email

This is a test email.
`

	// DKIM signing options
	options := &dkim.SignOptions{
		Domain:   "example.com",
		Selector: "darkpipe-2026q1",
		Signer:   privateKey,
	}

	// Sign the email
	var signed strings.Builder
	if err := dkim.Sign(&signed, strings.NewReader(message), options); err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, strings.NewReader(signed.String()))
}
```

### DNS Validation with miekg/dns

```go
// Source: https://github.com/miekg/dns

package main

import (
	"fmt"

	"github.com/miekg/dns"
)

func main() {
	// Query TXT record for DKIM public key
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("darkpipe-2026q1._domainkey.example.com"), dns.TypeTXT)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		panic(err)
	}

	if r.Rcode != dns.RcodeSuccess {
		fmt.Printf("DNS query failed: %s\n", dns.RcodeToString[r.Rcode])
		return
	}

	for _, ans := range r.Answer {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, s := range txt.Txt {
				fmt.Printf("DKIM TXT record: %s\n", s)
			}
		}
	}

	// Query MX records
	m.SetQuestion(dns.Fqdn("example.com"), dns.TypeMX)
	r, _, err = c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		panic(err)
	}

	for _, ans := range r.Answer {
		if mx, ok := ans.(*dns.MX); ok {
			fmt.Printf("MX record: priority=%d, host=%s\n", mx.Preference, mx.Mx)
		}
	}
}
```

### NS Record Lookup for Provider Auto-Detection

```go
// Source: https://pkg.go.dev/github.com/miekg/dns

package main

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func detectDNSProvider(domain string) (string, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return "", err
	}

	for _, ans := range r.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsHost := strings.ToLower(ns.Ns)

			// Detect Cloudflare
			if strings.Contains(nsHost, "cloudflare.com") {
				return "cloudflare", nil
			}

			// Detect AWS Route53
			if strings.Contains(nsHost, "awsdns") {
				return "route53", nil
			}
		}
	}

	return "unknown", nil
}

func main() {
	provider, err := detectDNSProvider("example.com")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Detected DNS provider: %s\n", provider)
}
```

### PTR Verification with mjl-/mox/iprev

```go
// Source: https://pkg.go.dev/github.com/mjl-/mox/iprev

package main

import (
	"context"
	"fmt"
	"net"

	"github.com/mjl-/mox/iprev"
)

func main() {
	// Verify PTR record for IP address
	ip := net.ParseIP("203.0.113.10")

	// iprev.Lookup performs:
	// 1. PTR lookup: 203.0.113.10 -> mail.example.com
	// 2. Forward lookup: mail.example.com -> 203.0.113.10
	// 3. Verify original IP is in forward lookup results

	result, err := iprev.Lookup(context.Background(), ip, nil)
	if err != nil {
		fmt.Printf("PTR verification failed: %v\n", err)
		return
	}

	if result.Hostname != "" {
		fmt.Printf("PTR verified: %s -> %s\n", ip, result.Hostname)
	} else {
		fmt.Printf("PTR verification failed: no matching hostname\n")
	}
}
```

### Authentication-Results Header Parsing

```go
// Source: https://github.com/emersion/go-msgauth

package main

import (
	"fmt"
	"strings"

	"github.com/emersion/go-msgauth/authres"
)

func main() {
	// Parse Authentication-Results header from received email
	header := `mx.google.com;
		dkim=pass header.i=@example.com header.s=darkpipe-2026q1 header.b=abc123;
		spf=pass (google.com: domain of sender@example.com designates 203.0.113.10 as permitted sender) smtp.mailfrom=sender@example.com;
		dmarc=pass (p=QUARANTINE sp=QUARANTINE dis=NONE) header.from=example.com`

	results, err := authres.Parse(strings.NewReader(header))
	if err != nil {
		panic(err)
	}

	for _, result := range results.Results {
		fmt.Printf("Method: %s, Result: %s\n", result.Method, result.Result)

		// Check for DKIM pass
		if result.Method == authres.MethodDKIM && result.Result == authres.ResultPass {
			fmt.Println("✓ DKIM authentication passed")
		}

		// Check for SPF pass
		if result.Method == authres.MethodSPF && result.Result == authres.ResultPass {
			fmt.Println("✓ SPF authentication passed")
		}

		// Check for DMARC pass
		if result.Method == authres.MethodDMARC && result.Result == authres.ResultPass {
			fmt.Println("✓ DMARC authentication passed")
		}
	}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 1024-bit DKIM keys | 2048-bit minimum | NIST 2025 recommendation | Security: 1024-bit keys are cryptographically weak and can be broken |
| RFC 7601 Authentication-Results | RFC 8601 (obsoletes 7601) | May 2019 | Parsing: Updated version tokens and method-specific semantics |
| AWS SDK Go v1 | AWS SDK Go v2 | EOL July 31, 2025 | Compatibility: v1 deprecated, v2 required for future AWS features |
| SPF + DKIM optional | SPF + DKIM + DMARC required | Google/Yahoo Feb 2024, enforced 2026 | Deliverability: Bulk senders (5000+/day) must have all three or emails rejected at SMTP |
| DMARC p=none tolerance | DMARC p=quarantine/reject enforcement | 2026 enforcement | Security: p=none is monitor-only; receivers now expect p=quarantine minimum |
| Manual DNS propagation wait | Automated polling with timeout | Best practice 2026 | Reliability: Prevents false validation failures from DNS caching delays |
| Annual DKIM rotation | Quarterly rotation recommended | 2026 best practice | Security: Reduces key compromise window |

**Deprecated/outdated:**
- **cloudflare-go v1/v2**: v6 is Stainless-generated with better API design
- **aws-sdk-go (v1)**: EOL July 2025, use aws-sdk-go-v2
- **RFC 7601 Authentication-Results**: Obsoleted by RFC 8601 (use emersion/go-msgauth which implements RFC 8601)
- **1024-bit DKIM keys**: Cryptographically weak, use 2048-bit minimum
- **SPF-only authentication**: Google/Yahoo/Microsoft now require SPF + DKIM + DMARC for bulk senders

## Open Questions

1. **VPS Provider PTR API Coverage**
   - What we know: DigitalOcean, Vultr, and Linode support reverse DNS API. DigitalOcean requires support ticket for some accounts.
   - What's unclear: Which VPS providers from Phase 1 guide support fully automated PTR creation vs verify-only?
   - Recommendation: Audit Phase 1 VPS provider guide and document PTR API status per provider. Start with verify-only for all, add automation incrementally.

2. **DKIM Key Storage Location**
   - What we know: Phase 3 mail server configs are in docker-compose, Stalwart/Maddy/Postfix+Dovecot each have different config paths.
   - What's unclear: Should DKIM keys live in mail server-specific paths or shared /etc/darkpipe/dkim/ volume?
   - Recommendation: Use shared volume /etc/darkpipe/dkim/ mounted into all mail server containers. Aligns with transport/pki pattern (/etc/darkpipe/certs).

3. **Multi-Domain Batch Operation**
   - What we know: Phase 3 supports multi-domain mail server configuration.
   - What's unclear: Should dns-setup process one domain per invocation or batch-process multiple domains from config file?
   - Recommendation: Start with single-domain per run (simpler UX, easier error handling). Add batch mode in Phase 9 automation tooling if needed.

4. **DKIM Rotation Command Placement**
   - What we know: DKIM rotation requires: generate new key, publish new selector DNS record, wait for propagation, switch mail server config, grace period, remove old selector.
   - What's unclear: Should rotation be `darkpipe dns-setup --rotate` or separate `darkpipe dkim-rotate` command?
   - Recommendation: Integrate into dns-setup as `--rotate-dkim` flag. Rotation is DNS-heavy operation and benefits from existing propagation polling and validation logic.

## Sources

### Primary (HIGH confidence)

- [Cloudflare Go SDK (Context7)](https://context7.com/cloudflare/cloudflare-go) - DNS API operations, record types, code examples
- [AWS SDK for Go v2 (Context7)](https://context7.com/websites/aws_amazon_sdk-for-go_v2_developer-guide) - Route53 API patterns, hosted zone management
- [Cloudflare Go GitHub](https://github.com/cloudflare/cloudflare-go) - Official library, v6 API reference
- [AWS SDK Go v2 Route53 pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/route53) - Service client documentation
- [emersion/go-msgauth GitHub](https://github.com/emersion/go-msgauth) - DKIM/DMARC/Authentication-Results implementation, RFC compliance statements
- [miekg/dns pkg.go.dev](https://pkg.go.dev/github.com/miekg/dns) - DNS library API, published 2026-01-22
- [crypto/rsa pkg.go.dev](https://pkg.go.dev/crypto/rsa) - RSA key generation, published 2026-02-10
- [RFC 8601](https://datatracker.ietf.org/doc/html/rfc8601) - Authentication-Results header specification (obsoletes RFC 7601)
- [RFC 6376](https://datatracker.ietf.org/doc/html/rfc6376) - DKIM specification
- [RFC 7489](https://datatracker.ietf.org/doc/html/rfc7489) - DMARC specification
- [RFC 7208](https://datatracker.ietf.org/doc/html/rfc7208) - SPF specification

### Secondary (MEDIUM confidence)

- [DKIM Selector Best Practices 2026](https://smartsmssolutions.com/resources/blog/business/dkim-selector-rotate-best-practices) - Rotation timing, selector naming conventions
- [M3AAWG DKIM Key Rotation](https://www.m3aawg.org/DKIMKeyRotation) - Industry best practices for key rotation
- [DNS Propagation Best Practices](https://www.networksolutions.com/blog/what-is-dns-propagation/) - TTL management, timeout recommendations
- [Email Authentication Setup Mistakes 2026](https://www.infraforge.ai/blog/spf-dkim-dmarc-common-setup-mistakes) - Common pitfalls, SPF lookup limits, DKIM key length
- [SPF DKIM DMARC 2026 Requirements](https://www.trulyinbox.com/blog/how-to-set-up-spf-dkim-and-dmarc/) - Google/Yahoo/Microsoft enforcement policies
- [Vultr Reverse DNS Docs](https://docs.vultr.com/reference/vultr-cli/instance/reverse-dns) - PTR API commands
- [mjl-/mox/iprev pkg.go.dev](https://pkg.go.dev/github.com/mjl-/mox/iprev) - PTR verification implementation

### Tertiary (LOW confidence)

- [VPS Provider API Comparison (LowEndTalk)](https://lowendtalk.com/discussion/148673/any-popular-vps-provider-with-great-api-like-digitalocean-vultr-linode-or-hetzner) - Community discussion on API quality (unverified claims)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Official SDKs (Cloudflare, AWS) verified via Context7 and pkg.go.dev, emersion/go-msgauth and miekg/dns verified via GitHub and pkg.go.dev with recent 2026 updates
- Architecture patterns: HIGH - Patterns aligned with existing DarkPipe codebase (transport/pki, cloud-relay/relay), verified via project file inspection
- Pitfalls: MEDIUM-HIGH - 2026 enforcement requirements verified with multiple sources, common mistakes documented in industry blogs and verified against RFCs
- DKIM rotation: MEDIUM - Best practices synthesized from industry sources (M3AAWG, SmartSMS), time-based selector convention not universally standardized but widely recommended
- PTR automation: MEDIUM - VPS provider API availability verified for major providers (DigitalOcean, Vultr, Linode) but not exhaustively tested

**Research date:** 2026-02-13
**Valid until:** 2026-05-13 (90 days — DNS/email auth stack is stable, but enforcement policies and provider APIs evolve)