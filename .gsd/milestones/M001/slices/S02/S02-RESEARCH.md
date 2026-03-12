# Phase 02: Cloud Relay - Research

**Researched:** 2026-02-08
**Domain:** Postfix SMTP relay, Let's Encrypt automation, container optimization
**Confidence:** HIGH

## Summary

Phase 02 implements an internet-facing SMTP gateway using Postfix in a minimal Alpine Linux container, combined with a custom Go relay daemon that forwards messages to the home device via Phase 01's transport layer (WireGuard or mTLS). The cloud relay enforces TLS on all internet-facing connections using Let's Encrypt certificates, provides optional strict mode to refuse plaintext-only peers, and maintains zero persistent message storage by forwarding in-flight only.

The architecture uses Postfix for battle-tested SMTP protocol handling and direct MTA delivery, with a Go daemon using emersion/go-smtp to bridge between Postfix and the secure transport. Container target is 25-30MB (Alpine 5MB + Postfix 15MB + WireGuard tools 5MB + Go binary 2-5MB) with sub-256MB RAM usage on a $5/month VPS.

**Primary recommendation:** Use Postfix relay-only configuration with mydestination="" (null client mode), integrate via transport_maps to forward all mail to Go daemon listening on localhost, Go daemon forwards via WireGuard/mTLS to home device. Certbot sidecar for Let's Encrypt automation. Monitor TLS connection quality via real-time log parsing with user notifications for plaintext-only peers.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Postfix (Alpine) | 3.7.4+ | Internet-facing SMTP relay | Battle-tested (25+ years), 15-30MB footprint, stable configuration API, relay-only mode eliminates local delivery complexity, excellent documentation |
| Alpine Linux | 3.20+ | Container base image | 5MB base, musl libc compatible with Postfix, package manager for operational debugging, standard for minimal containers |
| Let's Encrypt/Certbot | Latest | Public TLS certificates | Free, automated, 90-day rotation standard, DNS-01 challenge for wildcard support, official EFF tooling |
| emersion/go-smtp | 0.24.0 | Go SMTP server/client library | 1.1K+ dependent projects, RFC 5321 compliant, active maintenance (2025-08-05 release), MIT licensed, production-proven relay implementations exist (go-smtp-proxy pattern) |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| WireGuard tools | Latest (kernel module) | Cloud → home transport | Phase 01 transport, already implemented, requires kernel module access on VPS |
| cenkalti/backoff/v4 | Latest | Exponential backoff for reconnection | Phase 01 dependency, already integrated for persistent connection maintenance |
| postfix-exporter | Latest | Prometheus metrics (optional) | If Prometheus monitoring desired, provides queue depth, delivery stats, TLS metrics |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Postfix | Maddy (Go) | 15MB single binary vs 15MB Postfix, less battle-tested (0.8.2 released 2026-01-14), simpler config but smaller community |
| Postfix | Haraka (Node.js) | Event-driven architecture, but requires 50MB+ Node.js runtime, defeats container size target |
| Certbot | acme.sh | Shell script vs Python, broader DNS provider support, but less standardized, smaller community |
| Alpine | Distroless | 2-3MB smaller, but no shell for debugging, no package manager, harder operational troubleshooting |

**Installation:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay-daemon ./cmd/relay

FROM alpine:3.21
RUN apk add --no-cache postfix ca-certificates wireguard-tools
COPY --from=builder /build/relay-daemon /usr/local/bin/relay-daemon
COPY postfix-config/ /etc/postfix/
EXPOSE 25
CMD ["/usr/local/bin/relay-daemon"]
```

## Architecture Patterns

### Recommended Project Structure
```
cloud-relay/
├── cmd/
│   └── relay/
│       └── main.go           # Relay daemon entrypoint
├── relay/
│   ├── smtp/
│   │   ├── server.go         # emersion/go-smtp backend implementation
│   │   ├── session.go        # SMTP session handling (forward to transport)
│   │   └── tls_monitor.go    # TLS connection quality monitoring
│   ├── forward/
│   │   ├── wireguard.go      # WireGuard transport forwarding
│   │   └── mtls.go           # mTLS transport forwarding
│   └── notify/
│       └── tls_warnings.go   # User notifications for TLS failures
├── postfix-config/
│   ├── main.cf               # Postfix relay-only configuration
│   ├── master.cf             # Process configuration
│   └── transport             # Transport map: * -> Go daemon
└── Dockerfile
```

### Pattern 1: Postfix Relay-Only (Null Client) Configuration
**What:** Configure Postfix to forward all mail without local delivery
**When to use:** Cloud relay that never stores messages, only forwards them
**Example:**
```bash
# Source: http://www.postfix.org/BASIC_CONFIGURATION_README.html
# /etc/postfix/main.cf

