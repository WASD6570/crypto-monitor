# Dashboard Surface And Operator Usage

## Module Requirements And Scope

Target repo areas:

- `apps/web` for panel layout, rendering states, and operator messaging
- service query consumers in `apps/web` for slow-context read paths
- fixture-backed UI tests when implementation begins

This module defines how slower institutional context appears inside the visibility dashboard without diluting the main reading order.

In scope:

- placement of the slow-context panel inside the existing dashboard IA
- timestamp, cadence, freshness, and advisory labeling
- missing, delayed, stale, partial, and unavailable UI behavior
- operator usage rules that keep slow context explanatory and non-blocking

Out of scope:

- redesigning the core dashboard information architecture
- adding interactive historical analytics or deep drill-down workflows
- client-side derivation of market-state or freshness rules

## Placement And Reading Order

- Keep the dashboard reading order from `visibility-dashboard-core` intact: live state first, slow context second.
- Place the slow-context panel after the main symbol state and derivatives context, but before low-priority ancillary detail.
- Use a visually distinct section title such as `Slow USA Context` or `Institutional Context`, plus a persistent `Context only` badge.
- On mobile, the panel should stack as one compact section with tap-safe rows or cards and no hidden critical timestamp labels.

## Panel Content Model

For each asset or relevant shared section, display:

- metric name: CME volume, CME open interest, ETF daily net flow
- latest normalized value and unit
- `as of` timestamp in UTC
- freshness badge: `Fresh`, `Delayed`, `Stale`, or `Unavailable`
- cadence label such as `Daily publication` or `Session-based publication`
- concise operator message clarifying slower-cadence semantics

When helpful, show change versus prior published value, but only if services provide it directly. Do not derive comparison rules in the client.

## UI Behavior Rules

- Never block page render on slow-context fetch completion.
- If slow context is loading, render a reserved panel shell with explanatory placeholder text rather than shifting the rest of the dashboard.
- If slow context errors, show an isolated error state inside the panel while preserving the rest of the dashboard.
- If only some metrics are available, show partial availability explicitly instead of collapsing the entire panel.
- Use neutral styling that signals context and trust level without competing with `TRADEABLE/WATCH/NO-OPERATE` emphasis.

## Operator Messaging Defaults

- Baseline helper text: `These indicators update on a slower schedule than market-state feeds.`
- Fresh helper text: `Use as backdrop for USA participation and positioning, not as a live gate.`
- Delayed helper text: `Latest publication is later than expected; rely on realtime state first.`
- Stale helper text: `Latest slow context is stale; do not read it as current market confirmation.`
- Unavailable helper text: `No trusted slow context is available right now; market-state decisions still come from realtime services.`

## Non-Blocking Integration Rules

- The dashboard must render realtime state successfully when slow context is missing or stale.
- Do not gate summary badges, alert readiness language, or symbol state banners on slow-context presence in MVP.
- If product copy references CME or ETF context elsewhere, it must repeat that they are advisory inputs.
- Any future upgrade that allows slow context to affect gating requires a separate feature plan and explicit rollout safeguards.

## UI Test Expectations

- fresh panel renders values, cadence labels, and `Context only` messaging
- delayed panel renders age and delayed messaging without altering symbol state badges
- stale panel preserves last-known values with explicit stale labeling
- unavailable panel renders clean fallback copy and does not break layout
- partial metric availability renders present metrics and isolates missing ones
- mobile layout keeps timestamps, freshness badges, and helper text readable

## Summary

This module keeps the operator workflow honest: look at live state first, then consult slow USA context for explanation. The dashboard integration must stay visually clear, lightweight, and impossible to mistake for a hidden realtime gate.
