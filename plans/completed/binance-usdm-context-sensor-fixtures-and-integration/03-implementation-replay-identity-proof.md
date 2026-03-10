# Implementation: Replay Identity Proof

## Module Requirements

- Add replay-sensitive checks for at least one websocket-derived event and one REST-derived event.
- Prove repeated normalization preserves stable `sourceRecordId` values and raw duplicate behavior per stream family.
- Keep the proof targeted to Binance USD-M context sensors only.

## Target Repo Areas

- `tests/integration`
- `libs/go/ingestion`
- `services/normalizer`

## Unit Test Expectations

- repeated websocket normalization keeps the same canonical identity
- repeated REST normalization keeps the same canonical identity
- raw append duplicate facts remain scoped to the correct stream family

## Summary

This module closes the proof gap between isolated parser tests and later replay work by making repeated mixed-surface inputs deterministic now.
