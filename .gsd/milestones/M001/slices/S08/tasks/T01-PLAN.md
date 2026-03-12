# T01: 08-device-profiles-client-setup 01

**Slice:** S08 — **Milestone:** M001

## Description

Build the app password system and profile generation core libraries for DarkPipe device onboarding.

Purpose: Establishes the foundational packages that all other Phase 8 plans depend on -- app password generation/storage (PROF-05) and configuration profile generators for Apple .mobileconfig (PROF-01), Mozilla autoconfig (PROF-02/PROF-04), and Microsoft autodiscover (PROF-04). These are pure libraries with tests, no HTTP server yet.

Output: Go packages under home-device/profiles/ with comprehensive tests, ready for the HTTP server and webmail integration in subsequent plans.

## Must-Haves

- [ ] "App passwords can be generated with cryptographically secure randomness in XXXX-XXXX-XXXX-XXXX format"
- [ ] "App passwords can be stored and verified for each mail server backend (Stalwart, Maddy, Postfix+Dovecot)"
- [ ] "A .mobileconfig profile containing Email (IMAP+SMTP), CalDAV, and CardDAV payloads can be generated for any user"
- [ ] "Thunderbird autoconfig XML is generated with correct IMAP/SMTP settings for any domain"
- [ ] "Outlook autodiscover XML is generated with correct IMAP/SMTP settings for any domain"

## Files

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
