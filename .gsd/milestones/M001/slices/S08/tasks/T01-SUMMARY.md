---
id: T01
parent: S08
milestone: M001
provides:
  - App password generation with crypto/rand in XXXX-XXXX-XXXX-XXXX format
  - Bcrypt password hashing and verification
  - Store interface for Stalwart (REST API), Dovecot (JSON file), and Maddy (JSON file)
  - Apple .mobileconfig profile generator with Email/CalDAV/CardDAV payloads
  - Mozilla/Thunderbird autoconfig XML generator (v1.1 spec)
  - Microsoft Outlook autodiscover XML generator (POX protocol)
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 6.1min
verification_result: passed
completed_at: 2026-02-14
blocker_discovered: false
---
# T01: 08-device-profiles-client-setup 01

**# Phase 08 Plan 01: App Password & Profile Generation Core Summary**

## What Happened

# Phase 08 Plan 01: App Password & Profile Generation Core Summary

**Crypto-secure app passwords with bcrypt hashing, three mail server backends (Stalwart REST/Dovecot JSON/Maddy JSON), and profile generators for Apple (.mobileconfig), Mozilla (autoconfig), and Outlook (autodiscover)**

## Performance

- **Duration:** 6.1 min (368 seconds)
- **Started:** 2026-02-14T15:03:05Z
- **Completed:** 2026-02-14T15:09:13Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments
- App password generation produces XXXX-XXXX-XXXX-XXXX format using crypto/rand with confusing-character exclusion
- Three Store implementations (Stalwart REST API, Dovecot JSON file, Maddy JSON file) with bcrypt hashing
- Apple .mobileconfig generator with conditional Email+CalDAV+CardDAV payloads using typed structs
- Mozilla/Thunderbird autoconfig XML following v1.1 spec with IMAP/SMTP settings
- Microsoft Outlook autodiscover XML following POX protocol
- Comprehensive test coverage (100% package pass rate) with XML round-trip validation

## Task Commits

Each task was committed atomically:

1. **Task 1: App password generation, storage interface, and mail server backends** - `dfe8a40` (feat)
2. **Task 2: Profile generators (.mobileconfig, autoconfig XML, autodiscover XML)** - `611cc9d` (feat)

## Files Created/Modified

**Task 1: App Password System**
- `home-device/profiles/go.mod` - New module github.com/darkpipe/darkpipe/profiles
- `home-device/profiles/pkg/apppassword/generator.go` - GenerateAppPassword, HashPassword, VerifyPassword
- `home-device/profiles/pkg/apppassword/generator_test.go` - Tests for format, charset, uniqueness, hashing
- `home-device/profiles/pkg/apppassword/store.go` - Store interface with Create/List/Revoke/Verify
- `home-device/profiles/pkg/apppassword/stalwart.go` - Stalwart REST API backend with $app$ format
- `home-device/profiles/pkg/apppassword/dovecot.go` - JSON file backend with flock concurrency
- `home-device/profiles/pkg/apppassword/maddy.go` - JSON file backend with flock concurrency

**Task 2: Profile Generators**
- `home-device/profiles/pkg/mobileconfig/payloads.go` - Typed structs for Email/CalDAV/CardDAV payloads
- `home-device/profiles/pkg/mobileconfig/generator.go` - ProfileGenerator.GenerateProfile
- `home-device/profiles/pkg/mobileconfig/generator_test.go` - Tests with plist round-trip validation
- `home-device/profiles/pkg/autoconfig/autoconfig.go` - GenerateAutoconfig for Mozilla/Thunderbird
- `home-device/profiles/pkg/autoconfig/autoconfig_test.go` - Tests with XML parsing validation
- `home-device/profiles/pkg/autodiscover/autodiscover.go` - GenerateAutodiscover for Outlook
- `home-device/profiles/pkg/autodiscover/autodiscover_test.go` - Tests with XML parsing validation

## Decisions Made

**App Password Security:**
- Used crypto/rand exclusively (never math/rand) for password generation
- Charset excludes confusing characters (no 0/O/1/I) for human readability
- Bcrypt cost 12 balances security (2^12 iterations) with performance

**Backend Storage:**
- Stalwart uses REST API with $app$<device-name>$<bcrypt-hash> format (aligns with Stalwart conventions)
- Dovecot and Maddy use JSON file storage at configurable path (default /data/app-passwords.json)
- File-based stores use syscall.Flock for concurrent access safety

**Profile Generation:**
- Used micromdm/plist (renamed from groob/plist) for Apple plist serialization
- Typed structs with plist tags for compile-time safety (not map[string]interface{})
- Profiles are UNSIGNED for v1 (research shows unsigned profiles install on iOS/macOS without MDM)
- CalDAV/CardDAV payloads conditionally included when URLs provided

**Client Compatibility:**
- IMAP on 993 with SSL consistently across all formats
- SMTP on 587 with STARTTLS consistently across all formats
- Username is full email address (%EMAILADDRESS% for placeholders)
- Autoconfig and autodiscover endpoints designed as public (no auth) for maximum compatibility

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed os.Dir → filepath.Dir for directory creation**
- **Found during:** Task 1 (app password storage backends)
- **Issue:** Used os.Dir(path) which doesn't exist - should be filepath.Dir(path)
- **Fix:** Added `path/filepath` import and changed os.Dir to filepath.Dir in dovecot.go and maddy.go
- **Files modified:** home-device/profiles/pkg/apppassword/dovecot.go, home-device/profiles/pkg/apppassword/maddy.go
- **Verification:** go build succeeded after fix
- **Committed in:** dfe8a40 (part of Task 1 commit)

**2. [Rule 3 - Blocking] Updated groob/plist → micromdm/plist**
- **Found during:** Task 2 (profile generator dependencies)
- **Issue:** groob/plist module renamed to micromdm/plist, import failing with module path mismatch
- **Fix:** Updated imports in generator.go and generator_test.go to use github.com/micromdm/plist
- **Files modified:** home-device/profiles/pkg/mobileconfig/generator.go, home-device/profiles/pkg/mobileconfig/generator_test.go
- **Verification:** go mod tidy succeeded, all tests pass
- **Committed in:** 611cc9d (part of Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking dependency issue)
**Impact on plan:** Both auto-fixes necessary for code compilation. No scope creep.

## Issues Encountered

None - both deviations handled automatically via deviation rules.

## User Setup Required

None - no external service configuration required. This plan creates pure libraries for use in Plan 08-02 (HTTP server).

## Next Phase Readiness

**Ready for Plan 08-02:** All core libraries complete with comprehensive tests. Next plan will:
- Build HTTP server exposing profile endpoints
- Add QR code generation for easy mobile device onboarding
- Integrate with mail server backends for app password management

**Blockers:** None

**Concerns:** None

## Self-Check: PASSED

**Created files verified:**
- home-device/profiles/go.mod: FOUND
- home-device/profiles/pkg/apppassword/generator.go: FOUND
- home-device/profiles/pkg/mobileconfig/generator.go: FOUND
- home-device/profiles/pkg/autoconfig/autoconfig.go: FOUND
- home-device/profiles/pkg/autodiscover/autodiscover.go: FOUND

**Commits verified:**
- dfe8a40: FOUND
- 611cc9d: FOUND

**Tests verified:**
- All packages build: PASSED
- All tests pass: PASSED (16/16 tests)
- go vet clean: PASSED
- App password format validation: PASSED (XXXX-XXXX-XXXX-XXXX)

---
*Phase: 08-device-profiles-client-setup*
*Completed: 2026-02-14*