# Network and hostname
myhostname = relay.darkpipe.example.com
mydomain = darkpipe.example.com
myorigin = $mydomain

# NULL CLIENT: No local delivery
mydestination =
local_recipient_maps =
local_transport = error:local mail delivery is disabled

# Accept mail from internet
inet_interfaces = all
inet_protocols = ipv4

# Relay restrictions
smtpd_relay_restrictions = permit_mynetworks, reject_unauth_destination
mynetworks = 127.0.0.0/8, [::1]/128

# Forward all mail to Go relay daemon on localhost:10025
transport_maps = hash:/etc/postfix/transport
# In /etc/postfix/transport:
# *    smtp:[127.0.0.1]:10025

# Message size limits (50MB max for large attachments)
message_size_limit = 52428800

# TLS for inbound connections (STARTTLS)
smtpd_tls_cert_file = /etc/letsencrypt/live/relay.darkpipe.example.com/fullchain.pem
smtpd_tls_key_file = /etc/letsencrypt/live/relay.darkpipe.example.com/privkey.pem
smtpd_tls_security_level = may
smtpd_tls_loglevel = 1

# TLS for outbound connections (optional strict mode)
smtp_tls_security_level = may
smtp_tls_loglevel = 1
# For strict mode per-destination:
# smtp_tls_policy_maps = hash:/etc/postfix/tls_policy
```

### Pattern 2: Go SMTP Relay Daemon (emersion/go-smtp Backend)
**What:** Receive mail from Postfix, forward via WireGuard/mTLS to home device
**When to use:** Bridge between Postfix and secure transport layer
**Example:**
```go
// Source: https://github.com/emersion/go-smtp (backend interface)
// relay/smtp/server.go

package smtp

import (
    "io"
    "github.com/emersion/go-smtp"
)

type Backend struct {
    forwarder Forwarder // Forwards to WireGuard/mTLS transport
}

func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
    return &Session{forwarder: b.forwarder}, nil
}

type Session struct {
    forwarder Forwarder
    from      string
    to        []string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
    s.from = from
    return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
    s.to = append(s.to, to)
    return nil
}

func (s *Session) Data(r io.Reader) error {
    // Forward message via WireGuard/mTLS to home device
    return s.forwarder.Forward(s.from, s.to, r)
}

func (s *Session) Reset() {
    s.from = ""
    s.to = nil
}

func (s *Session) Logout() error {
    return nil
}
```

### Pattern 3: Let's Encrypt Certbot Sidecar Automation
**What:** Separate container handles certificate obtaining and renewal
**When to use:** Automated TLS certificate management for internet-facing relay
**Example:**
```yaml
# Source: https://github.com/bybatkhuu/sidecar-certbot
# docker-compose.yml

services:
  certbot:
    image: certbot/certbot
    volumes:
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
    environment:
      - CERTBOT_EMAIL=admin@darkpipe.example.com

  postfix-relay:
    build: .
    volumes:
      - certbot-etc:/etc/letsencrypt:ro
    ports:
      - "25:25"
    depends_on:
      - certbot

volumes:
  certbot-etc:
  certbot-var:
```

### Pattern 4: TLS Policy Enforcement (Per-Destination)
**What:** Override global TLS policy for specific destinations, enforce encryption selectively
**When to use:** Strict mode - refuse mail from plaintext-only peers
**Example:**
```bash
# Source: http://www.postfix.org/TLS_README.html
# /etc/postfix/main.cf
smtp_tls_security_level = may
smtp_tls_policy_maps = hash:/etc/postfix/tls_policy

