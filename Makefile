.PHONY: install update lint test dev build next merge

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