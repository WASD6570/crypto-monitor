#!/usr/bin/env python3

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DOC_PATH = ROOT / "docs/specs/canonical-contracts-and-fixtures/00-contract-families.md"
CONSUMER_DOC_PATH = ROOT / "docs/specs/canonical-contracts-and-fixtures/02-consumer-adoption.md"
FIXTURE_MANIFEST_PATH = ROOT / "tests/fixtures/manifest.v1.json"
REPLAY_MANIFEST_PATH = ROOT / "tests/replay/manifest.v1.json"
SCHEMA_FILENAME_PATTERN = "<schema-name>.v{major}.schema.json"
SCHEMA_REGEX = re.compile(r"^[a-z0-9-]+\.(v\d+)\.schema\.json$")
EXPECTED_FAMILIES = {
    "events": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "sourceSymbol",
            "quoteCurrency",
            "venue",
            "marketType",
            "exchangeTs",
            "recvTs",
            "timestampStatus",
            "sourceRecordId",
        },
        "implemented_schemas": {
            "market-trade.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "order-book-top.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "feed-health.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "funding-rate.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "open-interest-snapshot.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "mark-index.v1.schema.json": {
                "required_fields": [
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
                ]
            },
            "liquidation-print.v1.schema.json": {
                "required_fields": [
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
                ]
            },
        },
    },
    "features": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "venue",
            "compositeId",
            "marketType",
            "exchangeTs",
            "recvTs",
            "bucketSource",
            "feedHealthState",
            "replayRef",
        }
    },
    "alerts": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "marketType",
            "configVersion",
            "regimeTags",
            "feedHealthState",
            "replayRef",
        }
    },
    "outcomes": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "configVersion",
            "regimeTags",
            "replayRef",
            "sourceRecordId",
        }
    },
    "replay": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "exchangeTs",
            "recvTs",
            "timestampStatus",
            "replayRef",
            "sourceRecordId",
        },
        "implemented_schemas": {
            "replay-seed.v1.schema.json": {
                "required_fields": [
                    "schemaVersion",
                    "id",
                    "category",
                    "symbol",
                    "targetSchema",
                    "purpose",
                    "fixtureRefs",
                    "tags",
                    "expectedDeterminism",
                ]
            },
            "replay-window.v1.schema.json": {
                "required_fields": [
                    "schemaVersion",
                    "id",
                    "symbol",
                    "startTs",
                    "endTs",
                    "fixtureRefs",
                ]
            },
            "replay-run-result.v1.schema.json": {
                "required_fields": [
                    "schemaVersion",
                    "id",
                    "seedId",
                    "symbol",
                    "status",
                    "outputChecksum",
                ]
            },
        }
    },
    "simulation": {
        "required_reserved_fields": {
            "schemaVersion",
            "symbol",
            "marketType",
            "configVersion",
            "regimeTags",
            "replayRef",
        }
    },
}
REQUIRED_DOC_SECTIONS = (
    "## Family Inventory",
    "## Naming And Versioning Standard",
    "## Reserved Field Glossary",
)
REQUIRED_CONSUMER_DOC_SECTIONS = (
    "## Validation Workflow",
    "## Go Adoption Checklist",
    "## TypeScript Adoption Checklist",
    "## Optional Python Parity",
    "## Breaking Change Protocol",
    "## Required Diff Checklist",
)
REQUIRED_SUPPORT_READMES = (
    ROOT / "libs/go/README.md",
    ROOT / "libs/ts/README.md",
    ROOT / "libs/python/README.md",
    ROOT / "tests/parity/README.md",
)


