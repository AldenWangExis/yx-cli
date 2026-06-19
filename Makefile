VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/AldenWangExis/yx-cli/internal/cli.Version=$(VERSION) -X github.com/AldenWangExis/yx-cli/internal/cli.Commit=$(COMMIT) -X github.com/AldenWangExis/yx-cli/internal/cli.Date=$(DATE)

.PHONY: test build test-install release-check

test:
	sh scripts/test_install.sh
	go test ./...
	cd npm/yx-cli && npm test && npm pack --dry-run

build:
	go build -ldflags "$(LDFLAGS)" -o yx ./cmd/yx

test-install:
	sh scripts/test_install.sh

release-check:
	sh scripts/check_release_version.sh "$(VERSION)"
	sh scripts/test_install.sh
	go test ./...
	cd npm/yx-cli && npm test && npm pack --dry-run
