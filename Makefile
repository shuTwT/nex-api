GO ?= go
NPM ?= npm
BIN_DIR ?= bin
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
GOAMD64 ?= v1
GOARM64 ?= v8.0

.PHONY: doctor test-bootstrap generate openapi-lint build build-backend build-linux-amd64 build-current frontend-build test clean frontend-install frontend-typecheck

doctor:
	@command -v go >/dev/null || { printf '%s\n' 'go is required'; exit 1; }
	@printf 'Go: %s\n' "$$(go env GOVERSION)"
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
	go mod tidy -go=1.26
	go build ./cmd/server
	go build ./cmd/script-worker

generate:
	go generate ./ent
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --generalInfo cmd/server/swagger.go --dir . --output openapi/swagger --outputTypes yaml --parseInternal
	pnpm --package=swagger2openapi@7.0.8 dlx swagger2openapi openapi/swagger/swagger.yaml --patch --yaml --outfile openapi/openapi.yaml
	cd frontend && npx openapi-typescript@7.13.0 ../openapi/openapi.yaml --output src/api/generated/schema.ts
	go run ./tools/genoperations --in openapi/openapi.yaml --out frontend/src/api/generated/operations.ts

openapi-lint:
	pnpm --package=@redocly/cli@1.34.2 dlx redocly lint --config=openapi/redocly.yaml openapi/openapi.yaml

frontend-install:
	npm --prefix frontend ci

frontend-typecheck:
	npm --prefix frontend run typecheck

build:
	$(MAKE) build-backend
	$(MAKE) frontend-build

build-backend:
	@set -eu; \
	for platform in $(PLATFORMS); do \
		case "$$platform" in \
			*/*) ;; \
			*) printf 'invalid platform %s; expected GOOS/GOARCH\n' "$$platform" >&2; exit 1 ;; \
		esac; \
		os="$${platform%/*}"; \
		arch="$${platform#*/}"; \
		output_dir="$(BIN_DIR)/$${os}_$${arch}"; \
		mkdir -p "$$output_dir"; \
		printf 'Building server and script-worker for %s/%s\n' "$$os" "$$arch"; \
		CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" GOAMD64="$(GOAMD64)" GOARM64="$(GOARM64)" \
			$(GO) build -trimpath -o "$$output_dir/server" ./cmd/server; \
		CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" GOAMD64="$(GOAMD64)" GOARM64="$(GOARM64)" \
			$(GO) build -trimpath -o "$$output_dir/script-worker" ./cmd/script-worker; \
	done

build-linux-amd64:
	$(MAKE) build-backend PLATFORMS=linux/amd64

build-current:
	$(MAKE) build-backend PLATFORMS="$$($(GO) env GOOS)/$$($(GO) env GOARCH)"
	$(MAKE) frontend-build

frontend-build:
	$(NPM) --prefix frontend run build

test:
	go test -race -shuffle=on -count=1 ./...
	npm --prefix frontend run typecheck

clean:
	rm -rf bin
