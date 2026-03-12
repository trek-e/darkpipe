# T01: 07-build-system-deployment 01

**Slice:** S07 — **Milestone:** M001

## Description

Update all Dockerfiles for multi-architecture builds with size optimization and setup detection, then create GitHub Actions workflows for custom component selection builds, pre-built default stack publishing, and semantic version releases.

Purpose: Deliver BUILD-01 (GitHub Actions pipeline with component selection), BUILD-02 (multi-arch images), and BUILD-03 (pre-built full-featured images) while establishing the image optimization foundation (<100MB target).

Output: Updated Dockerfiles with TARGETARCH support, OCI labels, and setup detection entrypoints; three GitHub Actions workflows (custom build, prebuilt, release); .dockerignore files.

## Must-Haves

- [ ] "User forks the repo and triggers a GitHub Actions workflow_dispatch with component selection (mail server, webmail, calendar) to build custom multi-arch Docker images"
- [ ] "Pre-built default stack images (Stalwart + SnappyMail and Postfix+Dovecot + Roundcube + Radicale) are built and published to GHCR on every release tag"
- [ ] "All images build for linux/amd64 and linux/arm64 from a single workflow run"
- [ ] "All Docker images are under 100MB per architecture and use Alpine multi-stage builds with stripped Go binaries"
- [ ] "Container entrypoints detect missing setup configuration and exit with a helpful message instead of crash-looping"

## Files

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
