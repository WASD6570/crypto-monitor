# Implementation Module 1: Binance Profile Alignment

## Scope

- Update the checked-in Binance environment profiles so `local`, `dev`, and `prod` all declare valid open-interest polling defaults for the already-configured perpetual stream.
- Keep the change limited to profile data in `configs/*`; do not change provider startup semantics in this module.

## Target Repo Areas

- `configs/local/ingestion.v1.json`
- `configs/dev/ingestion.v1.json`
- `configs/prod/ingestion.v1.json`

## Requirements

- Preserve the existing `environment` labels and the fixed symbol list `BTC-USD`, `ETH-USD`.
- Keep the Binance stream set unchanged, including the perpetual `open-interest` stream.
- Ensure the Binance `rest` block has explicit positive `openInterestPollIntervalMs` and `openInterestPollsPerMinuteLimit` values in all three profiles.
- Keep the environment ladder conservative: `dev` must not be more aggressive than `local`, and `prod` must not be more aggressive than `dev`.
- Avoid unrelated edits to Bybit, Coinbase, Kraken, or non-Binance config fields.

## Key Decisions

- Treat missing Binance open-interest poll settings as a configuration bug to fix, not as a signal to remove the stream.
- Preserve the existing local profile as the reference point unless implementation proves a shared correction is necessary.
- Encode the chosen `dev` and `prod` values directly in the checked-in JSON so rollout defaults are visible without reading code.

## Unit Test Expectations

- Each environment file loads successfully through `LoadEnvironmentConfig(...)`.
- `RuntimeConfigFor(ingestion.VenueBinance)` returns positive open-interest poll settings for every environment.
- The chosen `dev` and `prod` values satisfy the planned monotonic pressure ladder.

## Contract / Fixture / Replay Impacts

- No public API, schema, or replay fixture changes are expected.
- The impact is limited to checked-in runtime defaults that the current live Binance path already consumes.

## Summary

This module makes the checked-in Binance rollout profiles honest and consumable: the perpetual open-interest stream stays enabled, and every environment declares the polling defaults required by the current live runtime.
