# Phase 03: Home Mail Server - Research

**Researched:** 2026-02-08
**Domain:** Home Mail Server (IMAP, SMTP Submission, Multi-User, Multi-Domain, Spam Filtering)
**Confidence:** HIGH

## Summary

Phase 03 implements the home mail server that receives mail from the cloud relay (via WireGuard/mTLS transport established in Phases 1-2) and provides IMAP/SMTP submission services to mail clients. The critical architectural decision is **delivery mechanism**: the Go relay daemon should deliver to the home mail server via **SMTP to port 25** (standard MTA relay), not LMTP, to preserve flexibility for all three mail server options (Stalwart, Maddy, Postfix+Dovecot).

All three mail server options (Stalwart, Maddy, Postfix+Dovecot) support multi-user, multi-domain, aliases, catch-all, and Rspamd integration. Stalwart is pre-v1.0 (v1.0 expected Q2 2026) but production-ready. Maddy is beta but stable. Postfix+Dovecot is battle-tested but requires more configuration.

**Primary recommendation:** Use **SMTP delivery to port 25** on home mail server. Implement mail server selection as docker-compose profiles or build-time selection (Phase 7 build system will formalize this). Default to **Stalwart** for modern stack, **Maddy** for minimal resources, **Postfix+Dovecot** for maximum compatibility.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **Stalwart Mail Server** | 0.15.4 (v1.0 Q2 2026) | All-in-one mail server (IMAP4rev2, SMTP, JMAP) | Rust, single binary, built-in CalDAV/CardDAV, memory-safe, REST API, production use globally, v1.0 finalizes schema |
| **Maddy Mail Server** | 0.8.2 | All-in-one Go mail server | Single 15MB binary, ~15MB RAM, replaces Postfix+Dovecot+OpenDKIM, simple config, beta but stable |
| **Postfix + Dovecot** | Postfix 3.7+, Dovecot 2.3+ | Traditional split MTA+IMAP | 20+ years production hardening, maximum flexibility, most documentation, ARM64 proven |
| **Rspamd** | Latest | Spam filtering and greylisting | Industry standard, integrates with all three mail servers via milter (Stalwart/Postfix) or LDA mode (Maddy/Dovecot) |
| **emersion/go-smtp** | 0.24.0 | SMTP client library for Go | Already in use by cloud relay daemon, will extend for home-side SMTP delivery, RFC 5321 compliant, 1.1K+ dependents |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **Redis** | Latest | Rspamd backend for greylisting | Required by Rspamd greylisting module, stores greylist hashes with TTL |
| **Let's Encrypt (optional)** | via Certbot | TLS certs for IMAP/SMTP if publicly accessible | Only if home device has public IP and DNS hostname |

### Mail Server Comparison

| Feature | Stalwart | Maddy | Postfix+Dovecot |
|---------|----------|-------|-----------------|
| **Binary Size** | ~50MB | ~15MB | ~80MB combined |
| **Memory Use** | ~50MB | ~15MB | ~100MB combined |
| **IMAP** | IMAP4rev2, IMAP4rev1 | IMAP4rev1 | IMAP4rev1 (Dovecot) |
| **Multi-User** | Yes (internal directory or LDAP/SQL) | Yes (virtual users by default) | Yes (virtual users) |
| **Multi-Domain** | Yes (multi-tenancy with isolation) | Yes (default: user@domain as account ID) | Yes (virtual mailbox domains) |
| **Aliases/Catch-All** | Yes (built-in) | Yes (configuration) | Yes (virtual_alias_maps) |
| **CalDAV/CardDAV** | Built-in | No | Requires Radicale/Baikal |
| **JMAP** | Yes | No | No |
| **Rspamd Integration** | Milter protocol | LDA mode or milter | Milter protocol |
| **Configuration** | Web UI + config files | Single config file | Multiple files (main.cf, dovecot.conf) |
| **Maturity** | Pre-v1.0 (production use globally) | Beta (use with caution) | 20+ years production |
| **ARM64 Support** | Excellent (official Docker multi-arch) | Excellent (Go native) | Excellent (standard packages) |

**Installation:**

