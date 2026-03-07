# Initiative Map

## Initiative 1: `crypto-market-copilot-visibility-foundation`

- Goal: give the first user a trustworthy live picture of BTC/ETH market state, feed health, WORLD vs USA divergence, and replayable context before any alerting promises are made.
- Primary repo areas: `services/venue-*`, `services/normalizer`, `services/replay-engine`, `services/feature-engine`, `services/regime-engine`, `apps/web`, `schemas/json`, `tests/fixtures`, `tests/replay`
- Exit standard: the user can open the product and understand current state, degraded feeds, and regime in under 60 seconds.

### What this initiative must deliver

- canonical contracts and deterministic fixtures
- resilient ingestion and feed health visibility
- raw append-only storage plus deterministic replay
- WORLD and USA composites plus market-quality features
- explicit `TRADEABLE/WATCH/NO-OPERATE` market state
- dashboards for symbol overview, microstructure, derivatives context, and feed/regime visibility

### What it must not pretend to solve yet

- setup quality
- push alerts
- outcome truth
- simulated execution
- threshold tuning from live alert history

## Initiative 2: `crypto-market-copilot-alerting-and-evaluation`

- Goal: turn trusted market state into bounded alerts, objective outcomes, and an operator review loop that proves whether alerts deserve ongoing attention.
- Primary repo areas: `services/alert-engine`, `services/risk-engine`, `services/outcome-engine`, `services/simulation-api`, `apps/web`, `schemas/json/alerts`, `schemas/json/outcomes`, `schemas/json/simulation`, `tests/integration`, `tests/e2e`
- Exit standard: every emitted alert can be delivered, explained, evaluated, and compared against baselines without guessing.

### What this initiative must deliver

- setups A/B/C with hygiene and permissions
- severity, dedupe, cooldown, and clustering behavior
- objective outcomes across 30s, 2m, and 5m
- saved simulated execution results
- operator feedback and review workflow
- baseline comparison and tuning workflow

### Dependency on initiative 1

Initiative 2 depends on initiative 1 for:

- stable contracts and fixtures
- deterministic replay
- trusted composite features and regime state
- feed health signals and degradation states
- dashboard shells and review surfaces that can host alert detail and outcome drill-downs

## Shared Program Tracks

### Contracts And Fixtures

- Required before either initiative can move safely.
- Must include event, feature, alert, outcome, replay, and simulation payload families.

### Config Versioning

- Thresholds, fee models, slippage assumptions, cooldowns, and baselines must all be configuration-driven and versioned.
- Later alerts and outcomes must store the config version used at evaluation time.

### Replay Discipline

- Replay is the bridge between both initiatives.
- Visibility uses replay to prove state correctness; alerting uses replay to prove alert and outcome correctness.

### Operator Trust

- Every screen and alert must answer "why" before it tries to answer "what to do."
