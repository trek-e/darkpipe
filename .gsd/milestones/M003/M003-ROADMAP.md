# M003: Container Runtime Compatibility

**Vision:** DarkPipe runs on Podman and Apple Containers in addition to Docker, with runtime-agnostic documentation, tested compose files, and platform guides — so users aren't locked into any single container runtime.

## Success Criteria

- `podman-compose` starts all cloud-relay and home-device services with health checks passing
- Full mail send/receive flow works on a Podman deployment
- Apple Containers platform guide enables running cloud relay components on macOS 26
- All existing Docker deployments continue to work identically (zero regression)
- Core documentation uses runtime-agnostic language with runtime-specific callouts
- Runtime compatibility check script validates Podman, Docker, and Apple Containers environments
- CI includes a Podman build/lint job that passes

## Key Risks / Unknowns

- Podman rootless networking with WireGuard tunnel — may need different network configuration
- Apple Containers has no compose equivalent — multi-service orchestration requires custom approach
- Apple Containers WireGuard kernel module availability — unknown in Apple's custom Linux kernel
- Podman SELinux volume labels (`:Z`) needed on Fedora/RHEL — current compose files lack them

## Proof Strategy

- Podman networking risk → retire in S01 by proving WireGuard tunnel connects between Podman containers and mail flows end-to-end
- Apple Containers orchestration risk → retire in S03 by proving cloud relay services start and accept SMTP connections
- SELinux volume labels → retire in S01 by testing on Fedora runner or documenting workaround

## Verification Classes

- Contract verification: `podman-compose config`, `podman build`, `go vet`, `go build`, compatibility check script
- Integration verification: mail send/receive on Podman, health checks across runtimes
- Operational verification: container restart cycles on Podman, WireGuard reconnection
- UAT / human verification: Apple Containers guide walkthrough on macOS 26

## Milestone Definition of Done

This milestone is complete only when all are true:

- All slices are complete with passing verification
- Podman deployment passes full mail flow test
- Apple Containers guide is published and tested
- Docker deployments show zero regression
- Documentation is runtime-agnostic throughout
- CI includes Podman validation
- Success criteria are verified against running systems

## Requirement Coverage

- Covers: All existing platform compatibility requirements (extended to new runtimes)
- Partially covers: None
- Leaves for later: Kubernetes deployment, LXC/LXD
- Orphan risks: Apple Containers API stability (macOS 26 is first release)

## Slices

- [x] **S01: Podman Compose Compatibility** `risk:high` `depends:[]`
  > After this: `podman-compose up` starts all cloud-relay and home-device services, health checks pass, and a full mail send/receive flow works over WireGuard on Podman. Compose files are validated dual-compatible.
- [x] **S02: Runtime-Agnostic Documentation & Tooling** `risk:medium` `depends:[S01]`
  > After this: All core docs use runtime-agnostic language, a runtime compatibility check script validates any supported runtime, Podman platform guide is published, and the FAQ "Can I use Podman?" answer is "Yes, fully supported."
- [ ] **S03: Apple Containers Support** `risk:high` `depends:[]`
  > After this: A macOS 26 platform guide documents running DarkPipe cloud relay on Apple Containers, images pull and start, and SMTP connectivity is verified. Limitations (no compose, manual orchestration) are clearly documented.
- [ ] **S04: CI & Regression Validation** `risk:low` `depends:[S01]`
  > After this: GitHub Actions includes a Podman build and compose validation job, existing Docker CI continues to pass, and the compatibility check script runs in CI for both runtimes.

## Boundary Map

### S01 (independent)

Produces:
- Podman-compatible compose files (dual-compatible with Docker)
- Documented Podman networking configuration for WireGuard
- SELinux volume label handling (`:Z` suffix documentation or conditional)
- Podman-specific Dockerfile adjustments (if any)

Consumes:
- nothing (independent)

### S02 (depends on S01)

Produces:
- Runtime-agnostic documentation across all docs/*.md files
- Runtime compatibility check script (`scripts/check-runtime.sh`)
- Podman platform guide (`deploy/platform-guides/podman.md`)
- Updated FAQ with official Podman support statement

Consumes:
- S01's tested Podman configuration and known issues/workarounds

### S03 (independent)

Produces:
- Apple Containers platform guide (`deploy/platform-guides/apple-containers.md`)
- Manual orchestration scripts or documentation for multi-service startup
- Tested image pull and container start on Apple Containers
- Documented limitations and workarounds

Consumes:
- nothing (independent)

### S04 (depends on S01)

Produces:
- GitHub Actions workflow job for Podman build and compose validation
- Regression test ensuring Docker compose continues to work
- CI integration of runtime compatibility check script

Consumes:
- S01's Podman-compatible compose files
- S02's compatibility check script