```bash
# Stalwart (Docker)
docker pull stalwartlabs/stalwart:0.15.4

# Maddy (Docker)
docker pull foxcpp/maddy:0.8.2

# Postfix+Dovecot (Alpine packages)
apk add postfix dovecot

# Rspamd (all platforms)
docker pull rspamd/rspamd:latest
# Requires Redis:
docker pull redis:alpine
```

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SMTP delivery to port 25 | LMTP delivery | LMTP is faster (no queue) and provides per-recipient responses, but not universally supported (Stalwart supports SMTP+LMTP, Maddy supports LMTP endpoint, Postfix+Dovecot standard is LMTP to Dovecot). SMTP to port 25 is most flexible for user-selectable mail server. |
| Stalwart | Maddy | Maddy has smaller footprint (~15MB vs ~50MB), simpler if only need email (no calendar), Go codebase. Choose if resource-constrained or prefer Go. |
| Stalwart | Postfix+Dovecot | Maximum compatibility, 20+ years production hardening, most documentation/examples. Choose if need complex routing or existing team expertise. |
| Rspamd | SpamAssassin | SpamAssassin is older, Perl-based, slower, less active development. Rspamd is modern (C++), faster, active, milter support. |

## Architecture Patterns

### Recommended Project Structure

```
home-device/
├── mail-server/
│   ├── stalwart/              # Stalwart option
│   │   ├── Dockerfile
│   │   └── stalwart.toml      # Configuration
│   ├── maddy/                 # Maddy option
│   │   ├── Dockerfile
│   │   └── maddy.conf         # Configuration
│   └── postfix-dovecot/       # Postfix+Dovecot option
│       ├── Dockerfile
│       ├── postfix/           # main.cf, master.cf
│       └── dovecot/           # dovecot.conf
├── spam-filter/
│   ├── rspamd/
│   │   ├── local.d/
│   │   │   ├── greylist.conf  # Greylisting config
│   │   │   └── worker-proxy.conf  # Milter settings
│   │   └── override.d/
│   └── redis/
│       └── redis.conf
├── transport/
│   ├── wireguard/             # WireGuard peer config (receives from relay)
│   └── smtp-receiver/         # Listens on WireGuard IP:25
└── docker-compose.yml         # Orchestrates selected mail server + Rspamd + Redis
```

### Pattern 1: SMTP Delivery from Cloud Relay to Home Mail Server

**What:** Go relay daemon delivers mail via SMTP to home mail server's port 25 over WireGuard/mTLS transport.

**When to use:** When supporting user-selectable mail servers (Stalwart, Maddy, or Postfix+Dovecot).

**Example (Go relay daemon extending existing Forward interface):**

```go
// Forward sends mail to home device SMTP server via WireGuard/mTLS
func (f *WireGuardForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
    // home device listens on WireGuard IP (e.g., 10.0.0.2:25)
    // emersion/go-smtp client already in use
    client, err := smtp.Dial(f.homeDeviceAddr) // e.g., "10.0.0.2:25"
    if err != nil {
        return fmt.Errorf("SMTP dial failed: %w", err)
    }
    defer client.Close()

    if err := client.Mail(from, nil); err != nil {
        return fmt.Errorf("MAIL FROM failed: %w", err)
    }

    for _, addr := range to {
        if err := client.Rcpt(addr, nil); err != nil {
            return fmt.Errorf("RCPT TO failed: %w", err)
        }
    }

    wc, err := client.Data()
    if err != nil {
        return fmt.Errorf("DATA failed: %w", err)
    }

    if _, err := io.Copy(wc, data); err != nil {
        return fmt.Errorf("message write failed: %w", err)
    }

    return wc.Close()
}
```

### Pattern 2: Multi-User Virtual Mailboxes

**What:** Each user has separate mailbox with independent credentials, no UNIX system users required.

**When to use:** All deployments (virtual users are standard for modern mail servers).

**Stalwart example:**
```toml
# stalwart.toml - internal directory
[directory.internal]
type = "internal"
store = "sqlite"

# Create users via REST API or web UI
# POST /api/v1/account
# { "email": "user@example.com", "password": "...", "quota": "10GB" }
```

**Maddy example:**
```
# maddy.conf - default behavior uses email as account ID
# user@example.org and user@example.com are separate accounts
$(local_domains) = example.org example.com

storage.imapsql local_mailboxes {
    driver sqlite3
    dsn imapsql.db
}

table.sqlite3 local_accounts {
    file /etc/maddy/accounts.db
    # Schema: email (username), password, (optional: name, quota)
}
```

**Postfix+Dovecot example:**
```
# Postfix main.cf
virtual_mailbox_domains = example.org, example.com
virtual_mailbox_base = /var/mail/vhosts
virtual_mailbox_maps = hash:/etc/postfix/vmailbox

# /etc/postfix/vmailbox
user@example.org  example.org/user/
admin@example.com example.com/admin/

# Dovecot dovecot.conf
passdb {
  driver = passwd-file
  args = /etc/dovecot/users
}
userdb {
  driver = static
  args = uid=vmail gid=vmail home=/var/mail/vhosts/%d/%n
}
```

### Pattern 3: Aliases and Catch-All Addresses

