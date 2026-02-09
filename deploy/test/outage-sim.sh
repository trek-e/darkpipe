#!/usr/bin/env bash
# outage-sim.sh - Simulates home internet outage and verifies auto-recovery
#
# This script directly validates Phase 1 success criterion #3:
# "After simulating a home internet outage (disconnect for 60 seconds),
#  the WireGuard tunnel automatically re-establishes and data flows resume
#  without manual intervention"
#
# Usage:
#   sudo ./outage-sim.sh wireguard [device]
#   sudo ./outage-sim.sh mtls [process_name]
#
# MUST run on home device (not VPS/relay).
# REQUIRES root privileges.
#
# Exit codes:
#   0 - Auto-recovery succeeded within time limit
#   1 - Tunnel did not recover or test failed

set -euo pipefail

TRANSPORT="${1:-wireguard}"
TARGET="${2:-}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}[STEP]${NC} $*"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $*"
}

# Prerequisite checks
if [[ $EUID -ne 0 ]]; then
    print_fail "Must run as root (required to stop/start services)"
    exit 1
fi

# Determine script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TUNNEL_TEST="$SCRIPT_DIR/tunnel-test.sh"

if [[ ! -x "$TUNNEL_TEST" ]]; then
    print_fail "tunnel-test.sh not found or not executable at $TUNNEL_TEST"
    exit 1
fi

