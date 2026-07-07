.PHONY: install update lint test dev build clean migrate migrate-create deploy ghcr-clean next

install:
	bun install

update:
	bun upgrade
	bun update -g
	bun update

lint: install
	bunx turbo run lint
	bunx markdownlint-cli2 "**/*.md" "#node_modules" "#apps/web/node_modules" "#apps/web/dist" "#apps/web/coverage" "#LICENSE.md" "#.github/instructions"

test: lint
	rm -rf coverage
	bunx turbo run test
	bun tools/merge-coverage.mjs

dev: install
	bunx turbo run dev

build: install
	rm -Rf dist
	bunx turbo run build

clean:
	rm -Rf node_modules
	rm -Rf dist
	rm -Rf coverage
	rm -Rf apps/web/node_modules
	rm -Rf apps/web/dist
	rm -Rf apps/web/coverage

deploy:
	bash -c 'set -a && source .env && set +a && \
	  cmd="$(or $(cmd),deploy)"; \
	  kamal "$$cmd"; status=$$?; \
	  if [ $$status -eq 0 ] && { [ "$$cmd" = "deploy" ] || [ "$$cmd" = "setup" ]; }; then \
	    $(MAKE) ghcr-clean; \
	  fi; \
	  exit $$status'

# Prune old ghcr.io/checkmeup/checkmeup image versions, keeping the last 5.
# Runs automatically after `make deploy` / `make deploy cmd=setup`; call
# directly (make ghcr-clean) to prune on demand, or `make ghcr-clean keep=10`
# to keep a different number.
ghcr-clean:
	bash -c 'set -a && source .env && set +a && bash scripts/ghcr-clean.sh $(or $(keep),5)'

# Load DATABASE_URL from apps/api/.env and run goose
DB_URL := $(shell grep '^DATABASE_URL=' apps/api/.env 2>/dev/null | cut -d= -f2-)

migrate:
	goose -dir apps/api/migrations postgres "$(DB_URL)" up

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=your_migration_name" && exit 1)
	goose -dir apps/api/migrations create $(name) sql

merge:
	git pull origin main
	git checkout main
	git merge next
	git push -u origin main