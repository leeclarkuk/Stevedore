VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/leeclarkuk/stevedore/internal/cli.version=$(VERSION)

.PHONY: build test golden lint e2e

build: ## Build ./bin/stevedore
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/stevedore ./cmd/stevedore

test: ## Unit and golden tests
	go test ./...

golden: ## Regenerate testdata/golden from examples/
	go test ./internal/convert -run TestGolden -update
	@git --no-pager diff --stat -- testdata/golden

lint: ## golangci-lint
	golangci-lint run

e2e: build ## Convert examples and apply to a kind cluster
	./scripts/e2e.sh
