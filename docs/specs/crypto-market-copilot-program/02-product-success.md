# Product Success And Baselines

## Primary User Outcome

The system succeeds when the first user does not need to stare at markets all day to stay informed.

The MVP should compress 24/7 market monitoring into a smaller loop:

- glance at current state
- receive high-signal alerts only when warranted
- later review whether those alerts were actually worth attention

## Program-Level Success Metrics

### 1. Alert Precision By Regime

- Definition: percentage of emitted alerts whose `target1` hits before invalidation, segmented by market regime and setup.
- Standard slices: `TRADEABLE`, `WATCH`, `NO-OPERATE`, fragmented vs non-fragmented, high vs normal leverage stress.
- Why it matters: the product should prove that regime gating improves signal quality rather than just labeling the tape.

### 2. Reduction In False Positives During Fragmented Markets

- Definition: compare the false-positive rate of the production alert stack against baseline controls specifically when USA vs WORLD divergence or fragmentation is elevated.
- Why it matters: one of the core product promises is to avoid bad states, not just detect good ones.

### 3. Percent Of Alerts With Positive Net Viability After Costs

- Definition: percentage of alerts whose simulated or evaluated net result remains positive after configured fees, slippage, and latency assumptions.
- Why it matters: raw directional correctness is not enough if costs erase the edge.

### 4. Time-To-Trust

- Definition: median time for the user to explain why an alert fired and whether it later worked, using only the product surfaces.
- MVP target: the product should make both answers available within 60 seconds for a recent alert.
- Why it matters: trust is a product outcome, not just a model property.

## Initiative-1 Success Gates

- live state and feed health can be understood in under 60 seconds
- replay reproduces the same state outputs for the same day and config
- degraded feeds and fragmented regimes are visible and explainable
- dashboards feel like a trustworthy monitoring surface, not a noisy chart wall

## Initiative-2 Success Gates

- every alert has a full payload, delivery record, and outcome record
- regime-aware alert quality is better than the defined baselines
- simulated execution shows whether edge survives costs and latency
- the user can save feedback on alerts and review patterns over time

## Baseline Controls

These baselines are required so later tuning has an honest comparison set.

### Baseline A: `naive-breakout`

- Logic: trigger when price breaks a chosen reference level with no 2m validator and no 5m market-state filter.
- Use: compare against setup A to prove that validation and regime gating add value.

### Baseline B: `naive-vwap-reversion`

- Logic: trigger when price deviates beyond a fixed threshold from composite intraday VWAP and mean-reversion is assumed without deeper context.
- Use: compare against setup B to prove fakeout and absorption logic add value.

### Baseline C: `single-venue-trigger`

- Logic: trigger from one venue only, without WORLD vs USA composite confirmation or fragmentation awareness.
- Use: prove that multi-venue confirmation reduces noise and improves trust.

## Metric Storage Requirements

- Store metric snapshots by config version.
- Keep baseline and production runs directly comparable by symbol, setup family, regime, and horizon.
- Do not overwrite historical evaluations when configs change; append new results with version context.
