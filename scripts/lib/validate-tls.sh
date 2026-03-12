#!/usr/bin/env bash
# TLS validation section for DarkPipe infrastructure validation.
# Validates: certificate chain, expiry, and domain match for HTTPS endpoints.
# Checks webmail, autoconfig, and autodiscover subdomains via openssl s_client.
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: RELAY_DOMAIN, RELAY_IP, DRY_RUN, VERBOSE (globals from parent script)
set -euo pipefail

# --- JSON helpers (local to this section) ---
_tls_check() {
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
_tls_dry_run_checks() {
  local domain="${RELAY_DOMAIN:-example.com}"
  local checks=""
  local subdomains=("mail" "autoconfig" "autodiscover")
  local labels=("webmail" "autoconfig" "autodiscover")

  for i in "${!subdomains[@]}"; do
    local fqdn="${subdomains[$i]}.${domain}"
    [[ -n "$checks" ]] && checks+=","
    checks+="$(_tls_check "${labels[$i]}_cert_chain" "pass" "dry-run: ${fqdn} certificate chain valid" "")"
    checks+=","
    checks+="$(_tls_check "${labels[$i]}_cert_expiry" "pass" "dry-run: ${fqdn} certificate expires in 89 days" "")"
    checks+=","
    checks+="$(_tls_check "${labels[$i]}_cert_match" "pass" "dry-run: ${fqdn} certificate subject matches domain" "")"
  done
  echo "$checks"
}

# --- Live TLS checks ---

# Validate a single HTTPS domain's TLS certificate.
# Args: label fqdn connect_host connect_port
_check_tls_domain() {
  local label="$1" fqdn="$2" connect_host="$3" connect_port="${4:-443}"
  local checks=""

  # Fetch certificate via openssl s_client
  local cert_output=""
  local cert_exit=0
  cert_output="$(echo | timeout 10 openssl s_client \
    -connect "${connect_host}:${connect_port}" \
    -servername "$fqdn" \
    -verify_return_error \
    2>&1)" || cert_exit=$?

  # Check 1: Certificate chain validation
  if echo "$cert_output" | grep -q "Verify return code: 0"; then
    checks+="$(_tls_check "${label}_cert_chain" "pass" "${fqdn}: certificate chain verified" "")"
  elif [[ $cert_exit -ne 0 ]] && ! echo "$cert_output" | grep -q "BEGIN CERTIFICATE"; then
    checks+="$(_tls_check "${label}_cert_chain" "fail" \
      "${fqdn}: TLS connection failed to ${connect_host}:${connect_port}" \
      "Verify Caddy is running and serving TLS for ${fqdn}")"
    # Can't check expiry or match if connection failed
    checks+=","
    checks+="$(_tls_check "${label}_cert_expiry" "skip" "skipped: TLS connection failed" "")"
    checks+=","
    checks+="$(_tls_check "${label}_cert_match" "skip" "skipped: TLS connection failed" "")"
    echo "$checks"
    return 1
  else
    local verify_code
    verify_code="$(echo "$cert_output" | grep "Verify return code:" | head -1 || echo "unknown")"
    checks+="$(_tls_check "${label}_cert_chain" "fail" \
      "${fqdn}: chain verification failed — ${verify_code}" \
      "Check Caddy TLS config and ACME certificate issuance for ${fqdn}")"
  fi

  # Extract the certificate for further checks
  local cert_pem=""
  cert_pem="$(echo "$cert_output" | sed -n '/BEGIN CERTIFICATE/,/END CERTIFICATE/p' | head -30)"

  if [[ -z "$cert_pem" ]]; then
    checks+=","
    checks+="$(_tls_check "${label}_cert_expiry" "skip" "skipped: could not extract certificate PEM" "")"
    checks+=","
    checks+="$(_tls_check "${label}_cert_match" "skip" "skipped: could not extract certificate PEM" "")"
    echo "$checks"
    return 1
  fi

  # Check 2: Certificate expiry (warn at 30 days, fail at 7 days)
  local days_left=""
  if echo "$cert_pem" | openssl x509 -checkend 0 -noout >/dev/null 2>&1; then
    # Certificate is not yet expired — compute days remaining
    local end_date
    end_date="$(echo "$cert_pem" | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)"
    local end_epoch now_epoch
    # macOS date vs GNU date
    if date -j >/dev/null 2>&1; then
      end_epoch="$(date -j -f "%b %e %T %Y %Z" "$end_date" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$end_date" +%s 2>/dev/null || echo 0)"
    else
      end_epoch="$(date -d "$end_date" +%s 2>/dev/null || echo 0)"
    fi
    now_epoch="$(date +%s)"

    if [[ "$end_epoch" -gt 0 ]]; then
      days_left=$(( (end_epoch - now_epoch) / 86400 ))
    fi

    if [[ -n "$days_left" ]] && [[ "$days_left" -lt 7 ]]; then
      checks+=","
      checks+="$(_tls_check "${label}_cert_expiry" "fail" \
        "${fqdn}: certificate expires in ${days_left} days (critical)" \
        "Renew certificate immediately — check Caddy ACME logs")"
    elif [[ -n "$days_left" ]] && [[ "$days_left" -lt 30 ]]; then
      checks+=","
      checks+="$(_tls_check "${label}_cert_expiry" "pass" \
        "${fqdn}: certificate expires in ${days_left} days (renewal soon)" "")"
    else
      checks+=","
      checks+="$(_tls_check "${label}_cert_expiry" "pass" \
        "${fqdn}: certificate expires in ${days_left:-unknown} days" "")"
    fi
  else
    checks+=","
    checks+="$(_tls_check "${label}_cert_expiry" "fail" \
      "${fqdn}: certificate has expired" \
      "Renew certificate — check Caddy ACME configuration and DNS challenge setup")"
  fi

  # Check 3: Domain match — verify certificate covers the requested domain
  local cert_subject cert_san
  cert_subject="$(echo "$cert_pem" | openssl x509 -noout -subject 2>/dev/null || echo "")"
  cert_san="$(echo "$cert_pem" | openssl x509 -noout -ext subjectAltName 2>/dev/null || echo "")"

  local domain_matched=false
  if echo "$cert_san" | grep -qi "DNS:${fqdn}"; then
    domain_matched=true
  elif echo "$cert_san" | grep -qi "DNS:\*.${RELAY_DOMAIN:-example.com}"; then
    domain_matched=true
  elif echo "$cert_subject" | grep -qi "CN.*=.*${fqdn}"; then
    domain_matched=true
  fi

  if [[ "$domain_matched" == "true" ]]; then
    checks+=","
    checks+="$(_tls_check "${label}_cert_match" "pass" \
      "${fqdn}: certificate subject/SAN matches domain" "")"
  else
    checks+=","
    checks+="$(_tls_check "${label}_cert_match" "fail" \
      "${fqdn}: certificate does not cover this domain. Subject: ${cert_subject}" \
      "Ensure Caddy is configured to request a certificate for ${fqdn}")"
  fi

  echo "$checks"

  # Return 1 if any check failed
  if echo "$checks" | grep -q '"status":"fail"'; then
    return 1
  fi
  return 0
}

# --- Main entry point ---
run_tls_validation() {
  local domain="${RELAY_DOMAIN:-example.com}"

  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _tls_dry_run_checks
    return 0
  fi

  # Determine connect host — use RELAY_IP if set, otherwise resolve domain
  local connect_host="${RELAY_IP:-}"
  if [[ -z "$connect_host" ]]; then
    connect_host="$(dig +short "${domain}" A @8.8.8.8 2>/dev/null | head -1)"
  fi
  if [[ -z "$connect_host" ]]; then
    connect_host="$domain"
  fi

  local checks=""
  local any_fail=false

  local subdomains=("mail" "autoconfig" "autodiscover")
  local labels=("webmail" "autoconfig" "autodiscover")

  for i in "${!subdomains[@]}"; do
    local fqdn="${subdomains[$i]}.${domain}"
    local result=""
    local exit_code=0
    result="$(_check_tls_domain "${labels[$i]}" "$fqdn" "$connect_host" 443)" || exit_code=$?

    [[ -n "$checks" ]] && checks+=","
    checks+="$result"
    [[ $exit_code -ne 0 ]] && any_fail=true
  done

  echo "$checks"

  if [[ "$any_fail" == "true" ]]; then
    return 1
  fi
  return 0
}
