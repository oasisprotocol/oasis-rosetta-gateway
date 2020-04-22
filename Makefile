#!/usr/bin/env gmake

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

# Built-in echo doesn't support '-e'.
ECHO = /bin/echo -e
else
# Don't use colors if not running interactively.
MAGENTA = ""
CYAN = ""
OFF = ""

# OK to use built-in echo.
ECHO = echo
endif

.PHONY: all build clean fmt lint nuke test

all: build
	@$(ECHO) "$(CYAN)*** Everything built successfully!$(OFF)"

build:
	@$(ECHO) "$(CYAN)*** Building...$(OFF)"
	@$(GO) build

test:
	@$(ECHO) "$(CYAN)*** Running tests...$(OFF)"
	@$(GO) test --timeout 2m -race -v ./...

fmt:
	@$(ECHO) "$(CYAN)*** Formatting code...$(OFF)"
	@$(GO) fmt ./...

lint:
	@$(ECHO) "$(CYAN)*** Linting code...$(OFF)"
	@$(GOLINT) run --timeout 1m

clean:
	@$(ECHO) "$(CYAN)*** Cleaning up...$(OFF)"
	@$(GO) clean

nuke:
	@$(ECHO) "$(CYAN)*** Cleaning up really well...$(OFF)"
	@$(GO) clean -cache -testcache -modcache
