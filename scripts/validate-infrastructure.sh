#!/usr/bin/env bash
# =============================================================================
# DarkPipe Infrastructure Validation Script
# =============================================================================
#
# Orchestrates five validation sections against a live (or simulated) DarkPipe
# deployment: DNS records, TLS certificates, WireGuard tunnel health, port
# reachability, and service stability.
#
# Usage:
#   validate-infrastructure.sh [OPTIONS]
#
# Options:
#   --json      Machine-readable JSON output (default: human-readable table)
#   --verbose   Emit timestamped diagnostic lines to stderr
#   --dry-run   Return mock pass results without contacting live infrastructure
#   --help      Print this help text and exit
#
# Environment Variables:
#   Name              Default          Description
#   ──────────────    ──────────────   ──────────────────────────────────────────
#   RELAY_DOMAIN      example.com      Primary mail domain to validate
#   HOME_DEVICE_IP    10.8.0.2         Home device IP on the WireGuard tunnel
#   RELAY_IP          (auto-detect)    Cloud relay public IPv4 address
#
# Prerequisites:
#   - bash 3.2+ (macOS default is fine)
#   - dig (DNS lookups)
#   - openssl (TLS certificate checks)
#   - nc or bash /dev/tcp (port connectivity)
#   - jq (optional, for pretty-printing JSON output)
#
# Exit Codes:
#   0   All checks passed (or --dry-run)
#   1   One or more checks failed
#   2   Script error or invalid arguments
#
# Sections:
#   dns        MX, A, SPF, DKIM, DMARC, SRV, and CNAME record checks
#   tls        Certificate chain, expiry, and domain match for HTTPS endpoints
#   tunnel     WireGuard tunnel connectivity via deploy/test/tunnel-test.sh
#   ports      TCP reachability for SMTP (25), HTTPS (443), submission (587),
#              and IMAP (993) on relay and tunnel targets
#   stability  Service recovery timing (requires root; skips otherwise)
#
# Examples:
#   # Quick dry-run to verify script logic
#   ./scripts/validate-infrastructure.sh --dry-run
#
#   # Full live check with JSON output piped to jq
#   RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --json | jq .
#
#   # Verbose human-readable output
#   RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --verbose
#
#   # Check a specific section's failures from JSON
#   ./scripts/validate-infrastructure.sh --json | jq '.sections.dns.checks[] | select(.status=="fail")'
#
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="${SCRIPT_DIR}/lib"

# --- Configuration (from env vars with defaults) ---
RELAY_DOMAIN="${RELAY_DOMAIN:-example.com}"
HOME_DEVICE_IP="${HOME_DEVICE_IP:-10.8.0.2}"
RELAY_IP="${RELAY_IP:-}"

# --- Globals ---
OUTPUT_JSON=false
VERBOSE=false
DRY_RUN=false
SECTIONS="dns tls tunnel ports stability"

# Temp dir for section results (portable alternative to associative arrays)
RESULTS_DIR=""

# --- Source section implementations ---
# Source only scripts that define a run_*_validation() function (not legacy stubs).
for _section_lib in "${LIB_DIR}"/validate-*.sh; do
  if [[ -f "$_section_lib" ]] && grep -q 'run_.*_validation()' "$_section_lib"; then
    source "$_section_lib"
  fi
done
unset _section_lib

cleanup() {
  if [[ -n "$RESULTS_DIR" && -d "$RESULTS_DIR" ]]; then
    rm -rf "$RESULTS_DIR"
  fi
}
trap cleanup EXIT

# --- Argument Parsing ---
usage() {
  cat <<'EOF'
Usage: validate-infrastructure.sh [OPTIONS]

Validate DarkPipe infrastructure: DNS, TLS, tunnel, ports, and stability.

Options:
  --json      Machine-readable JSON output (default: human-readable table)
  --verbose   Emit timestamped diagnostic lines to stderr
  --dry-run   Return mock pass results without contacting live infrastructure
  --help      Print this help text and exit

Environment Variables:
  Name              Default          Description
  ────────────────  ──────────────   ──────────────────────────────────────────
  RELAY_DOMAIN      example.com      Primary mail domain to validate
  HOME_DEVICE_IP    10.8.0.2         Home device IP on the WireGuard tunnel
  RELAY_IP          (auto-detect)    Cloud relay public IPv4 address

Sections:
  dns        MX, A, SPF, DKIM, DMARC, SRV, and CNAME record checks
  tls        Certificate chain, expiry, and domain match for HTTPS endpoints
  tunnel     WireGuard tunnel connectivity via deploy/test/tunnel-test.sh
  ports      TCP reachability for SMTP (25), HTTPS (443), submission (587),
             and IMAP (993) on relay and tunnel targets
  stability  Service recovery timing (requires root; skips otherwise)

Prerequisites:
  bash 3.2+, dig, openssl, nc or bash /dev/tcp, jq (optional)

Exit Codes:
  0   All checks passed (or --dry-run)
  1   One or more checks failed
  2   Script error or invalid arguments

Examples:
  # Dry-run to verify script logic
  ./scripts/validate-infrastructure.sh --dry-run

  # Full live check with JSON output
  RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --json | jq .

  # Verbose human-readable output
  RELAY_DOMAIN=darkpipe.email ./scripts/validate-infrastructure.sh --verbose
EOF
  exit 0
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --json)    OUTPUT_JSON=true; shift ;;
      --verbose) VERBOSE=true; shift ;;
      --dry-run) DRY_RUN=true; shift ;;
      --help)    usage ;;
      *)
        echo "Error: Unknown option '$1'" >&2
        echo "Run '$(basename "$0") --help' for usage." >&2
        exit 2
        ;;
    esac
  done
}

