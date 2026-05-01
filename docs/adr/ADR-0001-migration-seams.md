# ADR-0001: Deepen migration seams (source, destination, sync, flow)

- Status: Accepted
- Date: 2026-05-01

## Context
Migration logic was concentrated in wizard implementation. Callers needed to know provider auth details, destination parsing defaults, sync implementation details, and phase sequencing. This reduced locality and made test surfaces broad and brittle.

## Decision
Adopt four migration modules with explicit seams:
- Source module (`migrationsource`) for authentication, adapter opening, capabilities, metadata.
- Destination module (`migrationdest`) for immutable config validation, endpoint parsing, TLS policy, adapter connect.
- Sync module (`imapsync`) as adapter seam over IMAP migration implementation.
- Flow module (`wizard/flow`) for phase orchestration (auth/connect/preview/apply/report).

## Consequences
- Higher locality: migration policy and behavior are concentrated behind seams.
- Higher leverage: wizard caller complexity reduced to orchestration entrypoint.
- Better tests: interface-level tests for source/destination/sync/flow can evolve independently from implementations.

## Rejected alternatives
1. Keep wizard-centric implementation.
   - Rejected: shallow module shape; complexity spread across caller logic.
2. Build one monolithic migration module.
   - Rejected: lower locality for independent concerns (auth, destination config, sync, flow).
3. Force all providers to expose identical auth internals to wizard.
   - Rejected: seam leakage and weaker adapter isolation.
