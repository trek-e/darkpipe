# T02: 08-device-profiles-client-setup 02

**Slice:** S08 — **Milestone:** M001

## Description

Build the QR code system, profile HTTP server, Caddy autodiscovery routes, and DNS SRV record generation.

Purpose: Enables the two main device onboarding flows: (1) QR code scan from webmail/CLI triggers single-use token redemption and personalized profile download, and (2) desktop mail clients auto-discover settings via Thunderbird autoconfig and Outlook autodiscover endpoints served from the cloud relay. Also extends the Phase 4 DNS tool with RFC 6186 SRV records for universal email client autodiscovery.

Output: Working profile HTTP server on home device, Caddy routes on cloud relay, QR code generation library, and SRV record generation in DNS tool.

## Must-Haves

- [ ] "QR codes encode a single-use URL that links to a profile download endpoint"
- [ ] "Scanning a QR code and visiting the URL downloads a personalized .mobileconfig or redirects to setup instructions"
- [ ] "QR code tokens are single-use -- once redeemed, the same token cannot be used again"
- [ ] "QR code tokens expire after 15 minutes"
- [ ] "Caddy serves autoconfig and autodiscover XML from cloud relay for Thunderbird and Outlook auto-discovery"
- [ ] "SRV records for RFC 6186 email autodiscovery can be generated and deployed via darkpipe dns-setup"
- [ ] "Profile server exposes HTTP endpoints for profile download, autoconfig, and autodiscover on the home device"

## Files

- `home-device/profiles/pkg/qrcode/generator.go`
- `home-device/profiles/pkg/qrcode/generator_test.go`
- `home-device/profiles/pkg/qrcode/token.go`
- `home-device/profiles/pkg/qrcode/token_test.go`
- `home-device/profiles/cmd/profile-server/main.go`
- `home-device/profiles/cmd/profile-server/handlers.go`
- `home-device/profiles/cmd/profile-server/handlers_test.go`
- `cloud-relay/caddy/Caddyfile`
- `dns/records/srv.go`
- `dns/records/srv_test.go`
- `dns/cmd/dns-setup/main.go`
