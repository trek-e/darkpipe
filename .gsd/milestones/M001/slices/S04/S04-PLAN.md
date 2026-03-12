# S04: Dns Email Auth

**Goal:** Generate DKIM keys, build SPF/DKIM/DMARC/MX DNS records, implement DKIM message signing, and produce human-readable DNS setup guides for manual configuration.
**Demo:** Generate DKIM keys, build SPF/DKIM/DMARC/MX DNS records, implement DKIM message signing, and produce human-readable DNS setup guides for manual configuration.

## Must-Haves


## Tasks

- [x] **T01: 04-dns-email-auth 01**
  - Generate DKIM keys, build SPF/DKIM/DMARC/MX DNS records, implement DKIM message signing, and produce human-readable DNS setup guides for manual configuration.

Purpose: This plan creates the email authentication foundation. Without correct DNS records and DKIM signing, emails from DarkPipe will fail authentication at Gmail, Outlook, and Yahoo. This plan generates all required records and provides copy-paste templates for users whose DNS providers are not supported by the API integration (Plan 04-02).

Output: Go packages under dns/dkim/, dns/records/, and dns/config/ with full test coverage. DNS-RECORDS.md guide generation capability.
- [x] **T02: 04-dns-email-auth 02**
  - Implement DNS provider API integration for automated record creation via Cloudflare and Route53, with auto-detection, dry-run safety, and propagation polling.

Purpose: Automated DNS record creation eliminates the most error-prone step in email server setup. Auto-detection means users don't need to know or specify their DNS provider. Dry-run by default prevents accidental DNS modifications. Propagation polling prevents premature validation failures.

Output: Go packages under dns/provider/ with Cloudflare and Route53 implementations, provider auto-detection, dry-run wrapper, and propagation checker.
- [x] **T03: 04-dns-email-auth 03**
  - Build the DNS validation checker, PTR verification, end-to-end email auth test, and unified CLI that ties all Phase 4 components together into the `darkpipe dns-setup` command.

Purpose: Validation is the proof that everything works. The checker confirms DNS records are published correctly. PTR verification ensures reverse DNS is configured (required for deliverability). The email auth test sends a real message and parses Authentication-Results headers to confirm SPF/DKIM/DMARC pass at the receiver. The CLI is the single entry point for the entire DNS/auth workflow.

Output: Go packages under dns/validator/, dns/authtest/, and dns/cmd/dns-setup/ with the complete CLI workflow.

## Files Likely Touched

- `dns/dkim/keygen.go`
- `dns/dkim/keygen_test.go`
- `dns/dkim/selector.go`
- `dns/dkim/selector_test.go`
- `dns/dkim/signer.go`
- `dns/dkim/signer_test.go`
- `dns/records/spf.go`
- `dns/records/spf_test.go`
- `dns/records/dkim.go`
- `dns/records/dkim_test.go`
- `dns/records/dmarc.go`
- `dns/records/dmarc_test.go`
- `dns/records/mx.go`
- `dns/records/mx_test.go`
- `dns/records/guide.go`
- `dns/records/guide_test.go`
- `dns/config/config.go`
- `dns/provider/interface.go`
- `dns/provider/dryrun.go`
- `dns/provider/dryrun_test.go`
- `dns/provider/detector.go`
- `dns/provider/detector_test.go`
- `dns/provider/cloudflare/client.go`
- `dns/provider/cloudflare/client_test.go`
- `dns/provider/route53/client.go`
- `dns/provider/route53/client_test.go`
- `dns/provider/propagation.go`
- `dns/provider/propagation_test.go`
- `dns/validator/checker.go`
- `dns/validator/checker_test.go`
- `dns/validator/ptr.go`
- `dns/validator/ptr_test.go`
- `dns/authtest/sender.go`
- `dns/authtest/parser.go`
- `dns/authtest/parser_test.go`
- `dns/cmd/dns-setup/main.go`