def load_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def validate_schema_file(path: Path, required_fields: list[str], errors: list[str]) -> None:
    try:
        schema = load_json(path)
    except json.JSONDecodeError as exc:
        errors.append(f"Invalid JSON in {path.relative_to(ROOT)}: {exc}")
        return

    required_keys = {"$schema", "$id", "title", "type", "required", "properties"}
    missing_keys = sorted(required_keys.difference(schema))
    if missing_keys:
        errors.append(f"{path.relative_to(ROOT)} is missing schema keys: {', '.join(missing_keys)}")
        return

    if schema["type"] != "object":
        errors.append(f"{path.relative_to(ROOT)} must declare type 'object'")

    required = schema["required"]
    properties = schema["properties"]
    if not isinstance(required, list) or not required:
        errors.append(f"{path.relative_to(ROOT)} must declare a non-empty required list")
        return

    if not isinstance(properties, dict) or not properties:
        errors.append(f"{path.relative_to(ROOT)} must declare a non-empty properties object")
        return

    missing_required = [field for field in required_fields if field not in required]
    if missing_required:
        errors.append(
            f"{path.relative_to(ROOT)} is missing required field(s): {', '.join(missing_required)}"
        )

    missing_properties = [field for field in required_fields if field not in properties]
    if missing_properties:
        errors.append(
            f"{path.relative_to(ROOT)} is missing property definitions for: {', '.join(missing_properties)}"
        )

    schema_version = properties.get("schemaVersion")
    if not isinstance(schema_version, dict) or schema_version.get("const") != "v1":
        errors.append(f"{path.relative_to(ROOT)} must pin properties.schemaVersion.const to 'v1'")


def validate_docs(errors: list[str]) -> None:
    if not DOC_PATH.exists():
        errors.append(f"Missing contract inventory doc: {DOC_PATH.relative_to(ROOT)}")
        return

    content = DOC_PATH.read_text(encoding="utf-8")
    for section in REQUIRED_DOC_SECTIONS:
        if section not in content:
            errors.append(
                f"Contract inventory doc is missing required section {section!r}: {DOC_PATH.relative_to(ROOT)}"
            )

    if not CONSUMER_DOC_PATH.exists():
        errors.append(f"Missing consumer adoption doc: {CONSUMER_DOC_PATH.relative_to(ROOT)}")
        return

    consumer_content = CONSUMER_DOC_PATH.read_text(encoding="utf-8")
    for section in REQUIRED_CONSUMER_DOC_SECTIONS:
        if section not in consumer_content:
            errors.append(
                f"Consumer adoption doc is missing required section {section!r}: {CONSUMER_DOC_PATH.relative_to(ROOT)}"
            )

    for readme_path in REQUIRED_SUPPORT_READMES:
        if not readme_path.exists():
            errors.append(f"Missing support README: {readme_path.relative_to(ROOT)}")


