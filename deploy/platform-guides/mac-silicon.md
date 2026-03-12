# DarkPipe on Mac Silicon (Apple M-series)

Deploy DarkPipe on macOS with Apple Silicon (M1/M2/M3/M4) for development and testing.

## Prerequisites

**Hardware:**
- Mac with Apple Silicon (M1, M2, M3, M4, or later)
- 8GB+ RAM (16GB recommended for full stack)
- 20GB+ free disk space

**Software:**
- macOS 14+ (Sonoma or later)
- Docker Desktop for Mac **or** OrbStack
- Xcode Command Line Tools (for git, etc.)

> **Apple Containers:** macOS 26+ introduces Apple Containers as a native lightweight container runtime. For setup instructions and DarkPipe compatibility, see the [Apple Containers Guide](apple-containers.md) (coming soon).

**Purpose: Development/Testing Only**

DarkPipe on macOS is intended for:
- **Development**: Testing code changes before deployment
- **Testing**: Validating mail flow and configuration
- **Learning**: Understanding DarkPipe architecture

**Not for production** because:
- macOS blocks inbound port 25 by default (can't receive mail)
- macOS is typically behind NAT/firewall (no public IP)
- Power management may suspend services

For production, deploy cloud relay on a VPS and home device on Linux/Raspberry Pi/NAS.

## Quick Start (3 commands)

```bash
# 1. Install Docker Desktop or OrbStack (see detailed steps)

# 2. Download and run setup tool
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-arm64
chmod +x darkpipe-setup-darwin-arm64
./darkpipe-setup-darwin-arm64 setup

# 3. Start services
docker compose up -d
```

## Detailed Steps

### Step 1: Install Docker Runtime

**Option A: Docker Desktop (Official)**

1. Download [Docker Desktop for Mac](https://www.docker.com/products/docker-desktop/)
2. Choose **Apple Silicon** version
3. Open `.dmg` and drag to Applications
4. Launch Docker Desktop
5. Complete setup wizard
6. Verify installation:
   ```bash
   docker version
   docker compose version
   ```

**Option B: OrbStack (Lightweight Alternative)**

1. Download [OrbStack](https://orbstack.dev/)
2. Install and launch
3. OrbStack automatically provides `docker` and `docker compose` commands
4. Faster startup and lower resource usage than Docker Desktop

**Comparison:**

| Feature | Docker Desktop | OrbStack |
|---------|----------------|----------|
| **Performance** | Good | Excellent |
| **Startup time** | 15-30 seconds | <5 seconds |
| **RAM usage** | ~1-2GB | ~500MB |
| **License** | Free (personal use) | Free + Pro ($8/mo) |
| **Compatibility** | 100% Docker API | 100% Docker API |

Both work perfectly with DarkPipe. OrbStack is recommended for daily development.

### Step 2: Install Xcode Command Line Tools

```bash
# Check if already installed
xcode-select -p

# If not installed:
xcode-select --install
```

Required for `git`, `curl`, and other command-line tools.

### Step 3: Download Setup Tool

```bash
# Create directory for DarkPipe
mkdir -p ~/Development/darkpipe
cd ~/Development/darkpipe

# Download ARM64 macOS setup tool
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-arm64
chmod +x darkpipe-setup-darwin-arm64

# Run setup wizard
./darkpipe-setup-darwin-arm64 setup
```

**Setup Mode Selection:**

For development/testing, choose:
- **Quick mode**: Use defaults (Stalwart + SnappyMail)
- **Advanced mode**: Customize components

**Port 25 Note:**

Setup will warn about port 25 being blocked on macOS. Options:
1. **Test alternate port**: Use port 2525 for SMTP testing
2. **Deploy cloud relay on VPS**: Run cloud-relay on a VPS with port 25 open
3. **Skip cloud relay**: Test home device only (direct IMAP/webmail access)

### Step 4: Start Services

```bash
# Start all services in background
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f

# Access webmail
open http://localhost:8080
```

### Step 5: Test Mail Flow (Without Port 25)

**Option A: Test via SMTP submission (port 587)**

```bash
# Send test email via submission port
swaks --to test@example.com \
      --from you@example.com \
      --server localhost \
      --port 587 \
      --auth-user admin@example.com \
      --auth-password yourpassword \
      --tls
```

**Option B: Test via webmail**

1. Open http://localhost:8080
2. Log in with admin credentials
3. Compose and send test email
4. Check delivery in logs: `docker compose logs -f stalwart`

**Option C: Deploy cloud relay on VPS**

1. Provision a VPS with port 25 open
2. Run `darkpipe-setup` on VPS to generate cloud-relay config
3. Point home device to VPS relay via WireGuard tunnel
4. Full mail flow testing with real inbound mail

## Platform-Specific Notes

### Port 25 Blocked on macOS

macOS blocks inbound connections to port 25 by default (anti-spam measure).

**Cannot be unblocked** without kernel extensions or system modifications.

**Workaround for testing:**
```yaml
# Use alternate port in docker-compose.yml
services:
  stalwart:
    ports:
      - "2525:25"  # Map 2525 on host to 25 in container
      - "587:587"
      - "993:993"
```

Then send test emails to `localhost:2525`.

**Production setup:**
- Cloud relay on VPS (Linux with port 25)
- Home device on Raspberry Pi/NAS (receives via WireGuard)
- Mac for development only

### Docker Volumes on macOS

Docker volumes on macOS use bind mounts with VirtioFS (Docker Desktop) or virtiofs (OrbStack).

**Performance notes:**
- Slower than Linux native volumes (~70% throughput)
- Acceptable for development/testing
- Not suitable for production mail server

**Optimization:**
```yaml
volumes:
  mail-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ~/Development/darkpipe/mail-data
```

Use local directories for faster I/O during development.

### Apple Silicon Native Images

All DarkPipe images are built for `linux/arm64`, which runs natively on Apple Silicon (no Rosetta translation).

**Verify:**
```bash
docker inspect ghcr.io/trek-e/darkpipe/cloud-relay:latest | grep Architecture
# Should output: "Architecture": "arm64"
```

**Performance:**
- Native ARM64 execution (no x86 emulation)
- Same performance as Linux on ARM

### Memory Limits

Docker Desktop has configurable memory limits:

1. Open **Docker Desktop** → **Settings** → **Resources**
2. Adjust **Memory**: 4-8GB recommended for DarkPipe
3. Adjust **CPU**: 2-4 cores recommended
4. Click **Apply & Restart**

OrbStack automatically manages resources (no manual configuration needed).

### WireGuard on macOS

WireGuard kernel module is not available on macOS. DarkPipe uses WireGuard userspace implementation (wireguard-go).

**Installation** (if testing WireGuard separately from Docker):
```bash
brew install wireguard-tools wireguard-go
```

**Docker WireGuard:**
Works via `--cap-add=NET_ADMIN --device=/dev/net/tun` in docker-compose.yml (already configured).

### Firewall Configuration

macOS Firewall may block Docker port forwarding.

**Check firewall:**
```bash
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate
```

**Allow Docker:**
1. **System Settings** → **Network** → **Firewall** → **Options**
2. **Add** Docker Desktop or OrbStack
3. **Allow incoming connections**

Or disable firewall for testing (not recommended).

## Troubleshooting

### Docker daemon not running

**Symptom**: `Cannot connect to the Docker daemon`

**Cause**: Docker Desktop or OrbStack not started

**Fix**:
1. Launch Docker Desktop or OrbStack
2. Wait for Docker icon in menu bar to show "running"
3. Retry `docker` commands

### Containers fail to start with "port already allocated"

**Symptom**: `Bind for 0.0.0.0:25 failed: port is already allocated`

**Cause**: macOS blocking port 25, or another service using the port

**Fix**:
1. Change port mapping in docker-compose.yml:
   ```yaml
   ports:
     - "2525:25"  # Use port 2525 on host
   ```
2. Or: Find conflicting service: `sudo lsof -i :25`

### Services show "Unhealthy" status

**Symptom**: `docker compose ps` shows unhealthy containers

**Cause**: Services still starting, or health check failing

**Fix**:
1. Wait 2-3 minutes for services to initialize
2. Check logs: `docker compose logs [service-name]`
3. Verify DNS is configured (if using real domain)

### Slow performance

**Symptom**: Mail operations take seconds, high CPU usage

**Cause**: Docker filesystem overhead on macOS

**Fix**:
1. Use OrbStack instead of Docker Desktop (faster)
2. Increase Docker Desktop memory allocation
3. Use local bind mounts instead of named volumes
4. Disable unnecessary services (webmail, calendar)

### Can't access webmail on localhost:8080

**Symptom**: Browser shows "connection refused" on http://localhost:8080

**Cause**: Webmail container not running, or port not mapped

**Fix**:
1. Check container status: `docker compose ps`
2. Verify port mapping: `docker compose port snappymail 8888`
3. Check firewall: **System Settings** → **Network** → **Firewall**
4. Try alternate port: http://127.0.0.1:8080

## Development Workflow

### Live Code Editing

Mount source code into containers for live development:

```yaml
# docker-compose.override.yml
services:
  stalwart:
    volumes:
      - ./home-device/stalwart/config.toml:/opt/stalwart-mail/etc/config.toml:ro
```

Changes to `config.toml` apply immediately (after container restart).

### Debugging

**Attach to running container:**
```bash
docker exec -it darkpipe-mailserver /bin/sh
```

**View real-time logs:**
```bash
docker compose logs -f stalwart
docker compose logs -f rspamd
```

**Inspect service:**
```bash
docker inspect darkpipe-mailserver
```

### Testing Configuration Changes

```bash
# Make changes to docker-compose.yml or config files

# Rebuild and restart
docker compose down
docker compose up -d

# Or: Restart single service
docker compose restart stalwart
```

## Performance Benchmarks

**Mac Studio M2 Max (12 cores, 32GB RAM, Docker Desktop):**
- Mail delivery: <100ms
- IMAP FETCH: 10-20ms per message
- Webmail page load: 150-300ms
- RAM usage (Stalwart + SnappyMail): ~1.2GB
- Idle CPU: <3%

**MacBook Air M1 (8 cores, 8GB RAM, OrbStack):**
- Mail delivery: 100-200ms
- IMAP FETCH: 20-40ms per message
- Webmail page load: 300-600ms
- RAM usage (Maddy + no webmail): ~600MB
- Idle CPU: <5%

## Next Steps

1. **Configure DNS**: Set up test domain with DNS records
2. **Deploy cloud relay**: Provision VPS for cloud-relay component
3. **Test mail flow**: Send/receive test emails
4. **Develop features**: Contribute to DarkPipe development

## See Also

- [Raspberry Pi Guide](raspberry-pi.md) - Production home server deployment
- [TrueNAS Scale Guide](truenas-scale.md) - NAS platform deployment
- [DarkPipe Setup Tool](../setup/) - Interactive configuration wizard
- [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)
- [OrbStack](https://orbstack.dev/)
