# Phase 06: Webmail & Groupware - Research

**Researched:** 2026-02-14
**Domain:** Webmail, CalDAV/CardDAV, Reverse Proxy, Remote Access
**Confidence:** HIGH

## Summary

Phase 6 delivers web-based email access for non-technical household members and CalDAV/CardDAV calendar/contact sync across devices. The critical architectural decision is **webmail runs on home device with remote access via reverse proxy tunnel** through the cloud relay, keeping all mail content local while providing family-friendly access without VPN.

User decisions lock in: Roundcube and SnappyMail as user-selectable webmail options (Docker compose profiles), Stalwart built-in CalDAV/CardDAV when Stalwart is the mail server with standalone Radicale for Maddy/Postfix+Dovecot, subdomain access via `mail.example.com`, basic calendar sharing, and shared family address book.

**Primary recommendation:** Use **Caddy reverse proxy** on cloud relay for automatic HTTPS and simple tunnel forwarding to home webmail. Implement webmail selection as Docker compose profiles matching Phase 3 mail server pattern. Deploy **Radicale** for standalone CalDAV/CardDAV (lightweight, file-based, simple). Use IMAP passthrough authentication (no separate auth system). Configure well-known URLs for iOS/macOS/Android auto-discovery.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Webmail selection & access:**
- User-selectable webmail client: both Roundcube and SnappyMail available (matches MAIL-01 pattern of offering choices via Docker compose profiles)
- Webmail accessed via subdomain: `mail.example.com` (requires DNS record per domain)
- Webmail runs on home device (keeps mail content local, aligns with DarkPipe privacy model)
- Webmail accessible remotely through the cloud relay tunnel (non-technical family members don't need VPN)

**CalDAV/CardDAV server choice:**
- Use Stalwart built-in CalDAV/CardDAV when user picked Stalwart as mail server; deploy standalone server for Maddy/Postfix+Dovecot setups
- Same credentials as mail account — user@domain + mail password works for CalDAV/CardDAV (one set of credentials per person)
- Auto-create default calendar + address book per user on account creation (ready to sync immediately)
- Standard well-known URLs: `/.well-known/caldav` and `/.well-known/carddav` for auto-discovery on iOS/macOS/Android

**Multi-user experience:**
- Basic calendar sharing between household members (read-only or read-write)
- Shared family address book visible and editable by all household members, plus individual private address books
- CalDAV/CardDAV remotely accessible through tunnel (phone syncs whether at home or away)

### Claude's Discretion

- Webmail authentication approach (same IMAP credentials vs SSO — pick based on mail server capabilities)
- Standalone CalDAV/CardDAV server for non-Stalwart setups (Radicale vs Baikal — pick best fit for lightweight philosophy)
- Account onboarding flow (how household members get accounts — align with Phase 3 user management)
- Email isolation model (strictly private vs admin-accessible — pick based on privacy model)
- Reverse proxy choice (Caddy vs Nginx vs Traefik — pick best fit for DarkPipe's Go stack and auto-TLS needs)
- Webmail hosting location rationale and tunnel routing architecture

</user_constraints>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **Roundcube Webmail** | 1.6.13 (Feb 2026) | Feature-rich webmail with plugins | Industry standard, 20+ years development, extensive plugin ecosystem, Elastic responsive skin, multi-language support, address book integration |
| **SnappyMail** | 2.38.2 | Lightweight modern webmail | 66% smaller than RainLoop predecessor (~315KB JS vs 1.2MB), modern ES2020, PGP support, Sieve filtering, no tracking/social integrations, fast |
| **Radicale** | 3.6.0 (Jan 2026) | Lightweight CalDAV/CardDAV server | Python-based, file-system storage, minimal dependencies, out-of-box functionality, GPLv3, simple folder structure, works on UNIX/Windows |
| **Baikal** | 0.11.1 (Nov 2025) | PHP CalDAV/CardDAV server | Based on SabreDAV library, extensive web UI for user management, fast installation, basic PHP server requirements |
| **Stalwart CalDAV/CardDAV** | Built-in (0.15.4+) | All-in-one collaboration server | When Stalwart is mail server, provides first-class CalDAV/CardDAV/WebDAV support, JMAP for calendars/contacts, unified backend with email |
| **Caddy** | Latest | Reverse proxy with automatic HTTPS | Written in Go (aligns with DarkPipe stack), automatic Let's Encrypt certificates, simple Caddyfile syntax, built-in HTTPS by default, WebSocket/gRPC support |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **Nginx** | Latest | Traditional reverse proxy | Alternative to Caddy if manual TLS management preferred or existing Nginx expertise |
| **Traefik** | Latest | Cloud-native reverse proxy | Alternative for complex dynamic routing or Kubernetes integration (overkill for DarkPipe) |
| **Redis** | Alpine | Session storage for webmail | Optional for Roundcube/SnappyMail session backend (default is database/files) |
| **PostgreSQL/MariaDB** | Latest | Webmail database backend | Roundcube requires database for settings/contacts; SnappyMail can run database-free |

### Webmail Comparison

| Feature | Roundcube | SnappyMail |
|---------|-----------|------------|
| **Size** | ~2-3MB JS | ~315KB JS (66% smaller than RainLoop) |
| **Database** | Required (MySQL/PostgreSQL/SQLite) | Optional (can run without) |
| **Plugins** | Extensive ecosystem (200+ config options) | Minimal (focus on core features) |
| **Responsive Design** | Elastic skin (v1.4+) | Mobile-optimized (~138KB download) |
| **Authentication** | IMAP passthrough | IMAP passthrough + password_hash for admin |
| **Features** | Address book, spell check, filters, ManageSieve | PGP encryption, Sieve editor, minimal bloat |
| **Performance** | Resource-intensive on low-end servers | Lightspeed, 99% Lighthouse score |
| **Maturity** | 20+ years, battle-tested | RainLoop fork (active 2020+), less mature |

### CalDAV/CardDAV Server Comparison

| Feature | Radicale | Baikal | Stalwart Built-in |
|---------|----------|--------|-------------------|
| **Language** | Python | PHP | Rust |
| **Storage** | File-system (simple folders) | SabreDAV (database) | Integrated mail store |
| **Setup Complexity** | Minimal (preconfigured) | Web UI for management | REST API + Web UI |
| **Calendar Sharing** | Rights file (no integrated UI) | Complicated setup | JMAP sharing (v1.0+) |
| **Dependencies** | Few (Python 3) | PHP server + MySQL/PostgreSQL | None (part of Stalwart) |
| **Maturity** | Stable, 10+ years | Stable, based on SabreDAV | New (2023+), production use |
| **Development Status** | Active | Slow (volunteer-maintained) | Active (NLnet funded) |

**Installation:**

```bash
# Roundcube (Docker)
docker pull roundcube/roundcubemail:1.6.13

# SnappyMail (Docker)
docker pull djmaze/snappymail:2.38.2

# Radicale (Docker)
docker pull tomsquest/docker-radicale:3.6.0

# Baikal (Docker)
docker pull ckulka/baikal:0.11.1

# Caddy (Docker)
docker pull caddy:2-alpine

# Nginx (Docker)
docker pull nginx:alpine

# Traefik (Docker)
docker pull traefik:v3.2
```

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Caddy | Nginx | Nginx requires manual TLS setup, more configuration complexity, but 20+ years battle-tested and maximum performance. Choose if existing Nginx expertise or advanced routing needed. |
| Caddy | Traefik | Traefik excels in Kubernetes/microservices with automatic service discovery, but overkill for DarkPipe's simple tunnel-to-home architecture. Choose only if planning complex multi-service routing. |
| Radicale | Baikal | Baikal has web UI for user management, better for non-technical admins. Choose if prefer GUI over file/CLI management. WARNING: development has slowed, Radicale more active. |
| Roundcube | SnappyMail only | SnappyMail is faster and lighter, but Roundcube has more plugins and longer track record. Offering both gives users choice based on needs (features vs performance). |
| IMAP authentication | OAuth/SSO | OAuth adds complexity (external identity provider, token management). IMAP passthrough is simpler, aligns with mail server being source of truth, one set of credentials. |

## Architecture Patterns

### Recommended Project Structure

```
home-device/
├── webmail/
│   ├── roundcube/                 # Roundcube option
│   │   ├── Dockerfile
│   │   ├── config.inc.php         # IMAP/SMTP settings
│   │   └── plugins/               # Enabled plugins
│   └── snappymail/                # SnappyMail option
│       ├── Dockerfile
│       └── config/                # Application config
├── caldav-carddav/
│   ├── radicale/                  # Radicale option (Maddy/Postfix+Dovecot)
│   │   ├── Dockerfile
│   │   ├── config                 # Server config
│   │   ├── rights                 # ACL for sharing
│   │   └── collections/           # Calendar/contact data
│   └── stalwart-collections/      # Stalwart uses built-in (no separate container)
├── reverse-proxy/
│   ├── caddy/                     # Deployed on cloud relay
│   │   ├── Caddyfile              # Reverse proxy config
│   │   └── tunnel-forward.conf   # Forward mail.* to home device
│   └── well-known/                # Auto-discovery endpoints
│       ├── caldav                 # CalDAV discovery
│       └── carddav                # CardDAV discovery
└── docker-compose.yml             # Orchestrates webmail + CalDAV/CardDAV profiles
```

**Cloud relay structure:**
```
cloud-relay/
└── reverse-proxy/
    ├── caddy/
    │   ├── Caddyfile              # Main reverse proxy
    │   └── certs/                 # Let's Encrypt certs (auto-managed)
    └── docker-compose.yml         # Caddy service on cloud VPS
```

### Pattern 1: Webmail via Reverse Proxy Tunnel

**What:** Caddy on cloud relay proxies `mail.example.com` requests through WireGuard/mTLS tunnel to webmail running on home device.

**When to use:** Always (matches user requirement for remote access without VPN).

**Example (Caddyfile on cloud relay):**

```
# Source: Caddy official docs (https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)
# Cloud relay Caddyfile

mail.example.com {
    # Automatic HTTPS via Let's Encrypt
    # Caddy obtains and renews certificates automatically

    # Forward to home device via WireGuard tunnel
    reverse_proxy 10.0.0.2:8080 {
        # WireGuard peer IP (home device)
        # Assumes webmail listens on port 8080

        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}

        # Health check (optional)
        health_uri /health
        health_interval 30s
    }

    # Log access for debugging (optional)
    log {
        output file /var/log/caddy/mail.example.com.log
    }
}

# Well-known CalDAV/CardDAV discovery
mail.example.com {
    redir /.well-known/caldav /dav/ 301
    redir /.well-known/carddav /dav/ 301
}
```

### Pattern 2: IMAP Passthrough Authentication (Webmail)

**What:** Webmail authenticates users directly against IMAP server (no separate user database). User enters email@domain + password, webmail connects to IMAP with same credentials.

**When to use:** Always (simplest approach, mail server is source of truth).

**Roundcube example (config.inc.php):**

```php
// Source: Roundcube official docs (https://github.com/roundcube/roundcubemail/wiki/Configuration)

// IMAP server connection
$config['imap_host'] = 'ssl://imap:993';  // IMAP container name or IP
$config['imap_conn_options'] = array(
    'ssl' => array(
        'verify_peer' => false,        // Self-signed cert in WireGuard tunnel
        'verify_peer_name' => false,
    ),
);

// SMTP submission for sending
$config['smtp_host'] = 'tls://smtp:587';  // Submission port
$config['smtp_conn_options'] = array(
    'ssl' => array(
        'verify_peer' => false,
        'verify_peer_name' => false,
    ),
);

// Authentication
$config['smtp_user'] = '%u';              // Use IMAP username
$config['smtp_pass'] = '%p';              // Use IMAP password

// Auto-create user on first login
$config['auto_create_user'] = true;

// Default identity (email address)
$config['identities_level'] = 0;          // User can edit all identity fields
$config['username_domain'] = 'example.com'; // Default domain for login

// Database for Roundcube settings (not user auth)
$config['db_dsnw'] = 'sqlite:////var/roundcube/db/sqlite.db?mode=0640';
```

**SnappyMail example (application.ini):**

```ini
; Source: SnappyMail docs (https://github.com/the-djmaze/snappymail)

[defaults]
; IMAP settings
imap_host = "imap"
imap_port = 993
imap_secure = "SSL"

; SMTP settings
smtp_host = "smtp"
smtp_port = 587
smtp_secure = "TLS"
smtp_auth = "1"           ; Use IMAP credentials

; Authentication
auth_type = "IMAP"        ; Authenticate via IMAP
login_lowercase = 1       ; Normalize to lowercase

; No database required (settings in files)
storage_provider = "file"
```

### Pattern 3: CalDAV/CardDAV with Radicale

**What:** Radicale provides CalDAV/CardDAV server with file-based storage, rights file for sharing, and same credentials as mail account.

**When to use:** When user selected Maddy or Postfix+Dovecot (not Stalwart, which has built-in).

**Radicale config example:**

```ini
# Source: Radicale official docs (https://radicale.org/master.html)
# File: /etc/radicale/config

[server]
hosts = 0.0.0.0:5232

[auth]
type = htpasswd
htpasswd_filename = /etc/radicale/users
htpasswd_encryption = bcrypt

[rights]
type = from_file
file = /etc/radicale/rights

[storage]
filesystem_folder = /var/lib/radicale/collections

[web]
type = internal

[logging]
level = info
```

**Radicale rights file (calendar sharing):**

```
# Source: Radicale GitHub issues (https://github.com/Kozea/Radicale/issues/947)
# File: /etc/radicale/rights

# Owner has full access to their own collections
[owner]
user = .+
collection = ^/(?P<username>[^/]+)(/.*)?$
permissions = rw

# Family members can read each other's calendars
[family-read]
user = alice|bob|charlie
collection = ^/(alice|bob|charlie)/[^/]+\.ics$
permissions = r

# Shared family calendar (all can read/write)
[family-shared]
user = alice|bob|charlie
collection = ^/shared/family-calendar\.ics$
permissions = rw

# Shared family address book (all can read/write)
[family-contacts]
user = alice|bob|charlie
collection = ^/shared/family-contacts\.vcf$
permissions = rw

# Read-only public calendars (authenticated users)
[public-read]
user = .+
collection = ^/public/.*$
permissions = r
```

**Auto-create default calendar/address book (Python script):**

```python
# Source: Custom pattern based on Radicale storage structure
# Script: /opt/radicale/create-defaults.py

import os
import sys
from pathlib import Path

def create_default_collections(username, collections_path="/var/lib/radicale/collections"):
    user_path = Path(collections_path) / username
    user_path.mkdir(parents=True, exist_ok=True)

    # Create default calendar
    calendar_path = user_path / f"{username}-calendar.ics"
    if not calendar_path.exists():
        calendar_path.write_text("""BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//DarkPipe//CalDAV Server//EN
X-WR-CALNAME:My Calendar
X-WR-CALDESC:Default calendar for {username}
END:VCALENDAR
""".format(username=username))

    # Create default address book
    contacts_path = user_path / f"{username}-contacts.vcf"
    if not contacts_path.exists():
        contacts_path.write_text("""BEGIN:VCARD
VERSION:4.0
PRODID:-//DarkPipe//CardDAV Server//EN
END:VCARD
""")

    print(f"Created default collections for {username}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: create-defaults.py <username>")
        sys.exit(1)
    create_default_collections(sys.argv[1])
```

### Pattern 4: Well-Known Auto-Discovery URLs

**What:** Configure `/.well-known/caldav` and `/.well-known/carddav` redirects so iOS/macOS/Android clients auto-discover CalDAV/CardDAV servers.

**When to use:** Always (required for iOS/macOS auto-discovery, recommended for Android with DAVx5).

**Caddy reverse proxy example (on cloud relay):**

```
# Source: CalDAV auto-discovery docs (https://www.axigen.com/documentation/caldav-auto-discovery-p47120765)
# Caddyfile on cloud relay

mail.example.com {
    # Redirect well-known URLs to CalDAV/CardDAV server
    redir /.well-known/caldav /radicale/ 301
    redir /.well-known/carddav /radicale/ 301

    # Forward CalDAV/CardDAV requests to home device Radicale
    reverse_proxy /radicale/* 10.0.0.2:5232 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
    }

    # Forward webmail requests to home device
    reverse_proxy 10.0.0.2:8080
}
```

**Nginx alternative (if user prefers Nginx):**

```nginx
# Source: Nginx reverse proxy docs
# /etc/nginx/sites-available/mail.example.com

server {
    listen 443 ssl http2;
    server_name mail.example.com;

    ssl_certificate /etc/letsencrypt/live/mail.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mail.example.com/privkey.pem;

    # Well-known redirects
    location /.well-known/caldav {
        return 301 $scheme://$host/radicale/;
    }

    location /.well-known/carddav {
        return 301 $scheme://$host/radicale/;
    }

    # CalDAV/CardDAV reverse proxy
    location /radicale/ {
        proxy_pass http://10.0.0.2:5232/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Webmail reverse proxy
    location / {
        proxy_pass http://10.0.0.2:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Pattern 5: Docker Compose Profiles for Webmail Selection

**What:** Use Docker Compose profiles to allow user selection of Roundcube or SnappyMail webmail (mirrors Phase 3 mail server selection pattern).

**When to use:** Always (matches Phase 3 MAIL-01 pattern).

**Example (docker-compose.yml):**

```yaml
# Source: Docker Compose profiles docs (https://docs.docker.com/compose/how-tos/profiles/)
# home-device/docker-compose.yml

version: '3.8'

services:
  # Roundcube webmail (profile: roundcube)
  roundcube:
    image: roundcube/roundcubemail:1.6.13
    profiles: ["roundcube"]
    container_name: roundcube
    restart: unless-stopped
    ports:
      - "8080:80"  # Internal port (reverse proxy connects here)
    volumes:
      - ./webmail/roundcube/config.inc.php:/var/roundcube/config/config.inc.php:ro
      - roundcube-data:/var/roundcube/db
    environment:
      - ROUNDCUBEMAIL_DEFAULT_HOST=ssl://imap:993
      - ROUNDCUBEMAIL_DEFAULT_PORT=993
      - ROUNDCUBEMAIL_SMTP_SERVER=tls://smtp:587
      - ROUNDCUBEMAIL_SMTP_PORT=587
      - ROUNDCUBEMAIL_DB_TYPE=sqlite
    networks:
      - darkpipe
    depends_on:
      - stalwart  # or maddy or postfix-dovecot (based on mail server profile)
    deploy:
      resources:
        limits:
          memory: 256M

  # SnappyMail webmail (profile: snappymail)
  snappymail:
    image: djmaze/snappymail:2.38.2
    profiles: ["snappymail"]
    container_name: snappymail
    restart: unless-stopped
    ports:
      - "8080:8888"
    volumes:
      - snappymail-data:/var/lib/snappymail
    environment:
      - UPLOAD_MAX_SIZE=25M
    networks:
      - darkpipe
    depends_on:
      - stalwart  # or maddy or postfix-dovecot
    deploy:
      resources:
        limits:
          memory: 128M

  # Radicale CalDAV/CardDAV (profile: radicale)
  # Only used when mail server is Maddy or Postfix+Dovecot
  # (Stalwart has built-in CalDAV/CardDAV)
  radicale:
    image: tomsquest/docker-radicale:3.6.0
    profiles: ["radicale"]
    container_name: radicale
    restart: unless-stopped
    ports:
      - "5232:5232"
    volumes:
      - ./caldav-carddav/radicale/config:/config:ro
      - ./caldav-carddav/radicale/rights:/etc/radicale/rights:ro
      - radicale-data:/data
    environment:
      - RADICALE_CONFIG=/config/config
    networks:
      - darkpipe
    deploy:
      resources:
        limits:
          memory: 128M

networks:
  darkpipe:
    driver: bridge

volumes:
  roundcube-data:
    driver: local
  snappymail-data:
    driver: local
  radicale-data:
    driver: local
```

**Usage:**
```bash
# Start Stalwart + Roundcube + built-in CalDAV/CardDAV
docker compose --profile stalwart --profile roundcube up -d

# Start Maddy + SnappyMail + Radicale
docker compose --profile maddy --profile snappymail --profile radicale up -d

# Start Postfix+Dovecot + Roundcube + Radicale
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

### Anti-Patterns to Avoid

- **Don't terminate TLS on home device for webmail:** Cloud relay with public IP should terminate TLS (Let's Encrypt), then forward to home device over encrypted tunnel. Terminating on home device requires public DNS hostname for home IP (unlikely scenario).

- **Don't run separate auth system for webmail:** IMAP passthrough is simplest. Adding OAuth, LDAP, or separate user database creates complexity, credential duplication, and sync issues. Mail server is already source of truth.

- **Don't expose webmail directly on home device public port:** Always proxy through cloud relay tunnel. Direct exposure requires port forwarding (NAT traversal), home device public IP, and home device TLS certs — defeats DarkPipe privacy model.

- **Don't use Baikal if development velocity matters:** Baikal development is slow (volunteer-maintained). Radicale is more actively maintained. Choose Baikal only if web UI for user management is critical requirement.

- **Don't configure CalDAV/CardDAV without well-known URLs:** iOS/macOS require `/.well-known/caldav` and `/.well-known/carddav` for auto-discovery. Without these, users must manually enter full URLs (poor UX).

- **Don't share calendars without Radicale rights file:** Radicale has no web UI for sharing. Sharing requires rights file configuration. Document sharing setup clearly or users won't discover feature.

- **Don't use Traefik unless you need dynamic service discovery:** Traefik excels in Kubernetes/microservices with many services. DarkPipe has static routing (cloud relay → home device). Caddy is simpler for this architecture.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Webmail client | Custom web-based IMAP/SMTP UI | Roundcube or SnappyMail | Webmail requires HTML email rendering (security sandboxing), attachment handling, folder tree UI, MIME parsing, spell check, address book integration, keyboard shortcuts. Roundcube has 20+ years of edge cases. SnappyMail has modern security (CSP, SRI). Building custom webmail is months of work. |
| Reverse proxy | Custom Go HTTP proxy with TLS | Caddy | Caddy handles automatic Let's Encrypt certificates (ACME HTTP-01/DNS-01 challenges), certificate renewal, OCSP stapling, TLS cipher suite selection, HTTP/2, WebSocket proxying. Custom proxy requires implementing ACME protocol, managing cert renewal cron jobs, handling TLS edge cases. |
| CalDAV/CardDAV server | Custom calendar/contact sync API | Radicale or Stalwart built-in | CalDAV/CardDAV protocols are complex (WebDAV extensions, iCalendar/vCard formats, VTIMEZONE handling, recurring events, sync tokens). Radicale handles RFC compliance, conflict resolution, ETag management. Custom implementation will miss edge cases (timezone DST transitions, all-day events, calendar sharing ACLs). |
| Calendar sharing | Custom permissions system | Radicale rights file | Calendar sharing requires WebDAV ACLs, per-user/per-calendar permissions, read-only vs read-write, inheritance. Radicale rights file implements standard WebDAV ACL patterns. Custom system will reinvent permission model poorly. |
| iOS/Android auto-discovery | Custom device configuration generator | Standard well-known URLs | iOS/macOS auto-discovery uses `/.well-known/caldav` (RFC 6764), DNS SRV records, PROPFIND queries. Android DAVx5 follows same standards. Implementing custom discovery breaks standard clients. Use well-known redirects. |

**Key insight:** Webmail, reverse proxy, and CalDAV/CardDAV are mature problem domains with 15-20 years of RFC evolution and edge cases. Use battle-tested implementations (Roundcube, Caddy, Radicale) rather than building custom. Custom webmail or CalDAV server will be security liability and poor UX.

## Common Pitfalls

### Pitfall 1: Reverse Proxy TLS Termination Location

**What goes wrong:** User terminates TLS on home device instead of cloud relay, requiring public DNS hostname pointing to home IP and Let's Encrypt certificates on home device. This breaks if home IP changes or ISP blocks port 443.

**Why it happens:** Assumption that TLS should be terminated "closest to the data" or confusion about where certificates belong in tunnel architecture.

**How to avoid:**
1. Always terminate public TLS on cloud relay (static public IP, Let's Encrypt works)
2. Forward HTTPS requests over WireGuard/mTLS tunnel to HTTP on home device
3. Tunnel is already encrypted (WireGuard or mTLS), so HTTP inside tunnel is secure
4. Home device doesn't need public DNS hostname or certificates

**Warning signs:** Let's Encrypt failing on home device with "connection refused" or "DNS resolution failed", home IP changes breaking webmail access.

### Pitfall 2: Webmail Database Storage Location

**What goes wrong:** Roundcube stores settings/contacts in database on home device, but database volume isn't backed up. User loses settings after container recreation or volume deletion.

**Why it happens:** Default Docker volume is ephemeral or user doesn't configure volume mount for Roundcube database.

**How to avoid:**
1. Mount Roundcube database to persistent volume or bind mount to host path
2. For SQLite backend: `- ./roundcube/db:/var/roundcube/db`
3. For PostgreSQL/MariaDB: ensure database container has persistent volume
4. Document backup procedure for Roundcube database (contains user settings, custom identities)

**Warning signs:** Roundcube settings reset after container restart, address book entries disappear.

### Pitfall 3: CalDAV/CardDAV Authentication Without IMAP Sync

**What goes wrong:** Radicale users file (htpasswd) gets out of sync with IMAP user accounts. User can log in to webmail but CalDAV/CardDAV fails with "authentication error."

**Why it happens:** CalDAV/CardDAV server has separate user database not synchronized with mail server.

**How to avoid:**
1. Script to sync IMAP users to Radicale htpasswd on user creation
2. Or use HTTP basic auth proxy that validates against IMAP before forwarding to Radicale
3. Or configure Radicale to use IMAP auth (requires custom auth plugin)
4. **Recommended:** Document manual user sync procedure for admins when adding mail users

**Warning signs:** User receives "401 Unauthorized" on CalDAV/CardDAV but can access webmail, iOS Calendar setup fails with "invalid credentials."

**Correct pattern (user creation flow):**
1. Create user in mail server (Stalwart REST API, Maddy CLI, or Postfix vmailbox)
2. Add same user to Radicale htpasswd: `htpasswd -B /etc/radicale/users user@example.com`
3. Run auto-create script for default calendar/address book
4. User can now access webmail, IMAP, CalDAV, CardDAV with same credentials

### Pitfall 4: Well-Known URLs Not Configured

**What goes wrong:** iOS/macOS Calendar/Contacts setup requires manual URL entry instead of auto-discovery. User enters `mail.example.com` but setup fails with "cannot verify server."

**Why it happens:** `/.well-known/caldav` and `/.well-known/carddav` redirects not configured on reverse proxy.

**How to avoid:**
1. Configure well-known redirects on Caddy/Nginx for every domain
2. Test with iOS device: Settings → Calendar → Add Account → Enter email → should auto-discover without manual URL
3. For Radicale: redirect to `/radicale/` (with trailing slash)
4. Verify redirects with curl: `curl -I https://mail.example.com/.well-known/caldav` should return 301/302 to CalDAV server

**Warning signs:** iOS Calendar/Contacts setup asks for "Server Address" instead of auto-discovering, Android DAVx5 requires manual URL entry.

### Pitfall 5: Calendar Sharing Without Rights File Configuration

**What goes wrong:** User expects to share calendar with family member from iOS Calendar app, but sharing doesn't work. No error shown; calendar just doesn't appear for other user.

**Why it happens:** Radicale has no web UI for sharing. Sharing requires rights file configuration by admin. iOS Calendar sharing options assume CalDAV server supports programmatic sharing (like Nextcloud), but Radicale uses static rights file.

**How to avoid:**
1. Document that calendar sharing requires admin configuration (not user self-service)
2. Provide example rights file snippets for common patterns (read-only share, read-write share, family shared calendar)
3. Create shared collections in Radicale data directory before adding to rights file
4. Alternative: Use Stalwart built-in (JMAP sharing API supports programmatic sharing in future versions)

**Warning signs:** User reports "shared calendar not appearing," iOS Calendar share invite doesn't work, no error messages in logs.

**Correct pattern (admin-configured sharing):**
1. Admin edits `/etc/radicale/rights` to grant user2 read access to user1's calendar
2. User2 must manually add shared calendar URL in CalDAV client (not automatic)
3. Shared family calendar: Admin creates `/var/lib/radicale/collections/shared/family.ics` and adds rights for all family members

### Pitfall 6: Subdomain DNS Record Missing

**What goes wrong:** User configures Caddy reverse proxy for `mail.example.com` but domain doesn't resolve. Let's Encrypt certificate issuance fails with "DNS resolution failed."

**Why it happens:** DNS A/AAAA record for `mail.example.com` not created, or points to wrong IP (home device instead of cloud relay).

**How to avoid:**
1. Phase 4 DNS setup should create MX records; Phase 6 requires A/AAAA records for webmail subdomain
2. For each domain: `mail.example.com` → A record → cloud relay public IP
3. Test DNS resolution before starting Caddy: `dig +short mail.example.com` should return cloud relay IP
4. Caddy automatic HTTPS requires DNS resolution working BEFORE first start

**Warning signs:** Let's Encrypt failing with "dns: no such host", Caddy logs show "failed to obtain certificate", webmail unreachable from internet.

### Pitfall 7: Roundcube Plugin Compatibility

**What goes wrong:** User enables Roundcube plugin from community repository, but plugin breaks webmail with PHP errors or breaks IMAP connection.

**Why it happens:** Roundcube has 200+ plugins, not all tested with latest version. Some plugins require specific PHP versions or database backends.

**How to avoid:**
1. Only enable plugins from official Roundcube repository (avoid third-party without testing)
2. Test plugins in development environment before production
3. Check plugin compatibility with Roundcube version (1.6.x)
4. Prefer core features over plugins (e.g., built-in spell check vs plugin)

**Warning signs:** Roundcube shows white screen after plugin enable, PHP errors in logs, IMAP connection broken after plugin configuration.

### Pitfall 8: Mobile Webmail Viewport

**What goes wrong:** User accesses webmail on mobile phone, but layout broken (horizontal scrolling, tiny fonts, unusable on small screens).

**Why it happens:** Webmail not configured with responsive skin (Roundcube) or viewport meta tag missing.

**How to avoid:**
1. Roundcube: Enable Elastic skin (default in 1.4+): `$config['skin'] = 'elastic';` in config.inc.php
2. SnappyMail: Mobile-optimized by default, verify viewport meta tag present
3. Test on actual mobile device (iOS Safari, Android Chrome) before declaring "mobile-ready"
4. Check WEB-02 success criteria: "usable on mobile phone screen without horizontal scrolling"

**Warning signs:** User reports "webmail doesn't work on phone," horizontal scrolling on mobile, text unreadable without zoom.

### Pitfall 9: Webmail Session Timeout Too Short

**What goes wrong:** User logs into webmail, reads email for 5 minutes, tries to send reply, and session expired — forced to log in again.

**Why it happens:** Roundcube default session timeout (10 minutes inactivity) too aggressive for email reading workflow.

**How to avoid:**
1. Roundcube: Increase `$config['session_lifetime']` to 60 minutes (1 hour)
2. Roundcube: Enable `$config['refresh']` auto-refresh to keep session alive
3. SnappyMail: Configure session timeout in admin settings
4. Balance between security (short timeout) and UX (long timeout) — 30-60 minutes reasonable for webmail

**Warning signs:** User reports "constantly logging in again," session expired errors during email composition, lost draft emails.

## Code Examples

Verified patterns from official sources:

### Caddy Reverse Proxy with Automatic HTTPS

```
# Source: Caddy official docs (https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)
# File: /etc/caddy/Caddyfile (on cloud relay VPS)

# Webmail subdomain with automatic HTTPS
mail.example.com {
    # Caddy automatically obtains Let's Encrypt certificate
    # No manual TLS configuration required

    # Forward to home device via WireGuard tunnel
    reverse_proxy 10.0.0.2:8080 {
        # WireGuard peer IP (home device)
        # Preserves original client IP and headers

        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}

        # WebSocket support (for future real-time features)
        header_up Connection {>Connection}
        header_up Upgrade {>Upgrade}
    }

    # Access logging
    log {
        output file /var/log/caddy/mail.access.log
        format json
    }
}

# Multiple domain support (repeat for each domain)
mail.example.org {
    reverse_proxy 10.0.0.2:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

### Roundcube Docker Compose with IMAP Authentication

```yaml
# Source: Roundcube Docker official image (https://hub.docker.com/r/roundcube/roundcubemail/)
# File: home-device/docker-compose.yml

services:
  roundcube:
    image: roundcube/roundcubemail:1.6.13
    container_name: roundcube
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      - ./webmail/roundcube/config.inc.php:/var/roundcube/config/config.inc.php:ro
      - ./webmail/roundcube/plugins:/var/roundcube/plugins:ro
      - roundcube-db:/var/roundcube/db
    environment:
      # IMAP server (use container name if on same network)
      - ROUNDCUBEMAIL_DEFAULT_HOST=ssl://mail-server
      - ROUNDCUBEMAIL_DEFAULT_PORT=993

      # SMTP submission
      - ROUNDCUBEMAIL_SMTP_SERVER=tls://mail-server
      - ROUNDCUBEMAIL_SMTP_PORT=587

      # Database (SQLite for simplicity)
      - ROUNDCUBEMAIL_DB_TYPE=sqlite
      - ROUNDCUBEMAIL_DB_DIR=/var/roundcube/db

      # Auto-create users on first login
      - ROUNDCUBEMAIL_ENABLE_INSTALLER=false

      # Upload size limit
      - ROUNDCUBEMAIL_UPLOAD_MAX_FILESIZE=25M
    networks:
      - darkpipe
    depends_on:
      - stalwart  # or maddy or postfix-dovecot
    deploy:
      resources:
        limits:
          memory: 256M
        reservations:
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  roundcube-db:
    driver: local

networks:
  darkpipe:
    external: true
```

### SnappyMail Configuration

```yaml
# Source: SnappyMail Docker image (https://github.com/the-djmaze/snappymail)
# File: home-device/docker-compose.yml

services:
  snappymail:
    image: djmaze/snappymail:2.38.2
    container_name: snappymail
    restart: unless-stopped
    ports:
      - "8080:8888"
    volumes:
      - snappymail-data:/var/lib/snappymail
    environment:
      - UPLOAD_MAX_SIZE=25M
      - LOG_TO_STDOUT=true
      - MEMORY_LIMIT=128M
    networks:
      - darkpipe
    depends_on:
      - stalwart
    deploy:
      resources:
        limits:
          memory: 128M
        reservations:
          memory: 64M
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8888/"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  snappymail-data:
    driver: local

networks:
  darkpipe:
    external: true
```

**SnappyMail first-time setup (admin login):**
```
# Access admin interface: http://mail.example.com/?admin
# Default admin login: admin / 12345
# CHANGE IMMEDIATELY after first login

# Configure IMAP/SMTP in admin panel:
# - IMAP: mail-server, port 993, SSL
# - SMTP: mail-server, port 587, TLS
# - Auth: Use IMAP credentials
```

### Radicale Docker Compose with Calendar Sharing

```yaml
# Source: Radicale Docker image (https://github.com/tomsquest/docker-radicale)
# File: home-device/docker-compose.yml

services:
  radicale:
    image: tomsquest/docker-radicale:3.6.0
    container_name: radicale
    restart: unless-stopped
    init: true
    ports:
      - "5232:5232"
    volumes:
      - ./caldav-carddav/radicale/config:/config:ro
      - ./caldav-carddav/radicale/rights:/etc/radicale/rights:ro
      - ./caldav-carddav/radicale/users:/data/users:ro
      - radicale-collections:/data/collections
    environment:
      - RADICALE_CONFIG=/config/config
      - TZ=America/New_York
    networks:
      - darkpipe
    deploy:
      resources:
        limits:
          memory: 128M
        reservations:
          memory: 64M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5232/"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  radicale-collections:
    driver: local

networks:
  darkpipe:
    external: true
```

**Radicale config file:**
```ini
# Source: Radicale official docs (https://radicale.org/v3.html)
# File: caldav-carddav/radicale/config/config

[server]
hosts = 0.0.0.0:5232
max_connections = 20
max_content_length = 100000000
timeout = 30

[auth]
type = htpasswd
htpasswd_filename = /data/users
htpasswd_encryption = bcrypt
delay = 1

[rights]
type = from_file
file = /etc/radicale/rights

[storage]
type = multifilesystem
filesystem_folder = /data/collections

[web]
type = internal

[logging]
level = info
mask_passwords = true
```

**Radicale htpasswd users file (bcrypt):**
```bash
# Source: htpasswd command-line tool
# File: caldav-carddav/radicale/users

# Create users with bcrypt encryption (same as IMAP password)
# Sync with mail server user creation

# Example entries (generated with: htpasswd -B -c users alice@example.com)
alice@example.com:$2y$05$somehashedbcryptpassword
bob@example.com:$2y$05$anotherhashedbcryptpassword
charlie@example.com:$2y$05$yetanotherhashedbcryptpassword
```

### Well-Known Auto-Discovery Configuration

```
# Source: RFC 6764 - Locating Services for Calendaring (CalDAV)
# File: /etc/caddy/Caddyfile (on cloud relay)

# CalDAV/CardDAV auto-discovery for iOS/macOS/Android
mail.example.com {
    # Redirect well-known URLs to Radicale base path
    redir /.well-known/caldav /radicale/ 301
    redir /.well-known/carddav /radicale/ 301

    # Forward CalDAV/CardDAV to home device Radicale
    reverse_proxy /radicale/* 10.0.0.2:5232 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

    # Forward webmail to home device
    reverse_proxy 10.0.0.2:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

**iOS auto-discovery test:**
```bash
# Test well-known redirects
curl -I https://mail.example.com/.well-known/caldav
# Should return: HTTP/1.1 301 Moved Permanently
# Location: https://mail.example.com/radicale/

# Test CalDAV endpoint
curl -u alice@example.com:password -X PROPFIND \
  -H "Depth: 0" \
  https://mail.example.com/radicale/alice@example.com/
# Should return: 207 Multi-Status (WebDAV response)
```

### Shared Family Calendar Setup (Radicale)

```bash
# Source: Radicale rights file examples (https://github.com/Kozea/Radicale/blob/master/rights)
# Script to create shared family calendar

# 1. Create shared collection directory
mkdir -p /var/lib/radicale/collections/shared

# 2. Create family calendar
cat > /var/lib/radicale/collections/shared/family-calendar.ics <<EOF
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//DarkPipe//Radicale//EN
X-WR-CALNAME:Family Calendar
X-WR-CALDESC:Shared calendar for family events
X-WR-TIMEZONE:America/New_York
CALSCALE:GREGORIAN
METHOD:PUBLISH
END:VCALENDAR
EOF

# 3. Create shared family address book
cat > /var/lib/radicale/collections/shared/family-contacts.vcf <<EOF
BEGIN:VCARD
VERSION:4.0
FN:Family Contacts
PRODID:-//DarkPipe//Radicale//EN
END:VCARD
EOF

# 4. Set permissions (Radicale user owns files)
chown -R radicale:radicale /var/lib/radicale/collections/shared

# 5. Add rights in /etc/radicale/rights (already shown in Pattern 3)
```

**Family members access shared calendar:**
```
# iOS/macOS Calendar setup:
# 1. Settings → Calendar → Accounts → Add Account → Other
# 2. Add CalDAV Account
# 3. Server: mail.example.com
# 4. Username: alice@example.com
# 5. Password: <IMAP password>
# 6. Description: DarkPipe Calendar
# 7. Calendar app will auto-discover:
#    - alice@example.com/alice-calendar.ics (personal)
#    - shared/family-calendar.ics (family shared)

# Android DAVx5 setup:
# 1. Install DAVx5 from F-Droid or Play Store
# 2. Add account → Login with URL and credentials
# 3. Server: https://mail.example.com/radicale/
# 4. Username: bob@example.com
# 5. Password: <IMAP password>
# 6. DAVx5 discovers calendars and address books
# 7. Select which to sync (personal + family shared)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| RainLoop webmail | SnappyMail | 2020 (RainLoop abandoned) | SnappyMail fork actively maintained, 66% smaller, modern security (no tracking), PGP support. RainLoop unmaintained — use SnappyMail. |
| Roundcube Classic/Larry skins | Roundcube Elastic skin | 2018 (Roundcube 1.4.0) | Elastic is first responsive skin with mobile support. Classic/Larry not mobile-friendly. Use Elastic for WEB-02 requirement (mobile-responsive). |
| Manual TLS certificate management | Caddy automatic HTTPS | 2015+ (Caddy 1.0) | Caddy obtains/renews Let's Encrypt certificates automatically. No manual certbot cron jobs, no certificate expiry surprises. Use Caddy for auto-TLS unless existing Nginx expertise. |
| Separate auth systems (webmail, CalDAV) | IMAP passthrough authentication | 2010+ (standard pattern) | Webmail authenticates via IMAP, CalDAV via HTTP basic auth (synced with IMAP). No separate user database, one set of credentials. Simpler than OAuth/LDAP for single-server setup. |
| Desktop-only webmail | Mobile-responsive webmail | 2016+ (mobile-first era) | Roundcube Elastic (2018), SnappyMail mobile-optimized by default. Mobile access critical for non-technical users. Desktop-only webmail fails WEB-02. |
| CalDAV/CardDAV with proprietary extensions | Standard RFC-compliant CalDAV/CardDAV | 2010+ (RFC 4791/6352) | Radicale implements standard CalDAV/CardDAV without proprietary extensions. Works with iOS, macOS, Android, Thunderbird. Avoid proprietary extensions (breaks cross-platform sync). |
| Nginx reverse proxy manual config | Caddy automatic config | 2020+ (Caddy 2.0) | Nginx requires manual SSL config, certbot hooks, renewal cron. Caddy uses Caddyfile (5 lines vs 50). Choose Caddy unless Nginx expertise exists. |

**Deprecated/outdated:**
- **RainLoop webmail**: Abandoned 2020, use SnappyMail (active fork)
- **Roundcube Classic/Larry skins**: Use Elastic (mobile-responsive, default in 1.4+)
- **Owncloud Calendar/Contacts**: Use Nextcloud (active fork) or standalone Radicale/Baikal (lighter)
- **Manual Let's Encrypt with certbot cron**: Use Caddy automatic HTTPS (simpler, fewer failure modes)
- **WebDAV-only calendar sync**: Use CalDAV (standard protocol, better client support)

## Open Questions

1. **Should webmail session storage use Redis or file-based backend?**
   - What we know: Roundcube supports Redis for sessions (faster, shared across instances). SnappyMail uses file-based by default.
   - What's unclear: Performance impact on single-server DarkPipe deployment, Redis memory overhead vs disk I/O for sessions.
   - Recommendation: Use file-based sessions for Phase 6 (simpler, one less container). Document Redis migration path if user adds load balancer (Phase 7+).

2. **How to handle CalDAV/CardDAV user creation synchronization?**
   - What we know: Radicale has separate htpasswd file. Stalwart uses same user database as mail. User creation must sync Radicale htpasswd with mail server users.
   - What's unclear: Best automation approach (hook script on mail server, periodic sync cron, manual admin step).
   - Recommendation: Document manual sync procedure for Phase 6. Create hook script in Phase 8 (device profiles) when user onboarding formalized.

3. **Should Radicale storage be file-based or database-backed?**
   - What we know: Radicale default is file-based (simple folder structure, .ics/.vcf files). Radicale supports PostgreSQL backend (requires plugin).
   - What's unclear: File-based performance with many users/calendars, backup/restore simplicity.
   - Recommendation: Use file-based storage for Phase 6 (simpler, aligns with Radicale philosophy). File-based enables easy backup (rsync), version control (git), and manual inspection. Database backend adds complexity without clear benefit for household deployment.

4. **How to handle multiple domains with webmail (mail.example.com vs mail.example.org)?**
   - What we know: Each domain needs DNS A record pointing to cloud relay. Caddy can handle multiple domains with same reverse proxy target.
   - What's unclear: Should each domain have separate webmail subdomain, or single unified webmail subdomain for all domains?
   - Recommendation: Single webmail subdomain per domain (mail.example.com, mail.example.org) for clarity. User enters webmail.example.com → logs in as user@example.com OR user@example.org (webmail detects domain from username). Alternative: Single unified webmail.darkpipe.example.com → serves all domains (simpler DNS but less intuitive for users).

5. **What is the upgrade path for Stalwart v1.0 CalDAV/CardDAV?**
   - What we know: Stalwart 0.15.4 has CalDAV/CardDAV. v1.0 expected Q2 2026 will finalize collaboration features, add JMAP sharing API.
   - What's unclear: Will 0.15.4 → v1.0 migration require manual steps for calendar data? Will JMAP sharing API enable programmatic calendar sharing (vs Radicale static rights file)?
   - Recommendation: Document Stalwart upgrade procedure in Phase 6 docs. Test v1.0 upgrade in development environment. Monitor Stalwart blog for v1.0 release notes. JMAP sharing API (if available in v1.0) would eliminate Radicale rights file complexity for Stalwart users.

## Sources

### Primary (HIGH confidence)

**Webmail Official Documentation:**
- [Roundcube Webmail](https://roundcube.net/) - Official site, v1.6.13 release (Feb 2026)
- [Roundcube GitHub](https://github.com/roundcube/roundcubemail) - Source code, releases, configuration wiki
- [Roundcube Configuration](https://github.com/roundcube/roundcubemail/wiki/Configuration) - IMAP/SMTP settings, authentication
- [SnappyMail](https://snappymail.eu/) - Official site, modern webmail client
- [SnappyMail GitHub](https://github.com/the-djmaze/snappymail) - Source code, v2.38.2 release

**CalDAV/CardDAV Official Documentation:**
- [Radicale](https://radicale.org/) - Official site, v3.6.0 release (Jan 2026)
- [Radicale GitHub](https://github.com/Kozea/Radicale) - Source code, releases, issues
- [Radicale Documentation](https://radicale.org/master.html) - Configuration, storage, rights file
- [Baikal](https://sabre.io/baikal/) - Official site, v0.11.1 release (Nov 2025)
- [Baikal GitHub](https://github.com/sabre-io/Baikal) - Source code, releases
- [Stalwart Collaboration Server](https://stalw.art/blog/collaboration/) - CalDAV/CardDAV announcement
- [Stalwart Roadmap](https://stalw.art/blog/roadmap/) - Webmail planned 2026, JMAP collaboration

**Reverse Proxy Official Documentation:**
- [Caddy](https://caddyserver.com/) - Official site, automatic HTTPS
- [Caddy Reverse Proxy](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy) - Configuration docs
- [Caddy Automatic HTTPS](https://caddyserver.com/docs/automatic-https) - Let's Encrypt integration
- [Caddy GitHub](https://github.com/caddyserver/caddy) - Source code, Go-based server

**Protocol Specifications:**
- [RFC 6764 - Locating Services for CalDAV/CardDAV](https://datatracker.ietf.org/doc/html/rfc6764) - Well-known URLs, DNS SRV records
- [RFC 4791 - CalDAV](https://datatracker.ietf.org/doc/html/rfc4791) - Calendaring Extensions to WebDAV
- [RFC 6352 - CardDAV](https://datatracker.ietf.org/doc/html/rfc6352) - vCard Extensions to WebDAV

**Docker Documentation:**
- [Docker Compose Profiles](https://docs.docker.com/compose/how-tos/profiles/) - Service selection via profiles
- [Roundcube Docker Hub](https://hub.docker.com/r/roundcube/roundcubemail/) - Official Docker image
- [SnappyMail Docker Hub](https://hub.docker.com/r/djmaze/snappymail) - Official Docker image
- [Radicale Docker](https://github.com/tomsquest/docker-radicale) - Community Docker image

### Secondary (MEDIUM confidence)

**Comparisons and Tutorials:**
- [Reverse Proxy Comparison: Traefik vs Caddy vs Nginx](https://www.programonaut.com/reverse-proxies-compared-traefik-vs-caddy-vs-nginx-docker/) - Docker-focused comparison
- [Roundcube vs SnappyMail](https://www.saashub.com/compare-roundcube-vs-snappymail) - Feature/usage comparison
- [5 Amazing Open Source Email Clients for Webmail 2026](https://forwardemail.net/en/blog/open-source/webmail-email-clients) - Webmail options
- [Replacing Baikal with Radicale](https://petermolnar.net/article/replacing-baikal-with-radicale-for-carrdav-and-caldav/) - Migration guide, comparison
- [Caddy Reverse Proxy with WireGuard](https://geoff.tuxpup.com/posts/caddy_and_wireguard/) - Tunnel architecture example
- [DAVx5 Manual](https://manual.davx5.com/accounts_collections.html) - Android CalDAV/CardDAV setup

**Auto-Discovery Guides:**
- [CardDAV CalDAV autodiscovery on iOS](https://www.staze.org/carddav-caldav-autodiscovery-on-mac-os-x-and-ios/) - Well-known URLs for Apple devices
- [CalDAV Auto Discovery](https://www.axigen.com/documentation/caldav-auto-discovery-p47120765) - Server configuration
- [WebDAV System CalDAV Discovery](https://www.webdavsystem.com/server/creating_caldav_carddav/discovery/) - Discovery mechanisms

**Community Discussions:**
- [Radicale Calendar Sharing Support](https://github.com/Kozea/Radicale/discussions/1499) - Sharing limitations
- [Radicale Rights File Setup](https://github.com/Kozea/Radicale/issues/947) - ACL configuration
- [Stalwart JMAP for Calendars/Contacts](https://linuxiac.com/stalwart-moves-beyond-email-a-full-collaboration-server-is-on-the-horizon/) - Full collaboration platform

**Docker Mailserver Examples:**
- [Mailu Behind Reverse Proxy](https://mailu.io/master/reverse.html) - Reverse proxy patterns
- [Docker Mailserver Behind Proxy](https://docker-mailserver.github.io/docker-mailserver/latest/examples/tutorials/mailserver-behind-proxy/) - Nginx/Traefik examples
- [Running Mailcow Behind Reverse Proxy](https://www.bonfert.io/2021/02/running-mailcow-behind-a-reverse-proxy/) - Subdomain configuration

### Tertiary (LOW confidence - needs verification)

**Community Opinions:**
- [SoGo vs RoundCube vs SnappyMail](https://forum.cloudron.io/topic/10663/sogo-vs-roundcube-vs-snappymail-vs-nextcloud-recommendations-and-trade-offs-especially-aliases-forwarding) - Feature trade-offs
- [Hacker News: SnappyMail](https://brianlovin.com/hn/31257672) - Community reception
- [Best Practices for Caddy with WireGuard](https://caddy.community/t/best-practices-for-running-an-https-reverse-proxy-behind-a-wireguard-tunnel/25053) - Tunnel architecture discussion

## Metadata

**Confidence breakdown:**
- Standard stack (webmail/CalDAV/reverse proxy): HIGH - All options have official documentation, production use, active development
- Caddy for reverse proxy: HIGH - Official Go-based server, automatic HTTPS verified in docs, matches DarkPipe stack
- Radicale vs Baikal: MEDIUM-HIGH - Both stable, Radicale more active development (verified GitHub activity), Baikal slower updates
- IMAP passthrough authentication: HIGH - Standard webmail pattern, documented in Roundcube/SnappyMail official configs
- Well-known URL auto-discovery: HIGH - RFC 6764 standard, verified in iOS/macOS/Android documentation
- Calendar sharing with Radicale: MEDIUM - Rights file approach documented, no web UI (confirmed in GitHub discussions)

**Research date:** 2026-02-14
**Valid until:** 60 days (Stalwart v1.0 expected Q2 2026 may add JMAP sharing features)

**Critical dependencies:**
- Caddy 2.x - Go-based reverse proxy, automatic HTTPS, stable API
- Roundcube 1.6.x - Elastic skin for mobile support, stable release
- SnappyMail 2.x - RainLoop fork, active development, modern stack
- Radicale 3.x - Python 3, stable, file-based storage
- Stalwart 0.15.4+ - Built-in CalDAV/CardDAV when Stalwart is mail server
- Docker Compose profiles - Standard feature in Compose v2 (2021+)