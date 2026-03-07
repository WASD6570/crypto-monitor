#!/usr/bin/env python3

from __future__ import annotations

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CONFIG_PATHS = (
    ROOT / "configs/local/ingestion.v1.json",
    ROOT / "configs/dev/ingestion.v1.json",
    ROOT / "configs/prod/ingestion.v1.json",
)
REQUIRED_SYMBOLS = {"BTC-USD", "ETH-USD"}
REQUIRED_VENUES = {
    "BINANCE": {
        "service_path": "services/venue-binance",
        "required_streams": {
            "trades",
            "top-of-book",
            "order-book",
            "funding-rate",
            "open-interest",
            "mark-index",
            "liquidation",
        },
    },
    "BYBIT": {
        "service_path": "services/venue-bybit",
        "required_streams": {
            "trades",
            "top-of-book",
            "order-book",
            "funding-rate",
            "open-interest",
            "mark-index",
            "liquidation",
        },
    },
    "COINBASE": {
        "service_path": "services/venue-coinbase",
        "required_streams": {"trades", "top-of-book"},
    },
    "KRAKEN": {
        "service_path": "services/venue-kraken",
        "required_streams": {"trades", "top-of-book", "order-book"},
    },
}
POSITIVE_INT_FIELDS = (
    ("websocket", "heartbeatTimeoutMs"),
    ("websocket", "reconnectBackoffMinMs"),
    ("websocket", "reconnectBackoffMaxMs"),
    ("websocket", "reconnectLoopThreshold"),
    ("websocket", "connectsPerMinuteLimit"),
    ("health", "messageStaleAfterMs"),
    ("health", "snapshotStaleAfterMs"),
    ("health", "resyncLoopThreshold"),
    ("health", "clockOffsetWarningMs"),
    ("health", "clockOffsetDegradedMs"),
)


def load_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def validate_config(path: Path, errors: list[str]) -> None:
    if not path.exists():
        errors.append(f"Missing ingestion config: {path.relative_to(ROOT)}")
        return

    try:
        payload = load_json(path)
    except json.JSONDecodeError as exc:
        errors.append(f"Invalid JSON in {path.relative_to(ROOT)}: {exc}")
        return

    if payload.get("schemaVersion") != "v1":
        errors.append(f"{path.relative_to(ROOT)} must declare schemaVersion 'v1'")

    symbols = set(payload.get("symbols", []))
    if symbols != REQUIRED_SYMBOLS:
        errors.append(
            f"{path.relative_to(ROOT)} must declare symbols {sorted(REQUIRED_SYMBOLS)}, found {sorted(symbols)}"
        )

    handoff = payload.get("normalizerHandoff", {})
    if handoff.get("service") != "services/normalizer":
        errors.append(f"{path.relative_to(ROOT)} must hand off to services/normalizer")
    for key in ("preserveExchangeTs", "preserveRecvTs", "propagateDegradedReasons"):
        if handoff.get(key) is not True:
            errors.append(f"{path.relative_to(ROOT)} normalizerHandoff.{key} must be true")

    venues = payload.get("venues")
    if not isinstance(venues, dict):
        errors.append(f"{path.relative_to(ROOT)} must declare a venues object")
        return

    missing_venues = sorted(set(REQUIRED_VENUES).difference(venues))
    if missing_venues:
        errors.append(f"{path.relative_to(ROOT)} is missing venues: {', '.join(missing_venues)}")

    for venue_name, venue_rules in REQUIRED_VENUES.items():
        venue_config = venues.get(venue_name)
        if not isinstance(venue_config, dict):
            continue

        validate_service_path(path, venue_name, venue_rules["service_path"], venue_config, errors)
        validate_thresholds(path, venue_name, venue_config, errors)
        validate_snapshot_policy(path, venue_name, venue_config, errors)
        validate_streams(path, venue_name, venue_rules["required_streams"], venue_config, errors)


def validate_service_path(
    config_path: Path,
    venue_name: str,
    expected_path: str,
    venue_config: dict,
    errors: list[str],
) -> None:
    actual_path = venue_config.get("servicePath")
    if actual_path != expected_path:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} must use servicePath {expected_path!r}, found {actual_path!r}"
        )
        return

    service_dir = ROOT / actual_path
    if not service_dir.exists():
        errors.append(f"Missing service directory for {venue_name}: {actual_path}")
        return

    readme_path = service_dir / "README.md"
    if not readme_path.exists():
        errors.append(f"Missing service README for {venue_name}: {readme_path.relative_to(ROOT)}")


