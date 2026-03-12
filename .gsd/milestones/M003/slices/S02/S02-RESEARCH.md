# S02: Runtime-Agnostic Documentation & Tooling — Research

**Date:** 2026-03-12

## Summary

S02 transforms DarkPipe's documentation from Docker-only to runtime-agnostic, creates a runtime compatibility check script, publishes a Podman platform guide, and updates the FAQ. The scope spans 14 documentation files (~5,900 lines total) across `docs/` and `deploy/platform-guides/`, plus one new script and one new platform guide.

The core challenge is making documentation runtime-neutral without making it verbose or confusing for the 90% of users who will use Docker. The approach should use a "generic first, callout when different" pattern — keep the main flow Docker-friendly with callout blocks for Podman differences. The compatibility check script (`scripts/check-runtime.sh`) can largely reuse the pattern established by `scripts/verify-podman-compat.sh` and `scripts/verify-container-security.sh` from S01.

S01 already did the heavy lifting: compose files are dual-compatible, override files exist, per-component PODMAN.md docs are written, and the verification script validates compose syntax. S02 is a documentation and tooling polish pass, not a technical compatibility task.

## Recommendation

**Approach: Graduated runtime-agnostic language with runtime-specific callout blocks.**

1. **Core docs (`docs/*.md`):** Replace Docker-specific language with generic terms where the behavior is identical (e.g., "container runtime" instead of "Docker", `docker compose` commands shown with a note that `podman-compose` works identically). Use admonition/callout blocks for Podman-specific differences (rootful, SELinux, override files).

2. **Platform guides (`deploy/platform-guides/*.md`):** Keep Docker as the primary path since each guide is platform-specific. Add a "Using Podman" subsection to guides where Podman is relevant (raspberry-pi, proxmox-lxc, truenas-scale, unraid). Mac Silicon guide gets a brief note pointing to the Apple Containers guide (S03 scope).

3. **New Podman platform guide (`deploy/platform-guides/podman.md`):** Consolidate `cloud-relay/PODMAN.md` and `home-device/PODMAN.md` into a single comprehensive guide following the same structure as existing platform guides.

4. **Compatibility check script (`scripts/check-runtime.sh`):** Detect installed runtime, check version requirements, validate compose file compatibility, report network and permission readiness. Structured PASS/FAIL output matching existing script patterns.

5. **FAQ update:** Change the "Can I use Podman?" answer from "Probably, but not officially supported" to "Yes, fully supported" with links to the Podman platform guide.

**Why this approach:** Keeps docs approachable for Docker users (majority) while making Podman users feel first-class. Avoids the anti-pattern of duplicating every command block for each runtime.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| PASS/FAIL script structure | `scripts/verify-podman-compat.sh` | Established pattern with counters, emoji, skip handling — reuse the shell functions |
| Podman deployment instructions | `cloud-relay/PODMAN.md`, `home-device/PODMAN.md` | Already written and verified in S01; consolidate rather than rewrite |
| Platform guide structure | `deploy/platform-guides/raspberry-pi.md` | Consistent format: Prerequisites, Quick Start, Detailed Steps, Troubleshooting, Resources |

## Existing Code and Patterns

- `scripts/verify-podman-compat.sh` — S01's Podman compose validation script. 7 check categories, graceful skip when tools unavailable. Reuse `pass()/fail()/skip()` pattern for `check-runtime.sh`.
- `scripts/verify-container-security.sh` — Security audit script. Same output pattern. Shows how to iterate compose files and check properties.
- `cloud-relay/PODMAN.md` — Rootful Podman requirements, override file usage, firewalld config, known differences. Source material for the consolidated platform guide.
- `home-device/PODMAN.md` — Rootless option with sysctl, profile commands, pod mode warning. Source material for the consolidated platform guide.
- `docs/faq.md` — Lines ~118-130: "Can I use Podman?" section that needs updating. Currently says "Probably, but not officially supported."
- `docs/quickstart.md` — 25 Docker references, most in command blocks. Highest-touch file for runtime-agnostic rewrite.
- `docs/configuration.md` — 13 Docker references. Mentions `docker-compose.yml`, `docker compose` commands, Docker Compose profiles section.
- `docs/contributing.md` — 15 Docker references. Lists Docker 27+ as requirement, Docker build commands.
- `docs/security.md` — 8 Docker references. Mentions Docker HEALTHCHECK, container hardening, Docker volumes.
- `deploy/platform-guides/mac-silicon.md` — 54 Docker references. Heavily Docker Desktop / OrbStack focused. Needs Apple Containers mention (link to S03 guide).
- `deploy/platform-guides/raspberry-pi.md` — 30 Docker references. Good candidate for Podman subsection.
- `deploy/platform-guides/proxmox-lxc.md` — 28 Docker references. Docker install inside LXC container. Podman is viable alternative here.
- `deploy/platform-guides/unraid.md` — 38 Docker references. Unraid has native Docker; Podman unlikely but should mention.
- `deploy/platform-guides/synology-nas.md` — 26 Docker references. Synology Container Manager is Docker-based; Podman not practical.
- `deploy/platform-guides/truenas-scale.md` — 14 Docker references. TrueNAS SCALE uses Docker; low priority for Podman note.

