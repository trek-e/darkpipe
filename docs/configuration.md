# DarkPipe Configuration Reference

Complete reference for all configurable aspects of DarkPipe.

> **Container Runtime:** All examples use `docker compose` commands. Podman is fully supported via `podman-compose` with override files. See the [Podman platform guide](../deploy/platform-guides/podman.md) for details.

## Environment Variable Reference Files

Both deployment targets include `.env.example` files documenting all available environment variables with defaults, descriptions, and required/optional markers:

- `cloud-relay/.env.example` — All cloud relay variables (relay core, mTLS, TLS monitoring, queue, S3 overflow, Caddy)
- `home-device/.env.example` — All home device variables (domain, users, mail server, profile server, monitoring, webmail, CalDAV)

Copy these to `.env` and customize for your deployment:

```bash
cp cloud-relay/.env.example cloud-relay/.env
cp home-device/.env.example home-device/.env
```

## Environment Variables

Environment variables are set in `.env` files alongside `docker-compose.yml` files, or exported in the shell before running docker compose commands.

### Cloud Relay Variables

**Core Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_HOSTNAME` | relay.example.com | FQDN of the cloud relay server (must match PTR record) |
| `RELAY_DOMAIN` | example.com | Primary mail domain |
| `RELAY_LISTEN_ADDR` | 127.0.0.1:10025 | Address for Go relay service to listen on (receives from Postfix) |
| `RELAY_TRANSPORT` | wireguard | Transport type: `wireguard` or `mtls` |
| `RELAY_HOME_ADDR` | 10.8.0.2:25 | Address of home device mail server (via transport tunnel) |
| `RELAY_MAX_MESSAGE_BYTES` | 52428800 | Maximum message size in bytes (50MB default) |
| `RELAY_DEBUG` | false | Enable debug logging with full PII (email addresses, tokens) in SMTP session logs |

**WireGuard Transport**

| Variable | Default | Description |
|----------|---------|-------------|
| `WIREGUARD_CONFIG` | /etc/wireguard/wg0.conf | Path to WireGuard configuration file |
| `WIREGUARD_INTERFACE` | wg0 | WireGuard interface name |

**mTLS Transport**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_CA_CERT` | /certs/ca.crt | Path to internal CA certificate |
| `RELAY_CLIENT_CERT` | /certs/client.crt | Path to client certificate (for cloud relay) |
| `RELAY_CLIENT_KEY` | /certs/client.key | Path to client private key |

**TLS Monitoring**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_STRICT_MODE` | false | If true, reject connections that don't support TLS |
| `RELAY_WEBHOOK_URL` | (empty) | Webhook URL for TLS notification alerts |

**Ephemeral Storage Verification**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_EPHEMERAL_CHECK_INTERVAL` | 60 | Seconds between checks that no mail remains on disk |