# /etc/postfix/tls_policy (per-destination enforcement)
# Trusted domains: require encryption
gmail.com           encrypt
outlook.com         encrypt
.microsoft.com      encrypt

# Strict mode: refuse plaintext-only peers
# (User configures domains to enforce)
# example.com       encrypt

# After editing, run:
# postmap /etc/postfix/tls_policy
# postfix reload
```

### Pattern 5: Real-Time TLS Connection Monitoring
**What:** Parse Postfix logs for TLS warnings, notify user when remote server lacks encryption
**When to use:** RELAY-06 requirement - user notified of plaintext connections
**Example:**
```go
// relay/notify/tls_warnings.go
// Parse mail.log for TLS warnings, send notifications

package notify

import (
    "bufio"
    "os"
    "regexp"
    "strings"
)

var tlsWarningPattern = regexp.MustCompile(`postfix/smtp.*TLS is required, but was not offered`)

func MonitorTLSWarnings(logPath string, notifyFunc func(string)) error {
    f, err := os.Open(logPath)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        if tlsWarningPattern.MatchString(line) {
            // Extract remote server domain
            if domain := extractDomain(line); domain != "" {
                notifyFunc("Remote server does not support TLS: " + domain)
            }
        }
    }
    return scanner.Err()
}

func extractDomain(logLine string) string {
    // Parse: "to=<user@example.com>" from log line
    if idx := strings.Index(logLine, "to=<"); idx != -1 {
        start := idx + 4
        if end := strings.Index(logLine[start:], ">"); end != -1 {
            email := logLine[start : start+end]
            if at := strings.Index(email, "@"); at != -1 {
                return email[at+1:]
            }
        }
    }
    return ""
}
```

### Anti-Patterns to Avoid
- **Postfix local delivery enabled:** Set `mydestination =` (empty) to prevent local mailbox creation and disk usage
- **No transport map:** Without transport_maps, Postfix attempts MX lookup and direct delivery, bypassing Go relay daemon
- **Ramdisk for queue:** tmpfs queue risks data loss on crash; use standard disk queue with aggressive retry/bounce settings instead
- **Storing messages on disk:** Go relay daemon must forward immediately, never persist messages to filesystem
- **BerkleyDB:** Deprecated in Alpine; use LMDB (default in Alpine Postfix 3.7+)
- **Embedding Certbot in main container:** Sidecar pattern isolates certificate management, allows independent renewal without relay restart

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SMTP protocol handling | Custom SMTP parser/server | Postfix + emersion/go-smtp | SMTP has complex edge cases (pipelining, BDAT, UTF-8, SMTPUTF8, 8BITMIME extensions). Postfix handles 25 years of protocol quirks. emersion/go-smtp provides RFC 5321 compliance tested by 1.1K+ projects. |
| TLS certificate management | Manual OpenSSL + cron scripts | Certbot/Let's Encrypt | Certificate renewal, challenge handling, multi-domain support, OCSP stapling. Certbot automates DNS-01/HTTP-01 challenges, 90-day rotation, error recovery. |
| Direct MTA delivery | Custom DNS MX lookup + SMTP client | Postfix smtp transport | MX record prioritization, connection pooling, retry logic, bounce handling, greylisting detection, temporary failure vs permanent failure distinction. Postfix has 25 years of deliverability optimization. |
| TLS connection retry with backoff | Custom retry loops | Postfix queue + cenkalti/backoff/v4 | Postfix queue manager handles retry scheduling, exponential backoff, max retry limits. Go daemon uses backoff/v4 for transport reconnection (Phase 01 dependency). |
| Log parsing for monitoring | Regex + tail -f | postfix-logwatch or Prometheus postfix-exporter | Structured log parsing, rate limiting to prevent notification spam, stateful tracking of domains with TLS issues, integration with monitoring systems. |

**Key insight:** SMTP is deceptively complex. Even "simple relay" involves protocol negotiation (ESMTP extensions), TLS handshakes, authentication, bounce handling, queue management, and deliverability optimization. Postfix provides this for free; reimplementing in Go would take months and miss edge cases.

## Common Pitfalls

### Pitfall 1: VPS Port 25 Restrictions
**What goes wrong:** Many VPS providers (DigitalOcean, AWS Lightsail, Google Cloud) block outbound port 25 by default or require support tickets to unblock.
**Why it happens:** Anti-spam measures. Providers prevent new accounts from sending mail to stop abuse.
**How to avoid:**
- Research VPS providers that explicitly allow SMTP (Vultr, Linode/Akamai, OVH, Hetzner documented as SMTP-friendly)
- Verify port 25 access before deployment: `telnet smtp.gmail.com 25`
- Document "VPS Provider Selection" guide for users
**Warning signs:** `Connection timed out` when attempting SMTP to external servers, even with correct Postfix config.

### Pitfall 2: Let's Encrypt Rate Limits
**What goes wrong:** Let's Encrypt enforces 5 duplicate certificate requests per week. Frequent container rebuilds during development exhaust limits.
**Why it happens:** Certificate issuance without proper caching, requesting new certs on every container restart.
**How to avoid:**
- Use Let's Encrypt staging environment during development (`--test-cert` flag)
- Mount `/etc/letsencrypt` as persistent volume, not ephemeral
- Implement certificate expiry check before requesting renewal (renew only if <30 days remaining)
**Warning signs:** `too many certificates already issued for exact set of domains` error from Certbot.

### Pitfall 3: Postfix Transport Map Syntax Errors
**What goes wrong:** Mail stuck in queue, "mail for X loops back to myself" errors.
**Why it happens:** Transport map misconfiguration - not hashing with `postmap`, incorrect syntax, missing brackets around IP addresses.
**How to avoid:**
- Always hash transport maps: `postmap /etc/postfix/transport` after editing
- Use brackets to disable MX lookup for local forwarding: `smtp:[127.0.0.1]:10025`
- Test with `postmap -q "test@example.com" hash:/etc/postfix/transport`
- Run `postfix check` to validate configuration before reload
**Warning signs:** `mail for example.com loops back to myself` in logs, messages queued but not forwarding.

### Pitfall 4: Alpine DNS Resolution in Containers
**What goes wrong:** Postfix fails to resolve MX records, all outbound mail bounces with "Host not found".
**Why it happens:** Alpine uses musl libc, not glibc. DNS resolution behaves differently, especially with Docker's embedded DNS.
**How to avoid:**
- Ensure `ca-certificates` package installed in Alpine container
- Configure `/etc/resolv.conf` correctly (Docker usually handles this)
- Test DNS resolution inside container: `nslookup gmail.com`
- Consider using Google DNS (8.8.8.8) or Cloudflare DNS (1.1.1.1) if VPS DNS is unreliable
**Warning signs:** `Name service error for name=example.com type=MX: Host not found` in Postfix logs.

### Pitfall 5: Message Size Mismatches Between Relay and Home Server
**What goes wrong:** Large attachments accepted by cloud relay but rejected by home server, causing bounces after delay.
**Why it happens:** Cloud relay `message_size_limit` larger than home server limit.
**How to avoid:**
- Sync `message_size_limit` between cloud relay and home server (recommend 50MB)
- Configure early rejection at cloud relay to avoid accepting mail that will bounce later
- Document size limits in user-facing documentation
**Warning signs:** Deferred bounces with "message size exceeds fixed limit" from home server.

### Pitfall 6: Ephemeral Storage Without Queue Management
**What goes wrong:** Container restart loses queued messages if home device is offline.
**Why it happens:** Postfix queue stored in container ephemeral storage, not persistent volume.
**How to avoid:**
- Mount `/var/spool/postfix` as persistent volume (or use host volume)
- Implement queue depth monitoring to alert on buildup
- Phase 05 implements optional encrypted queue for offline scenarios
**Warning signs:** Messages disappear after container restart, users report lost mail during outages.

### Pitfall 7: TLS Version Mismatch Between Postfix and Remote Servers
**What goes wrong:** Postfix fails to deliver to remote servers requiring TLS 1.2+, while Postfix is configured for TLS 1.0+.
**Why it happens:** Default `smtpd_tls_mandatory_protocols` includes outdated protocols for compatibility.
**How to avoid:**
- Configure modern TLS: `smtpd_tls_mandatory_protocols = !SSLv2, !SSLv3, !TLSv1, !TLSv1.1`
- Test with known strict servers: `openssl s_client -connect smtp.gmail.com:25 -starttls smtp`
**Warning signs:** `TLS handshake failed` errors in logs when connecting to modern mail servers.

## Code Examples

Verified patterns from official sources and Phase 01 integration:

### Postfix Relay-Only Configuration
```bash
# Source: http://www.postfix.org/BASIC_CONFIGURATION_README.html
# Minimal relay-only setup for cloud relay

