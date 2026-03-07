from __future__ import annotations

from pathlib import Path

from libs.python.contracts import load_json, validate_fixture, validate_replay_seed


ROOT = Path(__file__).resolve().parents[2]


def test_python_parity_validates_fixture_corpus() -> None:
    manifest = load_json(ROOT / "tests/fixtures/manifest.v1.json")
    for entry in manifest["fixtures"]:
        fixture = load_json(ROOT / entry["path"])
        validate_fixture(fixture)


def test_python_parity_validates_replay_seed_catalog() -> None:
    manifest = load_json(ROOT / "tests/replay/manifest.v1.json")
    for entry in manifest["seeds"]:
        seed = load_json(ROOT / entry["path"])
        validate_replay_seed(seed)
