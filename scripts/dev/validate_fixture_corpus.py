#!/usr/bin/env python3

from __future__ import annotations

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
FIXTURE_DOC_PATH = ROOT / "docs/specs/canonical-contracts-and-fixtures/01-fixture-corpus.md"
FIXTURE_MANIFEST_PATH = ROOT / "tests/fixtures/manifest.v1.json"
REPLAY_MANIFEST_PATH = ROOT / "tests/replay/manifest.v1.json"

REQUIRED_DOC_SECTIONS = (
    "## Fixture Layout",
    "## Scenario Catalog",
    "## Replay Seed Catalog",
)
REQUIRED_FIXTURE_CATEGORIES = {
    "happy-trades",
    "happy-top-of-book",
    "happy-order-book-snapshot-delta",
    "happy-funding",
    "happy-open-interest",
    "happy-mark-index",
    "happy-liquidation",
    "edge-sequence-gap",
    "edge-forced-resync",
    "edge-stale-feed",
    "edge-timestamp-degraded",
    "edge-late-out-of-order",
    "edge-quote-variant",
}
REQUIRED_REPLAY_CATEGORIES = {
    "normal-microstructure",
    "fragmented-world-usa",
    "degraded-feed",
}
REQUIRED_FIXTURE_KEYS = {
    "fixtureVersion",
    "id",
    "family",
    "category",
    "scenarioClass",
    "venue",
    "symbol",
    "quoteCurrency",
    "targetSchema",
    "purpose",
    "checks",
    "rawMessages",
    "expectedCanonical",
}
REQUIRED_REPLAY_KEYS = {
    "schemaVersion",
    "id",
    "category",
    "symbol",
    "targetSchema",
    "purpose",
    "fixtureRefs",
    "tags",
    "expectedDeterminism",
}
REQUIRED_QUOTE_CURRENCIES = {"USD", "USDT", "USDC"}


def load_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def validate_docs(errors: list[str]) -> None:
    if not FIXTURE_DOC_PATH.exists():
        errors.append(f"Missing fixture corpus doc: {FIXTURE_DOC_PATH.relative_to(ROOT)}")
        return

    content = FIXTURE_DOC_PATH.read_text(encoding="utf-8")
    for section in REQUIRED_DOC_SECTIONS:
        if section not in content:
            errors.append(
                f"Fixture corpus doc is missing required section {section!r}: {FIXTURE_DOC_PATH.relative_to(ROOT)}"
            )


def validate_fixture_manifest(errors: list[str]) -> dict[str, dict]:
    if not FIXTURE_MANIFEST_PATH.exists():
        errors.append(f"Missing fixture manifest: {FIXTURE_MANIFEST_PATH.relative_to(ROOT)}")
        return {}

    try:
        manifest = load_json(FIXTURE_MANIFEST_PATH)
    except json.JSONDecodeError as exc:
        errors.append(f"Invalid JSON in {FIXTURE_MANIFEST_PATH.relative_to(ROOT)}: {exc}")
        return {}

    if manifest.get("manifestVersion") != "v1":
        errors.append(f"{FIXTURE_MANIFEST_PATH.relative_to(ROOT)} must declare manifestVersion 'v1'")

    categories = set(manifest.get("requiredCategories", []))
    missing_categories = sorted(REQUIRED_FIXTURE_CATEGORIES.difference(categories))
    if missing_categories:
        errors.append(
            f"{FIXTURE_MANIFEST_PATH.relative_to(ROOT)} is missing required categories: {', '.join(missing_categories)}"
        )

    fixtures = manifest.get("fixtures")
    if not isinstance(fixtures, list) or not fixtures:
        errors.append(f"{FIXTURE_MANIFEST_PATH.relative_to(ROOT)} must declare a non-empty fixtures list")
        return {}

    seen_ids: set[str] = set()
    seen_categories: set[str] = set()
    seen_quotes: set[str] = set()
    fixture_index: dict[str, dict] = {}

    for entry in fixtures:
        fixture_id = entry.get("id")
        path_value = entry.get("path")
        category = entry.get("category")

        if not fixture_id or not path_value or not category:
            errors.append(f"Fixture manifest entry is missing id, category, or path: {entry}")
            continue

        if fixture_id in seen_ids:
            errors.append(f"Duplicate fixture id in {FIXTURE_MANIFEST_PATH.relative_to(ROOT)}: {fixture_id}")
            continue
        seen_ids.add(fixture_id)
        seen_categories.add(category)

        fixture_path = ROOT / path_value
        if not fixture_path.exists():
            errors.append(f"Missing fixture file referenced by manifest: {path_value}")
            continue

        try:
            fixture = load_json(fixture_path)
        except json.JSONDecodeError as exc:
            errors.append(f"Invalid JSON in {path_value}: {exc}")
            continue

        fixture_index[fixture_id] = fixture

        missing_keys = sorted(REQUIRED_FIXTURE_KEYS.difference(fixture))
        if missing_keys:
            errors.append(f"{path_value} is missing required keys: {', '.join(missing_keys)}")
            continue

        if fixture["id"] != fixture_id:
            errors.append(f"{path_value} declares id {fixture['id']!r}, expected {fixture_id!r}")

        if fixture["category"] != category:
            errors.append(f"{path_value} declares category {fixture['category']!r}, expected {category!r}")

        raw_messages = fixture["rawMessages"]
        expected_canonical = fixture["expectedCanonical"]
        checks = fixture["checks"]

        if not isinstance(raw_messages, list) or not raw_messages:
            errors.append(f"{path_value} must declare a non-empty rawMessages list")

        if not isinstance(expected_canonical, list) or not expected_canonical:
            errors.append(f"{path_value} must declare a non-empty expectedCanonical list")

        if not isinstance(checks, list) or not checks:
            errors.append(f"{path_value} must declare a non-empty checks list")

        quote_currency = fixture.get("quoteCurrency")
        if quote_currency:
            seen_quotes.add(quote_currency)

    missing_seen_categories = sorted(REQUIRED_FIXTURE_CATEGORIES.difference(seen_categories))
    if missing_seen_categories:
        errors.append(
            f"Fixture corpus is missing scenario implementations for: {', '.join(missing_seen_categories)}"
        )

    missing_quotes = sorted(REQUIRED_QUOTE_CURRENCIES.difference(seen_quotes))
    if missing_quotes:
        errors.append(
            f"Fixture corpus is missing required quote currencies: {', '.join(missing_quotes)}"
        )

    return fixture_index


