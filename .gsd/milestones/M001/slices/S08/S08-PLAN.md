# S08: Device Profiles Client Setup

**Goal:** Build the app password system and profile generation core libraries for DarkPipe device onboarding.
**Demo:** Build the app password system and profile generation core libraries for DarkPipe device onboarding.

## Must-Haves


## Tasks

- [x] **T01: 08-device-profiles-client-setup 01** `est:6.1min`
  - Build the app password system and profile generation core libraries for DarkPipe device onboarding.

Purpose: Establishes the foundational packages that all other Phase 8 plans depend on -- app password generation/storage (PROF-05) and configuration profile generators for Apple .mobileconfig (PROF-01), Mozilla autoconfig (PROF-02/PROF-04), and Microsoft autodiscover (PROF-04). These are pure libraries with tests, no HTTP server yet.

Output: Go packages under home-device/profiles/ with comprehensive tests, ready for the HTTP server and webmail integration in subsequent plans.
- [x] **T02: 08-device-profiles-client-setup 02** `est:8.8min`
  - Build the QR code system, profile HTTP server, Caddy autodiscovery routes, and DNS SRV record generation.

Purpose: Enables the two main device onboarding flows: (1) QR code scan from webmail/CLI triggers single-use token redemption and personalized profile download, and (2) desktop mail clients auto-discover settings via Thunderbird autoconfig and Outlook autodiscover endpoints served from the cloud relay. Also extends the Phase 4 DNS tool with RFC 6186 SRV records for universal email client autodiscovery.

Output: Working profile HTTP server on home device, Caddy routes on cloud relay, QR code generation library, and SRV record generation in DNS tool.
- [x] **T03: 08-device-profiles-client-setup 03** `est:6.3min`
  - Integrate the profile server into Docker Compose, build the webmail-accessible device management UI, add the CLI QR code command, and create the phase integration test suite.

Purpose: This is the user-facing integration plan that ties together all Phase 8 components into a complete device onboarding experience. Users interact through webmail ("Add Device" button) or CLI (`darkpipe qr`), and the infrastructure (Docker containers, Caddy routes, profile server) delivers the sub-2-minute device setup promise. The phase test suite validates all five PROF requirements.

Output: Complete device onboarding system with Docker deployment, web UI, CLI, and comprehensive test suite.

## Files Likely Touched

- `home-device/profiles/go.mod`
- `home-device/profiles/go.sum`
- `home-device/profiles/pkg/apppassword/generator.go`
- `home-device/profiles/pkg/apppassword/generator_test.go`
- `home-device/profiles/pkg/apppassword/store.go`
- `home-device/profiles/pkg/apppassword/stalwart.go`
- `home-device/profiles/pkg/apppassword/dovecot.go`
- `home-device/profiles/pkg/apppassword/maddy.go`
- `home-device/profiles/pkg/mobileconfig/generator.go`
- `home-device/profiles/pkg/mobileconfig/generator_test.go`
- `home-device/profiles/pkg/mobileconfig/payloads.go`
- `home-device/profiles/pkg/autoconfig/autoconfig.go`
- `home-device/profiles/pkg/autoconfig/autoconfig_test.go`
- `home-device/profiles/pkg/autodiscover/autodiscover.go`
- `home-device/profiles/pkg/autodiscover/autodiscover_test.go`
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
- `home-device/profiles/Dockerfile`
- `home-device/docker-compose.yml`
- `cloud-relay/docker-compose.yml`
- `home-device/profiles/cmd/profile-server/cli.go`
- `home-device/profiles/cmd/profile-server/webui.go`
- `home-device/profiles/cmd/profile-server/templates/add_device.html`
- `home-device/profiles/cmd/profile-server/templates/device_list.html`
- `home-device/profiles/cmd/profile-server/static/style.css`
- `tests/test-device-profiles.sh`