## Constraints

- **Base compose files must not be modified** — S01 decision: Podman-specific settings live in override files only.
- **Existing Docker users must not be confused** — Docker is still the primary/default path. Runtime-agnostic language must not make Docker instructions harder to follow.
- **`docker compose` and `podman-compose` command syntax is nearly identical** — Most command examples work as-is for both runtimes. The key difference is the override file layering (`-f docker-compose.podman.yml`).
- **Platform guides are platform-specific by nature** — Some platforms (Synology, Unraid) are Docker-only; don't force Podman language where it doesn't apply.
- **S03 (Apple Containers) is a separate slice** — Mac Silicon guide should link to Apple Containers guide but not document it in S02.
- **File names remain as `Dockerfile` and `docker-compose.yml`** — Per DECISIONS.md, no renaming to Containerfile.
- **The per-component PODMAN.md files in cloud-relay/ and home-device/ should be kept** — They serve as operational quick-reference alongside the compose files. The new platform guide consolidates and contextualizes.

## Common Pitfalls

- **Over-genericizing Docker commands** — Don't replace every `docker compose` with `$RUNTIME compose`. This makes docs harder to copy-paste. Keep `docker compose` as the primary command with a note at the top of each relevant doc that `podman-compose` is also supported.
- **Duplicating Podman info in too many places** — The per-component PODMAN.md files, the platform guide, and inline doc callouts could diverge. Keep the platform guide as the single source of truth; per-component files are operational quick-reference only.
- **Compatibility script scope creep** — The check-runtime.sh script should detect and validate the runtime environment, NOT replicate what `verify-podman-compat.sh` already does (compose file validation). Keep them complementary: `check-runtime.sh` = "is my system ready?", `verify-podman-compat.sh` = "are the compose files correct?".
- **Forgetting the cloud-relay rootful requirement** — Podman docs must consistently state that cloud relay requires rootful Podman (port 25 + /dev/net/tun). Users expecting rootless-everywhere will hit this.
- **Missing the host-gateway difference** — Podman uses `host.containers.internal` where Docker uses `host-gateway`. The override file handles this, but docs must explain why the override is needed.

## Open Risks

- **Documentation churn during S03/S04** — S03 (Apple Containers) and S04 (CI) will also touch docs. Keep S02 edits focused on Docker→runtime-agnostic language and Podman; avoid touching areas S03/S04 will modify.
- **`cloud-relay/docker-compose.yml` has a pre-existing `docker compose config` failure** — Noted in every S01 task summary. The tmpfs/volume conflict on `/var/spool/postfix` will cause the compatibility check script to report a failure if it validates compose config. Either fix it (separate task) or document it as a known issue in the script output.
- **Version detection accuracy** — Detecting Podman version requires parsing `podman --version` output. Format may vary across distributions (e.g., `podman version 5.3.1` vs `podman 5.3.1-rhel`). Use permissive regex.
- **SELinux detection** — The check script needs to detect SELinux enforcement state to recommend the SELinux override file. `getenforce` may not exist on non-SELinux systems. Graceful handling required.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Bash scripting | No relevant skill found | none found |
| Technical docs | `github/awesome-copilot@documentation-writer` (8.1K installs) | available — generic documentation writing, not container-specific; likely not worth installing |
| Podman | No relevant skill found | none found |
| Docker | No relevant skill found (in available_skills) | none found |

No skills directly relevant to this slice's work (documentation rewriting, bash scripting, Podman/Docker compatibility). The work is domain-specific to DarkPipe's codebase patterns.

## Sources

- S01 task summaries (T01-T05) — Podman compatibility changes, override file structure, verification script patterns, known issues (source: `.gsd/milestones/M003/slices/S01/tasks/`)
- S01 PODMAN.md files — Existing Podman deployment docs to consolidate (source: `cloud-relay/PODMAN.md`, `home-device/PODMAN.md`)
- Existing verification scripts — Output patterns and shell function conventions (source: `scripts/verify-podman-compat.sh`, `scripts/verify-container-security.sh`)
- Current documentation files — Docker reference counts and locations across 14 files (source: `docs/*.md`, `deploy/platform-guides/*.md`)
- DECISIONS.md — Architectural constraints on compose files, naming, overrides (source: `.gsd/DECISIONS.md`)
