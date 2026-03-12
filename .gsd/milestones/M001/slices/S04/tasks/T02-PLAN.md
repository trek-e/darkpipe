# T02: 04-dns-email-auth 02

**Slice:** S04 — **Milestone:** M001

## Description

Implement DNS provider API integration for automated record creation via Cloudflare and Route53, with auto-detection, dry-run safety, and propagation polling.

Purpose: Automated DNS record creation eliminates the most error-prone step in email server setup. Auto-detection means users don't need to know or specify their DNS provider. Dry-run by default prevents accidental DNS modifications. Propagation polling prevents premature validation failures.

Output: Go packages under dns/provider/ with Cloudflare and Route53 implementations, provider auto-detection, dry-run wrapper, and propagation checker.

## Must-Haves

- [ ] "DNS provider is auto-detected from NS records without user specifying provider manually"
- [ ] "Cloudflare DNS records are created via official cloudflare-go v6 SDK using CLOUDFLARE_API_TOKEN env var"
- [ ] "Route53 DNS records are created via official aws-sdk-go-v2 using AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars"
- [ ] "Dry-run mode is default -- dns-setup shows what WOULD be created without --apply flag"
- [ ] "After record creation, propagation is polled across multiple public DNS servers with configurable timeout"
- [ ] "Unknown DNS providers fall back to manual guide output (no API calls, no error)"
- [ ] "DNSProvider interface allows community contributors to add new providers"

## Files

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
