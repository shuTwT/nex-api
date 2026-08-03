.PHONY: doctor test-bootstrap generate build test frontend-install frontend-typecheck

doctor:
	@$(MAKE) -C backend doctor
	@set -eu; \
	command -v node >/dev/null || { printf '%s\n' 'node is required'; exit 1; }; \
	command -v npm >/dev/null || { printf '%s\n' 'npm is required'; exit 1; }; \
	node_version="$$(node --version)"; \
	node_version="$${node_version#v}"; \
	node_major="$${node_version%%.*}"; \
	if [ "$$node_major" -lt 22 ]; then \
		printf 'Node 22+ is required; found v%s\n' "$$node_version"; \
		exit 1; \
	fi; \
	printf 'Node: v%s\n' "$$node_version"; \
	printf 'npm: %s\n' "$$(npm --version)"

test-bootstrap:
	$(MAKE) -C backend test-bootstrap

generate:
	$(MAKE) -C backend generate

frontend-install:
	npm --prefix frontend ci

frontend-typecheck:
	npm --prefix frontend run typecheck

build:
	$(MAKE) -C backend build
	npm --prefix frontend run build

test:
	$(MAKE) -C backend test
	npm --prefix frontend run typecheck
