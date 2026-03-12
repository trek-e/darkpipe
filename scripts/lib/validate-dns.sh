#!/usr/bin/env bash
# DNS validation section for DarkPipe infrastructure validation.
# Validates: MX, A, SPF, DKIM, DMARC, SRV (imaps, submission), autoconfig/autodiscover CNAMEs.
# Uses external resolvers (8.8.8.8, 1.1.1.1) — never local resolver.
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: RELAY_DOMAIN, DRY_RUN, VERBOSE (globals from parent script)
# Optional: RELAY_IP, RELAY_HOSTNAME, DKIM_SELECTOR (for live checks)
set -euo pipefail

SCRIPT_DIR_DNS="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR_DNS}/../.."

# External resolvers — never use local
DNS_RESOLVERS=("8.8.8.8" "1.1.1.1")

# --- JSON helpers (local to this section) ---
_dns_check() {
  local name="$1" status="$2" detail="${3:-}" fix="${4:-}"
  # Escape double quotes and backslashes in detail and fix for valid JSON
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
_dns_dry_run_checks() {
  local checks=""
  local record_types=("mx" "a" "spf" "dkim" "dmarc" "srv_imaps" "srv_submission" "autoconfig_cname" "autodiscover_cname")
  local descriptions=(
    "MX record points to mail.${RELAY_DOMAIN}"
    "A record for ${RELAY_DOMAIN} resolves to ${RELAY_IP:-203.0.113.1}"
    "SPF record contains ip4:${RELAY_IP:-203.0.113.1}"
    "DKIM record found for selector ${DKIM_SELECTOR:-darkpipe202601}"
    "DMARC record found with policy"
    "SRV _imaps._tcp.${RELAY_DOMAIN} -> port 993"
    "SRV _submission._tcp.${RELAY_DOMAIN} -> port 587"
    "autoconfig.${RELAY_DOMAIN} CNAME exists"
    "autodiscover.${RELAY_DOMAIN} CNAME exists"
  )

  for i in "${!record_types[@]}"; do
    [[ -n "$checks" ]] && checks+=","
    checks+="$(_dns_check "${record_types[$i]}" "pass" "dry-run: ${descriptions[$i]}" "")"
  done
  echo "$checks"
}

# --- Live dig queries ---

# Query a DNS record via external resolvers. Returns answer section or empty string.
# Usage: _dig_query <record_type> <name> [resolver]
_dig_query() {
  local rtype="$1" name="$2" resolver="${3:-${DNS_RESOLVERS[0]}}"
  dig +short +timeout=5 +tries=2 "@${resolver}" "$name" "$rtype" 2>/dev/null || echo ""
}

# Query from both resolvers and return results. Detects propagation issues.
# Sets: _DIG_RESULT_PRIMARY, _DIG_RESULT_SECONDARY
_dig_both() {
  local rtype="$1" name="$2"
  _DIG_RESULT_PRIMARY="$(_dig_query "$rtype" "$name" "${DNS_RESOLVERS[0]}")"
  _DIG_RESULT_SECONDARY="$(_dig_query "$rtype" "$name" "${DNS_RESOLVERS[1]}")"
}

# --- Individual record checks (live) ---

_check_mx() {
  local domain="$1"
  _dig_both "MX" "$domain"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "mx" "fail" "No MX record found for ${domain} via ${DNS_RESOLVERS[0]}" \
      "Add MX record: ${domain} -> mail.${domain} (priority 10)"
    return
  fi

  # Check if resolvers agree
  if [[ "$_DIG_RESULT_PRIMARY" != "$_DIG_RESULT_SECONDARY" ]]; then
    _dns_check "mx" "fail" \
      "MX propagation mismatch: ${DNS_RESOLVERS[0]} returned [${_DIG_RESULT_PRIMARY}] vs ${DNS_RESOLVERS[1]} returned [${_DIG_RESULT_SECONDARY}]" \
      "DNS propagation in progress — wait and retry"
    return
  fi

  _dns_check "mx" "pass" "MX: ${result}" ""
}

_check_a() {
  local domain="$1"
  _dig_both "A" "$domain"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "a" "fail" "No A record found for ${domain} via ${DNS_RESOLVERS[0]}" \
      "Add A record: ${domain} -> ${RELAY_IP:-<relay-ip>}"
    return
  fi

  # Check propagation
  if [[ "$_DIG_RESULT_PRIMARY" != "$_DIG_RESULT_SECONDARY" ]]; then
    _dns_check "a" "fail" \
      "A record propagation mismatch: ${DNS_RESOLVERS[0]}=${_DIG_RESULT_PRIMARY} vs ${DNS_RESOLVERS[1]}=${_DIG_RESULT_SECONDARY}" \
      "DNS propagation in progress — wait and retry"
    return
  fi

  # If RELAY_IP is set, verify it matches
  if [[ -n "${RELAY_IP:-}" ]]; then
    if [[ "$result" != *"$RELAY_IP"* ]]; then
      _dns_check "a" "fail" \
        "A record ${result} does not match expected RELAY_IP=${RELAY_IP}" \
        "Update A record to point to ${RELAY_IP}"
      return
    fi
  fi

  _dns_check "a" "pass" "A: ${domain} -> ${result}" ""
}

_check_spf() {
  local domain="$1"
  _dig_both "TXT" "$domain"
  local result="$_DIG_RESULT_PRIMARY"

  # Filter for SPF record
  local spf_record=""
  spf_record="$(echo "$result" | grep -i 'v=spf1' || true)"

  if [[ -z "$spf_record" ]]; then
    _dns_check "spf" "fail" "No SPF record found for ${domain}" \
      "Add TXT record: ${domain} -> v=spf1 ip4:${RELAY_IP:-<relay-ip>} -all"
    return
  fi

  # Check for multiple SPF records (RFC violation)
  local spf_count
  spf_count="$(echo "$spf_record" | wc -l | tr -d ' ')"
  if [[ "$spf_count" -gt 1 ]]; then
    _dns_check "spf" "fail" "Multiple SPF records found (${spf_count}) — RFC 7208 violation" \
      "Remove duplicate SPF records; keep only one v=spf1 TXT record"
    return
  fi

  _dns_check "spf" "pass" "SPF: ${spf_record}" ""
}

_check_dkim() {
  local domain="$1" selector="${DKIM_SELECTOR:-darkpipe202601}"
  local dkim_name="${selector}._domainkey.${domain}"
  _dig_both "TXT" "$dkim_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "dkim" "fail" "No DKIM record at ${dkim_name}" \
      "Add DKIM TXT record for selector ${selector} at ${dkim_name}"
    return
  fi

  # Verify it contains a DKIM key marker
  if ! echo "$result" | grep -qi 'v=DKIM1'; then
    _dns_check "dkim" "fail" "TXT record at ${dkim_name} does not contain v=DKIM1" \
      "Verify DKIM record content starts with v=DKIM1"
    return
  fi

  # Truncate key for display (no secrets in output)
  local truncated
  truncated="$(echo "$result" | head -c 60)..."
  _dns_check "dkim" "pass" "DKIM: ${dkim_name} -> ${truncated}" ""
}

_check_dmarc() {
  local domain="$1"
  local dmarc_name="_dmarc.${domain}"
  _dig_both "TXT" "$dmarc_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "dmarc" "fail" "No DMARC record at ${dmarc_name}" \
      "Add TXT record: ${dmarc_name} -> v=DMARC1; p=none; rua=mailto:dmarc@${domain}"
    return
  fi

  if ! echo "$result" | grep -qi 'v=DMARC1'; then
    _dns_check "dmarc" "fail" "TXT record at ${dmarc_name} does not contain v=DMARC1" \
      "Verify DMARC record starts with v=DMARC1"
    return
  fi

  _dns_check "dmarc" "pass" "DMARC: ${result}" ""
}

_check_srv_imaps() {
  local domain="$1"
  local srv_name="_imaps._tcp.${domain}"
  _dig_both "SRV" "$srv_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "srv_imaps" "fail" "No SRV record at ${srv_name}" \
      "Add SRV record: ${srv_name} -> 0 1 993 mail.${domain}"
    return
  fi

  # SRV format: priority weight port target
  if ! echo "$result" | grep -q '993'; then
    _dns_check "srv_imaps" "fail" "SRV ${srv_name} does not include port 993: ${result}" \
      "Update SRV record to use port 993"
    return
  fi

  _dns_check "srv_imaps" "pass" "SRV: ${srv_name} -> ${result}" ""
}

_check_srv_submission() {
  local domain="$1"
  local srv_name="_submission._tcp.${domain}"
  _dig_both "SRV" "$srv_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "srv_submission" "fail" "No SRV record at ${srv_name}" \
      "Add SRV record: ${srv_name} -> 0 1 587 mail.${domain}"
    return
  fi

  if ! echo "$result" | grep -q '587'; then
    _dns_check "srv_submission" "fail" "SRV ${srv_name} does not include port 587: ${result}" \
      "Update SRV record to use port 587"
    return
  fi

  _dns_check "srv_submission" "pass" "SRV: ${srv_name} -> ${result}" ""
}

_check_autoconfig_cname() {
  local domain="$1"
  local cname_name="autoconfig.${domain}"
  _dig_both "CNAME" "$cname_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "autoconfig_cname" "fail" "No CNAME record at ${cname_name}" \
      "Add CNAME record: ${cname_name} -> mail.${domain}"
    return
  fi

  _dns_check "autoconfig_cname" "pass" "CNAME: ${cname_name} -> ${result}" ""
}

_check_autodiscover_cname() {
  local domain="$1"
  local cname_name="autodiscover.${domain}"
  _dig_both "CNAME" "$cname_name"
  local result="$_DIG_RESULT_PRIMARY"

  if [[ -z "$result" ]]; then
    _dns_check "autodiscover_cname" "fail" "No CNAME record at ${cname_name}" \
      "Add CNAME record: ${cname_name} -> mail.${domain}"
    return
  fi

  _dns_check "autodiscover_cname" "pass" "CNAME: ${cname_name} -> ${result}" ""
}

# --- Main entry point ---
# Outputs: comma-separated JSON check objects (the checks array content)
# Returns: 0 if all pass, 1 if any fail
run_dns_validation() {
  local domain="${RELAY_DOMAIN:-example.com}"

  # Dry-run: return mocks without network calls
  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _dns_dry_run_checks
    return 0
  fi

  # Live mode: query external resolvers
  local checks=""
  local any_fail=false

  # Collect each check
  local check_functions=(
    "_check_mx"
    "_check_a"
    "_check_spf"
    "_check_dkim"
    "_check_dmarc"
    "_check_srv_imaps"
    "_check_srv_submission"
    "_check_autoconfig_cname"
    "_check_autodiscover_cname"
  )

  for fn in "${check_functions[@]}"; do
    local result
    result="$($fn "$domain")"
    [[ -n "$checks" ]] && checks+=","
    checks+="$result"

    # Track failures
    if echo "$result" | grep -q '"status":"fail"'; then
      any_fail=true
    fi
  done

  echo "$checks"

  if [[ "$any_fail" == "true" ]]; then
    return 1
  fi
  return 0
}
