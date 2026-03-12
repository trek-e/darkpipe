# T02: 07-build-system-deployment 02

**Slice:** S07 — **Milestone:** M001

## Description

Create a Go-based interactive setup CLI (darkpipe-setup) that collects user configuration with live validation, generates a tailored docker-compose.yml with Docker secrets, and supports upgrade-aware re-runs with config migration.

Purpose: Deliver UX-01 (tiered experience -- simple defaults for non-technical users, full control for power users) and the deployment experience locked decisions (interactive Go setup, live validation, Docker secrets, upgrade-aware).

Output: A standalone Go module at deploy/setup/ that compiles to a cross-platform binary, providing an interactive wizard for first-time setup and a migration path for upgrades.

## Must-Haves

- [ ] "User runs darkpipe-setup and interactively selects mail domain, relay hostname, mail server, webmail, and calendar components with live validation of DNS and SMTP at each step"
- [ ] "Setup script generates a complete docker-compose.yml and Docker secrets files based on user selections"
- [ ] "First run creates fresh config; re-run detects existing config and offers upgrade with setting preservation"
- [ ] "Non-technical user gets simple defaults (Stalwart + SnappyMail), power user can override every option"
- [ ] "Docker secrets are generated as files in ./secrets/ directory, referenced in docker-compose.yml via secrets: blocks"

## Files

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
