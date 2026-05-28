.PHONY: build test lint clean run install test-report coverage-all coverage-check bench-compare

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
	@echo ""
	@echo "=== Coverage Summary ==="
	@echo "Line:    $$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}')"
	@echo "Func:    $$(awk '!/total:/ && /[0-9]+\.[0-9]+%$$/ { total++; if ($$NF == "0.0%") uncovered++ } END { if (total > 0) printf "%.1f%% (%d/%d covered)\n", (total-uncovered)/total*100, total-uncovered, total; else print "N/A" }' coverage-func.txt)"
	@echo ""

coverage-check: test-cover
	@line=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	func=$$(awk '!/total:/ && /[0-9]+\.[0-9]+%$$/ { total++; if ($$NF == "0.0%") uncovered++ } END { if (total > 0) printf "%.0f", (total-uncovered)/total*100; else print "0" }' coverage-func.txt); \
	echo ""; \
	echo "=== Coverage Gates ==="; \
	fail=0; \
	if (( $$(echo "$$line < 80" | bc -l) )); then \
		echo "❌ Line coverage: $${line}% (gate: 80%)"; fail=1; \
	else \
		echo "✅ Line coverage: $${line}% (gate: 80%)"; \
	fi; \
	if [ "$$func" -lt 70 ]; then \
		echo "❌ Function coverage: $${func}% (gate: 70%)"; fail=1; \
	else \
		echo "✅ Function coverage: $${func}% (gate: 70%)"; \
	fi; \
	if [ $$fail -eq 1 ]; then echo ""; echo "Coverage gates failed."; exit 1; fi; \
	echo ""; echo "All coverage gates passed."; \
	echo "Reports: coverage.html coverage-func.txt"

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

# Fetch the latest betterleaks rules from upstream.
# Source: https://raw.githubusercontent.com/ailinter/ailinter/main/internal/secrets/betterleaks.toml
update-betterleaks:
	curl -sSL -o internal/secrets/betterleaks.toml \
		https://raw.githubusercontent.com/betterleaks/betterleaks/main/betterleaks.toml
	@echo "Fetched betterleaks.toml ($(shell wc -l < internal/secrets/betterleaks.toml) lines)"


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
