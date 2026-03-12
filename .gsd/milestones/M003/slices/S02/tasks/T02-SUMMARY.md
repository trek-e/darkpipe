---
id: T02
parent: S02
milestone: M003
provides:
  - deploy/platform-guides/podman.md — consolidated Podman deployment guide (242 lines)
key_files:
  - deploy/platform-guides/podman.md
key_decisions:
  - Structured guide with Quick Start for both components before detailed sections, matching user intent (get running fast, then understand details)
  - Troubleshooting organized by symptom rather than by component, since most Podman issues cut across both deployments
patterns_established:
  - Platform guide sections: Prerequisites → Quick Start → Key Differences → Component Deployments → SELinux → Troubleshooting → Resources
observability_surfaces:
  - none (documentation only)
duration: 10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Publish Podman platform guide

**Created comprehensive Podman deployment guide consolidating cloud-relay and home-device PODMAN.md content into a single 242-line reference at `deploy/platform-guides/podman.md`.**

## What Happened

Read both per-component PODMAN.md files and the Raspberry Pi platform guide for structure reference. Wrote `deploy/platform-guides/podman.md` covering:

- **Prerequisites** with version table, verification commands, and `check-runtime.sh` reference
- **Quick Start** with 2-3 command sequences for both cloud relay (rootful) and home device (rootless)
- **Key Differences from Docker** — override files, host-gateway, networking, security defaults, memory limits
- **Cloud Relay Deployment** — rootful requirement explained (port 25 + /dev/net/tun), firewalld configuration
- **Home Device Deployment** — rootless option with sysctl, all profile commands, firewalld
- **SELinux** — detection, override file layering with exact commands, `:z` label explanation
- **Troubleshooting** — 6 common issues: pod mode DNS, SELinux permissions, firewalld ports, memory limits, rootless port binding, WireGuard TUN failures
- **Resources** — links to per-component PODMAN.md files, check-runtime.sh, Podman docs

## Verification

All task-level checks passed:

- `deploy/platform-guides/podman.md` exists — PASS
- `wc -l` = 242 lines (>50 required) — PASS
- `grep -q "rootful\|rootless"` — PASS
- `grep -q "override"` — PASS
- `grep -q "SELinux\|selinux"` — PASS
- `grep -q "check-runtime.sh"` — PASS

Slice-level (`verify-s02-docs.sh`): 6/12 pass. The 6 remaining failures are for T03 (FAQ update, quickstart runtime-agnostic language) and T04 (platform guide Podman notes) — expected at this stage.

## Diagnostics

Read the file directly: `cat deploy/platform-guides/podman.md`
Check existence and topics: `grep -c "rootful\|rootless\|override\|SELinux\|check-runtime" deploy/platform-guides/podman.md`

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `deploy/platform-guides/podman.md` — new comprehensive Podman deployment guide (242 lines)