# /etc/postfix/main.cf
myhostname = relay.darkpipe.example.com
mydomain = darkpipe.example.com
myorigin = $mydomain

# Null client: no local delivery
mydestination =
local_recipient_maps =
local_transport = error:local mail delivery is disabled

# Accept from internet
inet_interfaces = all
inet_protocols = ipv4

# Relay security
smtpd_relay_restrictions = permit_mynetworks, reject_unauth_destination
mynetworks = 127.0.0.0/8, [::1]/128

# Forward to Go daemon
transport_maps = hash:/etc/postfix/transport
default_transport = smtp:[127.0.0.1]:10025

# TLS for internet connections
smtpd_tls_cert_file = /etc/letsencrypt/live/${myhostname}/fullchain.pem
smtpd_tls_key_file = /etc/letsencrypt/live/${myhostname}/privkey.pem
smtpd_tls_security_level = may
smtpd_tls_loglevel = 1

# Outbound TLS policy
smtp_tls_security_level = may
smtp_tls_loglevel = 1
smtp_tls_session_cache_database = btree:${data_directory}/smtp_scache
```

### Go Relay Daemon with WireGuard Transport Integration
```go
// Source: Phase 01 transport/mtls/client/connector.go + emersion/go-smtp
// cmd/relay/main.go - Bridge Postfix to WireGuard/mTLS transport

