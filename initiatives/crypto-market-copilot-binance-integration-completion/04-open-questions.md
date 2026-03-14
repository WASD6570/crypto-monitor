# Binance Integration Completion Open Questions

## Questions That Do Not Block Initiative Creation

1. Should USD-M context influence the effective regime directly, or first land as a bounded degrade/cap signal before any positive weighting is allowed?
2. Is an additive runtime-status endpoint preferable to expanding existing current-state payloads for operator visibility?
3. Should the final streaming runtime keep one combined Spot connection for the tracked symbols, or split stream families for clearer failure isolation?
4. What long-run validation window is enough to call the integration operationally finished for `local` and `dev`?

## Questions That Should Be Answered During Refinement

- whether the streaming runtime should persist a dedicated in-memory read model inside `cmd/market-state-api` or consume a narrower reusable service boundary from `services/venue-binance`
- whether `/healthz` should remain process-only while richer runtime status lives elsewhere
- whether `dev` and `prod` should keep the same symbol list and backoff profile as `local` for the first rollout
- whether USD-M semantic changes require additive API provenance or bucket metadata to stay operator-honest
