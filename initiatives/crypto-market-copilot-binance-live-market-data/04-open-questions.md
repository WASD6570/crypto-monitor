# Binance Live Market Data Open Questions

## Questions That Do Not Block Initiative Creation

1. Should the first live `services/market-state-api` cutover remain strictly Spot-driven for regime and composite inputs, with USD-M context exposed only as auxiliary context until a later slice proves the weighting rules?
2. What is the default `openInterest` polling cadence per environment, and should that cadence share the same stale/degraded vocabulary as WS-backed sensors or use a distinct freshness interpretation?
3. For the current small symbol set, is one combined Spot connection and one combined USD-M connection preferable, or is explicit separation by stream family worth the extra operational surface?
4. Do we want proactive reconnect at a fixed safety window before the 24h mark, or should refinement choose a randomized reconnect window to avoid synchronized churn in multi-instance environments?
5. Which live Binance outputs must be visible immediately in the current dashboard and API responses, and which can land first as raw/replay-safe backend surfaces before UI exposure?

## Questions That Should Be Answered During Refinement

- whether any source-record ID rules need to incorporate stream-family prefixes beyond the current canonical event IDs
- whether mark/index and funding data should share one normalized event or remain split by canonical event family
- whether depth snapshot refresh should be timer-driven only or also triggered by observed drift or prolonged book inactivity
- whether local/dev/prod configs should keep one shared backoff profile or diverge by environment once live traffic is observed
