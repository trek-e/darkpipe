#!/usr/bin/env bash
# Stability validation section for DarkPipe infrastructure validation.
# Invokes deploy/test/outage-sim.sh to simulate and recover from outages.
# Requires root — skips gracefully when not root.
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: DRY_RUN, VERBOSE (globals from parent script)
# Optional: TRANSPORT_TYPE (default: wireguard)
set -euo pipefail

SCRIPT_DIR_STABILITY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT_STABILITY="${SCRIPT_DIR_STABILITY}/../.."
OUTAGE_SIM_SCRIPT="${PROJECT_ROOT_STABILITY}/deploy/test/outage-sim.sh"

# --- JSON helpers (local to this section) ---
_stability_check() {
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
_stability_dry_run_checks() {
  local transport="${TRANSPORT_TYPE:-wireguard}"
  local checks=""
  checks+="$(_stability_check "stability_root_check" "pass" "dry-run: running as root" "")"
  checks+=","
  checks+="$(_stability_check "stability_outage_sim" "pass" "dry-run: ${transport} tunnel recovered in 8s after simulated outage" "")"
  checks+=","
  checks+="$(_stability_check "stability_recovery_time" "pass" "dry-run: recovery time 8s within 60s threshold" "")"
  echo "$checks"
}

# --- Main entry point ---
run_stability_validation() {
  local transport="${TRANSPORT_TYPE:-wireguard}"

  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _stability_dry_run_checks
    return 0
  fi

  local checks=""

  # Check 1: Root privilege check
  if [[ "$(id -u)" -ne 0 ]]; then
    checks+="$(_stability_check "stability_root_check" "skip" \
      "Not running as root — outage simulation requires root privileges" \
      "Run with sudo to enable stability testing")"
    checks+=","
    checks+="$(_stability_check "stability_outage_sim" "skip" \
      "Skipped: requires root" \
      "Run: sudo scripts/validate-infrastructure.sh")"
    echo "$checks"
    return 0  # skip is not a failure
  fi

  checks+="$(_stability_check "stability_root_check" "pass" "Running as root (uid 0)" "")"

  # Check 2: outage-sim.sh exists
  if [[ ! -f "$OUTAGE_SIM_SCRIPT" ]]; then
    checks+=","
    checks+="$(_stability_check "stability_outage_sim" "fail" \
      "outage-sim.sh not found at ${OUTAGE_SIM_SCRIPT}" \
      "Ensure deploy/test/outage-sim.sh exists")"
    echo "$checks"
    return 1
  fi

  if [[ ! -x "$OUTAGE_SIM_SCRIPT" ]]; then
    chmod +x "$OUTAGE_SIM_SCRIPT" 2>/dev/null || true
  fi

  # Check 3: Run outage simulation
  local sim_output=""
  local sim_exit=0
  sim_output="$(timeout 180 bash "$OUTAGE_SIM_SCRIPT" "$transport" 2>&1)" || sim_exit=$?

  # Parse results
  local pass_count fail_count
  pass_count="$(echo "$sim_output" | grep -c '\[PASS\]' || true)"
  fail_count="$(echo "$sim_output" | grep -c '\[FAIL\]' || true)"

  # Extract recovery time if present
  local recovery_detail=""
  recovery_detail="$(echo "$sim_output" | grep -i 'recover' | tail -1 || true)"

  if [[ $sim_exit -eq 0 ]] && [[ $fail_count -eq 0 ]]; then
    checks+=","
    checks+="$(_stability_check "stability_outage_sim" "pass" \
      "Outage simulation passed: ${pass_count} checks passed. ${recovery_detail}" "")"

    # Recovery time check
    local recovery_seconds=""
    recovery_seconds="$(echo "$sim_output" | grep -oE '[0-9]+ seconds' | head -1 | grep -oE '[0-9]+' || true)"
    if [[ -n "$recovery_seconds" ]]; then
      checks+=","
      if [[ "$recovery_seconds" -le 60 ]]; then
        checks+="$(_stability_check "stability_recovery_time" "pass" \
          "Recovery time ${recovery_seconds}s within 60s threshold" "")"
      else
        checks+="$(_stability_check "stability_recovery_time" "fail" \
          "Recovery time ${recovery_seconds}s exceeds 60s threshold" \
          "Check WireGuard PersistentKeepalive and reconnection settings")"
      fi
    fi
  elif [[ $sim_exit -eq 124 ]]; then
    checks+=","
    checks+="$(_stability_check "stability_outage_sim" "fail" \
      "Outage simulation timed out after 180s" \
      "Tunnel may not be recovering — check WireGuard persistent keepalive settings")"
  else
    local first_fail
    first_fail="$(echo "$sim_output" | grep '\[FAIL\]' | head -1 || echo "exit code ${sim_exit}")"
    first_fail="${first_fail//\"/\\\"}"
    checks+=","
    checks+="$(_stability_check "stability_outage_sim" "fail" \
      "Outage simulation failed: ${fail_count} checks failed. ${first_fail}" \
      "Review outage-sim.sh output; verify tunnel auto-recovery configuration")"
  fi

  echo "$checks"

  if echo "$checks" | grep -q '"status":"fail"'; then
    return 1
  fi
  return 0
}
