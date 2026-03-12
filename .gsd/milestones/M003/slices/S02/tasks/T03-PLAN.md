---
estimated_steps: 5
estimated_files: 5
---

# T03: Make core docs runtime-agnostic and update FAQ

**Slice:** S02 — Runtime-Agnostic Documentation & Tooling
**Milestone:** M003

## Description

Update the five core documentation files to use runtime-agnostic language while keeping Docker as the primary/default path. The key pattern: keep `docker compose` in command examples (copy-pasteable), replace generic references to "Docker" with "container runtime" where the statement applies to any runtime, and add Podman callout blocks where behavior differs.

The FAQ gets the most significant content change: the "Can I use Podman?" answer changes from "Probably, but not officially supported" to "Yes, fully supported" with links to the Podman platform guide.

## Steps

1. Update `docs/quickstart.md`:
   - Add a "Container Runtime" note near the top explaining examples use `docker compose` but Podman is fully supported via `podman-compose` with override files
   - Replace "Docker containers" / "Docker environment" with "containers" / "container environment" where runtime-generic
   - Keep all `docker compose` command examples as-is
   - Add "> **Podman users:**" callout block where override files are needed
2. Update `docs/configuration.md`:
   - Same pattern: add runtime note, genericize where appropriate, add Podman callouts for compose profile and override file sections
3. Update `docs/contributing.md`:
   - Update prerequisites section to list "Docker 24+ or Podman 5.3+" instead of Docker-only
   - Add note that `scripts/check-runtime.sh` validates the development environment
   - Keep Docker build commands as primary examples
4. Update `docs/security.md`:
   - Genericize "Docker HEALTHCHECK" to "container HEALTHCHECK" where appropriate
   - Keep Docker-specific security hardening language where it is Docker-specific
   - Add note that Podman's rootless mode provides additional security isolation
5. Update `docs/faq.md`:
   - Rewrite the "Can I use Podman?" section entirely: "Yes, fully supported" answer
   - Link to `deploy/platform-guides/podman.md`
   - List key differences (rootful for cloud relay, override files, SELinux)
   - Remove "Not tested" and "not officially supported" language

## Must-Haves

- [ ] `docs/quickstart.md` has a runtime note and Podman callout
- [ ] `docs/configuration.md` has runtime-agnostic language and Podman callouts
- [ ] `docs/contributing.md` lists Podman as alternative prerequisite
- [ ] `docs/security.md` genericized where appropriate
- [ ] `docs/faq.md` Podman answer says "Yes, fully supported" with link to guide
- [ ] No `docker compose` commands replaced with variables or generic placeholders
- [ ] All Docker command examples remain directly copy-pasteable

## Verification

- `grep -q "container runtime" docs/quickstart.md` confirms runtime-agnostic language
- `grep -qi "podman" docs/quickstart.md` confirms Podman callout present
- `grep -qi "fully supported" docs/faq.md` confirms FAQ update
- `! grep -qi "not tested\|not officially supported" docs/faq.md` confirms old language removed
- `grep -qi "podman 5.3\|podman" docs/contributing.md` confirms prerequisite update
- `bash scripts/verify-s02-docs.sh` — FAQ and quickstart checks now pass

## Observability Impact

- Signals added/changed: None (documentation only)
- How a future agent inspects this: grep for key phrases; run verify-s02-docs.sh
- Failure state exposed: None

## Inputs

- `docs/quickstart.md`, `docs/configuration.md`, `docs/contributing.md`, `docs/security.md`, `docs/faq.md` — current files with Docker-only language
- `deploy/platform-guides/podman.md` — link target for FAQ (created in T02)
- S02 research — Docker reference counts per file, constraint on keeping commands copy-pasteable

## Expected Output

- `docs/quickstart.md` — runtime-agnostic with Podman callouts
- `docs/configuration.md` — runtime-agnostic with Podman callouts
- `docs/contributing.md` — updated prerequisites
- `docs/security.md` — genericized where appropriate
- `docs/faq.md` — "fully supported" Podman answer with guide link
