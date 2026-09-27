.PHONY: fmt lint lint-fix lint-config test coverage coverage-check build vet nodevault-wire-sync-check

# Snapshot of NodeVault's certified-tools wire definitions at the NodeVault commit the
# golden JSON was produced from (see pkg/paletteclient/testdata).
NODEVAULT_WIRE_SNAPSHOT ?= pkg/paletteclient/testdata/nodevault_f45d4a11_catalogrest_wire.go.txt

LOCALBIN ?= $(CURDIR)/bin
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.11.3
COVERAGE_THRESHOLD ?= 70.0

fmt:
	go fmt ./...

$(GOLANGCI_LINT):
	@mkdir -p "$(LOCALBIN)"
	@test -x "$(GOLANGCI_LINT)" || bash -c '\
		set -euo pipefail; \
		OS="$$(uname | tr A-Z a-z)"; \
		ARCH="$$(uname -m)"; \
		case "$$ARCH" in x86_64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) echo "unsupported arch: $$ARCH"; exit 1 ;; esac; \
		VER="$(GOLANGCI_LINT_VERSION)"; \
		VER="$${VER#v}"; \
		FILE="golangci-lint-$$VER-$$OS-$$ARCH.tar.gz"; \
		URL="https://github.com/golangci/golangci-lint/releases/download/$(GOLANGCI_LINT_VERSION)/$$FILE"; \
		SUM_URL="https://github.com/golangci/golangci-lint/releases/download/$(GOLANGCI_LINT_VERSION)/golangci-lint-$$VER-checksums.txt"; \
		TMP="$$(mktemp -d)"; \
		curl -fsSL "$$URL" -o "$$TMP/lint.tgz"; \
		curl -fsSL "$$SUM_URL" -o "$$TMP/checksums.txt"; \
		EXPECTED="$$(awk -v f="$$FILE" "\$$2==f{print \$$1}" "$$TMP/checksums.txt")"; \
		ACTUAL="$$(sha256sum "$$TMP/lint.tgz" | awk "{print \$$1}")"; \
		if [ -z "$$EXPECTED" ] || [ "$$EXPECTED" != "$$ACTUAL" ]; then echo "checksum mismatch for $$FILE"; exit 1; fi; \
		tar -xzf "$$TMP/lint.tgz" -C "$$TMP"; \
		cp "$$TMP/golangci-lint-$$VER-$$OS-$$ARCH/golangci-lint" "$(GOLANGCI_LINT)"; \
		chmod +x "$(GOLANGCI_LINT)"; \
		rm -rf "$$TMP"'

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --config=.golangci.yml ./...

lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --config=.golangci.yml --fix ./...

lint-config: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) config verify --config=.golangci.yml

build:
	go build ./...

vet:
	go vet ./...

test:
	go test -v -race ./...

coverage:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

coverage-check: coverage
	@coverage="$$(go tool cover -func=coverage.out | awk '/^total:/ { gsub(/%/, "", $$3); print $$3 }')"; \
	awk -v got="$$coverage" -v want="$(COVERAGE_THRESHOLD)" 'BEGIN { \
		if (got + 0 < want + 0) { \
			printf("coverage %.1f%% below required %.1f%%\n", got, want); exit 1 \
		} \
		printf("coverage %.1f%% >= %.1f%%\n", got, want) \
	}'

# Fails when NodeVault's certified-tools wire (the owner) no longer matches the
# vendored snapshot; regenerate the golden and the snapshot from NodeVault together.
nodevault-wire-sync-check:
	@test -n "$(NODEVAULT_REPO_ROOT)" || { echo "ERROR: NODEVAULT_REPO_ROOT is required; set it to a NodeVault checkout (e.g. make $@ NODEVAULT_REPO_ROOT=/path/to/NodeVault)"; exit 1; }
	bash hack/nodevault-wire-extract.sh "$(NODEVAULT_REPO_ROOT)/pkg/catalogrest/server.go" | diff -u "$(NODEVAULT_WIRE_SNAPSHOT)" -
