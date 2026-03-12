#!/usr/bin/env bash
# Port reachability validation section for DarkPipe infrastructure validation.
# Validates: TCP connectivity to required mail and web ports.
#   - Ports 25 (SMTP) and 443 (HTTPS) on cloud relay (RELAY_IP)
#   - Ports 587 (submission) and 993 (IMAPS) through tunnel (HOME_DEVICE_IP)
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: RELAY_DOMAIN, RELAY_IP, HOME_DEVICE_IP, DRY_RUN, VERBOSE (globals from parent script)
set -euo pipefail

# --- JSON helpers (local to this section) ---
_port_check() {
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
_port_dry_run_checks() {
  local relay="${RELAY_IP:-203.0.113.1}"
  local home="${HOME_DEVICE_IP:-10.8.0.2}"
  local checks=""
  checks+="$(_port_check "port_25_smtp" "pass" "dry-run: ${relay}:25 reachable (SMTP)" "")"
  checks+=","
  checks+="$(_port_check "port_443_https" "pass" "dry-run: ${relay}:443 reachable (HTTPS)" "")"
  checks+=","
  checks+="$(_port_check "port_587_submission" "pass" "dry-run: ${home}:587 reachable via tunnel (submission)" "")"
  checks+=","
  checks+="$(_port_check "port_993_imaps" "pass" "dry-run: ${home}:993 reachable via tunnel (IMAPS)" "")"
  echo "$checks"
}

# --- TCP connect test ---
# Uses nc -z (preferred) with fallback to bash /dev/tcp
# Args: host port timeout_secs
# Returns: 0 if reachable, 1 if not
# Sets: _TCP_CONNECT_DETAIL with timing or error info
_tcp_connect() {
  local host="$1" port="$2" timeout_secs="${3:-5}"
  _TCP_CONNECT_DETAIL=""

  local start_time end_time elapsed_ms

  # Try nc -z first (most portable for this purpose)
  if command -v nc >/dev/null 2>&1; then
    start_time="$(date +%s)"
    if nc -z -w "$timeout_secs" "$host" "$port" 2>/dev/null; then
      end_time="$(date +%s)"
      elapsed_ms=$(( (end_time - start_time) * 1000 ))
      _TCP_CONNECT_DETAIL="connected via nc in ~${elapsed_ms}ms"
      return 0
    else
      _TCP_CONNECT_DETAIL="nc -z -w${timeout_secs} ${host} ${port} failed"
      return 1
    fi
  fi

  # Fallback: bash /dev/tcp (works in bash 3+ but not all environments)
  start_time="$(date +%s)"
  if (echo >/dev/tcp/"$host"/"$port") 2>/dev/null; then
    end_time="$(date +%s)"
    elapsed_ms=$(( (end_time - start_time) * 1000 ))
    _TCP_CONNECT_DETAIL="connected via /dev/tcp in ~${elapsed_ms}ms"
    return 0
  else
    _TCP_CONNECT_DETAIL="/dev/tcp connect to ${host}:${port} timed out or refused"
    return 1
  fi
}

# --- Main entry point ---
run_ports_validation() {
  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _port_dry_run_checks
    return 0
  fi

  local relay_ip="${RELAY_IP:-}"
  local home_ip="${HOME_DEVICE_IP:-10.8.0.2}"
  local domain="${RELAY_DOMAIN:-example.com}"

  # Auto-resolve relay IP if not set
  if [[ -z "$relay_ip" ]]; then
    relay_ip="$(dig +short "${domain}" A @8.8.8.8 2>/dev/null | head -1)"
  fi
  if [[ -z "$relay_ip" ]]; then
    local skip_msg="Cannot determine relay IP — set RELAY_IP or ensure ${domain} A record exists"
    local checks=""
    checks+="$(_port_check "port_25_smtp" "skip" "$skip_msg" "Set RELAY_IP environment variable")"
    checks+=","
    checks+="$(_port_check "port_443_https" "skip" "$skip_msg" "Set RELAY_IP environment variable")"
    checks+=","
    checks+="$(_port_check "port_587_submission" "skip" "skipped: relay IP unknown" "")"
    checks+=","
    checks+="$(_port_check "port_993_imaps" "skip" "skipped: relay IP unknown" "")"
    echo "$checks"
    return 1
  fi

  local checks=""
  local any_fail=false

  # Port definitions: name host port service_description fix_hint
  local -a port_defs=(
    "port_25_smtp|${relay_ip}|25|SMTP on relay|Check cloud relay firewall allows inbound TCP 25"
    "port_443_https|${relay_ip}|443|HTTPS on relay|Check Caddy is running and cloud firewall allows inbound TCP 443"
    "port_587_submission|${home_ip}|587|submission via tunnel|Check Postfix is listening on 587 and WireGuard tunnel is up"
    "port_993_imaps|${home_ip}|993|IMAPS via tunnel|Check Dovecot is listening on 993 and WireGuard tunnel is up"
  )

  for def in "${port_defs[@]}"; do
    IFS='|' read -r name host port desc fix_hint <<< "$def"
    [[ -n "$checks" ]] && checks+=","

    if _tcp_connect "$host" "$port" 5; then
      checks+="$(_port_check "$name" "pass" \
        "${host}:${port} reachable (${desc}) — ${_TCP_CONNECT_DETAIL}" "")"
    else
      checks+="$(_port_check "$name" "fail" \
        "${host}:${port} unreachable (${desc}) — ${_TCP_CONNECT_DETAIL}" \
        "$fix_hint")"
      any_fail=true
    fi
  done

  echo "$checks"

  if [[ "$any_fail" == "true" ]]; then
    return 1
  fi
  return 0
}
