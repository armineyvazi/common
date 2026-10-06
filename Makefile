.PHONY: test test-race lint build tidy tag release help

## Run unit tests
test:
	go test -count=1 -timeout 60s ./...

## Run unit tests with race detector
test-race:
	go test -race -count=1 -timeout 120s ./...

## Run integration tests (requires live services)
test-integration:
	go test -tags integration -count=1 -timeout 120s ./...

## Run golangci-lint
lint:
	golangci-lint run --timeout=5m

## Build all packages
build:
	go build ./...

## Tidy and verify go.mod
tidy:
	go mod tidy
	go mod verify

## Create and push a release tag.
## Usage: make tag VERSION=v0.2.0
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v<major>.<minor>.<patch>"; exit 1; \
	fi
	@echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$$' || \
		(echo "VERSION must be vMAJOR.MINOR.PATCH[-pre]"; exit 1)
	@git diff --quiet || (echo "Working tree is dirty — commit or stash first"; exit 1)
	@git diff --cached --quiet || (echo "Staged changes exist — commit or stash first"; exit 1)
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Tag $(VERSION) pushed. The release workflow will run on GitHub Actions."

## Show this help
help:
	@grep -E '^## ' Makefile | sed 's/^## //'
