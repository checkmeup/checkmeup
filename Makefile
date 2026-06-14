.PHONY: install update lint test dev build clean migrate migrate-create deploy

install:
	bun install

update:
	bun upgrade
	bun update -g
	bun update

lint: install
	bunx turbo run lint

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
	set -a && source .env && set +a && kamal $(cmd)

# Load DATABASE_URL from apps/api/.env and run goose
DB_URL := $(shell grep '^DATABASE_URL=' apps/api/.env 2>/dev/null | cut -d= -f2-)

migrate:
	goose -dir apps/api/migrations postgres "$(DB_URL)" up

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=your_migration_name" && exit 1)
	goose -dir apps/api/migrations create $(name) sql