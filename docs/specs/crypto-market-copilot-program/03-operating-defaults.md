# Operating Defaults

These defaults are intentionally conservative. Later agents should use them unless a child feature plan explicitly replaces them.

## Time Synchronization Policy

### Clock Discipline

- All live services should run with NTP or Chrony synchronization enabled.
- Warning threshold: local clock offset >= 100ms.
- Degraded threshold: local clock offset >= 250ms.
- If offset cannot be corrected quickly, log degradation and reduce trust in latency-sensitive evaluation.

### `exchangeTs` vs `recvTs`

- `exchangeTs` is the primary timestamp for canonical event ordering and bucket assignment when present and plausible.
- `recvTs` must always be stored and is the source of truth for feed staleness, transport latency, and ops health.
- If `exchangeTs` is missing, clearly invalid, or outside a sane skew window, fall back to `recvTs` and mark the event as timestamp-degraded.

### Bucket Assignment Policy

- Assign feature buckets in UTC using event time first.
- Default bucket source: `exchangeTs`.
- Fallback bucket source: `recvTs`, with an explicit degraded flag.
- Never discard the original timestamp choice; it must be auditable during replay.

### Late And Out-Of-Order Events

- Live processors should tolerate small lateness windows:
  - 30s features: 2s default watermark
  - 2m features: 5s default watermark
  - 5m features: 10s default watermark
- Events arriving after the live watermark are still persisted but should be marked late and handled through replay/backfill correction rather than silent mutation of already-emitted alert decisions.

## Data Retention And Storage Policy

### Hot Storage Defaults

- Raw canonical events: 30 days hot
- Derived features and regime outputs: 90 days hot
- Alerts, outcomes, simulations, and decision logs: keep hot for at least 180 days

### Cold Storage Defaults

- Raw canonical events: compressed cold retention for 365 days
- Features: compressed cold retention for 365 days
- Alerts, outcomes, simulations, and decision logs: retain for at least 2 years because they are smaller and drive tuning value

### Replay Cost Expectations

- A single symbol, single day replay should be runnable on demand on local/dev infrastructure in minutes, not hours.
- Safe MVP expectation: one day replay for a symbol should finish within 10 minutes under normal conditions.
- If replay is slower, optimize observability and storage layout before adding more alert complexity.

### Dashboard Query Targets

- Current state panels: target under 2 seconds
- Recent alert drill-down: target under 5 seconds
- 24h historical review views: target under 10 seconds
- If a view cannot meet these targets, prefer precomputed aggregates over expensive ad hoc queries.

## Delivery Surface Defaults

- Source of truth: the web UI in `apps/web`
- First push surface: Telegram notifications
- Secondary integration surface: generic webhook delivery
- Deferred by default: Slack and email, unless the user later proves a real need
- `NO-OPERATE` or `STOP` alerts stay `INFO` severity on every delivery surface

## Human Feedback Loop Defaults

Each alert should support these first-user actions:

- save
- dismiss
- thumbs up
- thumbs down
- mark `good setup bad timing`
- mark `useful context only`
- add a free-form operator note

Feedback rules:

- feedback is stored immutably with timestamp and config version
- feedback does not directly mutate live thresholds
- later tuning workflows may use feedback as review context, not as automatic ground truth

## Experiment And Tuning Defaults

- New thresholds and rules should be proposed only as new config versions.
- Every candidate change must be tested against:
  - pinned replay fixtures
  - a recent rolling live-data window
  - the baseline controls in `02-product-success.md`
- A candidate may graduate only if it improves or preserves alert precision and net viability without worsening fragmented-market false positives.
- Default rollback rule: revert the candidate if rolling performance degrades materially versus the active config or if fragmented-market false positives worsen noticeably.

## Source Of Truth For Tradeability

- Venue feed health produces venue-specific degraded flags.
- Venue degraded flags feed composite market-quality calculations.
- Symbol-specific 5m regime is the primary gate for BTC and ETH alerts.
- Global regime acts as a hard ceiling:
  - if global state is `NO-OPERATE`, all symbols are `INFO` only
  - if global state is `WATCH`, no symbol can escalate above `WATCH`
- If BTC is `TRADEABLE` and ETH is `NO-OPERATE`, BTC alerts may still be actionable while ETH remains informational only.
- Critical venue degradation can downgrade a symbol from `TRADEABLE` to `WATCH` or `NO-OPERATE` even if the global state is still normal.

## AI Policy For This Program

- Defer AI-generated scoring, ranking, or live decision authority.
- Keep MVP decisions deterministic, replayable, and operator-auditable.
- AI may be considered later for summarization, clustering, or review assistance once the deterministic stack is proven.
