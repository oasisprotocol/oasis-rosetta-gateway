include common.mk

OASIS_RELEASE := 20.11.3
ROSETTA_CLI_RELEASE := 0.4.0

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

all: build
	@$(ECHO) "$(CYAN)*** Everything built successfully!$(OFF)"

build:
	@$(ECHO) "$(CYAN)*** Building...$(OFF)"
	@$(GO) build

build-tests:
	@$(ECHO) "$(CYAN)*** Building tests...$(OFF)"
	@$(GO) build ./tests/...

tests/oasis_core_release.tar.gz:
	@$(ECHO) "$(MAGENTA)*** Downloading oasis-core release $(OASIS_RELEASE)...$(OFF)"
	@$(DOWNLOAD) $@ https://github.com/oasisprotocol/oasis-core/releases/download/v$(OASIS_RELEASE)/oasis_core_$(OASIS_RELEASE)_linux_amd64.tar.gz

tests/oasis-net-runner: tests/oasis_core_release.tar.gz
	@$(ECHO) "$(MAGENTA)*** Unpacking oasis-net-runner...$(OFF)"
	@tar -xf $< -C tests --strip-components=1 oasis_core_$(OASIS_RELEASE)_linux_amd64/oasis-net-runner

tests/oasis-node: tests/oasis_core_release.tar.gz
	@$(ECHO) "$(MAGENTA)*** Unpacking oasis-node...$(OFF)"
	@tar -xf $< -C tests --strip-components=1 oasis_core_$(OASIS_RELEASE)_linux_amd64/oasis-node

tests/rosetta-cli.tar.gz:
	@$(ECHO) "$(MAGENTA)*** Downloading rosetta-cli release $(ROSETTA_CLI_RELEASE)...$(OFF)"
	@$(DOWNLOAD) $@ https://github.com/coinbase/rosetta-cli/archive/v$(ROSETTA_CLI_RELEASE).tar.gz

tests/rosetta-cli: tests/rosetta-cli.tar.gz
	@$(ECHO) "$(MAGENTA)*** Building rosetta-cli...$(OFF)"
	@tar -xf $< -C tests
	@cd tests/rosetta-cli-$(ROSETTA_CLI_RELEASE) && go build
	@cp tests/rosetta-cli-$(ROSETTA_CLI_RELEASE)/rosetta-cli tests/.

test: build build-tests tests/oasis-net-runner tests/oasis-node tests/rosetta-cli
	@$(ECHO) "$(CYAN)*** Running tests...$(OFF)"
	@$(ROOT)/tests/test.sh

# Format code.
fmt:
	@$(ECHO) "$(CYAN)*** Running Go formatters...$(OFF)"
	@gofumpt -s -w .
	@gofumports -w -local github.com/oasisprotocol/oasis-core-rosetta-gateway .

# Lint code, commits and documentation.
lint-targets := lint-go lint-docs lint-git lint-go-mod-tidy

lint-go:
	@$(ECHO) "$(CYAN)*** Running Go linters...$(OFF)"
	@env -u GOPATH golangci-lint run

lint-git:
	@$(CHECK_GITLINT)

lint-docs:
	@$(ECHO) "$(CYAN)*** Runnint markdownlint-cli...$(OFF)"
	@npx markdownlint-cli '**/*.md' --ignore .changelog/

lint-go-mod-tidy:
	@$(ECHO) "$(CYAN)*** Checking go mod tidy...$(OFF)"
	@$(ENSURE_GIT_CLEAN)
	@$(CHECK_GO_MOD_TIDY)

lint: $(lint-targets)

clean:
	@$(ECHO) "$(CYAN)*** Cleaning up...$(OFF)"
	@$(GO) clean
	@-rm -f tests/oasis_core_release.tar.gz tests/oasis-net-runner tests/oasis-node
	@-rm -rf tests/oasis-core
	@-rm -f tests/rosetta-cli.tar.gz tests/rosetta-cli
	@-rm -rf tests/rosetta-cli-$(ROSETTA_CLI_RELEASE) tests/validator-data

nuke: clean
	@$(ECHO) "$(CYAN)*** Cleaning up really well...$(OFF)"
	@$(GO) clean -cache -testcache -modcache

# List of targets that are not actual files.
.PHONY: \
	all build build-tests \
	fmt \
	$(lint-targets) lint \
	test \
	clean nuke