**What:** Mail sent to alias@domain delivers to configured mailbox. Catch-all delivers all undefined addresses to single mailbox.

**When to use:** Multi-domain setups where users want single mailbox to receive mail for multiple addresses.

**Stalwart example:**
```toml
# stalwart.toml
[session.rcpt]
# Aliases via rewrite rules
relay = [ { if = "rcpt_domain", in-list = "sql/domains", then = "alias_map" } ]

[session.rcpt.rewrite]
# admin@example.org -> user@example.org
alias_map = "sql/aliases"

# Catch-all: *@example.org -> catchall@example.org
catch_all = [ { if = "rcpt_domain", eq = "example.org", then = "catchall@example.org" } ]
```

**Maddy example:**
```
# maddy.conf
table.file aliases {
    file /etc/maddy/aliases
}

# /etc/maddy/aliases
admin@example.org: user@example.org
support@example.org: user@example.org

# Catch-all in destination modifiers
destination postfix $(local_domains) {
    modify {
        # If address not found, deliver to catch-all
        replace_rcpt file /etc/maddy/catch-all
    }
    deliver_to &local_mailboxes
}
```

**Postfix+Dovecot example:**
```
# Postfix main.cf
virtual_alias_maps = hash:/etc/postfix/virtual

# /etc/postfix/virtual
# Aliases
admin@example.org   user@example.org
support@example.org user@example.org

# Catch-all for example.org
@example.org        catchall@example.org
```

**WARNING:** Catch-all addresses prevent early rejection of spam. Mail server must process all mail for domain, increasing load. Best practice: create specific aliases as needed instead of catch-all.

### Pattern 4: Rspamd Integration (Spam Filtering + Greylisting)

**What:** Rspamd scans incoming mail, assigns spam score, applies greylisting (temporary rejection for unknown senders).

**When to use:** All production deployments to reduce spam and abuse.

**Rspamd greylisting config:**
```
# local.d/greylist.conf
# Requires Redis backend
servers = "redis:6379";
timeout = 300;        # 5 minutes delay
expire = 86400;       # 24 hour expiry
greylist_min_score = 4.0;  # Only greylist messages with score >= 4.0

# Whitelist trusted networks (don't greylist internal mail)
whitelisted_ip = "/etc/rspamd/whitelist_ip.map";
# Example whitelist: 10.0.0.0/8, 192.168.0.0/16

# Don't greylist authenticated users or local networks
check_local = false;
check_authed = false;
```

**Stalwart + Rspamd (milter mode):**
```toml
# stalwart.toml
[session.data.milter."rspamd"]
enable = true
hostname = "rspamd"
port = 11332
timeout = "10s"
tempfail-on-error = true  # Reject on Rspamd failure (safe mode)
```

**Maddy + Rspamd (LDA mode or milter):**
```
# maddy.conf - LDA mode example
modify {
    # Call rspamc to add spam headers
    command /usr/bin/rspamc --mime {message_file}
}

# Or milter mode (requires Rspamd proxy worker in milter mode)
check {
    milter tcp://rspamd:11332
}
```

**Postfix + Rspamd (milter mode - recommended):**
```
# Postfix main.cf
smtpd_milters = inet:rspamd:11332
milter_default_action = accept  # or tempfail for strict mode
```

**Rspamd proxy worker config (milter mode):**
```
# local.d/worker-proxy.conf
milter = yes;
bind_socket = "*:11332";
upstream "local" {
  default = yes;
  self_scan = yes;
}
```

### Anti-Patterns to Avoid

- **Don't mix virtual_alias_domains and virtual_mailbox_domains in Postfix:** A domain cannot be both. Use virtual_mailbox_domains for final delivery, virtual_alias_domains only for forwarding to other domains.

- **Don't use catch-all without spam filtering:** Catch-all accepts all mail for domain, including spam and dictionary attacks. ALWAYS enable Rspamd + greylisting if using catch-all.

- **Don't configure IMAP/SMTP on public internet without TLS:** If home device has public IP, enforce TLS 1.2+ and use Let's Encrypt certificates. Never expose plaintext IMAP (port 143) or SMTP (port 25) without STARTTLS.

- **Don't use same hostname for mail.domain.com across multiple domains without proper certificates:** Each domain should have valid TLS certificate. Use wildcard cert or SAN certificate if serving mail.example.org and mail.example.com from same server.

