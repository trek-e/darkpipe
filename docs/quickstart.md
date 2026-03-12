# DarkPipe Quick Start Guide

This guide walks you through deploying DarkPipe from zero to sending and receiving email.

**Time required:** 30-60 minutes for initial setup, plus 4-6 weeks for IP warmup (gradual sending volume increase for deliverability)

## Container Runtime

All examples in this guide use `docker compose` commands, which you can copy-paste directly. **Podman is fully supported** as an alternative container runtime via `podman-compose` with override files. See the [Podman platform guide](../deploy/platform-guides/podman.md) for setup instructions and key differences.

## Prerequisites

Before you begin, ensure you have:

1. **A domain name you control**
   - With DNS API access for automation (Cloudflare or Route53), OR
   - Ability to manually edit DNS records

2. **A VPS with port 25 access**
   - See [docs/vps-providers.md](vps-providers.md) for compatible providers
   - Recommended: Hetzner, Vultr, OVH, Linode
   - Minimum specs: 1 vCPU, 1GB RAM, 20GB SSD
   - Cost: $3-6/month

3. **A home device running containers**
   - Raspberry Pi 4+ (4GB+ RAM), OR
   - Any x64/arm64 Linux system with Docker or Podman, OR
   - NAS platforms: TrueNAS Scale, Unraid, Synology, Proxmox
   - See [platform guides](../deploy/platform-guides/) for your platform

4. **Time**
   - Initial setup: 30-60 minutes
   - IP reputation warmup: 4-6 weeks (gradual sending volume increase)

## Step 1: Provision Your Cloud Relay

### 1.1 Choose a VPS Provider

Select a provider that allows port 25 access. See [docs/vps-providers.md](vps-providers.md) for full compatibility matrix.

**Recommended for beginners:**
- **Hetzner**: No restrictions, excellent price/performance, EU-based
- **Vultr**: Port 25 available after account verification (contact support)

**VPS requirements:**
- Operating system: Ubuntu 24.04 LTS or Debian 12
- Ports: 25 (SMTP), 80 (HTTP), 443 (HTTPS)
- IPv4 address (required)
- Reverse DNS (PTR record) capability

### 1.2 Set Up Your VPS

1. Provision the smallest available VPS (1 vCPU, 1GB RAM, 20GB SSD)

2. Set the hostname to match your relay hostname:
   ```bash
   hostnamectl set-hostname relay.yourdomain.com
   ```

3. Configure reverse DNS (PTR record):
   - Most providers have a control panel option for setting PTR records
   - Set PTR record for your VPS IP to: relay.yourdomain.com
   - This is required for email deliverability

4. Update the system:
   ```bash
   apt update && apt upgrade -y
   ```

5. Install Docker and Docker Compose:
   ```bash
   curl -fsSL https://get.docker.com | sh
   ```

   > **Podman users:** Skip this step. Install Podman 5.3+ and podman-compose instead. See the [Podman platform guide](../deploy/platform-guides/podman.md) for installation and override file setup.

### 1.3 Verify Port 25 Access

Test that port 25 is open and reachable:

```bash
# Install telnet if not present
apt install -y telnet

# Test outbound connection to Gmail's mail server
telnet gmail-smtp-in.l.google.com 25
```

You should see a response like:
```
220 mx.google.com ESMTP ...
```

Type `QUIT` and press Enter to exit.

If the connection times out or is refused, port 25 is blocked. Contact your VPS provider.

## Step 2: Set Up Transport Layer

Choose one of two transport options: WireGuard (recommended) or mTLS.

### Option A: WireGuard (Recommended)

WireGuard is simpler to set up and provides a full encrypted tunnel between cloud and home.

**On cloud relay VPS:**
```bash
# Install WireGuard
apt install -y wireguard

# Download setup script
curl -LO https://raw.githubusercontent.com/trek-e/darkpipe/main/deploy/wireguard/cloud-setup.sh
chmod +x cloud-setup.sh

# Run setup (generates keys and config)
./cloud-setup.sh

# This creates /etc/wireguard/wg0.conf
# Note the public key displayed - you'll need it for home device setup
```

**On home device:**
```bash
# Install WireGuard
# (Raspberry Pi OS / Ubuntu)
sudo apt install -y wireguard

# Download setup script
curl -LO https://raw.githubusercontent.com/trek-e/darkpipe/main/deploy/wireguard/home-setup.sh
chmod +x home-setup.sh

# Run setup with cloud relay's public key and endpoint
./home-setup.sh --cloud-endpoint YOUR_VPS_IP:51820 --cloud-pubkey CLOUD_PUBLIC_KEY

# Start WireGuard
sudo wg-quick up wg0
sudo systemctl enable wg-quick@wg0
```