def validate_thresholds(config_path: Path, venue_name: str, venue_config: dict, errors: list[str]) -> None:
    for section_name, field_name in POSITIVE_INT_FIELDS:
        value = venue_config.get(section_name, {}).get(field_name)
        if not isinstance(value, int) or value <= 0:
            errors.append(
                f"{config_path.relative_to(ROOT)} venue {venue_name} field {section_name}.{field_name} must be a positive integer"
            )

    reconnect_min = venue_config.get("websocket", {}).get("reconnectBackoffMinMs")
    reconnect_max = venue_config.get("websocket", {}).get("reconnectBackoffMaxMs")
    if isinstance(reconnect_min, int) and isinstance(reconnect_max, int) and reconnect_max < reconnect_min:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} reconnectBackoffMaxMs must be >= reconnectBackoffMinMs"
        )

    clock_warning = venue_config.get("health", {}).get("clockOffsetWarningMs")
    clock_degraded = venue_config.get("health", {}).get("clockOffsetDegradedMs")
    if isinstance(clock_warning, int) and isinstance(clock_degraded, int) and clock_degraded < clock_warning:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} clockOffsetDegradedMs must be >= clockOffsetWarningMs"
        )

    rest = venue_config.get("rest", {})
    snapshot_limit = rest.get("snapshotRecoveryPerMinuteLimit")
    snapshot_cooldown = rest.get("snapshotCooldownMs")
    if not isinstance(snapshot_limit, int) or snapshot_limit < 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} rest.snapshotRecoveryPerMinuteLimit must be >= 0"
        )
    if not isinstance(snapshot_cooldown, int) or snapshot_cooldown < 0:
        errors.append(f"{config_path.relative_to(ROOT)} venue {venue_name} rest.snapshotCooldownMs must be >= 0")

    resubscribe = venue_config.get("websocket", {}).get("resubscribeOnReconnect")
    if resubscribe is not True:
        errors.append(f"{config_path.relative_to(ROOT)} venue {venue_name} websocket.resubscribeOnReconnect must be true")


def validate_snapshot_policy(config_path: Path, venue_name: str, venue_config: dict, errors: list[str]) -> None:
    snapshot_policy = venue_config.get("snapshotRefreshPolicy", {})
    required = snapshot_policy.get("required")
    interval = snapshot_policy.get("refreshIntervalMs")
    if not isinstance(required, bool):
        errors.append(f"{config_path.relative_to(ROOT)} venue {venue_name} snapshotRefreshPolicy.required must be boolean")
        return
    if not isinstance(interval, int) or interval < 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} snapshotRefreshPolicy.refreshIntervalMs must be >= 0"
        )
        return
    if required and interval <= 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} snapshotRefreshPolicy.refreshIntervalMs must be positive when snapshots are required"
        )
    if not required and interval != 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} snapshotRefreshPolicy.refreshIntervalMs must be 0 when snapshots are not required"
        )


def validate_streams(
    config_path: Path,
    venue_name: str,
    required_streams: set[str],
    venue_config: dict,
    errors: list[str],
) -> None:
    streams = venue_config.get("streams")
    if not isinstance(streams, list) or not streams:
        errors.append(f"{config_path.relative_to(ROOT)} venue {venue_name} must declare a non-empty streams list")
        return

    seen_streams: set[str] = set()
    snapshot_required_count = 0
    for stream in streams:
        kind = stream.get("kind")
        market_type = stream.get("marketType")
        snapshot_required = stream.get("snapshotRequired")
        if not kind or not market_type or not isinstance(snapshot_required, bool):
            errors.append(
                f"{config_path.relative_to(ROOT)} venue {venue_name} has invalid stream definition {stream}"
            )
            continue
        if kind in seen_streams:
            errors.append(f"{config_path.relative_to(ROOT)} venue {venue_name} duplicates stream kind {kind!r}")
            continue
        seen_streams.add(kind)
        if snapshot_required:
            snapshot_required_count += 1

    missing_streams = sorted(required_streams.difference(seen_streams))
    if missing_streams:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} is missing required streams: {', '.join(missing_streams)}"
        )

    snapshot_policy_required = venue_config.get("snapshotRefreshPolicy", {}).get("required")
    if snapshot_policy_required and snapshot_required_count == 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} requires snapshots but no stream declares snapshotRequired=true"
        )
    if not snapshot_policy_required and snapshot_required_count != 0:
        errors.append(
            f"{config_path.relative_to(ROOT)} venue {venue_name} disables snapshot refresh but still marks snapshotRequired streams"
        )


def main() -> int:
    errors: list[str] = []
    for config_path in CONFIG_PATHS:
        validate_config(config_path, errors)

    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        return 1

    print("validated ingestion configs for local, dev, and prod")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
