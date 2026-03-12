---
id: T01
parent: S07
milestone: M001
provides: []
requires: []
affects: []
key_files: []
key_decisions: []
patterns_established: []
observability_surfaces: []
drill_down_paths: []
duration: 
verification_result: passed
completed_at: 
blocker_discovered: false
---
# T01: 07-build-system-deployment 01

**# Phase 07 Plan 01: Multi-Architecture Docker Images and CI/CD Workflows Summary**

## What Happened

# Phase 07 Plan 01: Multi-Architecture Docker Images and CI/CD Workflows Summary

**One-liner:** Multi-architecture Docker images (amd64/arm64) with OCI labels, setup detection, Docker secrets support, and GitHub Actions workflows for custom component selection and pre-built stack publishing.

## What Was Built

### Task 1: Multi-Architecture Dockerfiles with Optimization and Setup Detection

Updated all Dockerfiles for multi-architecture builds with size optimization and setup detection:

**Cloud Relay (cloud-relay/Dockerfile):**
- Added `ARG TARGETARCH` for automatic architecture detection via buildx
- Added `ARG VERSION=dev` for build-time version injection
- Updated Go build command: `GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w"`
- Added OCI labels: source, version, description, licenses
- Changed health check from `postfix status` to `nc -z localhost 25` (more reliable for orchestrators)
- Added netcat-openbsd package for health checks

**Home Device Dockerfiles:**
- **Stalwart:** Added OCI labels, VERSION arg, created entrypoint-wrapper.sh for setup detection and secrets
- **Maddy:** Added OCI labels, VERSION arg, created entrypoint-wrapper.sh for setup detection and secrets
- **Postfix+Dovecot:** Added TARGETARCH and VERSION args, OCI labels, updated COPY paths to use home-device/ prefix

**Entrypoint Updates:**
All entrypoints now include:
1. **Setup detection:** Checks `/config/.darkpipe-configured` and loads `/config/{relay,home}.env` if present
2. **Docker secrets support:** Reads `{VAR}_FILE` environment variables and loads secrets from files
3. **Missing setup error:** Exits with helpful message if required configuration is missing

Secrets supported via `_FILE` convention:
- Cloud relay: RELAY_WEBHOOK_URL, CERTBOT_EMAIL, RELAY_OVERFLOW_ACCESS_KEY, RELAY_OVERFLOW_SECRET_KEY
- Home device: ADMIN_PASSWORD, DKIM_PRIVATE_KEY

**Build Context Optimization:**
- Created `.dockerignore` files for root, cloud-relay, and home-device contexts
- Excludes: .git, .planning, secrets/, deploy/, .github/, tests/, *.md (except LICENSE)
- **CRITICAL:** Go source files (*.go, cmd/, pkg/, internal/) are NOT excluded (verified)
- Cloud relay .dockerignore excludes home-device/, home-device .dockerignore excludes cloud-relay/

### Task 2: GitHub Actions Workflows

Created three GitHub Actions workflows for automated builds and releases:

**1. build-custom.yml - Custom Stack Build:**
- Trigger: `workflow_dispatch` with component selection inputs
- Inputs:
  - `mail_server`: choice (stalwart, maddy, postfix-dovecot), default: stalwart
  - `webmail`: choice (none, roundcube, snappymail), default: snappymail
  - `calendar`: choice (none, radicale, builtin), default: builtin
  - `tag_suffix`: optional string for custom tagging
- Jobs run in parallel:
  - `build-relay`: Builds cloud-relay image for linux/amd64,linux/arm64
  - `build-home`: Builds home-{mail_server} image for linux/amd64,linux/arm64
- Tags: SHA, run number, latest (if default branch), custom suffix (if provided)
- Uses GitHub Actions cache (type=gha) for layer optimization
- Publishes to ghcr.io/{repo}/{component}

**2. build-prebuilt.yml - Pre-built Default Stack:**
- Triggers: push to v* tags, workflow_dispatch
- Builds two pre-configured stacks:
  - **Default:** Stalwart + SnappyMail + builtin CalDAV/CardDAV
  - **Conservative:** Postfix+Dovecot + Roundcube + Radicale
