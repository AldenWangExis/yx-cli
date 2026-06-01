.PHONY: test build

test:
	go test ./...

build:
	go build -o yx ./cmd/yx
