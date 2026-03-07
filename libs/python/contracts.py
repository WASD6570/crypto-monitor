from __future__ import annotations

import json
from typing import Any
from pathlib import Path


REQUIRED_FIELDS_BY_SCHEMA = {
    "market-trade.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
    "order-book-top.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
    "feed-health.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "feedHealthState",
        "sourceRecordId",
    ],
    "funding-rate.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
    "open-interest-snapshot.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
    "mark-index.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
    "liquidation-print.v1.schema.json": [
        "schemaVersion",
        "eventType",
        "symbol",
        "sourceSymbol",
        "quoteCurrency",
        "venue",
        "marketType",
        "exchangeTs",
        "recvTs",
        "timestampStatus",
        "sourceRecordId",
    ],
}

EXPECTED_EVENT_TYPE_BY_SCHEMA = {
    "market-trade.v1.schema.json": "market-trade",
    "order-book-top.v1.schema.json": "order-book-top",
    "feed-health.v1.schema.json": "feed-health",
    "funding-rate.v1.schema.json": "funding-rate",
    "open-interest-snapshot.v1.schema.json": "open-interest-snapshot",
    "mark-index.v1.schema.json": "mark-index",
    "liquidation-print.v1.schema.json": "liquidation-print",
}


def load_json(file_path: Path) -> dict[str, Any]:
    return json.loads(file_path.read_text(encoding="utf-8"))


def validate_fixture(fixture: dict[str, Any]) -> None:
    if fixture.get("fixtureVersion") != "v1":
        raise ValueError("fixtureVersion must be v1")

    target_schema = fixture.get("targetSchema")
    if not isinstance(target_schema, str):
        raise ValueError("fixture targetSchema must be a string")

    required_fields = REQUIRED_FIELDS_BY_SCHEMA.get(target_schema)
    if not required_fields:
        raise ValueError(f"unsupported target schema {target_schema}")

    expected_event_type = EXPECTED_EVENT_TYPE_BY_SCHEMA[target_schema]
    expected_canonical = fixture.get("expectedCanonical", [])
    if not isinstance(expected_canonical, list) or not expected_canonical:
        raise ValueError("expectedCanonical must not be empty")

    for index, payload in enumerate(expected_canonical):
        if not isinstance(payload, dict):
            raise ValueError(f"canonical payload {index} must be an object")
        for field in required_fields:
            if field not in payload:
                raise ValueError(f"canonical payload {index} missing field {field}")
        if payload.get("schemaVersion") != "v1":
            raise ValueError(f"canonical payload {index} has unsupported schemaVersion")
        if payload.get("eventType") != expected_event_type:
            raise ValueError(f"canonical payload {index} has unexpected eventType")


def validate_replay_seed(seed: dict[str, Any]) -> None:
    if seed.get("schemaVersion") != "v1":
        raise ValueError("schemaVersion must be v1")
    if seed.get("targetSchema") != "replay-seed.v1.schema.json":
        raise ValueError("unsupported replay seed target schema")

    expected_determinism = seed.get("expectedDeterminism")
    if not isinstance(expected_determinism, dict):
        raise ValueError("expectedDeterminism must be an object")

    ordered_ids = expected_determinism.get("orderedSourceRecordIds", [])
    event_count = expected_determinism.get("eventCount", 0)
    if not isinstance(ordered_ids, list) or not isinstance(event_count, int):
        raise ValueError("expectedDeterminism must include orderedSourceRecordIds and eventCount")
    if event_count <= 0 or len(ordered_ids) != event_count:
        raise ValueError("orderedSourceRecordIds length must match eventCount")
