# T03: 04-dns-email-auth 03

**Slice:** S04 — **Milestone:** M001

## Description

Build the DNS validation checker, PTR verification, end-to-end email auth test, and unified CLI that ties all Phase 4 components together into the `darkpipe dns-setup` command.

Purpose: Validation is the proof that everything works. The checker confirms DNS records are published correctly. PTR verification ensures reverse DNS is configured (required for deliverability). The email auth test sends a real message and parses Authentication-Results headers to confirm SPF/DKIM/DMARC pass at the receiver. The CLI is the single entry point for the entire DNS/auth workflow.

Output: Go packages under dns/validator/, dns/authtest/, and dns/cmd/dns-setup/ with the complete CLI workflow.

## Must-Haves

- [ ] "DNS validation checker queries live DNS and reports pass/fail for SPF, DKIM, DMARC, MX, and PTR records"
- [ ] "PTR verification confirms forward and reverse DNS match for the cloud relay IP"
- [ ] "Single darkpipe dns-setup command generates DKIM keys, creates/verifies DNS records"
- [ ] "darkpipe dns-setup --validate-only runs standalone validation without modifying DNS"
- [ ] "darkpipe dns-setup --json outputs machine-readable JSON results"
- [ ] "Test email sender and Authentication-Results header parser verify end-to-end DKIM/SPF/DMARC authentication"
- [ ] "Human-friendly colored checklist output shows pass/fail status for each record type"

## Files

- `dns/validator/checker.go`
- `dns/validator/checker_test.go`
- `dns/validator/ptr.go`
- `dns/validator/ptr_test.go`
- `dns/authtest/sender.go`
- `dns/authtest/parser.go`
- `dns/authtest/parser_test.go`
- `dns/cmd/dns-setup/main.go`