def validate_manifest(family: str, spec: dict, errors: list[str]) -> None:
    manifest_path = ROOT / f"schemas/json/{family}/family.v1.json"
    if not manifest_path.exists():
        errors.append(f"Missing family manifest: {manifest_path.relative_to(ROOT)}")
        return

    try:
        manifest = load_json(manifest_path)
    except json.JSONDecodeError as exc:
        errors.append(f"Invalid JSON in {manifest_path.relative_to(ROOT)}: {exc}")
        return

    required_keys = {
        "manifestVersion",
        "family",
        "directory",
        "filenamePattern",
        "schemaVersionPolicy",
        "supportedSchemaVersions",
        "purpose",
        "reservedFields",
        "plannedSchemas",
    }
    missing_keys = sorted(required_keys.difference(manifest))
    if missing_keys:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} is missing required keys: {', '.join(missing_keys)}"
        )
        return

    if manifest["manifestVersion"] != "v1":
        errors.append(f"{manifest_path.relative_to(ROOT)} must declare manifestVersion 'v1'")

    if manifest["family"] != family:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} declares family {manifest['family']!r}, expected {family!r}"
        )

    expected_directory = f"schemas/json/{family}"
    if manifest["directory"] != expected_directory:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} declares directory {manifest['directory']!r}, expected {expected_directory!r}"
        )

    if manifest["filenamePattern"] != SCHEMA_FILENAME_PATTERN:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} must use filenamePattern {SCHEMA_FILENAME_PATTERN!r}"
        )

    supported_versions = manifest["supportedSchemaVersions"]
    if not isinstance(supported_versions, list) or not supported_versions:
        errors.append(f"{manifest_path.relative_to(ROOT)} must declare a non-empty supportedSchemaVersions list")
        return

    unsupported_tokens = [token for token in supported_versions if not re.fullmatch(r"v\d+", str(token))]
    if unsupported_tokens:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} has invalid supported schema version token(s): {', '.join(map(str, unsupported_tokens))}"
        )

    reserved_fields = manifest["reservedFields"]
    if not isinstance(reserved_fields, list) or not reserved_fields:
        errors.append(f"{manifest_path.relative_to(ROOT)} must declare a non-empty reservedFields list")
        return

    missing_reserved = sorted(spec["required_reserved_fields"].difference(reserved_fields))
    if missing_reserved:
        errors.append(
            f"{manifest_path.relative_to(ROOT)} is missing reserved field(s): {', '.join(missing_reserved)}"
        )

    planned_schemas = manifest["plannedSchemas"]
    if not isinstance(planned_schemas, list) or not planned_schemas:
        errors.append(f"{manifest_path.relative_to(ROOT)} must declare a non-empty plannedSchemas list")
        return

    for schema_name in planned_schemas:
        match = SCHEMA_REGEX.fullmatch(str(schema_name))
        if not match:
            errors.append(
                f"{manifest_path.relative_to(ROOT)} has invalid planned schema filename {schema_name!r}"
            )
            continue

        version_token = match.group(1)
        if version_token not in supported_versions:
            errors.append(
                f"{manifest_path.relative_to(ROOT)} references unsupported schema version {version_token!r} in {schema_name!r}"
            )

    implemented_schemas = spec.get("implemented_schemas", {})
    for schema_name, schema_spec in implemented_schemas.items():
        schema_path = ROOT / expected_directory / schema_name
        if not schema_path.exists():
            errors.append(f"Missing concrete schema file: {schema_path.relative_to(ROOT)}")
            continue
        validate_schema_file(schema_path, schema_spec["required_fields"], errors)


def validate_fixture_schema_refs(errors: list[str]) -> None:
    if FIXTURE_MANIFEST_PATH.exists():
        fixture_manifest = load_json(FIXTURE_MANIFEST_PATH)
        for fixture_entry in fixture_manifest.get("fixtures", []):
            path_value = fixture_entry.get("path")
            if not path_value:
                continue
            fixture_path = ROOT / path_value
            if not fixture_path.exists():
                continue
            fixture = load_json(fixture_path)
            family = fixture.get("family")
            target_schema = fixture.get("targetSchema")
            if not family or not target_schema:
                continue
            schema_path = ROOT / f"schemas/json/{family}/{target_schema}"
            if not schema_path.exists():
                errors.append(
                    f"{fixture_path.relative_to(ROOT)} references missing schema {schema_path.relative_to(ROOT)}"
                )

    if REPLAY_MANIFEST_PATH.exists():
        replay_manifest = load_json(REPLAY_MANIFEST_PATH)
        for seed_entry in replay_manifest.get("seeds", []):
            path_value = seed_entry.get("path")
            if not path_value:
                continue
            seed_path = ROOT / path_value
            if not seed_path.exists():
                continue
            seed = load_json(seed_path)
            target_schema = seed.get("targetSchema")
            if not target_schema:
                errors.append(f"{seed_path.relative_to(ROOT)} is missing targetSchema")
                continue
            schema_path = ROOT / f"schemas/json/replay/{target_schema}"
            if not schema_path.exists():
                errors.append(
                    f"{seed_path.relative_to(ROOT)} references missing schema {schema_path.relative_to(ROOT)}"
                )


def main() -> int:
    errors: list[str] = []

    validate_docs(errors)
    for family, spec in EXPECTED_FAMILIES.items():
        validate_manifest(family, spec, errors)
    validate_fixture_schema_refs(errors)

    if errors:
        for error in errors:
            print(f"ERROR: {error}")
        return 1

    print("Contract family manifests and docs validate successfully.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