def validate_replay_manifest(fixture_index: dict[str, dict], errors: list[str]) -> None:
    if not REPLAY_MANIFEST_PATH.exists():
        errors.append(f"Missing replay manifest: {REPLAY_MANIFEST_PATH.relative_to(ROOT)}")
        return

    try:
        manifest = load_json(REPLAY_MANIFEST_PATH)
    except json.JSONDecodeError as exc:
        errors.append(f"Invalid JSON in {REPLAY_MANIFEST_PATH.relative_to(ROOT)}: {exc}")
        return

    if manifest.get("manifestVersion") != "v1":
        errors.append(f"{REPLAY_MANIFEST_PATH.relative_to(ROOT)} must declare manifestVersion 'v1'")

    categories = set(manifest.get("requiredSeedCategories", []))
    missing_categories = sorted(REQUIRED_REPLAY_CATEGORIES.difference(categories))
    if missing_categories:
        errors.append(
            f"{REPLAY_MANIFEST_PATH.relative_to(ROOT)} is missing required replay categories: {', '.join(missing_categories)}"
        )

    seeds = manifest.get("seeds")
    if not isinstance(seeds, list) or not seeds:
        errors.append(f"{REPLAY_MANIFEST_PATH.relative_to(ROOT)} must declare a non-empty seeds list")
        return

    seen_categories: set[str] = set()
    seen_ids: set[str] = set()

    for entry in seeds:
        seed_id = entry.get("id")
        seed_category = entry.get("category")
        path_value = entry.get("path")
        fixture_refs = entry.get("fixtureRefs")

        if not seed_id or not seed_category or not path_value:
            errors.append(f"Replay manifest entry is missing id, category, or path: {entry}")
            continue

        if seed_id in seen_ids:
            errors.append(f"Duplicate replay seed id in {REPLAY_MANIFEST_PATH.relative_to(ROOT)}: {seed_id}")
            continue
        seen_ids.add(seed_id)
        seen_categories.add(seed_category)

        seed_path = ROOT / path_value
        if not seed_path.exists():
            errors.append(f"Missing replay seed file referenced by manifest: {path_value}")
            continue

        try:
            seed = load_json(seed_path)
        except json.JSONDecodeError as exc:
            errors.append(f"Invalid JSON in {path_value}: {exc}")
            continue

        missing_keys = sorted(REQUIRED_REPLAY_KEYS.difference(seed))
        if missing_keys:
            errors.append(f"{path_value} is missing required keys: {', '.join(missing_keys)}")
            continue

        if seed["schemaVersion"] != "v1":
            errors.append(f"{path_value} must declare schemaVersion 'v1'")

        if seed["id"] != seed_id:
            errors.append(f"{path_value} declares id {seed['id']!r}, expected {seed_id!r}")

        if seed["category"] != seed_category:
            errors.append(f"{path_value} declares category {seed['category']!r}, expected {seed_category!r}")

        if fixture_refs != seed["fixtureRefs"]:
            errors.append(f"{path_value} fixtureRefs must match the replay manifest entry")

        if not isinstance(fixture_refs, list) or not fixture_refs:
            errors.append(f"{path_value} must declare a non-empty fixtureRefs list")
            continue

        for fixture_ref in fixture_refs:
            if fixture_ref not in fixture_index:
                errors.append(f"{path_value} references unknown fixture id {fixture_ref!r}")

        expected = seed["expectedDeterminism"]
        if not isinstance(expected, dict):
            errors.append(f"{path_value} must declare expectedDeterminism as an object")
            continue

        event_count = expected.get("eventCount")
        ordered_ids = expected.get("orderedSourceRecordIds")
        if not isinstance(event_count, int) or event_count <= 0:
            errors.append(f"{path_value} must declare a positive integer expectedDeterminism.eventCount")

        if not isinstance(ordered_ids, list) or not ordered_ids:
            errors.append(f"{path_value} must declare a non-empty expectedDeterminism.orderedSourceRecordIds list")
            continue

        if isinstance(event_count, int) and len(ordered_ids) != event_count:
            errors.append(
                f"{path_value} eventCount {event_count} does not match orderedSourceRecordIds length {len(ordered_ids)}"
            )

    missing_seen_categories = sorted(REQUIRED_REPLAY_CATEGORIES.difference(seen_categories))
    if missing_seen_categories:
        errors.append(
            f"Replay seed catalog is missing implementations for: {', '.join(missing_seen_categories)}"
        )


def main() -> int:
    errors: list[str] = []

    validate_docs(errors)
    fixture_index = validate_fixture_manifest(errors)
    validate_replay_manifest(fixture_index, errors)

    if errors:
      for error in errors:
          print(f"ERROR: {error}")
      return 1

    print("Fixture corpus and replay seeds validate successfully.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
