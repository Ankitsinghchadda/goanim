# Top-level convenience targets. Wraps `go ...` so contributors don't
# have to remember the flags.

GO         ?= go
PKG_LIB    := ./core/... ./mobjects/... ./internal/...
PKG_ALL    := ./...

.PHONY: all
all: fmt vet test build

.PHONY: build
build:
	$(GO) build $(PKG_ALL)

.PHONY: test
test:
	$(GO) test -race -count=1 $(PKG_LIB)

.PHONY: test-all
test-all:
	$(GO) test -race -count=1 $(PKG_ALL)

.PHONY: vet
vet:
	$(GO) vet $(PKG_ALL)

.PHONY: fmt
fmt:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
	  echo "gofmt diff in:"; echo "$$out"; exit 1; \
	fi

.PHONY: fmt-fix
fmt-fix:
	gofmt -w .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: bench
bench:
	$(GO) run ./cmd/bench --runs 3

.PHONY: bench-compare
bench-compare:
	$(GO) run ./cmd/bench --runs 3 --compare bench_baseline.json

.PHONY: bench-baseline
bench-baseline:
	$(GO) run ./cmd/bench --runs 3 --output bench_baseline.json

.PHONY: cover
cover:
	$(GO) test -coverprofile=coverage.out $(PKG_LIB)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"

.PHONY: clean
clean:
	rm -rf binary/ coverage.out coverage.html *.prof *.pprof
	$(GO) clean

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make build           — go build ./..."
	@echo "  make test            — race-tests for library packages"
	@echo "  make test-all        — race-tests including examples + cmd"
	@echo "  make vet             — go vet ./..."
	@echo "  make fmt             — verify gofmt clean"
	@echo "  make fmt-fix         — gofmt -w ."
	@echo "  make lint            — golangci-lint run"
	@echo "  make tidy            — go mod tidy"
	@echo "  make bench           — full bench suite"
	@echo "  make bench-compare   — bench vs bench_baseline.json"
	@echo "  make bench-baseline  — record new baseline"
	@echo "  make cover           — HTML coverage report"
	@echo "  make clean           — drop binary/ + temp artifacts"
