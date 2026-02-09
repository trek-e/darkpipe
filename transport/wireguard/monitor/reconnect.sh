#!/usr/bin/env bash
# reconnect.sh - WireGuard tunnel reconnection safety net
#
# Checks the latest handshake age for all peers on the WireGuard interface.
# If any handshake is older than MAX_AGE_SECONDS (default 300 = 5 minutes),
# the interface is restarted via systemctl.
#
# Intended to be called by a systemd timer or cron as a fallback alongside
# the Go health monitor. This ensures recovery even if the Go process itself
# is down.
#
# Usage:
#   ./reconnect.sh              # Uses defaults (wg0, 300s)
#   ./reconnect.sh wg1 600      # Custom interface and max age
#
# Requires: root privileges (for systemctl restart), wireguard-tools (wg)

set -euo pipefail

DEVICE="${1:-wg0}"
MAX_AGE_SECONDS="${2:-300}"

TAG="darkpipe-reconnect"

log_info() {
    logger -t "$TAG" -p daemon.info "$*"
    echo "[INFO] $*"
}

log_warn() {
    logger -t "$TAG" -p daemon.warning "$*"
    echo "[WARN] $*"
}

log_error() {
    logger -t "$TAG" -p daemon.err "$*"
    echo "[ERROR] $*" >&2
}

# Check prerequisites.
if ! command -v wg >/dev/null 2>&1; then
    log_error "wireguard-tools not installed (wg command not found)"
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    log_error "Must run as root (need systemctl restart and wg show access)"
    exit 1
fi

# Verify the interface exists.
if ! ip link show "$DEVICE" >/dev/null 2>&1; then
    log_error "WireGuard interface $DEVICE does not exist"
    exit 1
fi

# Read handshake timestamps for all peers.
HANDSHAKES=$(wg show "$DEVICE" latest-handshakes)

if [[ -z "$HANDSHAKES" ]]; then
    log_warn "No peers configured on $DEVICE"
    exit 0
fi

CURRENT_TIME=$(date +%s)
NEEDS_RESTART=false
STALE_PEERS=""

while IFS=$'\t' read -r PUBKEY TIMESTAMP; do
    if [[ -z "$TIMESTAMP" ]] || [[ "$TIMESTAMP" -eq 0 ]]; then
        log_warn "Peer ${PUBKEY:0:16}...: no handshake recorded"
        NEEDS_RESTART=true
        STALE_PEERS="${STALE_PEERS} ${PUBKEY:0:16}"
        continue
    fi

    AGE=$((CURRENT_TIME - TIMESTAMP))

    if [[ "$AGE" -gt "$MAX_AGE_SECONDS" ]]; then
        log_warn "Peer ${PUBKEY:0:16}...: handshake age ${AGE}s exceeds max ${MAX_AGE_SECONDS}s"
        NEEDS_RESTART=true
        STALE_PEERS="${STALE_PEERS} ${PUBKEY:0:16}"
    else
        log_info "Peer ${PUBKEY:0:16}...: handshake age ${AGE}s (healthy)"
    fi
done <<< "$HANDSHAKES"

if $NEEDS_RESTART; then
    log_warn "Stale peers detected:${STALE_PEERS}"
    log_info "Restarting wg-quick@${DEVICE}..."

    if systemctl restart "wg-quick@${DEVICE}"; then
        log_info "Successfully restarted wg-quick@${DEVICE}"

        # Wait briefly for handshake to re-establish.
        sleep 10

        # Verify recovery.
        NEW_HANDSHAKES=$(wg show "$DEVICE" latest-handshakes 2>/dev/null || true)
        if [[ -n "$NEW_HANDSHAKES" ]]; then
            log_info "Post-restart handshake state:"
            while IFS=$'\t' read -r PUBKEY TIMESTAMP; do
                if [[ -n "$TIMESTAMP" ]] && [[ "$TIMESTAMP" -ne 0 ]]; then
                    NEW_AGE=$(($(date +%s) - TIMESTAMP))
                    log_info "  Peer ${PUBKEY:0:16}...: handshake age ${NEW_AGE}s"
                else
                    log_warn "  Peer ${PUBKEY:0:16}...: still no handshake"
                fi
            done <<< "$NEW_HANDSHAKES"
        fi
    else
        log_error "Failed to restart wg-quick@${DEVICE}"
        exit 1
    fi
else
    log_info "All peers on $DEVICE are healthy"
fi

exit 0
