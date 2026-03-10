# Ingestion Feed Health Ops

## Metric Inventory

| Metric | Meaning | Trigger Signal |
|---|---|---|
| `message_lag_ms` | Age of the last market message versus current time | `message-stale` when freshness crosses venue threshold |
| `snapshot_lag_ms` | Age of the last book snapshot for snapshot-required streams | `snapshot-stale` when freshness crosses venue threshold |
| `reconnect_count` | Consecutive reconnect attempts in the active loop | `reconnect-loop` when threshold is reached |
| `resync_count` | Consecutive resync requests in the active loop | `resync-loop` when threshold is reached |
| `sequence_gap_count` | Number of detected sequence gaps in the current window | `sequence-gap` on any non-recoverable continuity loss |
| `snapshot_recovery_attempts_per_minute` | Snapshot recovery pressure inside the rolling minute window | operator warning before `resync-loop` or cooldown exhaustion |
| `rest_poll_attempts_per_minute` | REST poll pressure inside the rolling minute window | `rate-limit` when polling exceeds the configured budget |
| `clock_offset_ms` | Absolute local versus exchange clock skew | `clock-degraded` when degraded threshold is reached |

## Alert-Condition Matrix

| Feed State | Required Reasons | Operator Meaning | Immediate Action |
|---|---|---|---|
| `HEALTHY` | none | Adapter is current and trusted | continue monitoring |
| `DEGRADED` | `connection-not-ready` | Stream is reconnecting or not yet ready | inspect reconnect cadence and subscription health |
| `DEGRADED` | `sequence-gap` | Book continuity is broken | force resync, preserve degraded markers downstream |
| `DEGRADED` | `reconnect-loop` | Reconnects hit configured loop threshold | check venue/network reachability and backoff behavior |
| `DEGRADED` | `resync-loop` | Resync pressure is repeating | inspect sequence integrity and recovery preconditions |
| `DEGRADED` | `rate-limit` | REST poll cadence exceeded the configured budget | reduce poll pressure and confirm the next retry window |
| `DEGRADED` | `clock-degraded` | Local timestamps cannot be trusted | investigate host clock and exchange skew |
| `STALE` | `message-stale` | Messages stopped inside freshness window | confirm venue stream activity and reconnect state |
| `STALE` | `snapshot-stale` | Snapshot-required book is too old | trigger bounded snapshot recovery and confirm recovery pressure |

## Vocabulary Rules

- Use only shared state names: `HEALTHY`, `DEGRADED`, `STALE`.
- Use only shared degradation reasons: `connection-not-ready`, `message-stale`, `snapshot-stale`, `sequence-gap`, `reconnect-loop`, `resync-loop`, `rate-limit`, `clock-degraded`.
- Do not replace these names with ops aliases in logs, alerts, or runbooks.
