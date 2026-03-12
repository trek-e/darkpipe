---
estimated_steps: 4
estimated_files: 1
---

# T02: Publish Podman platform guide

**Slice:** S02 — Runtime-Agnostic Documentation & Tooling
**Milestone:** M003

## Description

Create the consolidated Podman platform guide at `deploy/platform-guides/podman.md`. This is the single source of truth for running DarkPipe on Podman, consolidating content from `cloud-relay/PODMAN.md` and `home-device/PODMAN.md` into a comprehensive guide following the structure of existing platform guides.

The per-component PODMAN.md files remain as operational quick-reference alongside compose files (per DECISIONS.md constraint). The platform guide provides the full deployment story: prerequisites, setup, configuration, troubleshooting.

## Steps

1. Read `cloud-relay/PODMAN.md` and `home-device/PODMAN.md` to gather all Podman-specific content
2. Read `deploy/platform-guides/raspberry-pi.md` for the platform guide structure pattern (Prerequisites, Quick Start, Detailed Steps, Troubleshooting, Resources)
3. Write `deploy/platform-guides/podman.md` covering:
   - Prerequisites: Podman 5.3+, podman-compose, `check-runtime.sh` validation
   - Key differences from Docker: rootful requirement for cloud relay (port 25 + /dev/net/tun), host-gateway (`host.containers.internal`), override file layering
   - Cloud relay deployment: rootful setup, override file usage, firewalld configuration
   - Home device deployment: rootless option with sysctl, profile commands
   - SELinux considerations: when to use the selinux override file, `:Z`/`:z` labels
   - Troubleshooting: common issues (network, permissions, pod mode warning)
   - Resources: links to per-component PODMAN.md files, Podman docs, check-runtime.sh
4. Verify the guide is >50 lines, self-contained, and covers all critical topics

## Must-Haves

- [ ] Guide follows existing platform guide structure (Prerequisites, Quick Start, Detailed Steps, Troubleshooting, Resources)
- [ ] Rootful requirement for cloud relay is clearly stated with explanation
- [ ] Override file usage explained with exact commands
- [ ] SELinux section with detection and override file guidance
- [ ] Troubleshooting section covers common Podman-specific issues
- [ ] References `scripts/check-runtime.sh` for prerequisites validation

## Verification

- `deploy/platform-guides/podman.md` exists
- `wc -l deploy/platform-guides/podman.md` shows >50 lines
- `grep -q "rootful\|rootless" deploy/platform-guides/podman.md` confirms rootful/rootless coverage
- `grep -q "override" deploy/platform-guides/podman.md` confirms override file documentation
- `grep -q "SELinux\|selinux" deploy/platform-guides/podman.md` confirms SELinux coverage

## Observability Impact

- Signals added/changed: None (documentation only)
- How a future agent inspects this: Read the file; check existence and content with grep
- Failure state exposed: None

## Inputs

- `cloud-relay/PODMAN.md` — rootful requirements, override file usage, firewalld config, known differences
- `home-device/PODMAN.md` — rootless option, sysctl configuration, pod mode warning
- `deploy/platform-guides/raspberry-pi.md` — structure pattern for platform guides
- `scripts/check-runtime.sh` — referenced in prerequisites (created in T01)

## Expected Output

- `deploy/platform-guides/podman.md` — comprehensive Podman deployment guide, single source of truth for Podman users
