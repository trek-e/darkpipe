# S07: Build System Deployment

**Goal:** Update all Dockerfiles for multi-architecture builds with size optimization and setup detection, then create GitHub Actions workflows for custom component selection builds, pre-built default stack publishing, and semantic version releases.
**Demo:** Update all Dockerfiles for multi-architecture builds with size optimization and setup detection, then create GitHub Actions workflows for custom component selection builds, pre-built default stack publishing, and semantic version releases.

## Must-Haves


## Tasks

- [x] **T01: 07-build-system-deployment 01**
  - Update all Dockerfiles for multi-architecture builds with size optimization and setup detection, then create GitHub Actions workflows for custom component selection builds, pre-built default stack publishing, and semantic version releases.

Purpose: Deliver BUILD-01 (GitHub Actions pipeline with component selection), BUILD-02 (multi-arch images), and BUILD-03 (pre-built full-featured images) while establishing the image optimization foundation (<100MB target).

Output: Updated Dockerfiles with TARGETARCH support, OCI labels, and setup detection entrypoints; three GitHub Actions workflows (custom build, prebuilt, release); .dockerignore files.
- [x] **T02: 07-build-system-deployment 02**
  - Create a Go-based interactive setup CLI (darkpipe-setup) that collects user configuration with live validation, generates a tailored docker-compose.yml with Docker secrets, and supports upgrade-aware re-runs with config migration.

Purpose: Deliver UX-01 (tiered experience -- simple defaults for non-technical users, full control for power users) and the deployment experience locked decisions (interactive Go setup, live validation, Docker secrets, upgrade-aware).

Output: A standalone Go module at deploy/setup/ that compiles to a cross-platform binary, providing an interactive wizard for first-time setup and a migration path for upgrades.
- [x] **T03: 07-build-system-deployment 03**
  - Create native app templates for TrueNAS Scale and Unraid, write platform-specific deployment guides for all target platforms, and build the Phase 7 integration test suite.

Purpose: Deliver UX-03 (runs on RPi4, x64 Docker, TrueNAS Scale, Unraid) with native platform integration and clear guides for all target platforms. The phase test suite validates all Phase 7 artifacts work correctly.

Output: TrueNAS Scale YAML app template, Unraid XML template, six platform deployment guides, and a comprehensive integration test script.

## Files Likely Touched

- `cloud-relay/Dockerfile`
- `cloud-relay/entrypoint.sh`
- `cloud-relay/.dockerignore`
- `home-device/stalwart/Dockerfile`
- `home-device/maddy/Dockerfile`
- `home-device/postfix-dovecot/Dockerfile`
- `home-device/postfix-dovecot/entrypoint.sh`
- `home-device/.dockerignore`
- `.github/workflows/build-custom.yml`
- `.github/workflows/build-prebuilt.yml`
- `.github/workflows/release.yml`
- `.dockerignore`
- `deploy/setup/go.mod`
- `deploy/setup/go.sum`
- `deploy/setup/cmd/darkpipe-setup/main.go`
- `deploy/setup/pkg/config/config.go`
- `deploy/setup/pkg/validate/dns.go`
- `deploy/setup/pkg/validate/smtp.go`
- `deploy/setup/pkg/validate/ports.go`
- `deploy/setup/pkg/compose/generate.go`
- `deploy/setup/pkg/compose/templates.go`
- `deploy/setup/pkg/secrets/secrets.go`
- `deploy/setup/pkg/migrate/migrate.go`
- `deploy/templates/truenas-scale/app.yaml`
- `deploy/templates/truenas-scale/questions.yaml`
- `deploy/templates/unraid/darkpipe.xml`
- `deploy/platform-guides/raspberry-pi.md`
- `deploy/platform-guides/truenas-scale.md`
- `deploy/platform-guides/unraid.md`
- `deploy/platform-guides/proxmox-lxc.md`
- `deploy/platform-guides/synology-nas.md`
- `deploy/platform-guides/mac-silicon.md`
- `tests/test-phase-07.sh`
