# DarkPipe on Proxmox LXC

Deploy DarkPipe in a Proxmox LXC container for lightweight virtualization.

## Prerequisites

**Hardware:**
- Proxmox VE 8.x installation
- 2GB+ RAM allocated to LXC container
- 20GB+ disk space for container

**Software:**
- Proxmox VE 8.0+
- LXC container support enabled

**Why LXC over VM?**
- **Lighter**: LXC uses ~10% of VM overhead (no hypervisor, shared kernel)
- **Faster**: Near-native performance for mail operations
- **Easier**: Direct filesystem access, simpler backups

## Quick Start (3 commands)

```bash
# 1. Create Debian LXC container via Proxmox UI (see detailed steps)

# 2. SSH into container and install Docker
curl -fsSL https://get.docker.com | sh

# 3. Download setup tool and deploy
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64
./darkpipe-setup-linux-amd64 setup
docker compose up -d
```

## Detailed Steps

### Step 1: Create LXC Container

**Via Proxmox Web UI:**

1. Navigate to **Datacenter** → **Node** → **Create CT**

**General:**
- **CT ID**: (auto-assigned or choose)
- **Hostname**: `darkpipe-mail`
- **Unprivileged container**: ✓ (recommended for security)
- **Nesting**: ✓ (required for Docker in unprivileged container)
- **Password**: Set root password

**Template:**
- **Storage**: Select storage for container
- **Template**: Debian 12 (Bookworm) or Ubuntu 24.04 LTS

**Root Disk:**
- **Disk size**: 20GB minimum (50GB recommended for mail storage)
- **Storage**: Choose SSD/NVMe if available

**CPU:**
- **Cores**: 2+ (4 recommended for spam filtering)

**Memory:**
- **Memory**: 2048 MB minimum (4096 MB recommended)
- **Swap**: 1024 MB

**Network:**
- **Bridge**: vmbr0 (or your preferred bridge)
- **IPv4**: Static or DHCP
- **IPv6**: (optional)

**DNS:**
- **Use host settings**: ✓

2. Click **Finish** (do not start yet)

**Step 2: Configure Container for Docker**

Edit the container configuration to enable Docker support:

**Via Proxmox Shell:**
```bash
# Replace 100 with your container ID
nano /etc/pve/lxc/100.conf
```

Add these lines:
```
# Enable nesting for Docker
lxc.apparmor.profile: unconfined
lxc.cgroup2.devices.allow: a
lxc.cap.drop:
lxc.mount.auto: proc:rw sys:rw cgroup:rw

# For WireGuard support (required for DarkPipe)
lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file
```

Save and exit.

**Step 3: Start Container**

In Proxmox UI:
1. Select container
2. Click **Start**
3. Click **Console** to access shell

**Step 4: Install Docker Inside Container**

```bash
# Update system
apt update && apt upgrade -y

# Install Docker via convenience script
curl -fsSL https://get.docker.com | sh

# Verify Docker
docker version
docker compose version
```

**Step 5: Download and Run Setup Tool**

```bash
# Create directory for DarkPipe
mkdir -p /opt/darkpipe
cd /opt/darkpipe

# Download setup tool
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64

# Run setup wizard
./darkpipe-setup-linux-amd64 setup
```

Setup will:
- Validate DNS records
- Test SMTP port 25 connectivity
- Generate `docker-compose.yml`
- Create Docker secrets

**Step 6: Start Services**

```bash
docker compose up -d
```

**Step 7: Verify Services**

```bash
# Check status
docker compose ps

# View logs
docker compose logs -f

# Test mail server
docker exec -it darkpipe-mailserver nc -z localhost 25
```

## Platform-Specific Notes

### Privileged vs Unprivileged Containers

**Unprivileged (recommended):**
- Better security (no root on host)
- Requires `nesting=1` for Docker
- Requires `/dev/net/tun` mount for WireGuard

**Privileged:**
- Simpler Docker setup (no nesting required)
- Less secure (root in container = root on host)
- Use only if unprivileged fails

To create privileged container:
- Uncheck "Unprivileged container" in Create CT wizard
- No need for apparmor/cgroup configuration

### WireGuard in LXC

DarkPipe uses WireGuard for cloud relay <-> home device communication.

**Unprivileged container:**
- Requires `/dev/net/tun` mounted from host
- Add to `/etc/pve/lxc/100.conf`:
  ```
  lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file
  ```

**Privileged container:**
- `/dev/net/tun` available automatically
- No additional configuration needed

**Verify WireGuard support:**
```bash
ls -l /dev/net/tun
# Should show: crw-rw-rw- 1 root root 10, 200 ...
```

### Storage: Directory vs ZFS

**Directory storage (ext4/xfs):**
- Simpler, works everywhere
- No snapshot support

**ZFS storage:**
- Snapshots for instant backups
- Compression to save space
- Recommended for Proxmox

To use ZFS:
1. Create ZFS pool in Proxmox
2. Select ZFS pool when creating container
3. Enable compression: `zfs set compression=lz4 pool/subvol-100-disk-0`

### Backups

