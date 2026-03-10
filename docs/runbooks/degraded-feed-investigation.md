# Degraded Feed Investigation

## Entry Conditions

- Start this runbook whenever a venue feed leaves `HEALTHY` and enters `DEGRADED` or `STALE`.
- Keep the emitted state and reasons intact; do not translate `connection-not-ready`, `message-stale`, `snapshot-stale`, `sequence-gap`, `reconnect-loop`, `resync-loop`, `rate-limit`, or `clock-degraded` into alternate labels.

## Quick Triage

1. Record the current feed state: `HEALTHY`, `DEGRADED`, or `STALE`.
2. Record the emitted degradation reasons exactly as shown.
3. Check `message_lag_ms`, `snapshot_lag_ms`, `reconnect_count`, `resync_count`, `sequence_gap_count`, `snapshot_recovery_attempts_per_minute`, `rest_poll_attempts_per_minute`, and `clock_offset_ms`.
4. Confirm whether the current condition is bounded by retry policy or requires operator intervention.

## Reason-Specific Actions

| Reason | Meaning | Operator Action |
|---|---|---|
| `connection-not-ready` | transport is disconnected, connecting, reconnecting, or resyncing | inspect recent reconnect cadence and confirm the adapter can resubscribe cleanly |
| `message-stale` | the last message exceeded the freshness threshold | verify venue traffic, heartbeat behavior, and whether reconnect is already in progress |
| `snapshot-stale` | the last snapshot exceeded the freshness threshold | confirm snapshot-required stream health and trigger bounded recovery |
| `sequence-gap` | order-book continuity is uncertain | force resync immediately; never infer missing Kraken or Bybit L2 continuity |
| `reconnect-loop` | reconnect attempts crossed the configured threshold | verify upstream availability, host networking, and backoff clamp behavior |
| `resync-loop` | resync attempts crossed the configured threshold | inspect parser integrity, snapshot inputs, and gap frequency |
| `rate-limit` | REST polling exceeded the configured minute budget | confirm the configured cadence, wait for the retry window, and avoid manual burst polling |
| `clock-degraded` | clock skew crossed the degraded threshold | check host NTP/clock source and validate timestamp trust before downstream use |

## Exit Criteria

- Return to `HEALTHY` with no degradation reasons.
- Confirm canonical outputs still preserve `exchangeTs`, `recvTs`, venue, symbol, market type, and degradation markers during recovery.
- If the feed remains `DEGRADED` or `STALE`, keep the condition visible and hand off with the active reasons unchanged.
