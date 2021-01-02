PKG := "github.com/labstack/echo"
PKG_LIST := $(shell go list ${PKG}/...)

tag:
	@git tag `grep -P '^\tversion = ' echo.go|cut -f2 -d'"'`
	@git tag|grep -v ^v

.DEFAULT_GOAL := check
check: lint vet race ## Check project

init:
	@go get -u golang.org/x/lint/golint

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run tests
	@go test -short ${PKG_LIST}

race: ## Run tests with data race detector
	@go test -race ${PKG_LIST}

benchmark: ## Run benchmarks
	@go test -run="-" -bench=".*" ${PKG_LIST}

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

goversion ?= "1.12"
test_version: ## Run tests inside Docker with given version (defaults to 1.12 oldest supported). Example: make test_version goversion=1.13
	@docker run --rm -it -v $(shell pwd):/project golang:$(goversion) /bin/sh -c "cd /project && make check"
