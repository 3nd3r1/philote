.PHONY: check lint fmt test ideation index new setup prototypes

check: lint test ideation

lint:
	@command -v golangci-lint >/dev/null || { echo "golangci-lint not found; run inside nix-shell"; exit 1; }
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

test:
	go test ./...

ideation:
	python3 tools/ideation.py check

index:
	python3 tools/ideation.py index

new:
	python3 tools/ideation.py new "$(KIND)" "$(TITLE)"

setup:
	git config core.hooksPath .githooks

prototypes:
	cd prototypes/frontend-v0 && go test ./...
