#!/usr/bin/env bash
# Tunnel validation section for DarkPipe infrastructure validation.
# Invokes deploy/test/tunnel-test.sh and parses pass/fail results.
# Supports WireGuard and mTLS transport types.
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: DRY_RUN, VERBOSE (globals from parent script)
# Optional: TRANSPORT_TYPE (default: wireguard), HOME_DEVICE_IP
set -euo pipefail

SCRIPT_DIR_TUNNEL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT_TUNNEL="${SCRIPT_DIR_TUNNEL}/../.."
TUNNEL_TEST_SCRIPT="${PROJECT_ROOT_TUNNEL}/deploy/test/tunnel-test.sh"

# --- JSON helpers (local to this section) ---
_tunnel_check() {
  local name="$1" status="$2" detail="${3:-}" fix="${4:-}"
  detail="${detail//\\/\\\\}"
  detail="${detail//\"/\\\"}"
  detail="${detail//$'\n'/\\n}"
  fix="${fix//\\/\\\\}"
  fix="${fix//\"/\\\"}"
  fix="${fix//$'\n'/\\n}"
  printf '{"name":"%s","status":"%s","detail":"%s","suggested_fix":"%s"}' \
    "$name" "$status" "$detail" "$fix"
}

# --- Dry-run mock results ---
_tunnel_dry_run_checks() {
  local transport="${TRANSPORT_TYPE:-wireguard}"
  local checks=""
  checks+="$(_tunnel_check "tunnel_script_exists" "pass" "dry-run: tunnel-test.sh found" "")"
  checks+=","
  checks+="$(_tunnel_check "tunnel_${transport}_health" "pass" "dry-run: ${transport} tunnel healthy, handshake age 45s" "")"
  checks+=","
  checks+="$(_tunnel_check "tunnel_${transport}_connectivity" "pass" "dry-run: ${transport} peer reachable, latency 12ms" "")"
  echo "$checks"
}

# --- Live tunnel check ---
run_tunnel_validation() {
  local transport="${TRANSPORT_TYPE:-wireguard}"

  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _tunnel_dry_run_checks
    return 0
  fi

  local checks=""
  local any_fail=false

  # Check 1: tunnel-test.sh script exists and is executable
  if [[ ! -f "$TUNNEL_TEST_SCRIPT" ]]; then
    checks+="$(_tunnel_check "tunnel_script_exists" "fail" \
      "tunnel-test.sh not found at ${TUNNEL_TEST_SCRIPT}" \
      "Ensure deploy/test/tunnel-test.sh exists")"
    echo "$checks"
    return 1
  fi

  if [[ ! -x "$TUNNEL_TEST_SCRIPT" ]]; then
    chmod +x "$TUNNEL_TEST_SCRIPT" 2>/dev/null || true
  fi

  checks+="$(_tunnel_check "tunnel_script_exists" "pass" "tunnel-test.sh found at ${TUNNEL_TEST_SCRIPT}" "")"

  # Check 2: Run tunnel-test.sh with a warm-up timeout
  # Allow 30s for fresh tunnels to establish handshake
  local tunnel_output=""
  local tunnel_exit=0
  tunnel_output="$(timeout 60 bash "$TUNNEL_TEST_SCRIPT" "$transport" 2>&1)" || tunnel_exit=$?

  # Parse [PASS] / [FAIL] lines from tunnel-test.sh output
  local pass_count=0 fail_count=0
  pass_count="$(echo "$tunnel_output" | grep -c '\[PASS\]' || true)"
  fail_count="$(echo "$tunnel_output" | grep -c '\[FAIL\]' || true)"

  # Extract handshake age if present
  local handshake_detail=""
  handshake_detail="$(echo "$tunnel_output" | grep -i 'handshake.*age\|Handshake age' | head -1 || true)"
  if [[ -z "$handshake_detail" ]]; then
    handshake_detail="$(echo "$tunnel_output" | grep '\[PASS\].*[Hh]andshake' | head -1 || true)"
  fi

  # Extract connectivity/latency info
  local connectivity_detail=""
  connectivity_detail="$(echo "$tunnel_output" | grep -i 'ping\|latency\|succeeded' | head -1 || true)"
  if [[ -z "$connectivity_detail" ]]; then
    connectivity_detail="$(echo "$tunnel_output" | grep '\[PASS\].*[Pp]ing' | head -1 || true)"
  fi

  # Build health check result
  if [[ $tunnel_exit -eq 0 ]] && [[ $fail_count -eq 0 ]]; then
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_health" "pass" \
      "${transport} tunnel healthy: ${pass_count} checks passed. ${handshake_detail}" "")"
  elif [[ $tunnel_exit -eq 124 ]]; then
    # timeout
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_health" "fail" \
      "${transport} tunnel test timed out after 60s" \
      "Check WireGuard interface status and peer configuration")"
    any_fail=true
  else
    # Extract first failure line for detail
    local first_fail
    first_fail="$(echo "$tunnel_output" | grep '\[FAIL\]' | head -1 || echo "exit code ${tunnel_exit}")"
    first_fail="${first_fail//\"/\\\"}"
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_health" "fail" \
      "${transport} tunnel unhealthy: ${fail_count} checks failed. ${first_fail}" \
      "Review tunnel-test.sh output; check WireGuard config and peer connectivity")"
    any_fail=true
  fi

  # Build connectivity check result
  if [[ -n "$connectivity_detail" ]] && ! echo "$connectivity_detail" | grep -q '\[FAIL\]'; then
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_connectivity" "pass" \
      "${connectivity_detail}" "")"
  elif [[ $tunnel_exit -eq 0 ]]; then
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_connectivity" "pass" \
      "${transport} connectivity verified (${pass_count} checks passed)" "")"
  else
    checks+=","
    checks+="$(_tunnel_check "tunnel_${transport}_connectivity" "fail" \
      "${transport} connectivity failed" \
      "Verify tunnel peer IP is reachable and firewall rules allow traffic")"
    any_fail=true
  fi

  echo "$checks"

  if [[ "$any_fail" == "true" ]]; then
    return 1
  fi
  return 0
}