- **Don't configure Postfix virtual_transport to skip Rspamd:** Always route inbound mail through spam filter before final delivery. Correct flow: Postfix SMTP (port 25) → Rspamd (milter) → Postfix virtual delivery OR Dovecot LMTP.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Spam filtering | Custom Bayesian filter, regex rules | Rspamd | Spam filtering requires constant rule updates, machine learning, DNS blacklist integration, phishing detection. Rspamd has dedicated team, daily rule updates, neural network module, comprehensive RBL support. |
| Greylisting | Custom temporary rejection logic | Rspamd greylisting module | Greylisting requires persistent state (Redis), per-triplet tracking (IP, sender, recipient), whitelisting logic, TTL management. Rspamd handles all edge cases (IPv6, forwarding, retry detection). |
| IMAP server | Custom IMAP4 protocol implementation | Dovecot or Stalwart or Maddy | IMAP4rev1 has 50+ extensions (IDLE, COMPRESS, NOTIFY, QRESYNC), complex folder subscription, concurrent session handling, mailbox locking. Dovecot has 20+ years of edge case handling. |
| SMTP server | Custom SMTP protocol stack | Postfix or Stalwart or Maddy | SMTP has extensive extensions (ESMTP, PIPELINING, CHUNKING, SIZE, DSN), queue management, retry logic, bounce handling, TLS negotiation. Postfix is battle-tested MTA used by millions. |
| Virtual user database | Custom auth system | Mail server built-in directory or LDAP | Virtual user management requires secure password hashing (bcrypt/Argon2), quota tracking, per-user settings, mailbox provisioning. Stalwart/Maddy have built-in user management with REST API. |
| Multi-domain routing | Custom domain lookup and routing | Mail server virtual domain config | Multi-domain requires per-domain transport maps, catch-all handling, domain-specific quotas, DNS verification. All modern mail servers handle this natively. |

**Key insight:** Email protocols (SMTP, IMAP) are deceptively complex with decades of RFCs and extensions. Use battle-tested implementations (Postfix, Dovecot) or modern all-in-one servers (Stalwart, Maddy). Never build custom mail server components.

## Common Pitfalls

### Pitfall 1: LMTP vs SMTP Delivery from Cloud Relay

**What goes wrong:** Using LMTP for delivery from cloud relay to home mail server works for some configurations (Postfix can receive LMTP) but creates tight coupling and breaks Stalwart (which expects SMTP on port 25 for MX delivery).

**Why it happens:** LMTP is more efficient (no queue, per-recipient responses) and standard for local delivery (Postfix → Dovecot). Developers assume it's best for all mail delivery.

**How to avoid:** Use **SMTP to port 25** for cloud relay → home mail server delivery. This is the MX delivery pattern, works with all three mail server options, and preserves standard mail server behavior.

**Warning signs:** If planning to use LMTP, verify compatibility with ALL mail server options (Stalwart, Maddy, Postfix+Dovecot).

### Pitfall 2: Virtual Alias Domains vs Virtual Mailbox Domains (Postfix)

**What goes wrong:** Adding domain to both `virtual_alias_domains` and `virtual_mailbox_domains` in Postfix causes "mail loops back to myself" errors or delivery failures.

**Why it happens:** Postfix address classes are exclusive. A domain is EITHER alias (forwarding to other domains) OR mailbox (final destination), never both.

**How to avoid:** Use `virtual_mailbox_domains` for domains where you store mail. Use `virtual_alias_domains` only for domains that forward to other domains. For aliases within a virtual mailbox domain, use `virtual_alias_maps` instead.

**Warning signs:** Postfix log shows "mail for <domain> loops back to myself" or "User unknown in virtual mailbox table" for addresses that should exist.

**Correct pattern:**
```
# main.cf
virtual_mailbox_domains = example.org, example.com
virtual_mailbox_maps = hash:/etc/postfix/vmailbox
virtual_alias_maps = hash:/etc/postfix/virtual  # Aliases WITHIN virtual mailbox domains

# /etc/postfix/vmailbox (actual mailboxes)
user@example.org  example.org/user/

# /etc/postfix/virtual (aliases WITHIN example.org)
admin@example.org  user@example.org
```

### Pitfall 3: Catch-All Without Rate Limiting

**What goes wrong:** Catch-all addresses accept all mail for domain, including spam and dictionary attacks. Without rate limiting, mail server becomes spam processing machine, consuming CPU/disk/bandwidth.

**Why it happens:** Users want convenience of "any email to my domain reaches me" without understanding spam implications.

**How to avoid:**
1. Enable Rspamd + greylisting BEFORE enabling catch-all
2. Implement rate limiting (max messages per hour from unknown senders)
3. Monitor catch-all mailbox size and spam score distribution
4. Prefer specific aliases over catch-all when possible

**Warning signs:** Catch-all mailbox receives 100+ spam messages per day, disk space fills rapidly, Rspamd processing queue backs up.

### Pitfall 4: TLS Certificate Hostname Mismatch for IMAP/SMTP

