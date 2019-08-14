SHELL := /bin/bash

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

SRC = main.go pkg

# These will be provided to the target
VERSION := 1.0.0
BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

GOIMPORTS := $(shell which goimports)
GOLINTER := $(shell which golangci-lint)

.PHONY: help all build clean install uninstall fmt simplify check run

all: vet test install

help: ## print Makefile targets doc's
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET) ## build partybox binary
	@true

test: lint ## run unit tests
	@go test $$(go list ./...)

clean: ## remove partybox binary
	@rm -f $(TARGET)

install: ## install partybox
	@go install $(LDFLAGS)

uninstall: clean ## unistall partybox
	@rm -f $$(which ${TARGET})

vet: ## vet partybox
	go vet $(SRC)

fmt: ## format partybox
	@gofmt -l -w $(SRC)

simplify: ## auot-fix format/import and lint issues whenever possible
	# fix formatting issues
	@gofmt -s -l -w $(SRC)
	# fix linting issues (whenever possible)
	@$(GOLINTER) run --fix

imports: $(GOIMPORTS) ## check partybox formatting or die
	@goimports_out=$$($(GOIMPORTS) -d -e $(SRC) 2>&1) && [ -z "$${goimports_out}" ] || (echo "$${goimports_out}" 1>&2 && false)

lint: $(GOLINTER) ## run linter on partybox
	@$(GOLINTER) run

run: install ## run partybox app
	@$(TARGET) < sample_csv.txt

$(GOIMPORTS):
	@go get -u golang.org/x/tools/cmd/goimports

$(GOLINTER):
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint