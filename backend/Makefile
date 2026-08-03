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
	$(GO) generate ./ent
	$(GO) run github.com/swaggo/swag/cmd/swag@v1.16.6 init --generalInfo cmd/server/swagger.go --dir . --output openapi/swagger --outputTypes yaml --parseInternal
	pnpm --package=swagger2openapi@7.0.8 dlx swagger2openapi openapi/swagger/swagger.yaml --patch --yaml --outfile openapi/openapi.yaml
	cd ../frontend && npx openapi-typescript@7.13.0 ../backend/openapi/openapi.yaml --output src/api/generated/schema.ts

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