**What goes wrong:** Mail client shows "certificate name mismatch" or "cannot verify server identity" errors when connecting to IMAP/SMTP.

**Why it happens:** TLS certificate CN or SAN doesn't match hostname used by client (e.g., certificate is for `mail.example.org` but client connects to `example.org` or bare IP).

**How to avoid:**
1. If home device has public DNS hostname, get Let's Encrypt certificate for that hostname
2. Ensure all mail clients use SAME hostname as certificate (e.g., mail.example.org)
3. Use wildcard certificate (`*.example.org`) if supporting multiple domains
4. For local-only access (no public DNS), use self-signed cert and instruct clients to accept it (less secure)

**Warning signs:** iOS Mail shows "Cannot Verify Server Identity", Thunderbird shows "Add Security Exception" prompt, K-9 Mail shows "Certificate error".

### Pitfall 5: Stalwart Pre-v1.0 Schema Changes

**What goes wrong:** Upgrading Stalwart from 0.14.x to 0.15.x or to v1.0 (Q2 2026) may require database schema migration. Automated migration may fail if custom schema modifications exist.

**Why it happens:** Stalwart is pre-v1.0 and schema is not yet finalized. Breaking changes possible until v1.0.

**How to avoid:**
1. Read upgrade notes before upgrading Stalwart (especially 0.14 → 0.15 and pre-v1.0 → v1.0)
2. Backup Stalwart database before upgrades (`/opt/stalwart-mail/data`)
3. Test upgrade on non-production instance first
4. Wait for v1.0 (Q2 2026) if schema stability is critical
5. After v1.0, Stalwart promises no more database migrations

**Warning signs:** Stalwart fails to start after upgrade with "database schema mismatch" or "migration failed" errors.

### Pitfall 6: Maddy Beta Status

**What goes wrong:** Maddy developers explicitly state "should currently be regarded as a beta product" and advise against production use. Edge cases may exist, data loss possible.

**Why it happens:** Maddy is younger project (vs Postfix/Dovecot with 20+ years), smaller user base, less battle-testing.

**How to avoid:**
1. Use Maddy only for non-critical email or with external backups
2. Monitor Maddy GitHub issues for bug reports before upgrading
3. Test thoroughly before production use
4. Prefer Stalwart (larger community, more production use) or Postfix+Dovecot (maximum maturity) if stability is critical

**Warning signs:** Maddy GitHub issues show database corruption, mail loss, or crash reports.

### Pitfall 7: Rspamd Greylisting Without Redis

**What goes wrong:** Rspamd greylisting module fails to start or doesn't greylist any mail.

**Why it happens:** Greylisting requires persistent state (greylist triplets: sender IP + from + to). Rspamd stores this in Redis.

**How to avoid:** Always run Redis alongside Rspamd if enabling greylisting. Configure Rspamd greylisting module with `servers = "redis:6379"`.

**Warning signs:** Rspamd logs show "failed to connect to Redis" or greylisting module is disabled.

### Pitfall 8: Multi-Domain Without Proper DNS

**What goes wrong:** Mail for second domain (example.com) bounces even though mail server is configured for multiple domains.

**Why it happens:** DNS MX records point to wrong server or missing entirely. Remote MTAs can't find where to deliver mail.

**How to avoid:**
1. Configure MX record for EACH domain pointing to cloud relay's public hostname
2. Verify MX records with `dig MX example.com` before testing
3. Add ALL domains to mail server's `$(local_domains)` (Maddy) or `virtual_mailbox_domains` (Postfix) or Stalwart domain list
4. Test delivery to each domain separately

**Warning signs:** Remote MTAs bounce with "Name or service not known" or "Host not found" errors. Mail delivery works for first domain but not others.

## Code Examples

Verified patterns from official sources:

### SMTP Delivery from Go Relay Daemon to Home Mail Server