**Proxmox backup:**
```bash
# Manual backup
vzdump 100 --mode snapshot --compress zstd --storage local

# Scheduled backup
# Navigate to Datacenter → Backup → Add
```

**Backup includes:**
- Container filesystem
- Docker volumes
- Configuration files

**Restore:**
1. Navigate to storage with backup
2. Select backup file
3. Click **Restore**

### Networking: Bridge vs NAT

**Bridge (default):**
- Container gets IP on host network
- Direct access from LAN
- Simpler port forwarding

**NAT:**
- Container on isolated network
- Port forwarding via iptables
- More isolation

DarkPipe works with both. Bridge is recommended for ease of access.

### Resource Limits

Set resource limits in Proxmox UI:

**CPU:**
- Navigate to container → **Resources** → **CPU**
- Set **CPU limit**: 2-4 cores

**Memory:**
- Navigate to container → **Resources** → **Memory**
- Set **Memory**: 2048-4096 MB
- Set **Swap**: 512-1024 MB

Limits prevent DarkPipe from consuming all host resources.

## Troubleshooting

### Docker fails to start inside container

**Symptom**: `docker: command not found` or `docker daemon not running`

**Cause**: Docker not installed or container missing nesting support

**Fix**:
1. Verify nesting enabled:
   ```bash
   cat /etc/pve/lxc/100.conf | grep nesting
   # Should show: features: nesting=1
   ```
2. Restart container
3. Reinstall Docker if needed

### WireGuard tunnel fails to establish

**Symptom**: Relay can't reach home device, logs show "connection refused"

**Cause**: `/dev/net/tun` not available in container

**Fix**:
1. Add `/dev/net/tun` mount to container config (see Step 2)
2. Restart container
3. Verify: `ls -l /dev/net/tun`

### Out of memory errors

**Symptom**: Containers restart, logs show OOM kills

**Cause**: Insufficient memory allocated to LXC

**Fix**:
1. Navigate to container → **Resources** → **Memory**
2. Increase **Memory** to 4096 MB
3. Increase **Swap** to 2048 MB
4. Or: Reduce Docker memory limits in docker-compose.yml

### Can't access webmail from LAN

**Symptom**: Can ping container, but can't access port 8080

**Cause**: Firewall blocking ports

**Fix**:
```bash
# In container
iptables -L
# Check for DROP rules on port 8080

# Allow port 8080
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

Or disable firewall for testing:
```bash
iptables -F
```

## Performance Benchmarks

**Proxmox LXC on typical hardware (4-core CPU, 4GB RAM, SSD):**
- Mail delivery: <50ms
- IMAP FETCH: 5-15ms per message
- Webmail page load: 100-200ms
- RAM usage (Stalwart + SnappyMail): ~1.2GB
- Idle CPU: <2%

**Proxmox LXC vs VM:**
- LXC: 95-99% of bare-metal performance
- VM: 85-90% of bare-metal performance
- LXC advantage: Lower RAM overhead (~100MB vs ~500MB)

## Security Considerations

1. **Unprivileged containers**: Prefer unprivileged for security (root in container ≠ root on host)
2. **Firewall**: Use Proxmox firewall to restrict access to container
3. **Backups**: Enable scheduled backups with encryption
4. **Updates**: Keep Proxmox and container OS updated

## Next Steps

1. **Configure DNS**: Set up DKIM, SPF, DMARC (see Phase 4 guide)
2. **Test mail flow**: Send test email from Gmail to your@domain.com
3. **Set up backups**: Configure Proxmox scheduled backups
4. **Monitor logs**: `docker compose logs -f rspamd` to watch spam filtering

## Using Podman

Podman is a viable alternative to Docker inside an LXC container. Because Podman is daemonless, it can be simpler to set up in unprivileged LXC containers where running a persistent Docker daemon adds complexity.

### Installing Podman in LXC

```bash
# Debian 12 / Ubuntu 24.04 LXC container
apt update && apt install -y podman

# Verify
podman --version
podman compose version   # requires podman 4.7+ with compose provider
```

### Why Podman in LXC?

- **No daemon**: Docker requires `dockerd` running inside the container, which needs cgroup delegation and nesting. Podman runs without a daemon, reducing the surface area of LXC configuration.
- **Rootless option**: If your LXC container runs unprivileged, Podman's native rootless mode is a natural fit. The cloud relay still needs rootful mode for port 25.
- **Same Compose files**: Use DarkPipe's `docker-compose.yml` directly with `podman compose`.

### Override Files

Use the Podman-specific compose override for volume mount flags and health-check adjustments:

```bash
podman compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

### Runtime Validation

```bash
bash scripts/check-runtime.sh
```

For full Podman deployment details — including SELinux, networking, and troubleshooting — see the [Podman Platform Guide](podman.md).

## See Also

- [Podman Platform Guide](podman.md) - Full Podman deployment reference
- [Raspberry Pi Guide](raspberry-pi.md) - Alternative home server platform
- [TrueNAS Scale Guide](truenas-scale.md) - NAS platform
- [DarkPipe Setup Tool](../setup/) - Interactive configuration wizard
- [Proxmox LXC Documentation](https://pve.proxmox.com/wiki/Linux_Container)
