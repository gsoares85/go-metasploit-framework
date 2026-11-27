# Developer convenience targets. CI runs the authoritative checks.
.PHONY: all build test race vet fmt fmtcheck cover vuln lint check tidy run

all: check

build:
	go build ./...

test:
	go test ./...

race:
	go test ./... -race

vet:
	go vet ./...

fmt:
	gofmt -w .

fmtcheck:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then echo "Not gofmt-formatted:"; echo "$$unformatted"; exit 1; fi

cover:
	go test ./... -covermode=atomic -coverprofile=coverage.out
	go tool cover -func=coverage.out

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

lint:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

tidy:
	go mod tidy

# Full local gate — run before opening a PR.
check: fmtcheck vet build race

run:
	go run ./cmd
