.PHONY: build clean test generate lint help

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/rfancn/prism/cmd.Version=$(VERSION) \
	-X github.com/rfancn/prism/cmd.GitCommit=$(GIT_COMMIT) \
	-X github.com/rfancn/prism/cmd.BuildDate=$(BUILD_DATE)"

# Default target
all: build

## build: Build the binary
build:
	go build $(LDFLAGS) -o prism .

## clean: Clean build artifacts
clean:
	rm -f prism
	go clean

## test: Run tests
test:
	go test ./... -v

## generate: Generate sqlc code
generate:
	go generate ./...

## lint: Run linters
lint:
	go vet ./...

## sqlc: Run sqlc generate
sqlc:
	sqlc generate -f assets/sqlc.yaml

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'