**Verify connectivity:**
```bash
# On cloud relay
ping 10.8.0.2

# On home device
ping 10.8.0.1
```

Both pings should succeed. If not, check firewall rules (allow UDP port 51820).

### Option B: mTLS (Minimal Footprint Alternative)

mTLS uses mutual TLS authentication without requiring a VPN. Suitable for minimal systems or when WireGuard is not available.

**Set up internal PKI with step-ca:**
```bash
# On home device (acts as CA)
curl -LO https://github.com/trek-e/darkpipe/raw/main/deploy/pki/step-ca-setup.sh
chmod +x step-ca-setup.sh
./step-ca-setup.sh

# Generate certificates for cloud relay
step ca certificate relay.internal relay-cert.pem relay-key.pem

# Copy relay-cert.pem, relay-key.pem, and ca.crt to cloud relay VPS
```

**Configure mTLS in docker-compose.yml:**
- Uncomment mTLS environment variables
- Mount certificate files into containers
- See [docs/configuration.md](configuration.md) for full mTLS setup

## Step 3: Run the Setup Wizard

The darkpipe-setup wizard generates configuration files for both cloud relay and home device.

### 3.1 Download Setup Tool

Download the appropriate binary for your platform:

**Linux (amd64):**
```bash
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64
```

**Linux (arm64, Raspberry Pi):**
```bash
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-arm64
chmod +x darkpipe-setup-linux-arm64
```

**macOS (Intel):**
```bash
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-amd64
chmod +x darkpipe-setup-darwin-amd64
```

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-arm64
chmod +x darkpipe-setup-darwin-arm64
```

### 3.2 Run the Wizard

```bash
./darkpipe-setup-linux-amd64
```

The wizard will ask:

1. **Domain name**: Your primary mail domain (e.g., example.com)
2. **Relay hostname**: FQDN for your cloud relay (e.g., relay.example.com)
3. **Mail server**: Stalwart (default), Maddy, or Postfix+Dovecot
4. **Webmail**: Roundcube or SnappyMail
5. **CalDAV/CardDAV**: Radicale (if not using Stalwart) or Stalwart built-in
6. **Transport type**: WireGuard or mTLS
7. **Admin email**: Your admin email address
8. **Admin password**: Password for admin account

The wizard generates:
- `cloud-relay/.env` and `cloud-relay/docker-compose.yml`
- `home-device/.env` and `home-device/docker-compose.yml`
- Configuration files for selected components

### 3.3 Review Generated Configuration

Check the generated .env files and adjust if needed. Reference `.env.example` files for all available variables with descriptions and defaults:

```bash
# See all available variables and their defaults
cat cloud-relay/.env.example
cat home-device/.env.example
```

**cloud-relay/.env:**
```bash
RELAY_HOSTNAME=relay.example.com
RELAY_DOMAIN=example.com
RELAY_TRANSPORT=wireguard
RELAY_HOME_ADDR=10.8.0.2:25
```

**home-device/.env:**
```bash
MAIL_DOMAIN=example.com
MAIL_HOSTNAME=mail.example.com
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme  # Change this!
```

## Step 4: Configure DNS

DarkPipe requires several DNS records for email authentication and deliverability.

### 4.1 Download DNS Setup Tool

```bash
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/dns-setup-linux-amd64
chmod +x dns-setup-linux-amd64
```

### 4.2 Run DNS Setup (Dry-Run)

First, run in dry-run mode to preview changes:

```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip YOUR_VPS_IP
```

This will display:
- MX record (points to relay.example.com)
- A record for relay.example.com (your VPS IP)
- SPF record (authorized senders)
- DKIM record (public key for signature verification)
- DMARC record (policy enforcement)

### 4.3 Apply DNS Changes

**Option A: Automatic (Cloudflare or Route53)**

Set up API credentials:

*Cloudflare:*
```bash
export CLOUDFLARE_API_TOKEN=your_token_here
```

*Route53:*
```bash
export AWS_ACCESS_KEY_ID=your_key_here
export AWS_SECRET_ACCESS_KEY=your_secret_here
```

Apply changes:
```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip YOUR_VPS_IP \
  --provider cloudflare \
  --apply
