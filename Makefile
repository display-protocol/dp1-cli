# dp1-cli — local development helpers
# Lint: golangci-lint v2 on PATH (see .golangci.yml; same family as CI)
# Markdown: npx markdownlint-cli2 (same as .github/workflows/lint.yaml)

GOLANGCI_LINT ?= golangci-lint
BIN            ?= dp1

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make build         - build CLI binary to ./$(BIN)"
	@echo "  make install       - go install (uses module path; binary name may match repo dir)"
	@echo "  make test          - go test ./... -race -count=1"
	@echo "  make test-coverage - tests with coverage.out (atomic mode)"
	@echo "  make lint          - golangci-lint + markdownlint-cli2 on *.md"
	@echo "  make lint-fix      - golangci-lint run --fix"
	@echo "  make fmt           - go fmt ./..."
	@echo "  make fmt-imports   - golangci-lint fmt (goimports per .golangci.yml)"
	@echo "  make vet           - go vet ./..."
	@echo "  make tidy          - go mod tidy"
	@echo "  make verify        - go mod verify"
	@echo "  make check         - lint + test (local gate similar to CI)"
	@echo "  make clean         - remove ./$(BIN)"
	@echo "  make gitleaks      - gitleaks detect (if gitleaks is on PATH)"

.PHONY: build
build:
	go build -o $(BIN) .

.PHONY: install
install:
	go install .

.PHONY: test
test:
	go test ./... -race -count=1 -timeout=5m

.PHONY: test-coverage
test-coverage:
	go test ./... -race -count=1 -timeout=5m -coverprofile=coverage.out -covermode=atomic
	@echo ""
	go tool cover -func=coverage.out | tail -25

.PHONY: lint
lint:
	@$(GOLANGCI_LINT) version >/dev/null 2>&1 || { echo "golangci-lint not found; install: https://golangci-lint.run/welcome/install/"; exit 1; }
	$(GOLANGCI_LINT) run ./...
	@command -v npx >/dev/null 2>&1 || { echo "npx not found; install Node.js for markdown lint (https://nodejs.org/)"; exit 1; }
	npx --yes markdownlint-cli2 "**/*.md" "#node_modules"

.PHONY: lint-fix
lint-fix:
	@$(GOLANGCI_LINT) version >/dev/null 2>&1 || { echo "golangci-lint not found; install: https://golangci-lint.run/welcome/install/"; exit 1; }
	$(GOLANGCI_LINT) run ./... --fix

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: fmt-imports
fmt-imports:
	@$(GOLANGCI_LINT) version >/dev/null 2>&1 || { echo "golangci-lint not found; install: https://golangci-lint.run/welcome/install/"; exit 1; }
	$(GOLANGCI_LINT) fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: verify
verify:
	go mod verify

.PHONY: check
check: lint test

.PHONY: clean
clean:
	rm -f $(BIN)

.PHONY: gitleaks
gitleaks:
	@command -v gitleaks >/dev/null 2>&1 || { echo "gitleaks not found; see https://github.com/gitleaks/gitleaks or CI workflow for install"; exit 1; }
	gitleaks detect --source . --redact --verbose
