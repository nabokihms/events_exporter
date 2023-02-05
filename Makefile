SHELL=/bin/bash -o pipefail
$( shell mkdir -p bin )

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GOLANGCI_VERSION = 1.51.0
HELM_DOCS_VERSION = 1.11.0

ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif

COMMIT=$(shell git rev-parse --verify HEAD)

###########
# BUILDING
###########
bin/events_exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -mod=readonly -o bin/events_exporter

build: bin/events_exporter

###########
# LINTING
###########
bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint

bin/golangci-lint-${GOLANGCI_VERSION}:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

bin/helm-docs: bin/helm-docs-${HELM_DOCS_VERSION}
	@ln -sf helm-docs-${HELM_DOCS_VERSION} bin/helm-docs
bin/helm-docs-${HELM_DOCS_VERSION}:
	@mkdir -p bin
	curl -L https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION}/helm-docs_${HELM_DOCS_VERSION}_$(shell uname)_x86_64.tar.gz | tar -zOxf - helm-docs > ./bin/helm-docs-${HELM_DOCS_VERSION} && chmod +x ./bin/helm-docs-${HELM_DOCS_VERSION}

.PHONY: lint fix
lint: bin/golangci-lint
	bin/golangci-lint run

fix: bin/golangci-lint
	bin/golangci-lint run --fix

.PHONY: docs
docs: bin/helm-docs
	bin/helm-docs -s file -c charts/ -t ../docs/templates/overrides.gotmpl -t README.md.gotmpl

###########
# TESTING
###########
test:
	go test -race -cover -v ./...
