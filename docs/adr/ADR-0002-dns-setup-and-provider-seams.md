# ADR-0002: Deepen DNS setup and provider seams

- Status: Accepted
- Date: 2026-05-01

## Context
DNS setup behavior was concentrated in CLI implementation. Provider behavior varied by adapter capabilities. Apply reporting lacked per-record machine-readable outcomes.

## Decision
Adopt deep DNS seams:
- DNS setup module (`dns/setup`) owns Plan, Apply, Validate, RotateDKIM, SendAuthTest.
- Provider adapter seam (`dns/setup/provider_adapter.go`) is capability-aware.
- Unsupported record types return typed outcomes.
- Apply returns aggregate + per-record results with reason code/message/retryable.
- CLI remains adapter for human/json output; json always carries full apply details.

## Consequences
- Higher locality: DNS policy and provider handling are concentrated in setup/provider modules.
- Higher leverage: CLI and automation reuse same interface behavior.
- Better tests: provider adapter contract tests validate action/reason/retryability semantics.

## Rejected alternatives
1. Keep apply logic in CLI.
   - Rejected: shallow seam; poor reuse and weak test surface.
2. Force strict full-CRUD provider support now.
   - Rejected: does not match current adapter reality; blocks incremental depth.
3. Silent skip on unsupported record types.
   - Rejected: hides failure modes; weak operational visibility.
