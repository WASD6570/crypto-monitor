.DEFAULT_GOAL := help

.PHONY: help repo-check compose-config compose-dev-config compose-smoke compose-dev-smoke contracts-validate fixtures-validate replay-smoke

help:
	@printf "Available targets:\n"
	@printf "  make help            Show available commands\n"
	@printf "  make repo-check      Verify the scaffolded repo layout\n"
	@printf "  make compose-config  Validate docker-compose.yml\n"
	@printf "  make compose-dev-config  Validate dev compose overlay\n"
	@printf "  make compose-smoke   Run compose rollout smoke proof\n"
	@printf "  make compose-dev-smoke   Run dev overlay smoke proof\n"
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

compose-dev-config:
	@docker compose -f docker-compose.yml -f docker-compose.dev.yml config >/dev/null
	@printf "dev compose overlay is valid.\n"

compose-smoke:
	@./scripts/dev/compose_smoke.sh

compose-dev-smoke:
	@./scripts/dev/compose_dev_smoke.sh

contracts-validate:
	@python3 scripts/dev/validate_contract_families.py

fixtures-validate:
	@python3 scripts/dev/validate_fixture_corpus.py

replay-smoke:
	@python3 scripts/dev/replay_smoke.py