**Queue Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_QUEUE_ENABLED` | true | Enable offline queue (queue vs bounce when home unreachable) |
| `RELAY_QUEUE_KEY_PATH` | /data/queue-keys/identity | Path to age identity key for queue encryption |
| `RELAY_QUEUE_SNAPSHOT_PATH` | /data/queue-state/snapshot.json | Path to queue state snapshot |
| `RELAY_QUEUE_TIMEOUT` | 604800 | Queue timeout in seconds (7 days default) |

**S3 Overflow Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_OVERFLOW_ENABLED` | false | Enable S3 overflow when local queue exceeds threshold |
| `RELAY_OVERFLOW_THRESHOLD` | 104857600 | Threshold in bytes before using S3 (100MB default) |
| `RELAY_OVERFLOW_ENDPOINT` | gateway.storjshare.io | S3-compatible endpoint URL |
| `RELAY_OVERFLOW_BUCKET` | darkpipe-queue | S3 bucket name |
| `RELAY_OVERFLOW_ACCESS_KEY` | (required if enabled) | S3 access key ID |
| `RELAY_OVERFLOW_SECRET_KEY` | (required if enabled) | S3 secret access key |
| `RELAY_OVERFLOW_REGION` | auto | S3 region (or "auto" for providers that don't need it) |

**Certbot (Let's Encrypt)**

| Variable | Default | Description |
|----------|---------|-------------|
| `CERTBOT_EMAIL` | (required) | Email address for Let's Encrypt certificate notifications |
| `CERTBOT_DOMAINS` | ${RELAY_HOSTNAME} | Comma-separated list of domains for certificate |

**Caddy Reverse Proxy**

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBMAIL_DOMAINS` | mail.${RELAY_DOMAIN} | Domains for webmail reverse proxy |
| `AUTOCONFIG_DOMAINS` | autoconfig.${RELAY_DOMAIN} | Domain for Thunderbird autoconfig |
| `AUTODISCOVER_DOMAINS` | autodiscover.${RELAY_DOMAIN} | Domain for Outlook autodiscover |

### Home Device Variables

**Core Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `MAIL_DOMAIN` | example.com | Primary mail domain |
| `MAIL_HOSTNAME` | mail.example.com | FQDN of mail server |
| `ADMIN_EMAIL` | admin@example.com | Admin email address (first user created) |
| `ADMIN_PASSWORD` | changeme | Admin password (CHANGE THIS!) |

**Mail Server Selection**

| Variable | Default | Description |
|----------|---------|-------------|
| `MAIL_SERVER_TYPE` | stalwart | Mail server type: `stalwart`, `maddy`, or `postfix-dovecot` |

**Postfix+Dovecot Specific**

| Variable | Default | Description |
|----------|---------|-------------|
| `VIRTUAL_DOMAINS` | ${MAIL_DOMAIN} | Comma-separated list of virtual domains |

**CalDAV/CardDAV Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `CALDAV_URL` | (auto-configured) | CalDAV server URL (for profile generation) |
| `CARDDAV_URL` | (auto-configured) | CardDAV server URL (for profile generation) |

**Stalwart Specific**

| Variable | Default | Description |
|----------|---------|-------------|
| `STALWART_ADMIN_USER` | ${ADMIN_EMAIL} | Admin username for Stalwart web UI |
| `STALWART_ADMIN_PASSWORD` | ${ADMIN_PASSWORD} | Admin password for Stalwart web UI |

**Profile Server**

| Variable | Default | Description |
|----------|---------|-------------|
| `PROFILE_SERVER_PORT` | 8090 | Port for profile server to listen on |
| `PROFILE_DEBUG` | false | Enable debug logging with full PII (email addresses) in profile server logs |

**Monitoring Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `MONITOR_ALERT_EMAIL` | (empty) | Email address for monitoring alerts |
| `MONITOR_WEBHOOK_URL` | (empty) | Webhook URL for monitoring alerts (Slack, Discord, etc.) |
| `MONITOR_HEALTHCHECK_URL` | (empty) | Healthcheck.io ping URL for uptime monitoring |
| `MONITOR_CLI_ALERT_PATH` | /data/monitoring/cli-alerts.json | Path to CLI alert state file |
| `MONITOR_LOG_PATH` | /var/log/mail.log | Path to mail log for parsing |
| `MONITOR_CERT_PATHS` | (empty) | Comma-separated paths to certificates to monitor for expiry |

**Rspamd Configuration**

Rspamd is configured via files in `home-device/spam-filter/rspamd/local.d/` and `override.d/`. No environment variables needed for basic operation.

**Redis Configuration**

Redis is configured via `home-device/spam-filter/redis/redis.conf`. No environment variables needed.

## Container Security Configuration

All compose services are configured with security hardening directives by default. These are set in the compose files and generally should not be modified:

- `security_opt: [no-new-privileges:true]` — Prevents privilege escalation
- `cap_drop: [ALL]` — Drops all Linux capabilities
- `cap_add: [...]` — Selectively re-adds only required capabilities
- `read_only: true` — Read-only root filesystem
- `tmpfs: [...]` — Explicit writable tmpfs mounts

All custom Dockerfiles include HEALTHCHECK instructions for container health monitoring.

Run `bash scripts/verify-container-security.sh` to audit these directives.

## Compose Profiles

DarkPipe uses compose profiles to select which components to run. Profiles are specified with the `--profile` flag.

> **Podman users:** When using `podman-compose`, add the override file (`-f docker-compose.podman.yml`) to each command. Override files adjust volume mounts and security options for Podman compatibility. See the [Podman platform guide](../deploy/platform-guides/podman.md).

### Mail Server Profiles

Select exactly one mail server profile:

| Profile | Mail Server | Description |
|---------|-------------|-------------|
| `stalwart` | Stalwart 0.15.4 | Modern all-in-one (IMAP4rev2, JMAP, built-in CalDAV/CardDAV) |
| `maddy` | Maddy 0.8.2 | Minimal Go-based single binary |
| `postfix-dovecot` | Postfix + Dovecot | Traditional MTA + IMAP (battle-tested) |

### Webmail Profiles

Select exactly one webmail profile:

| Profile | Webmail | Description |
|---------|---------|-------------|
| `roundcube` | Roundcube 1.6.13 | Traditional, feature-rich, PHP-based |
| `snappymail` | SnappyMail 2.38.2 | Modern, fast, lightweight |

### CalDAV/CardDAV Profiles

Select `radicale` profile ONLY if using `maddy` or `postfix-dovecot`. Stalwart has built-in CalDAV/CardDAV.

| Profile | Server | Description |
|---------|--------|-------------|
| `radicale` | Radicale 3.6.0 | Standalone CalDAV/CardDAV server (for Maddy/Postfix+Dovecot) |

### Example Profile Combinations

**Default Stack (Stalwart + SnappyMail):**
```bash
docker compose --profile stalwart --profile snappymail up -d
```

**Conservative Stack (Postfix+Dovecot + Roundcube + Radicale):**
```bash
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

**Minimal Stack (Maddy + SnappyMail + Radicale):**
```bash
docker compose --profile maddy --profile snappymail --profile radicale up -d
```

**Custom (Stalwart + Roundcube, no separate CalDAV server):**
```bash
docker compose --profile stalwart --profile roundcube up -d
```

## Mail Server Configuration

Each mail server has its own configuration files located in the `home-device/` directory.

### Stalwart Configuration

**Config file:** `home-device/stalwart/config.toml`

Key sections:

```toml
[server]
hostname = "mail.example.com"

[server.listener."smtp"]
bind = ["0.0.0.0:25"]
protocol = "smtp"

[server.listener."submission"]
bind = ["0.0.0.0:587"]
protocol = "smtp"

[server.listener."imaps"]
bind = ["0.0.0.0:993"]
protocol = "imap"

[server.listener."http"]
bind = ["0.0.0.0:8080"]
protocol = "http"
```

**User management:**
- Web UI: http://HOME_DEVICE_IP:8080
- Users, domains, aliases managed via web interface

**CalDAV/CardDAV:**
- Built-in, enabled by default
- Access via same credentials as mail
- URLs: https://mail.example.com/dav/ (CalDAV and CardDAV)

### Maddy Configuration

**Config file:** `home-device/maddy/maddy.conf`

Key sections:

```
hostname mail.example.com

tls file /data/certs/fullchain.pem /data/certs/privkey.pem

smtp tcp://0.0.0.0:25 {
    # Inbound SMTP
}

submission tcp://0.0.0.0:587 {
    # SMTP submission (authenticated)
}

imap tcp://0.0.0.0:993 {
    # IMAP server
    tls mandatory
}
```

**User management:**
- Edit `home-device/maddy/setup-users.sh`
- Add user lines: `maddyctl creds create user@example.com`
- Restart Maddy container: `docker compose restart maddy`

**CalDAV/CardDAV:**
- NOT built-in, use Radicale profile
- Add `--profile radicale` when starting containers

### Postfix+Dovecot Configuration

**Postfix config:** `home-device/postfix-dovecot/postfix/`
**Dovecot config:** `home-device/postfix-dovecot/dovecot/`

Key files:

**Virtual mailboxes:** `home-device/postfix-dovecot/postfix/vmailbox`
```
user1@example.com example.com/user1/
user2@example.com example.com/user2/
```

**Virtual aliases:** `home-device/postfix-dovecot/postfix/virtual`
```
# Aliases
info@example.com user1@example.com
sales@example.com user1@example.com

# Catch-all
@example.com user1@example.com
```

**Dovecot users:** `home-device/postfix-dovecot/dovecot/users`
```
user1@example.com:{PLAIN}password1
user2@example.com:{PLAIN}password2
```

**After modifying config files:**
```bash
docker compose restart postfix-dovecot
```

**CalDAV/CardDAV:**
- NOT built-in, use Radicale profile

## Transport Configuration

### WireGuard Configuration

**Cloud relay:** `/etc/wireguard/wg0.conf`
```ini
[Interface]
Address = 10.8.0.1/24
PrivateKey = <cloud-private-key>
ListenPort = 51820

[Peer]
PublicKey = <home-public-key>
AllowedIPs = 10.8.0.2/32
```

**Home device:** `/etc/wireguard/wg0.conf`
```ini
[Interface]
Address = 10.8.0.2/24
PrivateKey = <home-private-key>

[Peer]
PublicKey = <cloud-public-key>
Endpoint = YOUR_VPS_IP:51820
AllowedIPs = 10.8.0.1/32
PersistentKeepalive = 25
```

**Management commands:**
```bash
# Start WireGuard
sudo wg-quick up wg0

# Stop WireGuard
sudo wg-quick down wg0

# Check status
sudo wg show

# Enable on boot
sudo systemctl enable wg-quick@wg0
```

### mTLS Configuration

**Set environment variables in cloud-relay/.env:**
```bash
RELAY_TRANSPORT=mtls
RELAY_CA_CERT=/certs/ca.crt
RELAY_CLIENT_CERT=/certs/relay.crt
RELAY_CLIENT_KEY=/certs/relay.key
```

**Mount certificates in docker-compose.yml:**
```yaml
volumes:
  - ./certs:/certs:ro
```

**Certificate management with step-ca:**

Generate certificates:
```bash
# On home device (CA server)
step ca certificate relay.internal relay.crt relay.key

# Copy to cloud relay
scp relay.crt relay.key ca.crt user@cloud-relay-vps:/path/to/certs/
```

Automatic rotation (configured in step-ca):
- Default: 30 days
- Configurable: 30, 60, or 90 days
- Renewal happens automatically via step-ca

## DNS Configuration

### Supported DNS Providers

**Cloudflare (API)**
- Requires: API token with Zone.DNS.Edit permission
- Set: `export CLOUDFLARE_API_TOKEN=your_token`
- Usage: `dns-setup --provider cloudflare --apply`

**AWS Route53 (API)**
- Requires: AWS credentials with route53:ChangeResourceRecordSets permission
- Set: `export AWS_ACCESS_KEY_ID=your_key` and `export AWS_SECRET_ACCESS_KEY=your_secret`
- Usage: `dns-setup --provider route53 --apply`

**Manual (Any Provider)**
- No credentials needed
- Usage: `dns-setup` (outputs records to copy manually)

### DNS Records Created

**MX Record:**
```
example.com.  IN  MX  10  relay.example.com.
```

**A Record (for relay):**
```
relay.example.com.  IN  A  YOUR_VPS_IP
```

**SPF Record:**
```
example.com.  IN  TXT  "v=spf1 mx -all"
```

**DKIM Record:**
```
default._domainkey.example.com.  IN  TXT  "v=DKIM1; k=rsa; p=PUBKEY_BASE64"
```

**DMARC Record:**
```
_dmarc.example.com.  IN  TXT  "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com"
```

### DKIM Key Rotation

Generate new DKIM key:
```bash
dns-setup --domain example.com --rotate-dkim
```

This generates a new key pair and outputs the new DNS record. Update DNS, wait for propagation, then update mail server config.

## Offline Queue Configuration

### Queue Behavior

**Enable queueing (default):**
```bash
RELAY_QUEUE_ENABLED=true
```
Mail is queued encrypted when home device is offline, delivered when it reconnects.

**Disable queueing (bounce immediately):**
```bash
RELAY_QUEUE_ENABLED=false
```
Mail is bounced immediately if home device is unreachable.

### Queue Encryption

Queue encryption uses age (filippo.io/age) with a recipient key.

**Key generation (done automatically by setup wizard):**
```bash
age-keygen -o /data/queue-keys/identity
```

Public key is derived from identity and used to encrypt queued messages.

### S3 Overflow

**Enable S3 overflow:**
```bash
RELAY_OVERFLOW_ENABLED=true
RELAY_OVERFLOW_THRESHOLD=104857600  # 100MB
RELAY_OVERFLOW_ENDPOINT=gateway.storjshare.io
RELAY_OVERFLOW_BUCKET=darkpipe-queue
RELAY_OVERFLOW_ACCESS_KEY=your_access_key
RELAY_OVERFLOW_SECRET_KEY=your_secret_key
```

**Supported S3 providers:**
- Storj (recommended, decentralized, $4/TB)
- AWS S3
- MinIO (self-hosted)
- Any S3-compatible service

**How it works:**
1. Local queue accumulates encrypted messages
2. When queue size exceeds threshold, overflow to S3
3. Local queue drains first, then S3 overflow
4. Messages deleted from S3 after successful delivery

## Monitoring Configuration

### Alert Channels

**Email alerts:**
```bash
MONITOR_ALERT_EMAIL=your-email@gmail.com
```

**Webhook alerts (Slack, Discord, etc.):**
```bash
MONITOR_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

**Healthcheck.io integration:**
```bash
MONITOR_HEALTHCHECK_URL=https://hc-ping.com/your-uuid-here
```

### Certificate Monitoring

Monitor custom certificates for expiry:
```bash
MONITOR_CERT_PATHS=/etc/letsencrypt/live/example.com/fullchain.pem,/certs/relay.crt
```

Alerts sent when certificates are within 7 days of expiry.

### Health Check Intervals

Health checks are defined in docker-compose.yml per service:

```yaml
healthcheck:
  test: ["CMD", "nc", "-z", "localhost", "25"]
  interval: 30s   # Check every 30 seconds
  timeout: 10s    # Timeout after 10 seconds
  retries: 3      # 3 failures before marking unhealthy
```

## Custom Builds (GitHub Actions)

Build custom stack images with component selection via GitHub Actions.

### Trigger Custom Build

1. Fork https://github.com/trek-e/darkpipe
2. Go to Actions > "Build Custom Stack"
3. Click "Run workflow"
4. Select components:
   - Mail server: stalwart, maddy, postfix-dovecot
   - Webmail: roundcube, snappymail
   - Groupware: radicale, stalwart-builtin, none
5. Enter custom tag (e.g., "my-stack")
6. Run workflow

Resulting image: `ghcr.io/YOUR_USERNAME/darkpipe/home-device:my-stack`

### Use Custom Image

Update docker-compose.yml to use your custom image:

```yaml
services:
  # Remove individual service definitions
  # Use custom combined image instead
  home-device:
    image: ghcr.io/YOUR_USERNAME/darkpipe/home-device:my-stack
    ports:
      - "25:25"
      - "587:587"
      - "993:993"
      - "8080:8080"
    # ... rest of config
```

---

Last Updated: 2026-03-12

License: AGPLv3 - See [LICENSE](../LICENSE)
