# T01: 04-dns-email-auth 01

**Slice:** S04 — **Milestone:** M001

## Description

Generate DKIM keys, build SPF/DKIM/DMARC/MX DNS records, implement DKIM message signing, and produce human-readable DNS setup guides for manual configuration.

Purpose: This plan creates the email authentication foundation. Without correct DNS records and DKIM signing, emails from DarkPipe will fail authentication at Gmail, Outlook, and Yahoo. This plan generates all required records and provides copy-paste templates for users whose DNS providers are not supported by the API integration (Plan 04-02).

Output: Go packages under dns/dkim/, dns/records/, and dns/config/ with full test coverage. DNS-RECORDS.md guide generation capability.

## Must-Haves

- [ ] "2048-bit RSA DKIM key pair can be generated and saved to disk in PEM format"
- [ ] "SPF record is generated with the cloud relay IP and correct -all qualifier"
- [ ] "DKIM TXT record is generated with the public key in base64 format under the correct selector._domainkey.domain name"
- [ ] "DMARC record is generated with configurable policy (none/quarantine/reject) and rua/ruf reporting addresses"
- [ ] "MX record is generated pointing to the cloud relay FQDN"
- [ ] "Time-based DKIM selector follows {prefix}-{YYYY}q{Q} format for automated rotation"
- [ ] "All DNS records are printed to terminal AND saved to a DNS-RECORDS.md file with explanations"
- [ ] "DKIM signing wraps emersion/go-msgauth and signs email messages with the generated private key"

## Files

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
