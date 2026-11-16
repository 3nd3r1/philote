.PHONY: check lint fmt test setup smoke

check: lint test

lint:
	@command -v golangci-lint >/dev/null || { echo "golangci-lint not found; run inside nix-shell"; exit 1; }
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

test:
	go test ./...

setup:
	git config core.hooksPath .githooks

smoke:
	sh tools/smoke.sh