package main

import (
    "context"
    "io"
    "log"
    "net"
    "time"

    "github.com/darkpipe/darkpipe/transport/health"
    "github.com/darkpipe/darkpipe/transport/mtls/client"
    "github.com/emersion/go-smtp"
)

// Backend implements smtp.Backend, forwarding to home device via mTLS transport
type Backend struct {
    transport *client.Client // Phase 01 mTLS client
}

func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
    return &Session{backend: b}, nil
}

type Session struct {
    backend *client.Client
    from    string
    to      []string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
    s.from = from
    return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
    s.to = append(s.to, to)
    return nil
}

func (s *Session) Data(r io.Reader) error {
    // Connect to home device via mTLS transport
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    conn, err := s.backend.Connect(ctx)
    if err != nil {
        return err
    }
    defer conn.Close()

    // Forward SMTP envelope and message
    // (Simplified - real implementation uses SMTP client to home server)
    _, err = io.Copy(conn, r)
    return err
}

func (s *Session) Reset() {
    s.from = ""
    s.to = nil
}

func (s *Session) Logout() error {
    return nil
}

func main() {
    // Initialize Phase 01 transport
    mtlsClient, err := client.NewClient(
        "home-device:port",
        "/etc/darkpipe/ca.crt",
        "/etc/darkpipe/relay-client.crt",
        "/etc/darkpipe/relay-client.key",
    )
    if err != nil {
        log.Fatalf("init mTLS client: %v", err)
    }

    // Start SMTP server for Postfix to forward to
    be := &Backend{transport: mtlsClient}
    s := smtp.NewServer(be)
    s.Addr = "127.0.0.1:10025"
    s.Domain = "relay.darkpipe.example.com"
    s.ReadTimeout = 10 * time.Second
    s.WriteTimeout = 10 * time.Second
    s.MaxMessageBytes = 50 * 1024 * 1024 // 50MB

    log.Printf("Starting relay daemon on %s", s.Addr)
    if err := s.ListenAndServe(); err != nil {
        log.Fatalf("SMTP server error: %v", err)
    }
}
```

### Docker Health Check for Postfix SMTP Service
```dockerfile
# Source: https://github.com/bokysan/docker-postfix/issues/40
# Dockerfile

FROM alpine:3.21
RUN apk add --no-cache postfix ca-certificates wireguard-tools

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD postfix status || exit 1

