#!/usr/bin/env bash
set -euo pipefail

# WireGuard Cloud (Hub) Setup Script
# Installs WireGuard, applies configuration, enables systemd service with auto-restart
# Requires: Debian/Ubuntu-based system, root access

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check for root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

# Check for config file argument
if [[ $# -ne 1 ]]; then
    log_error "Usage: $0 <path-to-wg0.conf>"
    log_error "Example: $0 /tmp/cloud-wg0.conf"
    exit 1
fi

CONFIG_FILE="$1"

# Verify config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    log_error "Config file not found: $CONFIG_FILE"
    exit 1
fi

log_info "Starting WireGuard cloud (hub) setup"

# Install WireGuard packages
log_info "Installing WireGuard packages..."
apt-get update -qq
apt-get install -y wireguard wireguard-tools

# Verify kernel module can load
log_info "Verifying WireGuard kernel module..."
if ! modprobe wireguard; then
    log_error "Failed to load WireGuard kernel module"
    log_error "Your kernel may not support WireGuard. Consider upgrading kernel or using wireguard-dkms."
    exit 1
fi

# Copy config to system location with secure permissions
log_info "Installing WireGuard configuration..."
cp "$CONFIG_FILE" /etc/wireguard/wg0.conf
chmod 0600 /etc/wireguard/wg0.conf
chown root:root /etc/wireguard/wg0.conf
log_info "Config installed at /etc/wireguard/wg0.conf with mode 0600"

# Create systemd override directory and install override
log_info "Installing systemd override for auto-restart..."
mkdir -p /etc/systemd/system/wg-quick@wg0.service.d
cat > /etc/systemd/system/wg-quick@wg0.service.d/override.conf <<'OVERRIDE_EOF'
[Service]
Restart=on-failure
RestartSec=30
OVERRIDE_EOF
log_info "Systemd override installed"

# Reload systemd to pick up override
systemctl daemon-reload

# Enable IP forwarding (required for routing traffic)
log_info "Enabling IP forwarding..."
sysctl -w net.ipv4.ip_forward=1 > /dev/null
echo "net.ipv4.ip_forward=1" > /etc/sysctl.d/99-wireguard.conf
log_info "IP forwarding enabled and persisted"

# Configure firewall if ufw is active
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    log_info "UFW is active, opening UDP port 51820..."
    ufw allow 51820/udp
    log_info "Firewall rule added"
else
    log_warn "UFW not active or not installed. Ensure UDP port 51820 is open in your firewall."
fi

# Enable and start WireGuard service
log_info "Enabling and starting wg-quick@wg0 service..."
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

# Wait a moment for interface to come up
sleep 2

# Verify service status
if systemctl is-active --quiet wg-quick@wg0; then
    log_info "WireGuard service is running"
else
    log_error "WireGuard service failed to start"
    systemctl status wg-quick@wg0 --no-pager
    exit 1
fi

# Show interface status
log_info "WireGuard interface status:"
echo "=================================================="
wg show wg0
echo "=================================================="

log_info "Cloud WireGuard setup complete!"
log_info "Interface wg0 is up and listening on port 51820"
log_info "Auto-restart is enabled with 30s delay on failure"
