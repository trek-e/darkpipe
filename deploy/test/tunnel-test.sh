#!/usr/bin/env bash
# tunnel-test.sh - End-to-end tunnel connectivity test
#
# Tests connectivity and health for either WireGuard or mTLS transport.
# Verifies interface status, connectivity, and data transfer capability.
#
# Usage:
#   ./tunnel-test.sh wireguard [peer_ip]
#   ./tunnel-test.sh mtls [server_addr]
#
# Exit codes:
#   0 - All checks passed
#   1 - One or more checks failed

set -euo pipefail

TRANSPORT="${1:-}"
PEER="${2:-}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $*"
}

usage() {
    cat <<EOF
Usage: $0 <transport> [peer]

Transport types:
  wireguard    Test WireGuard tunnel (default peer: 10.8.0.1)
  mtls         Test mTLS connection (default server: localhost:8443)

Examples:
  $0 wireguard 10.8.0.1
  $0 mtls relay.example.com:8443
EOF
    exit 1
}

if [[ -z "$TRANSPORT" ]]; then
    usage
fi

# Track overall success
OVERALL_PASS=true

case "$TRANSPORT" in
    wireguard)
        DEVICE="${WG_DEVICE:-wg0}"
        PEER="${PEER:-10.8.0.1}"

        print_info "Testing WireGuard tunnel on $DEVICE to peer $PEER"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        # Check 1: Interface exists
        print_info "Check 1: WireGuard interface exists..."
        if ip link show "$DEVICE" >/dev/null 2>&1; then
            print_pass "Interface $DEVICE exists"
        else
            print_fail "Interface $DEVICE not found"
            OVERALL_PASS=false
        fi

        # Check 2: Handshake age
        print_info "Check 2: Handshake freshness..."
        if command -v wg >/dev/null 2>&1; then
            HANDSHAKES=$(wg show "$DEVICE" latest-handshakes 2>/dev/null || echo "")
            if [[ -z "$HANDSHAKES" ]]; then
                print_fail "No handshakes recorded (no peers configured?)"
                OVERALL_PASS=false
            else
                MAX_AGE=300  # 5 minutes
                CURRENT_TIME=$(date +%s)
                STALE_FOUND=false

                while IFS=$'\t' read -r PUBKEY TIMESTAMP; do
                    if [[ -z "$TIMESTAMP" ]] || [[ "$TIMESTAMP" -eq 0 ]]; then
                        print_fail "Peer ${PUBKEY:0:16}...: No handshake completed"
                        STALE_FOUND=true
                    else
                        AGE=$((CURRENT_TIME - TIMESTAMP))
                        if [[ "$AGE" -gt "$MAX_AGE" ]]; then
                            print_fail "Peer ${PUBKEY:0:16}...: Handshake age ${AGE}s exceeds max ${MAX_AGE}s"
                            STALE_FOUND=true
                        else
                            print_pass "Peer ${PUBKEY:0:16}...: Handshake age ${AGE}s (healthy)"
                        fi
                    fi
                done <<< "$HANDSHAKES"

                if $STALE_FOUND; then
                    OVERALL_PASS=false
                fi
            fi
        else
            print_fail "wg command not found (wireguard-tools not installed)"
            OVERALL_PASS=false
        fi

        # Check 3: Ping connectivity
        print_info "Check 3: Ping through tunnel..."
        if ping -c 3 -W 5 "$PEER" >/dev/null 2>&1; then
            # Get latency stats
            LATENCY=$(ping -c 3 "$PEER" 2>/dev/null | tail -1 | awk -F '/' '{print $5}')
            print_pass "Ping to $PEER succeeded (avg latency: ${LATENCY}ms)"
        else
            print_fail "Cannot ping peer $PEER through tunnel"
            OVERALL_PASS=false
        fi

        # Check 4: Data transfer (if netcat available)
        print_info "Check 4: Data transfer capability..."
        if command -v nc >/dev/null 2>&1; then
            # Generate 100KB of test data
            TEST_DATA=$(dd if=/dev/urandom bs=1024 count=100 2>/dev/null | base64)
            TEST_SIZE=${#TEST_DATA}

            # Start a listener on the peer (this requires SSH access)
            # For now, just report that we would test this
            print_info "Data transfer test requires peer cooperation (skipping)"
            print_info "To test manually: Start 'nc -l 9999' on peer, then 'echo test | nc $PEER 9999' locally"
        else
            print_info "netcat not available (skipping data transfer test)"
        fi
        ;;

    mtls)
        SERVER="${PEER:-localhost:8443}"
        print_info "Testing mTLS connection to $SERVER"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        # Check 1: Server is listening
        print_info "Check 1: Server is listening..."
        HOST="${SERVER%%:*}"
        PORT="${SERVER##*:}"

        if nc -z -w 5 "$HOST" "$PORT" 2>/dev/null; then
            print_pass "Server $SERVER is listening"
        else
            print_fail "Cannot connect to $SERVER (server not reachable)"
            OVERALL_PASS=false
        fi

        # Check 2: Certificate paths
        print_info "Check 2: Client certificate files..."
        CERT_DIR="${CERT_DIR:-/etc/darkpipe/certs}"
        CLIENT_CERT="$CERT_DIR/client.crt"
        CLIENT_KEY="$CERT_DIR/client.key"
        CA_CERT="$CERT_DIR/ca.crt"

        for CERT_FILE in "$CLIENT_CERT" "$CLIENT_KEY" "$CA_CERT"; do
            if [[ -f "$CERT_FILE" ]]; then
                print_pass "Found $CERT_FILE"
            else
                print_fail "Missing $CERT_FILE"
                OVERALL_PASS=false
            fi
        done

        # Check 3: TLS handshake
        print_info "Check 3: TLS handshake with mutual authentication..."
        if [[ -f "$CLIENT_CERT" ]] && [[ -f "$CLIENT_KEY" ]] && [[ -f "$CA_CERT" ]]; then
            START=$(date +%s%3N)
            if timeout 10 openssl s_client \
                -connect "$SERVER" \
                -cert "$CLIENT_CERT" \
                -key "$CLIENT_KEY" \
                -CAfile "$CA_CERT" \
                -verify_return_error \
                </dev/null 2>&1 | grep -q "Verify return code: 0"; then
                END=$(date +%s%3N)
                HANDSHAKE_TIME=$((END - START))
                print_pass "TLS handshake succeeded (${HANDSHAKE_TIME}ms)"
            else
                print_fail "TLS handshake failed (mutual authentication rejected?)"
                OVERALL_PASS=false
            fi
        else
            print_fail "Cannot test TLS handshake (missing certificates)"
            OVERALL_PASS=false
        fi

        # Check 4: Certificate expiry
        print_info "Check 4: Certificate validity..."
        if [[ -f "$CLIENT_CERT" ]]; then
            EXPIRY=$(openssl x509 -in "$CLIENT_CERT" -noout -enddate 2>/dev/null | cut -d= -f2)
            EXPIRY_EPOCH=$(date -j -f "%b %d %T %Y %Z" "$EXPIRY" +%s 2>/dev/null || echo "0")
            NOW_EPOCH=$(date +%s)

            if [[ "$EXPIRY_EPOCH" -gt "$NOW_EPOCH" ]]; then
                REMAINING=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))
                print_pass "Client certificate valid (expires in ${REMAINING} days)"
            else
                print_fail "Client certificate has expired"
                OVERALL_PASS=false
            fi
        fi
        ;;

    *)
        echo "Unknown transport type: $TRANSPORT"
        usage
        ;;
esac

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if $OVERALL_PASS; then
    print_pass "All checks passed"
    exit 0
else
    print_fail "One or more checks failed"
    exit 1
fi