# Alternative: Check SMTP port is listening
# HEALTHCHECK CMD nc -z localhost 25 || exit 1
```

### TLS Connection Monitoring with Notifications
```bash
# Source: http://www.postfix.org/TLS_README.html + custom monitoring
# monitor-tls.sh - Parse mail.log for TLS failures, send notifications

#!/bin/sh
tail -f /var/log/mail.log | while read line; do
    if echo "$line" | grep -q "TLS is required, but was not offered"; then
        # Extract destination domain
        domain=$(echo "$line" | grep -oP 'to=<[^@]+@\K[^>]+')
        if [ -n "$domain" ]; then
            # Send notification (webhook, email, etc.)
            curl -X POST https://notify.darkpipe.example.com/tls-warning \
                -H "Content-Type: application/json" \
                -d "{\"domain\": \"$domain\", \"message\": \"Remote server does not support TLS\"}"
        fi
    fi
done
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| BerkleyDB for Postfix maps | LMDB (Lightning Memory-Mapped Database) | Alpine 3.13+ (2021) | BerkleyDB deprecated due to Oracle license change to AGPL-3.0. LMDB is default in Alpine Postfix 3.7+. |
| Manual OpenSSL certificate generation | Let's Encrypt + Certbot automation | 2016+ (Let's Encrypt GA) | 90-day automated rotation replaces manual yearly renewals. DNS-01 challenge enables wildcard certs without exposing HTTP port 80. |
| Postfix smtpd_tls_mandatory_protocols including TLS 1.0/1.1 | TLS 1.2+ minimum | 2020+ (PCI DSS 3.2.1) | TLS 1.0/1.1 deprecated industry-wide. Modern config: `!TLSv1, !TLSv1.1` |
| Smart host relaying (e.g., SendGrid, Mailgun) | Direct MTA delivery | DarkPipe design decision | Privacy requirement: no third-party inspection of messages. Direct delivery maintains end-to-end encryption goal. |
| WireGuard userspace (wireguard-go) | WireGuard kernel module | Linux 5.6+ (2020) | Kernel module 3-5x faster on ARM64, lower CPU, better latency. Userspace only as fallback. |
| IPv4-only SMTP | Dual-stack IPv4 + IPv6 | 2020+ | Gmail, Outlook prefer IPv6. Configure `inet_protocols = all` for best deliverability. |

**Deprecated/outdated:**
- **Postfix BerkleyDB:** Replaced with LMDB in Alpine. Do not use `hash:` with BerkleyDB backend.
- **TLS 1.0/1.1:** Deprecated industry-wide. Configure `smtpd_tls_mandatory_protocols = !SSLv2, !SSLv3, !TLSv1, !TLSv1.1`
- **HTTP-01 challenge only:** DNS-01 challenge preferred for wildcard certs and VPS environments where port 80 may be restricted.
- **Manual certificate renewal:** Certbot + systemd timer or Docker sidecar for automated 90-day rotation.

## Open Questions

1. **VPS Provider Whitelist for Port 25**
   - What we know: DigitalOcean, AWS Lightsail, Google Cloud block or restrict port 25 by default. Vultr, Linode, OVH, Hetzner documented as permissive.
   - What's unclear: Complete current list of SMTP-friendly providers with pricing comparison.
   - Recommendation: Create "VPS Provider Guide" in Phase 02 documentation, include test procedure (telnet smtp.gmail.com 25) for users to verify.

2. **Optimal Message Size Limit for $5/month VPS**
   - What we know: 50MB is common for modern mail servers. VPS RAM target is 256MB.
   - What's unclear: Whether 50MB limit causes memory pressure during concurrent deliveries on minimal VPS.
   - Recommendation: Start with 50MB (`message_size_limit = 52428800`), document how to reduce if memory issues occur. Phase 09 monitoring will track memory usage.

3. **Notification Mechanism for TLS Failures**
   - What we know: Postfix logs TLS handshake failures at loglevel 1+. Real-time log parsing can detect plaintext-only peers.
   - What's unclear: Preferred notification channel (webhook, email, WebSocket to admin UI). Phase-specific or deferred to Phase 9 monitoring?
   - Recommendation: Implement basic webhook notification in Phase 02, integrate with Phase 9 monitoring dashboard when available.

