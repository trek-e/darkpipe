# Phase 4: DNS & Email Authentication - Context

**Gathered:** 2026-02-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Email sent from DarkPipe passes SPF, DKIM, and DMARC authentication at all major providers (Gmail, Outlook, Yahoo). DNS records are automated for supported providers (Cloudflare, Route53) or clearly documented for manual setup. A validation checker confirms all records are correct. PTR verification ensures the cloud relay IP resolves to the mail server FQDN.

</domain>

<decisions>
## Implementation Decisions

### DNS Provider Scope
- Cloudflare and Route53 only at launch. Add more providers post-v1 based on user demand.
- Go interface pattern (DNSProvider interface) so community contributors can add providers later
- Credentials via environment variables only (CLOUDFLARE_API_TOKEN, AWS_ACCESS_KEY_ID, etc.) — 12-factor, CI-friendly
- Auto-detect DNS provider via NS record lookup — no need for user to specify provider manually
- Dry-run by default — show what WOULD be created, require --apply flag to make changes
- PTR records: automate via VPS provider API where possible, verify-only for others

### DKIM Key Management
- 2048-bit RSA keys as default
- Automated key rotation: tool rotates keys on schedule, publishes new selector, waits for propagation, switches signing key

### CLI Tool Workflow
- Single `darkpipe dns-setup` command does everything: generates DKIM keys, creates/verifies DNS records
- Validation built into dns-setup (runs after apply). Add --validate-only flag for standalone checks
- Human-friendly colored checklist output by default, --json flag for machine-readable output
- DNS + headers check: after DNS validation, send test email and inspect Authentication-Results headers for end-to-end verification
- Built-in DNS propagation wait: after creating records, poll DNS until propagation confirmed (with timeout), then validate

### Manual Setup Guide
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

</decisions>

<specifics>
## Specific Ideas

- Dry-run by default is a safety requirement — never modify DNS without explicit --apply
- Auto-detection of DNS provider via NS records reduces config burden
- Propagation polling should have a reasonable timeout and inform the user during the wait
- Test email verification should check Authentication-Results headers, not rely on external services
- Manual guide should be self-contained (terminal output + file) so it works offline

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-dns-email-auth*
*Context gathered: 2026-02-13*
