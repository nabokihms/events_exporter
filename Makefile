SHELL=/bin/bash -o pipefail
$( shell mkdir -p bin )

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GOLANGCI_VERSION = 1.40.1
ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif

COMMIT=$(shell git rev-parse --verify HEAD)

###########
# BUILDING
###########
.PHONY: build
bin/events_exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -mod=readonly -o bin/events_exporter

build: bin/events_exporter

###########
# LINTING
###########
bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint

bin/golangci-lint-${GOLANGCI_VERSION}:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint fix
lint: bin/golangci-lint
	bin/golangci-lint run

fix: bin/golangci-lint
	bin/golangci-lint run --fix

###########
# TESTING
###########
