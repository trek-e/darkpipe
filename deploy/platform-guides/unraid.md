# DarkPipe on Unraid

Deploy DarkPipe on Unraid using Community Applications or Docker Compose.

## Prerequisites

**Hardware:**
- Unraid 6.12+ installation
- 4GB+ RAM allocated to Docker
- Array or cache pool with 20GB+ free space

**Software:**
- Unraid 6.12+ with Docker enabled
- Community Applications plugin (for Method 1)
- Docker Compose Manager plugin (for Method 2)

## Quick Start (Community Applications)

1. **Apps** → **Community Applications** → Search "DarkPipe"
2. **Install** DarkPipe cloud-relay template
3. **Configure** mail domain, relay hostname, admin password
4. **Apply**

**Note**: Community Applications template covers **cloud-relay only** (for VPS deployment). For the full home mail server stack, use Method 2 (Docker Compose).

## Detailed Steps

### Method 1: Community Applications XML Template (Cloud Relay Only)

This method installs the cloud-relay container on your Unraid VPS (if you're running Unraid as a VPS, which is rare).

**Step 1: Install Community Applications Plugin**

If not already installed:
1. Navigate to **Apps** → **Settings**
2. Enable **Community Applications**
3. Install the plugin

**Step 2: Search for DarkPipe**

1. Navigate to **Apps** → **Community Applications**
2. Search for `DarkPipe`
3. Click **DarkPipe-CloudRelay**

**Step 3: Configure Template**

The template will present configuration fields:

**Ports:**
- **SMTP Port**: 25 (required for mail delivery)
- **HTTP Port**: 80 (for Let's Encrypt challenges)
- **HTTPS Port**: 443 (for webmail reverse proxy)

**Paths:**
- **Postfix Queue**: `/mnt/user/appdata/darkpipe/postfix-queue` (persistent mail queue)
- **Let's Encrypt Certificates**: `/mnt/user/appdata/darkpipe/certbot` (TLS certificates)
- **Queue Data**: `/mnt/user/appdata/darkpipe/queue` (message queue and encryption keys)
- **Configuration**: `/mnt/user/appdata/darkpipe/config` (DarkPipe config files)

**Environment Variables:**
- **Relay Hostname**: Your cloud relay hostname (e.g., `relay.example.com`)
- **Mail Domain**: Your primary mail domain (e.g., `example.com`)
- **Relay Listen Address**: `127.0.0.1:10025` (default)
- **Transport Protocol**: `wireguard` (default) or `mtls`
- **Home Device Address**: WireGuard IP of home device (e.g., `10.8.0.2:25`)
- **Max Message Size**: `52428800` (50MB default)
- **Enable Queue**: `true` (recommended)
- **Certbot Email**: Your email for Let's Encrypt notifications

**Step 4: Apply and Start**

1. Click **Apply**
2. Navigate to **Docker** tab
3. Verify **DarkPipe-CloudRelay** is running

**Limitations:**
- This template covers cloud-relay only
- Home mail server (Stalwart/Maddy/Postfix+Dovecot) requires Docker Compose (Method 2)
- Webmail, CalDAV, Rspamd require Docker Compose

### Method 2: Docker Compose Manager (Full Stack)

This method deploys the complete DarkPipe stack (cloud relay + mail server + webmail + spam filtering) on your Unraid home server.

**Step 1: Install Docker Compose Manager Plugin**

1. Navigate to **Apps** → **Community Applications**
2. Search for `Compose Manager` or `Docker Compose Manager`
3. Install the plugin

**Step 2: Create appdata Directory**

Via SSH or Unraid terminal:
```bash
mkdir -p /mnt/user/appdata/darkpipe
cd /mnt/user/appdata/darkpipe
```

**Step 3: Download Setup Tool**

```bash
# For x64 systems (most Unraid servers)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64

# Run setup
./darkpipe-setup-linux-amd64 setup
```

Setup will:
- Validate DNS records
- Test SMTP connectivity
- Generate `docker-compose.yml` with your configuration
- Create Docker secrets with secure permissions

**Step 4: Deploy via Compose Manager**

**Option A: Via Web UI**
1. Navigate to **Apps** → **Compose Manager**
2. Click **Add New Stack**
3. **Stack Name**: `darkpipe`
4. **Compose File**: Upload the generated `docker-compose.yml`
5. Click **Deploy**

**Option B: Via Terminal**
```bash
cd /mnt/user/appdata/darkpipe
docker compose up -d
```

**Step 5: Verify Services**

```bash
# Check service status
docker compose ps

# View logs
docker compose logs -f

# Check specific service
docker compose logs -f stalwart
```

## Platform-Specific Notes

### Network Mode: Bridge vs Host

**Bridge mode (default)** is recommended:
- Isolated networking
- Port mapping via Unraid UI
- Easier to manage

**Host mode** may be required if:
- SMTP port 25 has connectivity issues
- You need direct network access

To use host mode, edit `docker-compose.yml`:
```yaml
services:
  stalwart:
    network_mode: host
```

**Trade-off**: Host mode exposes all container ports directly. Less isolated.

### Storage Location: Array vs Cache

**Cache pool (SSD)** recommended for:
- `mail-data` (frequent random I/O)
- `rspamd-data` (database operations)
- `redis-data` (in-memory cache with persistence)

**Array (HDDs)** acceptable for:
- `postfix-queue` (sequential writes)
- `certbot-etc` (infrequent access)
- Backups

To use cache pool, ensure appdata share is set to:
- **Primary Storage**: Cache
- **Use Cache**: Yes

### Docker Secrets on Unraid

Unraid Docker Compose Manager supports Docker secrets via file-based secrets:

```yaml
secrets:
  admin_password:
    file: ./secrets/admin_password.txt
```

The setup tool automatically creates secrets files with 0600 permissions.

**Alternative (less secure)**: Use environment variables directly in Unraid Docker UI when creating containers via Community Applications template.

### Memory Limits

Unraid respects Docker Compose `deploy.resources.limits`:

```yaml
deploy:
  resources:
    limits:
      memory: 512M
```

Monitor memory usage via:
```bash
docker stats
```

If services are OOM killed, increase limits or add more RAM.

### IPv6 Support

Unraid supports IPv6 via Docker bridge networks. If your ISP provides IPv6:

1. Navigate to **Settings** → **Docker** → **IPv6**
2. Enable IPv6 support
3. Configure subnet (default: `fd00::/80`)

DarkPipe will automatically use IPv6 for cloud relay <-> home device communication if both support it.

## Troubleshooting

### Port 25 conflict with Unraid mail server

**Symptom**: Container fails to start with "port already in use"

**Cause**: Unraid may have a mail notification service on port 25.

**Fix**:
1. Navigate to **Settings** → **Notifications**
2. Change SMTP server to use alternate port (587 for submission)
3. Or disable local mail server

### Appdata on array instead of cache

**Symptom**: Slow mail operations, high disk activity

**Cause**: Mail data on array HDDs instead of cache SSD

**Fix**:
1. Navigate to **Shares** → **appdata**
2. Set **Primary Storage**: Cache
3. Set **Use Cache**: Yes
4. Run **Mover** to migrate existing data

### Docker Compose not found

**Symptom**: `docker compose` command not found

**Cause**: Older Unraid versions use `docker-compose` (v1) instead of `docker compose` (v2).

**Fix**:
```bash
# Install Docker Compose v2
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Verify
docker-compose version
```

Or use Compose Manager plugin which handles this automatically.

### Services unhealthy after Unraid reboot

**Symptom**: Containers show "Unhealthy" in Docker tab

**Cause**: Services may start before network is fully ready.

**Fix**:
1. Wait 2-3 minutes for network to stabilize
2. Navigate to **Docker** → **DarkPipe** → **Restart**
3. If persists, check logs for errors

## Performance Benchmarks

**Unraid on typical hardware (4-core CPU, 16GB RAM, SSD cache):**
- Mail delivery: <50ms
- IMAP FETCH: 5-20ms per message
- Webmail page load: 100-300ms
- RAM usage (Stalwart + SnappyMail): ~1.2GB
- Idle CPU: <2%

**Unraid on older hardware (dual-core, 8GB RAM, HDD only):**
- Mail delivery: 200-500ms
- IMAP FETCH: 50-100ms per message
- Webmail page load: 1-2 seconds
- RAM usage (Maddy + no webmail): ~600MB
- Idle CPU: <5%

## Security Considerations

1. **Network isolation**: Use custom Docker networks to isolate DarkPipe
2. **Backups**: Use Unraid appdata backup plugin to back up `/mnt/user/appdata/darkpipe`
3. **Updates**: Check for DarkPipe updates regularly, redeploy via Compose Manager
4. **Firewall**: Configure Unraid firewall to only allow required ports

## Next Steps

1. **Configure DNS**: Set up DKIM, SPF, DMARC (see Phase 4 guide)
2. **Test mail flow**: Send test email from Gmail to your@domain.com
3. **Set up backups**: Use Unraid appdata backup or rsync to external drive
4. **Monitor logs**: `docker compose logs -f rspamd` to watch spam filtering

## Alternative Runtimes

Unraid's Docker integration is tightly coupled with its built-in Docker engine and Community Applications ecosystem. Podman is **not applicable** on Unraid — the platform does not support alternative container runtimes. Use the native Docker engine via Community Applications or Docker Compose Manager as described in this guide.

For Podman deployment on supported platforms, see the [Podman Platform Guide](podman.md).

## See Also

- [Podman Platform Guide](podman.md) - Podman deployment (for supported platforms)
- [Raspberry Pi Guide](raspberry-pi.md) - Alternative home server platform
- [TrueNAS Scale Guide](truenas-scale.md) - Similar NAS platform
- [DarkPipe Setup Tool](../setup/) - Interactive configuration wizard
- [Unraid Docker Docs](https://docs.unraid.net/unraid-os/manual/vm-management/)