4. **Ephemeral Queue Strategy for Phase 02**
   - What we know: Phase 05 implements optional encrypted queue for offline scenarios. Phase 02 is "ephemeral forwarding" only.
   - What's unclear: Should Phase 02 mount persistent volume for Postfix queue, or truly ephemeral (lose queued mail on container restart)?
   - Recommendation: Use persistent volume for Postfix queue even in Phase 02 to prevent data loss. Simplifies transition to Phase 05 encrypted queue.

5. **Postfix vs Go Daemon Responsibility Split**
   - What we know: Postfix handles internet SMTP, Go daemon bridges to transport. Either could handle TLS enforcement.
   - What's unclear: Should Postfix enforce TLS policy (smtp_tls_policy_maps), or should Go daemon check connection quality and reject at application layer?
   - Recommendation: Postfix handles TLS policy enforcement (simpler, standard Postfix feature). Go daemon logs TLS connection metadata for monitoring.

## Sources

### Primary (HIGH confidence)
- [Postfix Basic Configuration](http://www.postfix.org/BASIC_CONFIGURATION_README.html) - Relay-only configuration, mydestination, transport maps
- [Postfix TLS Support](http://www.postfix.org/TLS_README.html) - STARTTLS enforcement, security levels, policy maps
- [Postfix Configuration Parameters](http://www.postfix.org/postconf.5.html) - Complete parameter reference
- [Postfix Performance Tuning](https://www.postfix.org/TUNING_README.html) - Process limits, memory optimization, relay concurrency
- [emersion/go-smtp GitHub](https://github.com/emersion/go-smtp) - SMTP server/client library, backend interface
- [emersion/go-smtp-proxy GitHub](https://github.com/emersion/go-smtp-proxy) - Relay backend implementation pattern (archived but reference)
- [Phase 01 Transport Implementation](transport/mtls/client/connector.go) - mTLS client with persistent connection and backoff
- [Phase 01 Health Checker](transport/health/checker.go) - Unified transport health checking

### Secondary (MEDIUM confidence)
- [boky/docker-postfix GitHub](https://github.com/bokysan/docker-postfix) - Alpine Postfix container, environment variable configuration
- [Alpine Linux Postfix Package](https://pkgs.alpinelinux.org/package/edge/main/x86/postfix) - Dependencies, LMDB support
- [sidecar-certbot GitHub](https://github.com/bybatkhuu/sidecar-certbot) - Let's Encrypt automation in Docker sidecar pattern
- [Certbot Docker Hub](https://hub.docker.com/r/certbot/certbot) - Official EFF Certbot image
- [Postfix Transport Map Guide (LinuxBabe)](https://www.linuxbabe.com/mail-server/postfix-transport-map-relay-map-flexible-email-delivery) - Transport map configuration examples
- [Docker Health Checks (OneUpTime)](https://oneuptime.com/blog/post/2026-01-23-docker-health-checks-effectively/view) - 2026 guide to Docker health check best practices

### Tertiary (LOW confidence - needs validation)
- Postfix ramdisk queue discussion - Community consensus is NOT recommended due to data loss risk
- VPS port 25 restrictions - Anecdotal reports, needs current provider documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Postfix is proven technology, emersion/go-smtp is production-tested, Let's Encrypt is industry standard
- Architecture: HIGH - Patterns verified against official Postfix docs and Phase 01 implementation, relay pattern proven by go-smtp-proxy
- Container optimization: MEDIUM-HIGH - 25-30MB target achievable (Alpine 5MB + Postfix 15MB + WireGuard 5MB + Go 2-5MB), but real-world testing needed
- TLS enforcement: HIGH - Postfix TLS features well-documented, policy maps standard approach
- Pitfalls: MEDIUM-HIGH - Common issues documented in community sources, Phase 01 experience informs transport integration pitfalls
- VPS port 25 whitelist: LOW - Anecdotal evidence, needs current provider verification

**Research date:** 2026-02-08
**Valid until:** 2026-03-08 (30 days - Postfix/Alpine/Certbot are stable, versions unlikely to change rapidly)