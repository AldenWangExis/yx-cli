VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/AldenWangExis/yx-cli/internal/cli.Version=$(VERSION) -X github.com/AldenWangExis/yx-cli/internal/cli.Commit=$(COMMIT) -X github.com/AldenWangExis/yx-cli/internal/cli.Date=$(DATE)

.PHONY: test build

test:
	go test ./...

build:
	go build -ldflags "$(LDFLAGS)" -o yx ./cmd/yx