- Matrix strategy for home device images
- Tags include semantic versioning: {{version}}, {{major}}.{{minor}}, stack suffix
- Cloud relay shared across both stacks (no stack suffix)
- Default stack gets `latest` tag

**3. release.yml - Semantic Version Releases:**
- Triggers: push to v* tags
- Jobs:
  - `create-release`: Creates GitHub release with auto-generated notes
    - Pre-release detection for -rc/-beta/-alpha tags
    - Includes deployment instructions in release body
  - `build-setup-tool`: Builds cross-platform setup binaries
    - Matrix: linux/darwin on amd64/arm64
    - Conditional: only runs if deploy/setup/cmd/darkpipe-setup/main.go exists
    - Embeds version via ldflags: `-X main.version=${{ github.ref_name }}`
    - Uploads binaries as release assets

All workflows:
- Use official docker/* actions (setup-qemu@v3, setup-buildx@v3, login@v3, metadata@v5, build-push@v6)
- Target linux/amd64,linux/arm64 platforms
- Publish to ghcr.io only (no Docker Hub)
- Authenticate with GITHUB_TOKEN (no additional secrets needed)
- Use GitHub Actions cache for layer optimization

## Deviations from Plan

None - plan executed exactly as written.

## Key Decisions Made

1. **Default stack selection:** Stalwart + SnappyMail (modern, smallest footprint, built-in calendar) as default; Postfix+Dovecot + Roundcube + Radicale as conservative alternative
2. **Image architecture:** Separate images per component (cloud-relay, home-stalwart, home-maddy, home-postfix-dovecot) for flexibility and size optimization
3. **Health check command:** Netcat-based (`nc -z localhost 25`) instead of process-based (`postfix status`) for better container orchestrator compatibility
4. **Setup tool build:** Conditional in release workflow (only if source exists) to prevent failures before Plan 07-02 creates the setup tool

## Requirements Traceability

### BUILD-01: GitHub Actions pipeline with component selection
**Status:** ✅ COMPLETE
- build-custom.yml workflow accepts mail_server, webmail, calendar inputs
- workflow_dispatch trigger allows fork users to build custom stacks
- Multi-arch images (amd64, arm64) published to GHCR

### BUILD-02: Multi-architecture image support
**Status:** ✅ COMPLETE
- All Dockerfiles use TARGETARCH build arg
- buildx automatically sets TARGETARCH based on platform target
- All workflows specify `platforms: linux/amd64,linux/arm64`
- QEMU emulation via docker/setup-qemu-action for cross-compilation

### BUILD-03: Pre-built full-featured images
**Status:** ✅ COMPLETE
- build-prebuilt.yml creates two pre-built stacks
- Default stack: Stalwart + SnappyMail (latest tag)
- Conservative stack: Postfix+Dovecot + Roundcube + Radicale
- Semantic version tags on all pre-built images

### UX-03: Setup detection and helpful error messages
**Status:** ✅ COMPLETE
- All entrypoints check /config/.darkpipe-configured
- Missing setup triggers exit with helpful message
- Error message includes darkpipe-setup command
- No crash-looping on missing configuration

### Related requirements addressed:
- **SEC-05 (Docker secrets):** _FILE suffix convention implemented across all containers
- **BUILD-04 (Image size):** Alpine base, multi-stage builds, stripped binaries (foundation for <100MB target)

## Verification Results

All verification steps passed:

1. ✅ All Dockerfiles have OCI labels (org.opencontainers.image.source, version, description, licenses)
2. ✅ All entrypoints have setup detection (checks /config/.darkpipe-configured)
3. ✅ All entrypoints have Docker secrets support (reads *_FILE environment variables)
4. ✅ .dockerignore files exclude .git, .planning, secrets, deploy, tests
5. ✅ Go source files NOT excluded from build context (verified: only *_test.go excluded)
6. ✅ build-custom.yml has workflow_dispatch with mail_server, webmail, calendar choice inputs
7. ✅ build-prebuilt.yml triggers on v* tags
8. ✅ All workflows use ghcr.io registry (not docker.io)
9. ✅ All workflows target linux/amd64,linux/arm64 platforms
10. ✅ All workflows use official docker/* actions (setup-qemu, setup-buildx, login, metadata, build-push)

## Technical Implementation Notes

### Multi-Architecture Build Pattern

Dockerfiles use `ARG TARGETARCH` which is automatically set by buildx based on the `--platform` flag:
```dockerfile
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build ...
```

GitHub Actions uses `platforms: linux/amd64,linux/arm64` to trigger builds for both architectures in a single workflow run. QEMU emulation (via docker/setup-qemu-action) allows building arm64 images on amd64 runners.

### Docker Secrets Pattern

Entrypoints implement the `_FILE` suffix convention used by official Docker images:
```bash
for var in ADMIN_PASSWORD DKIM_PRIVATE_KEY; do
  file_var="${var}_FILE"
  eval file_path="\$$file_var"
  if [ -n "$file_path" ] && [ -f "$file_path" ]; then
    eval export "$var=\$(cat \"\$file_path\" | tr -d '\n')"
  fi
done
```

This allows both direct environment variables (less secure) and Docker secrets (recommended):
```yaml
# docker-compose.yml
services:
  stalwart:
    environment:
      ADMIN_PASSWORD_FILE: /run/secrets/admin_password
    secrets:
      - admin_password

secrets:
  admin_password:
    file: ./secrets/admin_password.txt
```

### GitHub Actions Cache Strategy

All workflows use GitHub Actions cache for Docker layer caching:
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

This persists Docker layers across workflow runs, significantly speeding up subsequent builds (especially for Go dependency downloads and multi-arch compilation).

### OCI Labels

All images include standard OCI labels for metadata:
- `org.opencontainers.image.source`: https://github.com/trek-e/darkpipe
- `org.opencontainers.image.version`: Set via VERSION build arg (commit SHA for custom builds, semver for releases)
- `org.opencontainers.image.description`: Component-specific description
- `org.opencontainers.image.licenses`: AGPL-3.0

## Files Changed

**Created:**
- .github/workflows/build-custom.yml (116 lines)
- .github/workflows/build-prebuilt.yml (114 lines)
- .github/workflows/release.yml (71 lines)
- cloud-relay/.dockerignore (69 lines)
- home-device/.dockerignore (69 lines)
- home-device/stalwart/entrypoint-wrapper.sh (48 lines)
- home-device/maddy/entrypoint-wrapper.sh (48 lines)

**Modified:**
- cloud-relay/Dockerfile (added TARGETARCH, VERSION args; OCI labels; netcat health check)
- cloud-relay/entrypoint.sh (added setup detection and Docker secrets support)
- home-device/stalwart/Dockerfile (added OCI labels, entrypoint wrapper)
- home-device/maddy/Dockerfile (added OCI labels, entrypoint wrapper)
- home-device/postfix-dovecot/Dockerfile (added TARGETARCH, VERSION args; OCI labels)
- home-device/postfix-dovecot/entrypoint.sh (added setup detection and Docker secrets support)
- .dockerignore (added secrets/, deploy/, tests/ exclusions)

## Next Steps

**Immediate (Plan 07-02):**
- Implement interactive setup script (darkpipe-setup) in Go
- Live validation during setup (DNS, SMTP port testing)
- Docker Compose file generation
- Upgrade-aware configuration migration

**Future:**
- Plan 07-03: Platform deployment templates (TrueNAS Scale, Unraid, Proxmox LXC)
- Phase 08: Monitoring and alerting
- Phase 09: Documentation and guides

## Self-Check: PASSED

Verified all created files exist:
- [x] .github/workflows/build-custom.yml
- [x] .github/workflows/build-prebuilt.yml
- [x] .github/workflows/release.yml
- [x] cloud-relay/.dockerignore
- [x] home-device/.dockerignore
- [x] home-device/stalwart/entrypoint-wrapper.sh
- [x] home-device/maddy/entrypoint-wrapper.sh

Verified all commits exist:
- [x] 5d8ad41 (Task 1: Dockerfiles and entrypoints)
- [x] 3a1c9f9 (Task 2: GitHub Actions workflows)

All files created successfully, all commits recorded in git history.
