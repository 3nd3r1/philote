.PHONY: check index new test

check:
	python3 tools/ideation.py check

index:
	python3 tools/ideation.py index

new:
	python3 tools/ideation.py new "$(KIND)" "$(TITLE)"

test: check
	cd prototypes/frontend-v0 && go test ./...