```go
// Source: emersion/go-smtp library (https://github.com/emersion/go-smtp)
// Extends existing cloud-relay/relay/forward/wireguard.go

import (
    "context"
    "fmt"
    "io"
    "net/smtp"

    gosmtp "github.com/emersion/go-smtp"
)

// Forward delivers mail to home device SMTP server via WireGuard/mTLS transport
// homeDeviceAddr example: "10.0.0.2:25" (WireGuard peer IP)
func (f *WireGuardForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
    // Dial home mail server SMTP port 25 over WireGuard tunnel
    client, err := gosmtp.Dial(f.homeDeviceAddr)
    if err != nil {
        return fmt.Errorf("SMTP dial to %s failed: %w", f.homeDeviceAddr, err)
    }
    defer client.Close()

    // Send MAIL FROM
    if err := client.Mail(from, nil); err != nil {
        return fmt.Errorf("MAIL FROM %s failed: %w", from, err)
    }

    // Send RCPT TO for each recipient
    for _, addr := range to {
        if err := client.Rcpt(addr, nil); err != nil {
            return fmt.Errorf("RCPT TO %s failed: %w", addr, err)
        }
    }

    // Send DATA
    wc, err := client.Data()
    if err != nil {
        return fmt.Errorf("DATA command failed: %w", err)
    }

    // Copy message body
    if _, err := io.Copy(wc, data); err != nil {
        wc.Close()
        return fmt.Errorf("message write failed: %w", err)
    }

    // Close DATA, commits message
    if err := wc.Close(); err != nil {
        return fmt.Errorf("message commit failed: %w", err)
    }

    return nil
}
```

### Rspamd Greylisting Configuration

```
# Source: Rspamd official documentation (https://rspamd.com/doc/modules/greylisting.html)
# File: /etc/rspamd/local.d/greylist.conf

# Redis backend (required)
servers = "redis:6379";

# Greylisting parameters
timeout = 300;              # Delay for 5 minutes (standard retry interval)
expire = 86400;             # Expire greylist entries after 24 hours
greylist_min_score = 4.0;   # Only greylist messages with spam score >= 4.0

# Whitelist configuration
whitelisted_ip = "/etc/rspamd/whitelist_ip.map";    # IP/CIDR whitelist
whitelist_domains_url = "/etc/rspamd/whitelist_domains.map";  # Domain whitelist

# Skip greylisting for these cases
check_local = false;    # Don't greylist mail from local networks
check_authed = false;   # Don't greylist authenticated users

# Response message (sent to remote MTA)
message = "Try again later";  # 4xx response message
report_time = true;           # Tell when greylisting expires
```

### Postfix Virtual Mailbox Domains with Dovecot LMTP

```
# Source: Dovecot official documentation (https://doc.dovecot.org/main/howto/lmtp/postfix.html)
# File: /etc/postfix/main.cf

# Virtual mailbox domains (final destination)
virtual_mailbox_domains = example.org, example.com
virtual_transport = lmtp:unix:private/dovecot-lmtp

# Virtual aliases WITHIN virtual mailbox domains
virtual_alias_maps = hash:/etc/postfix/virtual

# File: /etc/dovecot/dovecot.conf
protocols = imap lmtp

service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    group = postfix
    mode = 0600
    user = postfix
  }
}

protocol lmtp {
  postmaster_address = postmaster@example.org
  mail_plugins = $mail_plugins sieve quota
}
```

### Maddy Multi-Domain Virtual Users

```
# Source: Maddy official documentation (https://maddy.email/multiple-domains/)
# File: /etc/maddy/maddy.conf

# Local domains (final destination)
$(local_domains) = example.org example.com

# Virtual users stored in SQLite
table.sqlite3 local_accounts {
    file /etc/maddy/accounts.db
    # Accounts are email addresses: user@example.org, admin@example.com
}

auth.pass_table local_authdb {
    table local_accounts
}

storage.imapsql local_mailboxes {
    driver sqlite3
    dsn imapsql.db
}

# IMAP endpoint
imap tcp://0.0.0.0:993 tls://cert.pem,key.pem {
    auth &local_authdb
    storage &local_mailboxes
}

# SMTP submission endpoint (port 587)
submission tcp://0.0.0.0:587 tls://cert.pem,key.pem {
    auth &local_authdb
    deliver_to &local_mailboxes
}

# Inbound SMTP endpoint (port 25 - receives from cloud relay)
smtp tcp://0.0.0.0:25 {
    destination $(local_domains) {
        deliver_to &local_mailboxes
    }
}
```

### Stalwart Multi-User Configuration