# --- Logging ---
log() {
  if [[ "$VERBOSE" == true ]]; then
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" >&2
  fi
}

log_info() {
  if [[ "$OUTPUT_JSON" == false ]]; then
    echo "$*" >&2
  fi
}

# --- JSON Helpers ---
json_check() {
  local name="$1" status="$2" detail="${3:-}" fix="${4:-}"
  printf '{"name":"%s","status":"%s","detail":"%s","suggested_fix":"%s"}' \
    "$name" "$status" "$detail" "$fix"
}

json_section() {
  local status="$1" timestamp="$2"
  shift 2
  local checks_json="$*"
  printf '{"status":"%s","checks":[%s],"timestamp":"%s"}' \
    "$status" "$checks_json" "$timestamp"
}

# --- Section Result Storage (file-based, bash 3 compatible) ---
set_section_result() {
  local section="$1" result="$2"
  echo "$result" > "${RESULTS_DIR}/${section}.json"
}

get_section_result() {
  local section="$1"
  cat "${RESULTS_DIR}/${section}.json" 2>/dev/null || echo ""
}

# --- Section Runner ---
run_section() {
  local section="$1"
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  local fn_name="run_${section}_validation"

  # If a sourced function exists for this section, use it (handles both dry-run and live)
  if type "$fn_name" &>/dev/null; then
    log "Running section '${section}' via ${fn_name}()"
    local section_output
    local section_exit=0
    section_output="$($fn_name 2>/dev/null)" || section_exit=$?

    local status="pass"
    [[ $section_exit -ne 0 ]] && status="fail"
    set_section_result "$section" "$(json_section "$status" "$timestamp" "$section_output")"
    return 0
  fi

  # Fallback: generic dry-run or subprocess execution
  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: Generating mock results for section '${section}'"
    local mock_check
    mock_check="$(json_check "${section}-mock" "pass" "dry-run mock check" "")"
    set_section_result "$section" "$(json_section "pass" "$timestamp" "$mock_check")"
    return 0
  fi

  local section_script="${LIB_DIR}/validate-${section}.sh"
  if [[ ! -x "$section_script" ]]; then
    log "Section script not found or not executable: ${section_script}"
    local skip_check
    skip_check="$(json_check "${section}-missing" "skip" "section script not found: ${section_script}" "create ${section_script}")"
    set_section_result "$section" "$(json_section "skip" "$timestamp" "$skip_check")"
    return 0
  fi

  log "Running section '${section}' from ${section_script}"
  local section_output
  local section_exit=0
  section_output="$("$section_script" 2>&1)" || section_exit=$?

  if [[ $section_exit -eq 0 ]]; then
    set_section_result "$section" "$(json_section "pass" "$timestamp" "$section_output")"
  else
    local fail_check
    fail_check="$(json_check "${section}-error" "fail" "section exited with code ${section_exit}: ${section_output}" "check ${section_script} logs")"
    set_section_result "$section" "$(json_section "fail" "$timestamp" "$fail_check")"
  fi
}

# --- Compute Overall Status ---
compute_overall_status() {
  local section
  for section in $SECTIONS; do
    local result
    result="$(get_section_result "$section")"
    if echo "$result" | grep -q '"status":"fail"'; then
      echo "fail"
      return
    fi
  done
  echo "pass"
}

# --- Output ---
emit_json() {
  local overall_status="$1"
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  local sections_json=""
  local section
  for section in $SECTIONS; do
    if [[ -n "$sections_json" ]]; then
      sections_json+=","
    fi
    sections_json+="\"${section}\":$(get_section_result "$section")"
  done

  printf '{"overall_status":"%s","timestamp":"%s","config":{"relay_domain":"%s","home_device_ip":"%s","dry_run":%s},"sections":{%s}}\n' \
    "$overall_status" "$timestamp" "$RELAY_DOMAIN" "$HOME_DEVICE_IP" "$DRY_RUN" "$sections_json"
}

emit_human() {
  local overall_status="$1"
  echo ""
  echo "=== DarkPipe Infrastructure Validation ==="
  echo ""

  local section
  for section in $SECTIONS; do
    local result
    result="$(get_section_result "$section")"
    local status
    status="$(echo "$result" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)"
    local icon="✓"
    [[ "$status" == "fail" ]] && icon="✗"
    [[ "$status" == "skip" ]] && icon="○"
    printf "  %s %-12s %s\n" "$icon" "$section" "$status"
  done

  echo ""
  local icon="✓"
  [[ "$overall_status" == "fail" ]] && icon="✗"
  echo "  ${icon} Overall: ${overall_status}"
  echo ""
}

# --- Main ---
main() {
  parse_args "$@"

  RESULTS_DIR="$(mktemp -d)"

  log_info "DarkPipe Infrastructure Validation"
  [[ "$DRY_RUN" == true ]] && log_info "  (dry-run mode — no live checks)"
  log "Config: RELAY_DOMAIN=${RELAY_DOMAIN} HOME_DEVICE_IP=${HOME_DEVICE_IP}"

  local section
  for section in $SECTIONS; do
    run_section "$section"
  done

  local overall_status
  overall_status="$(compute_overall_status)"

  if [[ "$OUTPUT_JSON" == true ]]; then
    emit_json "$overall_status"
  else
    emit_human "$overall_status"
  fi

  if [[ "$overall_status" == "fail" ]]; then
    exit 1
  fi
  exit 0
}

main "$@"
