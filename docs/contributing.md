# Contributing to DarkPipe

Thank you for your interest in contributing to DarkPipe! This document provides guidelines for contributing code, documentation, bug reports, and other improvements.

## Welcome

DarkPipe is AGPLv3 licensed open source software. We welcome contributions from everyone, whether you're fixing a typo, reporting a bug, suggesting a feature, or submitting code.

All contributions must be licensed under AGPLv3 to maintain the project's license compatibility.

## Ways to Contribute

### Report Bugs

Found a bug? Report it via [GitHub Issues](https://github.com/trek-e/darkpipe/issues).

**Good bug reports include:**
- Clear description of the problem
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- Environment details (OS, Docker version, DarkPipe version, mail server choice)
- Relevant logs (docker compose logs)
- Screenshots (if applicable)

**Before reporting:**
- Search existing issues to avoid duplicates
- Verify you're using the latest DarkPipe release
- Check the [FAQ](faq.md) and [troubleshooting sections](quickstart.md#troubleshooting)

### Suggest Features

Have an idea for a new feature? Start a discussion in [GitHub Discussions](https://github.com/trek-e/darkpipe/discussions).

**Good feature suggestions include:**
- Clear description of the feature and its purpose
- Use case or problem it solves
- How it fits with DarkPipe's core mission (email sovereignty, privacy, simplicity)
- Alternatives you've considered
- Willingness to contribute code (if applicable)

**Note:** Feature suggestions are not guarantees of implementation. Maintainers prioritize features that align with project goals and have broad user benefit.

### Submit Code

Code contributions are welcome via Pull Requests (PRs).

**Types of code contributions:**
- Bug fixes
- New features (discuss first in GitHub Discussions)
- Performance improvements
- Test coverage improvements
- Refactoring for clarity or maintainability
- Documentation improvements

See "Pull Request Process" section below for detailed workflow.

### Improve Documentation

Documentation improvements are highly valued and easier to contribute than code.

**Documentation areas:**
- User-facing docs (docs/)
- Platform guides (deploy/platform-guides/)
- Code comments
- README improvements
- Examples and tutorials

### Test on New Platforms

DarkPipe supports many platforms, but not all are tested by maintainers.

**Help by:**
- Testing DarkPipe on your platform (NAS, SBC, custom Linux)
- Reporting success or failure via GitHub Discussions
- Writing platform-specific guides (see deploy/platform-guides/)
- Contributing Docker compose configurations for new platforms

### Help Other Users

Active in the community? Help other users in [GitHub Discussions](https://github.com/trek-e/darkpipe/discussions).

- Answer questions
- Share your deployment experience
- Provide troubleshooting tips
- Create tutorials and guides

## Development Setup

### Prerequisites

- Go 1.25 or later
- Docker 24+ and Docker Compose v2, **or** Podman 5.3+ with podman-compose
- Git
- A text editor or IDE (VS Code, GoLand, etc.)

Run `bash scripts/check-runtime.sh` to validate your container runtime and development environment prerequisites.

### Clone the Repository

```bash
git clone https://github.com/trek-e/darkpipe.git
cd darkpipe
```

### Build from Source

**Go services (cloud relay, CLIs):**
```bash
# Build all Go binaries
go build ./...

# Build specific tools
go build -o darkpipe-setup ./deploy/setup/cmd/darkpipe-setup
go build -o dns-setup ./dns/cmd/dns-setup
go build -o migrate ./migration/cmd/migrate

# Run tests
go test ./...

# Run linters
go vet ./...
gofmt -s -w .
```

**Docker images:**
```bash
# Build cloud relay image
docker build -f cloud-relay/Dockerfile -t darkpipe/cloud-relay:dev .

# Build home device images (requires profiles selection)
docker compose -f home-device/docker-compose.yml build
```

### Run Locally

**Cloud relay:**
```bash
cd cloud-relay
docker compose up -d
docker compose logs -f
```

**Home device:**
```bash
cd home-device
docker compose --profile stalwart --profile snappymail up -d
docker compose logs -f
```

## Code Conventions

### Go Code Style

Follow standard Go conventions:

- **Formatting:** Use `gofmt` (or `goimports`)
- **Linting:** Code must pass `go vet` without warnings
- **Naming:** Use camelCase for unexported, PascalCase for exported
- **Error handling:** Return errors, don't panic (except in main/init for unrecoverable errors)
- **Comments:** Exported functions, types, and constants must have doc comments
- **Logging:** Use structured logging (log/slog or equivalent)

**Example:**
```go
// ProcessMessage handles incoming SMTP messages and forwards to home device.
// Returns an error if the message cannot be processed or forwarded.
func ProcessMessage(msg *SMTPMessage) error {
    if msg == nil {
        return errors.New("message is nil")
    }

    // Process message...

    return nil
}
```

### SPDX Copyright Header

**CRITICAL:** Every `.go` file MUST include the following SPDX header at the top:

```go
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 The Artificer of Ciphers, LLC
```

This is a legal requirement for AGPLv3 compliance. PRs without this header will not be merged.

**Full example:**
```go
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 The Artificer of Ciphers, LLC

package relay

import (
    "errors"
)

// ProcessMessage handles incoming SMTP messages...
func ProcessMessage(msg *SMTPMessage) error {
    // ...
}
```

### Docker Best Practices

- **Multi-stage builds:** Separate build and runtime stages
- **Minimal base images:** Use alpine or distroless for final images
- **Security:** Run as non-root user where possible. All services must include `security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, and `read_only: true` in compose files. Add `cap_add` only for documented capabilities. Include `HEALTHCHECK` in all custom Dockerfiles.
- **Layer optimization:** Combine RUN commands to reduce layers
- **Cache optimization:** Copy go.mod/go.sum before source code
- **Verification:** Run `bash scripts/verify-container-security.sh` to validate all security directives

**Example:**
```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o relay ./cloud-relay/cmd/relay

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/relay /usr/local/bin/relay
USER nobody
ENTRYPOINT ["/usr/local/bin/relay"]
```

### Configuration Files

- **YAML:** 2-space indentation
- **TOML:** 2-space indentation
- **Shell scripts:** 2-space indentation, use shellcheck
- **Markdown:** Consistent formatting, 80-character line length preferred (not strict)

## Pull Request Process

### 1. Fork and Branch

1. Fork the DarkPipe repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/darkpipe.git
   cd darkpipe
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

**Branch naming:**
- `feature/feature-name` for new features
- `fix/bug-description` for bug fixes
- `docs/topic` for documentation improvements

### 2. Make Changes

- Write your code following conventions above
- Add or update tests as needed
- Update documentation if behavior changes
- Ensure all tests pass: `go test ./...`
- Ensure code is formatted: `gofmt -s -w .`
- Ensure no lint warnings: `go vet ./...`
- If modifying Dockerfiles or compose files: `bash scripts/verify-container-security.sh`
- If adding log statements: ensure PII (email addresses, tokens) is redacted using `logutil.RedactEmail()`. Run `bash scripts/verify-log-redaction.sh` to check.
- If adding environment variables: update the relevant `.env.example` file (`cloud-relay/.env.example` or `home-device/.env.example`)

### 3. Update THIRD-PARTY-LICENSES.md

If you added new Go dependencies:

1. List the new dependency in THIRD-PARTY-LICENSES.md
2. Include the dependency's license type and copyright notice
3. Ensure the dependency uses a compatible license (MIT, BSD, Apache 2.0, etc.)

**AGPLv3 is NOT compatible with GPL-2.0-only dependencies.** Avoid adding GPL-2.0-only dependencies.

### 4. Commit Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "fix(relay): handle connection timeout gracefully

- Add timeout handling to SMTP relay connections
- Retry failed connections up to 3 times
- Log connection failures with structured logging

Fixes #123"
```

**Commit message format:**
- `type(scope): subject` on first line
- Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`
- Scopes: `relay`, `dns`, `migration`, `setup`, `docs`, etc.
- Subject: imperative mood ("add feature" not "adds feature" or "added feature")
- Body: explain what and why (not how - code shows how)
- Footer: reference issues if applicable

### 5. Push and Open PR

```bash
git push origin feature/your-feature-name
```

Open a Pull Request on GitHub:

1. Go to https://github.com/trek-e/darkpipe
2. Click "Pull requests" > "New pull request"
3. Click "compare across forks"
4. Select your fork and branch
5. Fill out PR template:
   - Clear description of changes
   - Motivation and context
   - How has this been tested?
   - Checklist: tests pass, docs updated, SPDX headers present

### 6. Code Review

Maintainers will review your PR and may request changes.

**During review:**
- Respond to feedback promptly and constructively
- Make requested changes in new commits (don't force-push during review)
- Mark conversations as resolved when addressed
- Be patient - maintainers are volunteers with limited time

**After approval:**
- Maintainer will merge your PR
- Your contribution will be included in the next release

### 7. Celebrate

You're now a DarkPipe contributor! Thank you for helping improve email sovereignty for everyone.

## Testing Guidelines

### Unit Tests

- Write unit tests for new functions and packages
- Use Go's built-in testing package
- Aim for 70%+ code coverage on new code
- Test both success and error cases

**Example:**
```go
func TestProcessMessage(t *testing.T) {
    msg := &SMTPMessage{
        From: "sender@example.com",
        To:   []string{"recipient@example.com"},
        Body: "Test message",
    }

    err := ProcessMessage(msg)
    if err != nil {
        t.Errorf("ProcessMessage() failed: %v", err)
    }
}

func TestProcessMessage_NilMessage(t *testing.T) {
    err := ProcessMessage(nil)
    if err == nil {
        t.Error("ProcessMessage(nil) should return error")
    }
}
```

### Integration Tests

For changes that affect multiple components:

- Test with docker compose up
- Verify end-to-end flows (send/receive email)
- Test on multiple platforms if possible (amd64, arm64)

### Manual Testing Checklist

Before submitting PR:

- [ ] Cloud relay builds successfully
- [ ] Home device builds successfully
- [ ] Setup wizard runs without errors
- [ ] DNS setup tool works (dry-run)
- [ ] Docker compose up succeeds for cloud and home
- [ ] SMTP connection to relay works (telnet test)
- [ ] Webmail loads and login works
- [ ] No errors in docker compose logs
- [ ] Container security audit passes: `bash scripts/verify-container-security.sh`
- [ ] Log redaction audit passes: `bash scripts/verify-log-redaction.sh`
- [ ] `.env.example` files updated if new env vars added

## Code of Conduct

DarkPipe is an inclusive project. We expect all contributors to:

- **Be respectful:** Treat others with respect and empathy
- **Be constructive:** Provide helpful feedback and criticism
- **Be inclusive:** Welcome contributors of all backgrounds and skill levels
- **Be collaborative:** Work together toward common goals

**Not tolerated:**
- Harassment, discrimination, or personal attacks
- Trolling, insulting comments, or inflammatory language
- Publishing others' private information without permission
- Any conduct that would be inappropriate in a professional setting

**Enforcement:**
- Maintainers reserve the right to remove contributions that violate these principles
- Repeat or severe violations may result in ban from the project
- Reports of violations: Email conduct@darkpipe.org (maintainers only)

## License

By contributing to DarkPipe, you agree that your contributions will be licensed under the GNU Affero General Public License v3.0 or later (AGPLv3+).

This means:
- Your code can be freely used, modified, and distributed
- Derivative works must also be AGPLv3 licensed
- Network use triggers the same license obligations as distribution
- You retain copyright to your contributions, but grant DarkPipe and all users the rights defined by AGPLv3

See [LICENSE](../LICENSE) for full license text.

## Questions?

- General questions: [GitHub Discussions](https://github.com/trek-e/darkpipe/discussions)
- Bug reports: [GitHub Issues](https://github.com/trek-e/darkpipe/issues)
- Security concerns: See [docs/security.md](security.md)

Thank you for contributing to DarkPipe and helping build a more sovereign internet!

---

Last Updated: 2026-03-12

License: AGPLv3 - See [LICENSE](../LICENSE)
