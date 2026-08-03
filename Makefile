.PHONY: doctor test-bootstrap generate build test

doctor:
	@set -eu; \
	command -v go >/dev/null || { printf '%s\n' 'go is required'; exit 1; }; \
	command -v pnpm >/dev/null || { printf '%s\n' 'pnpm is required'; exit 1; }; \
	command -v node >/dev/null || { printf '%s\n' 'node is required'; exit 1; }; \
	go_version="$$(go env GOVERSION)"; \
	go_version="$${go_version#go}"; \
	go_major="$${go_version%%.*}"; \
	go_minor="$${go_version#*.}"; \
	go_minor="$${go_minor%%.*}"; \
	if [ "$$go_major" -lt 1 ] || { [ "$$go_major" -eq 1 ] && [ "$$go_minor" -lt 26 ]; }; then \
		printf 'Go 1.26+ is required; found go%s\n' "$$go_version"; \
		exit 1; \
	fi; \
	pnpm_version="$$(pnpm --version)"; \
	pnpm_major="$${pnpm_version%%.*}"; \
	if [ "$$pnpm_major" -lt 9 ]; then \
		printf 'pnpm 9+ is required; found %s\n' "$$pnpm_version"; \
		exit 1; \
	fi; \
	node_version="$$(node --version)"; \
	node_version="$${node_version#v}"; \
	node_major="$${node_version%%.*}"; \
	if [ "$$node_major" -lt 22 ]; then \
		printf 'Node 22+ is required; found v%s\n' "$$node_version"; \
		exit 1; \
	fi; \
	printf 'Go: go%s\n' "$$go_version"; \
	printf 'pnpm: %s\n' "$$pnpm_version"; \
	printf 'Node: v%s\n' "$$node_version"

test-bootstrap:
	$(MAKE) -C backend test-bootstrap

generate:
	$(MAKE) -C backend generate

build:
	$(MAKE) -C backend build

test:
	$(MAKE) -C backend test
