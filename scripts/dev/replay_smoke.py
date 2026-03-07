#!/usr/bin/env python3

from __future__ import annotations

import hashlib
import json
import os
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def materialize_order(seed: dict, fixtures_by_id: dict[str, dict]) -> list[str]:
    ordered: list[str] = []
    for fixture_id in seed["fixtureRefs"]:
        fixture = fixtures_by_id[fixture_id]
        ordered.extend(item["sourceRecordId"] for item in fixture["expectedCanonical"])
    return ordered


def deterministic_checksum(seed: dict, fixtures_by_id: dict[str, dict]) -> str:
    materialized = {
        "seed": seed["id"],
        "orderedSourceRecordIds": materialize_order(seed, fixtures_by_id),
    }
    return hashlib.sha256(json.dumps(materialized, sort_keys=True).encode("utf-8")).hexdigest()


def main() -> int:
    if os.environ.get("CONTRACT_FIXTURES") != "1":
        print("ERROR: replay smoke requires CONTRACT_FIXTURES=1", file=sys.stderr)
        return 1

    fixture_manifest = load_json(ROOT / "tests/fixtures/manifest.v1.json")
    replay_manifest = load_json(ROOT / "tests/replay/manifest.v1.json")

    fixtures_by_id = {
        entry["id"]: load_json(ROOT / entry["path"])
        for entry in fixture_manifest["fixtures"]
    }

    for entry in replay_manifest["seeds"]:
        seed = load_json(ROOT / entry["path"])
        expected = seed["expectedDeterminism"]
        ordered_ids = materialize_order(seed, fixtures_by_id)

        if ordered_ids != expected["orderedSourceRecordIds"]:
            print(f"ERROR: replay ordering mismatch for {seed['id']}", file=sys.stderr)
            return 1

        if len(ordered_ids) != expected["eventCount"]:
            print(f"ERROR: replay event count mismatch for {seed['id']}", file=sys.stderr)
            return 1

        first_checksum = deterministic_checksum(seed, fixtures_by_id)
        second_checksum = deterministic_checksum(seed, fixtures_by_id)
        if first_checksum != second_checksum:
            print(f"ERROR: replay checksum mismatch for {seed['id']}", file=sys.stderr)
            return 1

    print("Replay smoke checks passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
