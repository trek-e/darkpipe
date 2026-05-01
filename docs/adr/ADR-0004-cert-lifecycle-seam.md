# ADR-0004: Deepen certificate lifecycle seam

- Status: Accepted
- Date: 2026-05-01

## Context
Certificate lifecycle behavior (check, renew, reload) was spread across separate modules with direct command execution coupling in tests.

## Decision
Introduce lifecycle manager seam (`monitoring/cert/lifecycle.go`) with explicit adapters:
- CertInspector
- Renewer
- Reloader

Lifecycle orchestrates: check -> renew (if needed) -> reload.
Existing retry renew behavior is adapted via `RetryRenewer`.

## Consequences
- Higher locality: lifecycle state transitions in one module.
- Higher leverage: deterministic lifecycle tests with fakes; no external tool dependency for core orchestration tests.
- Better failure handling: explicit step-level error modes.

## Rejected alternatives
1. Keep orchestration spread across callers.
   - Rejected: low locality; repeated flow logic.
2. Test lifecycle via real certbot/step in unit tests.
   - Rejected: flaky/non-deterministic; poor test surface.
3. Merge renewal command implementation into lifecycle module.
   - Rejected: weak seam; reduces adapter swappability.