```toml
# Source: Stalwart official documentation (https://stalw.art/docs/)
# File: /opt/stalwart-mail/etc/config.toml

[server.listener."smtp"]
bind = ["0.0.0.0:25"]
protocol = "smtp"

[server.listener."submission"]
bind = ["0.0.0.0:587"]
protocol = "smtp"
tls.implicit = false

[server.listener."imap"]
bind = ["0.0.0.0:993"]
protocol = "imap"
tls.implicit = true

[directory.internal]
type = "internal"
store = "sqlite"

[session.rcpt]
# Relay for local domains only
relay = [ { if = "rcpt_domain", in-list = "sql/domains" } ]

# Catch-all example
catch-all = [ { if = "rcpt_domain", eq = "example.org", then = "catchall@example.org" } ]

# Multi-tenancy with domain isolation
[storage]
data = "sqlite"
blob = "fs"

[storage.data]
path = "/opt/stalwart-mail/data/index.db"

[storage.blob]
path = "/opt/stalwart-mail/data/blobs"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Separate MTA + MDA + Auth + Webmail stack (5+ containers) | All-in-one mail server (Stalwart, Maddy) | 2020-2024 | Single binary reduces complexity, memory usage, configuration drift. Trade-off: less flexibility, newer/less battle-tested. |
| SpamAssassin for spam filtering | Rspamd for spam filtering + greylisting | 2015-2020 | Rspamd is faster (C++ vs Perl), more active development, milter integration standard. SpamAssassin still maintained but slower innovation. |
| Berkeley DB for Postfix maps | LMDB for Postfix maps | Alpine 3.13+ (2021) | Oracle changed Berkeley DB license to AGPL-3.0, incompatible with Alpine. LMDB is default replacement, API-compatible. |
| IMAP4rev1 | IMAP4rev2 (Stalwart) | RFC 9051 (2021) | IMAP4rev2 simplifies obsolete features, adds SAVEDATE, clarifies METADATA. Backward-compatible with rev1. Only Stalwart implements rev2 currently. |
| IMAP/SMTP only | JMAP + IMAP/SMTP | JMAP RFC 8620 (2019), Stalwart 2023+ | JMAP is modern alternative to IMAP with better mobile support, efficient sync. Stalwart is only open-source JMAP mail server. |
| Manual user management | REST API for user management (Stalwart) | 2023+ | Stalwart provides REST API and web UI for account/domain management, eliminating manual config file editing. |

**Deprecated/outdated:**
- **RainLoop webmail**: Abandoned in 2020, forked to SnappyMail (active). Use SnappyMail instead.
- **Postfix Berkeley DB maps**: Deprecated in Alpine 3.13+. Use LMTP maps (default).
- **Dovecot dict auth with passwd-file**: Use SQL backend for production (passwd-file for testing only).
- **Greylisting with postgrey**: Use Rspamd greylisting module for better integration and performance.

## Open Questions

1. **Should home mail server run on same container network as cloud relay daemon or separate?**
   - What we know: Cloud relay Go daemon needs to reach home mail server SMTP port 25. Can be same Docker network or separate with WireGuard overlay.
   - What's unclear: Performance implications of Docker network vs WireGuard overlay for local delivery.
   - Recommendation: Use WireGuard overlay even for local delivery to match production architecture and test transport layer resilience.

2. **How to handle mail server selection in build system?**
   - What we know: Phase 7 (Build System) will implement GitHub Actions with user-selectable components. Phase 3 needs intermediate solution.
   - What's unclear: Should Phase 3 implement docker-compose profiles (select at runtime) or multiple docker-compose files (select at deploy time)?
   - Recommendation: Use docker-compose profiles for Phase 3 interim solution. Phase 7 will formalize as build-time selection via GitHub Actions.

3. **Should Rspamd run as sidecar to each mail server or shared service?**
   - What we know: Rspamd integrates via milter protocol (network connection). Can be dedicated container or embedded.
   - What's unclear: Resource implications of shared Rspamd instance vs per-mail-server instance.
   - Recommendation: Shared Rspamd + Redis service for Phase 3. Reduces memory usage (single Redis, single Rspamd) and simplifies configuration.

4. **What is the upgrade path from Stalwart 0.15.4 to v1.0?**
   - What we know: v1.0 expected Q2 2026, will finalize database schema, eliminate future migrations.
   - What's unclear: Will 0.15.4 → v1.0 migration be automatic or require manual intervention?
   - Recommendation: Document upgrade procedure in Phase 3 docs. Test v1.0 upgrade in development environment before production rollout. Monitor Stalwart GitHub releases for upgrade notes.

## Sources

### Primary (HIGH confidence)

**Mail Server Official Documentation:**
- [Stalwart Mail Server](https://stalw.art/) - Official site, features overview
- [Stalwart GitHub](https://github.com/stalwartlabs/stalwart) - Source code, releases, issues
- [Stalwart Roadmap](https://stalw.art/blog/roadmap/) - v1.0 timeline (Q2 2026), post-v1.0 plans
- [Maddy Mail Server](https://maddy.email/) - Official documentation
- [Maddy GitHub](https://github.com/foxcpp/maddy) - Source code, releases, beta status warning
- [Maddy Multiple Domains](https://maddy.email/multiple-domains/) - Multi-domain configuration
- [Maddy SMTP/LMTP Endpoint](https://maddy.email/reference/endpoints/smtp/) - SMTP and LMTP support
- [Dovecot Postfix LMTP](https://doc.dovecot.org/main/howto/lmtp/postfix.html) - Postfix → Dovecot LMTP integration
- [Dovecot Virtual Users](https://doc.dovecot.org/main/core/config/auth/users/virtual.html) - Virtual user configuration
- [Postfix Virtual Domain Hosting](http://www.postfix.org/VIRTUAL_README.html) - Virtual mailbox vs virtual alias domains

**Rspamd Official Documentation:**
- [Rspamd Greylisting Module](https://rspamd.com/doc/modules/greylisting.html) - Greylisting configuration, Redis requirement
- [Rspamd MTA Integration](https://docs.rspamd.com/tutorials/integration/) - Milter and LDA integration methods
- [Rspamd Quickstart](https://rspamd.com/doc/tutorials/quickstart.html) - Installation and basic config

**Go Libraries:**
- [emersion/go-smtp GitHub](https://github.com/emersion/go-smtp) - SMTP client/server library, LMTP support, v0.24.0
- [emersion/go-imap GitHub](https://github.com/emersion/go-imap) - IMAP client/server library (for future IMAP integration)

**Protocol Specifications:**
- [RFC 5321 - SMTP](https://datatracker.ietf.org/doc/html/rfc5321) - SMTP protocol
- [RFC 2033 - LMTP](https://datatracker.ietf.org/doc/html/rfc2033) - LMTP specification
- [RFC 9051 - IMAP4rev2](https://datatracker.ietf.org/doc/html/rfc9051) - IMAP4rev2 (Stalwart implements)

### Secondary (MEDIUM confidence)

**Comparisons and Tutorials:**
- [Four modern mail systems for self-hosting](https://www.sidn.nl/en/news-and-blogs/four-modern-mail-systems-for-self-hosting) - SIDN comparison (Stalwart, Maddy, others)
- [LMTP vs SMTP - GeeksforGeeks](https://www.geeksforgeeks.org/computer-networks/difference-between-smtp-and-lmtp/) - Protocol comparison
- [Postfix vs Dovecot Key Differences](https://dev.to/shrsv/postfix-vs-dovecot-key-differences-for-building-email-systems-55k2) - MTA vs MDA roles
- [Email Greylisting Guide - MailerCheck](https://www.mailercheck.com/articles/email-greylisting) - Greylisting best practices
- [Postfix Virtual Aliases and Catchall 2026](https://copyprogramming.com/howto/postfix-virtual-aliases-and-catchall-for-undefined-addresses) - Catch-all configuration

**Docker Images:**
- [Stalwart Docker Hub](https://hub.docker.com/r/stalwartlabs/stalwart) - Official multi-arch images
- [Maddy Docker Hub](https://hub.docker.com/r/foxcpp/maddy) - Official image
- [Rspamd Docker Hub](https://hub.docker.com/r/rspamd/rspamd) - Official image
- [Redis Docker Hub](https://hub.docker.com/_/redis) - Official Redis (required for Rspamd greylisting)

**Raspberry Pi Compatibility:**
- [Docker Maddy Stack for Raspberry Pi](https://github.com/zerotens/docker-maddy-stack) - Maddy on RPi4, ARM64 build notes
- [Dockerized Stalwart Mail Server](https://github.com/tiredofit/docker-stalwart) - ARM64 variant (unsupported, use official image)

### Tertiary (LOW confidence - needs verification)

**Community Discussions:**
- [Lobsters: What email server do you run?](https://lobste.rs/s/xisnkd/what_email_server_do_you_run) - Community opinions on mail server choices
- [Hacker News: Stalwart Mail Server with Web UI](https://news.ycombinator.com/item?id=39983018) - Stalwart discussion, production use feedback
- [Best Open Source Email Servers 2026](https://webshanks.com/open-source-email-servers/) - Feature comparison (treat as starting point, verify with official docs)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All three mail servers have official documentation, production use, Docker images
- SMTP vs LMTP delivery: HIGH - Verified with official docs for all three mail servers, emersion/go-smtp library supports both
- Multi-user/multi-domain: HIGH - Documented in official docs for Stalwart, Maddy, Postfix+Dovecot
- Rspamd integration: MEDIUM-HIGH - Verified with Rspamd official docs, community tutorials for specific mail server integrations
- Pitfalls: MEDIUM - Based on official docs warnings, community experience, verified where possible

**Research date:** 2026-02-08
**Valid until:** 60 days (Stalwart v1.0 expected Q2 2026 may change recommendations)

**Critical dependencies:**
- Stalwart v1.0 release (Q2 2026) - will finalize schema, may require migration from 0.15.4
- emersion/go-smtp library - already in use by cloud relay, well-maintained
- Rspamd - stable, active development, de facto standard for modern mail servers