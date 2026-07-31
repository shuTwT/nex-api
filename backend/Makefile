GO := go
SERVER_BINARY := bin/server
WORKER_BINARY := bin/script-worker

.PHONY: doctor test-bootstrap generate openapi-lint build test clean

doctor:
	@command -v $(GO) >/dev/null || { printf '%s\n' 'go is required'; exit 1; }
	@printf 'Go: %s\n' "$$($(GO) env GOVERSION)"

test-bootstrap:
	$(GO) mod tidy -go=1.26
	$(GO) build ./cmd/server
	$(GO) build ./cmd/script-worker

generate:
	$(GO) generate ./internal/httpapi

openapi-lint:
	pnpm --package=@redocly/cli@1.34.2 dlx redocly lint --config=openapi/redocly.yaml openapi/openapi.yaml

build:
	mkdir -p bin
	$(GO) build -trimpath -o $(SERVER_BINARY) ./cmd/server
	$(GO) build -trimpath -o $(WORKER_BINARY) ./cmd/script-worker

test:
	$(GO) test -race -shuffle=on -count=1 ./...

clean:
	rm -rf bin
