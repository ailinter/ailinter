.PHONY: build test lint clean run install test-report coverage-all bench-compare

NAME := ailinter
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(NAME) ./cmd/ailinter

run: build
	./bin/$(NAME)

test:
	go test ./... -v -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=count
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out > coverage-func.txt
	@echo "Coverage: $$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}')"

test-report: test-cover
	go test ./... -json -count=1 > test-report.json
	go run ./scripts/test-report/main.go test-report.json > test-report.html
	@echo "Reports: coverage.html coverage-func.txt test-report.html"

coverage-all: test-report
	@echo "All reports generated."

lint:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

install:
	go install $(LDFLAGS) ./cmd/ailinter

clean:
	rm -rf bin/ coverage.out coverage.html coverage-func.txt test-report.json test-report.html benchmarks/data/ benchmarks/repos/

bench:
	go test ./internal/... -bench=. -benchmem

bench-compare:
	go run ./benchmarks/runner/main.go

# Build for all platforms
release:
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(NAME)-darwin-amd64   ./cmd/ailinter
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(NAME)-darwin-arm64   ./cmd/ailinter
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(NAME)-linux-amd64    ./cmd/ailinter
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/$(NAME)-linux-arm64    ./cmd/ailinter
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(NAME)-windows-amd64.exe ./cmd/ailinter
