#!/usr/bin/env gmake

OASIS_RELEASE := 20.11.3
ROSETTA_CLI_RELEASE := 0.4.0

OASIS_GO ?= go
GO := env -u GOPATH $(OASIS_GO)
GOLINT := env -u GOPATH golangci-lint

# Check if we're running in an interactive terminal.
ISATTY := $(shell [ -t 0 ] && echo 1)

ifdef ISATTY
# Running in interactive terminal, OK to use colors!
MAGENTA = \e[35;1m
CYAN = \e[36;1m
OFF = \e[0m
else
# Don't use colors if not running interactively.
MAGENTA = ""
CYAN = ""
OFF = ""
endif

# Check which tool to use for downloading.
HAVE_WGET := $(shell which wget > /dev/null && echo 1)
ifdef HAVE_WGET
DOWNLOAD := wget --quiet --show-progress --progress=bar:force:noscroll -O
else
HAVE_CURL := $(shell which curl > /dev/null && echo 1)
ifdef HAVE_CURL
DOWNLOAD := curl --progress-bar --location -o
else
$(error Please install wget or curl)
endif
endif

ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: all build clean fmt lint nuke test

all: build
	@printf "$(CYAN)*** Everything built successfully!$(OFF)\n"

build:
	@printf "$(CYAN)*** Building...$(OFF)\n"
	@$(GO) build

tests/oasis_core_release.tar.gz:
	@printf "$(MAGENTA)*** Downloading oasis-core release $(OASIS_RELEASE)...$(OFF)\n"
	@$(DOWNLOAD) $@ https://github.com/oasisprotocol/oasis-core/releases/download/v$(OASIS_RELEASE)/oasis_core_$(OASIS_RELEASE)_linux_amd64.tar.gz

tests/oasis-net-runner: tests/oasis_core_release.tar.gz
	@printf "$(MAGENTA)*** Unpacking oasis-net-runner...$(OFF)\n"
	@tar -xf $< -C tests --strip-components=1 oasis_core_$(OASIS_RELEASE)_linux_amd64/oasis-net-runner

tests/oasis-node: tests/oasis_core_release.tar.gz
	@printf "$(MAGENTA)*** Unpacking oasis-node...$(OFF)\n"
	@tar -xf $< -C tests --strip-components=1 oasis_core_$(OASIS_RELEASE)_linux_amd64/oasis-node

tests/rosetta-cli.tar.gz:
	@printf "$(MAGENTA)*** Downloading rosetta-cli release $(ROSETTA_CLI_RELEASE)...$(OFF)\n"
	@$(DOWNLOAD) $@ https://github.com/coinbase/rosetta-cli/archive/v$(ROSETTA_CLI_RELEASE).tar.gz

tests/rosetta-cli: tests/rosetta-cli.tar.gz
	@printf "$(MAGENTA)*** Building rosetta-cli...$(OFF)\n"
	@tar -xf $< -C tests
	@cd tests/rosetta-cli-$(ROSETTA_CLI_RELEASE) && go build
	@cp tests/rosetta-cli-$(ROSETTA_CLI_RELEASE)/rosetta-cli tests/.

test: build tests/oasis-net-runner tests/oasis-node tests/rosetta-cli
	@printf "$(CYAN)*** Running tests...$(OFF)\n"
	@$(ROOT)/tests/test.sh

fmt:
	@printf "$(CYAN)*** Formatting code...$(OFF)\n"
	@$(GO) fmt ./...

lint:
	@printf "$(CYAN)*** Linting code...$(OFF)\n"
	@$(GOLINT) run --timeout 1m

clean:
	@printf "$(CYAN)*** Cleaning up...$(OFF)\n"
	@$(GO) clean
	@-rm -f tests/oasis_core_release.tar.gz tests/oasis-net-runner tests/oasis-node
	@-rm -rf tests/oasis-core
	@-rm -f tests/rosetta-cli.tar.gz tests/rosetta-cli
	@-rm -rf tests/rosetta-cli-$(ROSETTA_CLI_RELEASE) tests/validator-data

nuke: clean
	@printf "$(CYAN)*** Cleaning up really well...$(OFF)\n"
	@$(GO) clean -cache -testcache -modcache