```

**Option B: Manual**

Copy the displayed records and add them manually via your DNS provider's control panel.

### 4.4 Validate DNS

After DNS propagation (5-60 minutes), validate:

```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --validate-only
```

This checks that all required records are present and correct.

## Step 5: Deploy Services

### 5.1 Deploy Cloud Relay

On your VPS:

```bash
cd cloud-relay
docker compose up -d
```

> **Podman users:** Use `podman-compose` with the override file: `podman-compose -f docker-compose.yml -f docker-compose.podman.yml up -d`. The cloud relay requires rootful Podman for port 25 binding. See the [Podman platform guide](../deploy/platform-guides/podman.md).

Verify services are running:
```bash
docker compose ps
docker compose logs -f
```

Expected containers:
- darkpipe-relay (Postfix + Go relay service)
- caddy (reverse proxy)

### 5.2 Deploy Home Device

On your home device:

```bash
cd home-device

# Start with your selected profiles
# Example for Stalwart + SnappyMail:
docker compose --profile stalwart --profile snappymail up -d

# Example for Postfix+Dovecot + Roundcube + Radicale:
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

> **Podman users:** Add `-f docker-compose.podman.yml` after the base compose file. The home device can run rootless. See the [Podman platform guide](../deploy/platform-guides/podman.md).

Verify services are running:
```bash
docker compose ps
docker compose logs -f
```

Expected containers (for Stalwart + SnappyMail):
- stalwart (mail server)
- snappymail (webmail)
- rspamd (spam filter)
- redis (Rspamd backend)
- profile-server (device onboarding)
- caddy (reverse proxy)

## Step 6: Test Email Delivery

### 6.1 Send Test Email to Your Domain

Use the dns-setup tool's test feature:

```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --send-test \
  --from external@gmail.com \
  --to admin@example.com
```

Or send manually from an external email account (Gmail, Outlook, etc.) to admin@example.com.

### 6.2 Check Webmail for Received Mail

1. Access webmail at https://mail.example.com (or http://HOME_DEVICE_IP:8080 if DNS not configured)

2. Log in with admin credentials:
   - Email: admin@example.com
   - Password: (from home-device/.env ADMIN_PASSWORD)

3. Check inbox for test message

### 6.3 Send Outbound Test Email

From webmail:

1. Compose a new email to an external address (your personal Gmail, etc.)
2. Send the message
3. Check the external account for delivery

**Check spam folder** if not in inbox (new IP reputation takes time).

### 6.4 Verify Email Authentication

Forward the received test email (sent from your DarkPipe) to check-auth@verifier.port25.com.

You'll receive an automated reply showing:
- SPF: PASS
- DKIM: PASS
- DMARC: PASS

If any fail, review DNS configuration with `dns-setup --validate-only`.

## Step 7: Onboard Devices

### 7.1 Access Profile Server

Navigate to http://HOME_DEVICE_IP:8090 in your browser.

### 7.2 iOS/macOS Configuration

1. Click "iOS/macOS Profile"
2. Download the .mobileconfig file
3. On iOS: Open the file, follow prompts to install profile
4. On macOS: Open the file, System Preferences > Profiles > Install

This configures:
- IMAP (mail.example.com:993)
- SMTP (mail.example.com:587)
- CalDAV (if enabled)
- CardDAV (if enabled)

### 7.3 Android Configuration

1. Click "QR Code"
2. Scan with Android mail app (Gmail app, K-9 Mail, FairEmail)
3. Follow prompts to add account

### 7.4 Desktop Mail Clients (Thunderbird, Outlook)

Thunderbird and Outlook support autodiscover.

**Manual configuration if autodiscover fails:**

*IMAP Settings:*
- Server: mail.example.com
- Port: 993
- Security: SSL/TLS
- Authentication: Normal password
- Username: your-email@example.com

*SMTP Settings:*
- Server: mail.example.com
- Port: 587
- Security: STARTTLS
- Authentication: Normal password
- Username: your-email@example.com

### 7.5 Generate App Passwords (Optional)

For mail clients that don't support regular passwords or for enhanced security:

1. Access profile server: http://HOME_DEVICE_IP:8090
2. Click "App Passwords"
3. Generate a new password for specific device/app
4. Use this password instead of your main password when configuring mail client

## Step 8: Monitor and Maintain

### 8.1 Access Monitoring Dashboard

Navigate to http://HOME_DEVICE_IP:8090/monitoring

