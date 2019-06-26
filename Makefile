SHELL=/bin/bash -o pipefail

GO ?= go
GOLINT ?= golangci-lint

COMMIT_HASH := $(shell git rev-parse --short HEAD 2> /dev/null || true)
GIT_TAG := $(shell git describe --tags --dirty --always)

LDFLAGS := -ldflags '-X main.commit=${COMMIT_HASH} -X main.date=$(shell date +%s) -X main.version=${GIT_TAG}'
TESTPACKAGES := $(shell go list ./... | grep -v /constants | grep -v /cmd/)

kubectl_free ?= _output/kubectl-free

.PHONY: build
build: clean ${kubectl_free}

${kubectl_free}:
	GO111MODULE=on CGO_ENABLED=0 $(GO) build ${LDFLAGS} -o $@ ./cmd/kubectl-free/root.go

.PHONY: clean
clean:
	rm -Rf _output

.PHONY: test
test:
	GO111MODULE=on $(GO) test -count=1 -v -race $(TESTPACKAGES)

.PHONY: lint-install
lint-install:
	GO111MODULE=on ${GO} install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: 
	GO111MODULE=on ${GOLINT} run -E stylecheck -E gocritic

.PHONY: fmt
fmt: 
	${GO} fmt ./cmd/... ./pkg/...
