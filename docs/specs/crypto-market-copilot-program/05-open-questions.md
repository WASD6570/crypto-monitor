# Open Questions

These do not block the two-initiative split, but later feature plans should answer them explicitly.

## Infrastructure Choices

- What storage engines back raw events, derived features, alerts, outcomes, simulations, and decision logs?
- Which views require low-latency query paths versus batch or cold-storage access?

## Slow Context Providers

- Which exact CME and ETF flow sources are acceptable for MVP use?
- What cadence, licensing, and retry constraints do those sources impose?

## Threshold Locking

- What exact starting thresholds define fragmentation persistence, volatility shock, depth degradation, feed staleness, and derivatives stress?
- Which thresholds should be shared across BTC and ETH versus symbol-specific?

## Delivery Preferences

- Telegram is the default push channel, but should a webhook be enabled in the same first release or only after the UI and Telegram path are stable?

## Simulation Confidence Policy

- When L2 is degraded, should simulation always fallback to a spread proxy or refuse once confidence drops below a minimum threshold?

## Future Execution Policy

- If assisted execution is ever introduced later, what human approval model should clear `DE-RISK` and `STOP` restrictions?