Dashboard shows:
- Mail queue status (cloud and home)
- Service health (all containers)
- Certificate expiry dates
- Recent delivery logs

### 8.2 Configure Alerts

Edit home-device/.env:

```bash
# Email alerts
MONITOR_ALERT_EMAIL=your-personal-email@gmail.com

# Webhook alerts (Slack, Discord, etc.)
MONITOR_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Healthcheck.io integration
MONITOR_HEALTHCHECK_URL=https://hc-ping.com/your-uuid-here
```

Restart profile-server to apply:
```bash
docker compose restart profile-server
```

### 8.3 Check Logs

By default, logs redact email addresses for privacy (e.g., `s***r@example.com`). To see full email addresses for troubleshooting, enable debug mode:

```bash
# Cloud relay — add to cloud-relay/.env
RELAY_DEBUG=true

# Profile server — add to home-device/.env
PROFILE_DEBUG=true
```

Restart services after changing debug settings. **Disable debug mode after troubleshooting** — it logs full PII.

**Cloud relay logs:**
```bash
cd cloud-relay
docker compose logs -f relay
```

**Home device logs:**
```bash
cd home-device
docker compose logs -f stalwart  # or maddy, postfix-dovecot
docker compose logs -f rspamd
```

### 8.4 IP Warmup

New IP addresses have no sending reputation. Warm up gradually over 4-6 weeks:

**Week 1:** Send 10-20 emails/day to engaged recipients
**Week 2:** 50-100 emails/day
**Week 3:** 200-500 emails/day
**Week 4-6:** Gradually increase to your normal sending volume

**Tips:**
- Start with emails to people you know who will open and respond
- Avoid sending to large lists immediately
- Monitor spam folder placement
- Use [mail-tester.com](https://www.mail-tester.com/) to check deliverability score

## What's Next

**Add Users and Domains:**
- Stalwart: Access web UI at http://HOME_DEVICE_IP:8080
- Maddy: Edit home-device/maddy/setup-users.sh and restart container
- Postfix+Dovecot: Edit vmailbox and virtual files, restart container

**Migrate Existing Email:**
- See [docs/migration.md](migration.md) for detailed migration guide
- Supports Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, generic IMAP

**Configure Spam Filtering:**
- Access Rspamd web UI: http://HOME_DEVICE_IP:11334
- Train Bayesian filter with spam/ham samples
- Adjust greylisting and threshold settings

**Set Up Backups:**
- Back up home device mail data: docker volumes or mail-data directory
- Cloud relay has no mail to back up (by design)
- Recommended: Daily automated backups with retention

**Customize Configuration:**
- See [docs/configuration.md](configuration.md) for all environment variables
- Adjust memory limits, ports, queue behavior, etc.

**Improve Security:**
- Enable firewall on both cloud and home (allow only required ports)
- Set up fail2ban on cloud relay to prevent brute-force attacks
- Use strong passwords for all accounts
- Enable full-disk encryption on home device

## Troubleshooting

**Mail not being received:**
1. Check DNS records: `dns-setup --validate-only`
2. Verify port 25 is open on cloud relay VPS
3. Check cloud relay logs: `docker compose logs relay`
4. Verify WireGuard/mTLS tunnel is up: `ping 10.8.0.2` from cloud

**Mail not being sent:**
1. Check SMTP submission port 587 is accessible
2. Review home device mail server logs
3. Verify DKIM signing is working (check-auth@verifier.port25.com)
4. Check if IP is on blocklists: [MXToolbox](https://mxtoolbox.com/blacklists.aspx)

**Webmail not loading:**
1. Verify Caddy is running: `docker compose ps caddy`
2. Check Caddy logs: `docker compose logs caddy`
3. Ensure webmail container is healthy: `docker compose ps`

**Transport tunnel down:**
1. Check WireGuard status: `sudo wg show`
2. Verify firewall allows UDP 51820 (WireGuard) or TCP 443 (mTLS)
3. Check monitoring dashboard for connectivity alerts

**Services using too much memory:**
1. Review resource limits in docker-compose.yml
2. Consider switching to lighter components (Maddy instead of Stalwart, SnappyMail instead of Roundcube)
3. See platform guides for memory optimization tips

**Need help?**
- GitHub Discussions: https://github.com/trek-e/darkpipe/discussions
- GitHub Issues: https://github.com/trek-e/darkpipe/issues (for bugs)

---

Last Updated: 2026-03-12

License: AGPLv3 - See [LICENSE](../LICENSE)
