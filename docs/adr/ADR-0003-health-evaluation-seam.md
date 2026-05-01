# ADR-0003: Deepen system health evaluation seam

- Status: Accepted
- Date: 2026-05-01

## Context
System status aggregation mixed collection and policy evaluation. Threshold/rule behavior lived in aggregator implementation, reducing locality and making policy changes risky.

## Decision
Create dedicated health evaluation module (`monitoring/status/healtheval`) with:
- Input: normalized snapshot.
- Output: overall status + reasons + triggered rules.
- Policy: constructor config object.
- Rule model: evaluate all rules; severity determines final status.

Aggregator now builds snapshot and delegates evaluation through seam.

## Consequences
- Higher locality: policy logic isolated from data collection.
- Higher leverage: dashboards/alerts/scripts consume consistent evaluation output.
- Better tests: rule behavior tested directly at evaluation interface.

## Rejected alternatives
1. Keep `computeOverallStatus` in aggregator.
   - Rejected: shallow module, mixed responsibilities.
2. First-match rule evaluation.
   - Rejected: loses explanatory depth and triggered-rule visibility.
3. String-only output.
   - Rejected: insufficient machine-readable detail for automation.
