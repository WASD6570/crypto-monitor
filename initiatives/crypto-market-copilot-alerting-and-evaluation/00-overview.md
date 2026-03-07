# Alerting And Evaluation Overview

## Objective

After trusted market state exists, add alerts and review workflows that help the user pay attention only when conditions justify it.

## User Outcome

The user should be able to:

- receive a bounded alert without babysitting the market
- immediately see why it fired and why it was allowed
- later review whether it worked after costs and latency assumptions
- compare the live alert stack against simple baselines instead of trusting it blindly

## In Scope

- setups A, B, and C with explicit permissions and hygiene
- global and symbol-level alert permissioning using market state and risk state
- delivery via UI, Telegram, and optional webhook
- objective outcomes across 30s, 2m, and 5m
- saved simulated execution runs
- operator feedback, notes, and review workflow
- baseline comparison and config-versioned tuning workflow
- replay and analytics UI for alert review

## Out Of Scope

- live order submission
- AI-led live ranking or explanation
- exchange credentials and private endpoints
- automatic threshold mutation from user feedback alone

## Prerequisites From Initiative 1

- stable contracts and fixtures
- deterministic replay foundation
- composite features and market state outputs
- visible feed health and degradation states
- dashboard and query surfaces that can host alert and review data

## Exit Criteria

- every alert has payload, delivery record, and outcome record
- alert precision and fragmented-market false positives are measurable against baselines
- saved simulations show net viability after latency, fees, and slippage assumptions
- the user can review a past alert end to end in under 60 seconds

## Ordered Slice Queue

1. `alert-generation-and-hygiene`
2. `tactical-risk-state-and-permissioning`
3. `alert-delivery-and-routing`
4. `outcome-evaluation`
5. `simulated-execution`
6. `operator-feedback-and-notes`
7. `replay-and-analytics-ui`
8. `baseline-comparison-and-tuning`