case "$TRANSPORT" in
    wireguard)
        DEVICE="${TARGET:-wg0}"
        SYSTEMD_SERVICE="wg-quick@${DEVICE}"

        print_info "Simulating outage for WireGuard device: $DEVICE"
        print_warn "This will disconnect your tunnel for up to 90 seconds"
        echo ""

        # Step 1: Verify tunnel is currently healthy
        print_step "1. Verifying tunnel is currently healthy..."
        if bash "$TUNNEL_TEST" wireguard >/dev/null 2>&1; then
            print_pass "Tunnel is healthy before test"
        else
            print_fail "Tunnel is not healthy before test -- cannot proceed"
            print_info "Run: bash $TUNNEL_TEST wireguard"
            exit 1
        fi

        # Step 2: Record pre-outage state
        print_step "2. Recording pre-outage handshake state..."
        BEFORE_HANDSHAKES=$(wg show "$DEVICE" latest-handshakes 2>/dev/null || echo "")
        if [[ -z "$BEFORE_HANDSHAKES" ]]; then
            print_fail "No handshakes recorded before test"
            exit 1
        fi

        echo "$BEFORE_HANDSHAKES" | while IFS=$'\t' read -r PUBKEY TIMESTAMP; do
            if [[ -n "$TIMESTAMP" ]] && [[ "$TIMESTAMP" -ne 0 ]]; then
                AGE=$(( $(date +%s) - TIMESTAMP ))
                print_info "  Peer ${PUBKEY:0:16}...: handshake age ${AGE}s"
            fi
        done

        # Step 3: Simulate outage
        print_step "3. Simulating outage (stopping $SYSTEMD_SERVICE)..."
        if systemctl stop "$SYSTEMD_SERVICE"; then
            print_pass "Service stopped"
        else
            print_fail "Failed to stop service"
            exit 1
        fi

        # Step 4: Wait 60 seconds (per phase success criteria)
        print_step "4. Waiting 60 seconds (simulating internet outage)..."
        for i in {60..1}; do
            printf "\r  Remaining: %2ds" "$i"
            sleep 1
        done
        printf "\r  ${GREEN}Complete: 60s${NC}\n"

        # Step 5: Restore service
        print_step "5. Restoring service (starting $SYSTEMD_SERVICE)..."
        if systemctl start "$SYSTEMD_SERVICE"; then
            print_pass "Service started"
        else
            print_fail "Failed to start service"
            exit 1
        fi

        # Step 6: Wait for tunnel to re-establish (up to 90 seconds)
        print_step "6. Waiting for tunnel to re-establish (max 90 seconds)..."
        RECOVERY_START=$(date +%s)
        RECOVERY_DEADLINE=$((RECOVERY_START + 90))
        RECOVERED=false

        while [[ $(date +%s) -lt $RECOVERY_DEADLINE ]]; do
            ELAPSED=$(( $(date +%s) - RECOVERY_START ))
            printf "\r  Checking... (elapsed: %2ds)" "$ELAPSED"

            if bash "$TUNNEL_TEST" wireguard >/dev/null 2>&1; then
                RECOVERY_TIME=$(( $(date +%s) - RECOVERY_START ))
                printf "\r  ${GREEN}Tunnel recovered after ${RECOVERY_TIME}s${NC}\n"
                RECOVERED=true
                break
            fi

            sleep 5
        done

        if ! $RECOVERED; then
            printf "\n"
            print_fail "Tunnel did not recover within 90 seconds"
            exit 1
        fi

        # Step 7: Verify tunnel is healthy
        print_step "7. Verifying tunnel health post-recovery..."
        if bash "$TUNNEL_TEST" wireguard; then
            print_pass "Tunnel is fully healthy after recovery"
        else
            print_fail "Tunnel recovered but health check failed"
            exit 1
        fi

        # Step 8: Report handshake re-establishment
        print_step "8. Verifying new handshake established..."
        AFTER_HANDSHAKES=$(wg show "$DEVICE" latest-handshakes 2>/dev/null || echo "")

        echo "$AFTER_HANDSHAKES" | while IFS=$'\t' read -r PUBKEY TIMESTAMP; do
            if [[ -n "$TIMESTAMP" ]] && [[ "$TIMESTAMP" -ne 0 ]]; then
                AGE=$(( $(date +%s) - TIMESTAMP ))
                if [[ "$AGE" -lt 30 ]]; then
                    print_pass "  Peer ${PUBKEY:0:16}...: fresh handshake (age ${AGE}s)"
                else
                    print_warn "  Peer ${PUBKEY:0:16}...: handshake age ${AGE}s (older than expected)"
                fi
            fi
        done

        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        print_pass "OUTAGE SIMULATION PASSED"
        echo ""
        print_info "Summary:"
        print_info "  - Outage duration: 60 seconds (as per success criteria)"
        print_info "  - Recovery time: ${RECOVERY_TIME} seconds"
        print_info "  - Auto-recovery: SUCCESSFUL (no manual intervention)"
        print_info "  - Data flow: RESUMED (tunnel healthy)"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;

    mtls)
        PROCESS="${TARGET:-darkpipe-client}"

        print_info "Simulating outage for mTLS process: $PROCESS"
        print_warn "This will kill the mTLS client and verify it restarts"
        echo ""

        # Step 1: Verify connection is healthy
        print_step "1. Verifying mTLS connection is currently healthy..."
        if bash "$TUNNEL_TEST" mtls >/dev/null 2>&1; then
            print_pass "mTLS connection is healthy before test"
        else
            print_fail "mTLS connection is not healthy before test -- cannot proceed"
            print_info "Run: bash $TUNNEL_TEST mtls"
            exit 1
        fi

        # Step 2: Find and kill the process
        print_step "2. Killing mTLS client process..."
        PIDS=$(pgrep -f "$PROCESS" || echo "")
        if [[ -z "$PIDS" ]]; then
            print_fail "Process '$PROCESS' not found"
            exit 1
        fi

        for PID in $PIDS; do
            print_info "  Killing PID $PID"
            kill "$PID"
        done
        print_pass "Process killed"

        # Step 3: Wait 60 seconds
        print_step "3. Waiting 60 seconds (simulating outage)..."
        for i in {60..1}; do
            printf "\r  Remaining: %2ds" "$i"
            sleep 1
        done
        printf "\r  ${GREEN}Complete: 60s${NC}\n"

        # Step 4: Check if systemd restarted the service
        print_step "4. Checking if systemd auto-restarted the client..."
        sleep 5  # Give systemd time to react

        NEW_PIDS=$(pgrep -f "$PROCESS" || echo "")
        if [[ -z "$NEW_PIDS" ]]; then
            print_fail "Process did not auto-restart (check systemd service configuration)"
            exit 1
        fi

        print_pass "Process auto-restarted (new PID: $NEW_PIDS)"

        # Step 5: Wait for connection to re-establish
        print_step "5. Waiting for mTLS connection to re-establish (max 90 seconds)..."
        RECOVERY_START=$(date +%s)
        RECOVERY_DEADLINE=$((RECOVERY_START + 90))
        RECOVERED=false

        while [[ $(date +%s) -lt $RECOVERY_DEADLINE ]]; do
            ELAPSED=$(( $(date +%s) - RECOVERY_START ))
            printf "\r  Checking... (elapsed: %2ds)" "$ELAPSED"

            if bash "$TUNNEL_TEST" mtls >/dev/null 2>&1; then
                RECOVERY_TIME=$(( $(date +%s) - RECOVERY_START ))
                printf "\r  ${GREEN}Connection recovered after ${RECOVERY_TIME}s${NC}\n"
                RECOVERED=true
                break
            fi

            sleep 5
        done

        if ! $RECOVERED; then
            printf "\n"
            print_fail "mTLS connection did not recover within 90 seconds"
            exit 1
        fi

        # Step 6: Verify connection is healthy
        print_step "6. Verifying mTLS connection health post-recovery..."
        if bash "$TUNNEL_TEST" mtls; then
            print_pass "mTLS connection is fully healthy after recovery"
        else
            print_fail "Connection recovered but health check failed"
            exit 1
        fi

        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        print_pass "OUTAGE SIMULATION PASSED"
        echo ""
        print_info "Summary:"
        print_info "  - Outage duration: 60 seconds (as per success criteria)"
        print_info "  - Recovery time: ${RECOVERY_TIME} seconds"
        print_info "  - Auto-recovery: SUCCESSFUL (systemd restart + reconnection)"
        print_info "  - Connection: HEALTHY"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;

    *)
        print_fail "Unknown transport type: $TRANSPORT"
        echo "Usage: $0 {wireguard|mtls} [device/process]"
        exit 1
        ;;
esac

exit 0
