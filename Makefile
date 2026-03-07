.DEFAULT_GOAL := help

.PHONY: help repo-check compose-config contracts-validate fixtures-validate replay-smoke

help:
	@printf "Available targets:\n"
	@printf "  make help            Show available commands\n"
	@printf "  make repo-check      Verify the scaffolded repo layout\n"
	@printf "  make compose-config  Validate docker-compose.yml\n"
	@printf "  make contracts-validate  Validate contract family manifests\n"
	@printf "  make fixtures-validate   Validate fixture corpus and replay seeds\n"
	@printf "  make replay-smoke        Run deterministic replay smoke checks\n"

repo-check:
	@test -d apps/web
	@test -d apps/research
	@test -d services
	@test -d libs
	@test -d schemas
	@test -d tests
	@printf "Repo scaffold looks good.\n"

compose-config:
	@docker compose config >/dev/null
	@printf "docker-compose.yml is valid.\n"

contracts-validate:
	@python3 scripts/dev/validate_contract_families.py

fixtures-validate:
	@python3 scripts/dev/validate_fixture_corpus.py

replay-smoke:
	@python3 scripts/dev/replay_smoke.py
