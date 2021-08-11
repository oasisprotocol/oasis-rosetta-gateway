include common.mk

OASIS_RELEASE := 21.2.8
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

# Check if Go's linkers flags are set in common.mk and add them as extra flags.
ifneq ($(GOLDFLAGS),)
	GO_EXTRA_FLAGS += -ldflags $(GOLDFLAGS)
endif

ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: build
	@$(ECHO) "$(CYAN)*** Everything built successfully!$(OFF)"

build:
	@$(ECHO) "$(CYAN)*** Building...$(OFF)"
	@$(GO) build $(GOFLAGS) $(GO_EXTRA_FLAGS)

build-tests:
	@$(ECHO) "$(CYAN)*** Building tests...$(OFF)"
	@$(GO) build $(GOFLAGS) $(GO_EXTRA_FLAGS) ./tests/...

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
lint-targets := lint-go lint-docs lint-changelog lint-git lint-go-mod-tidy

lint-go:
	@$(ECHO) "$(CYAN)*** Running Go linters...$(OFF)"
	@env -u GOPATH golangci-lint run

lint-git:
	@$(CHECK_GITLINT)

lint-docs:
	@$(ECHO) "$(CYAN)*** Runnint markdownlint-cli...$(OFF)"
	@npx markdownlint-cli '**/*.md' --ignore .changelog/ --ignore tests/rosetta-cli-*/

lint-changelog:
	@$(CHECK_CHANGELOG_FRAGMENTS)

lint-go-mod-tidy:
	@$(ECHO) "$(CYAN)*** Checking go mod tidy...$(OFF)"
	@$(ENSURE_GIT_CLEAN)
	@$(CHECK_GO_MOD_TIDY)

lint: $(lint-targets)

# Fetch all the latest changes (including tags) from the canonical upstream git
# repository.
fetch-git:
	@$(ECHO) "Fetching the latest changes (including tags) from $(GIT_ORIGIN_REMOTE) remote..."
	@git fetch $(GIT_ORIGIN_REMOTE) --tags

# Private target for bumping project's version using the Punch tool.
# NOTE: It should not be invoked directly.
_version-bump: fetch-git
	@$(ENSURE_VALID_RELEASE_BRANCH_NAME)
	@$(PUNCH_BUMP_VERSION)
	@git add $(PUNCH_VERSION_FILE)

# Private target for assembling the Change Log.
# NOTE: It should not be invoked directly.
_changelog:
	@$(ECHO) "$(CYAN)*** Generating Change Log for version $(PUNCH_VERSION)...$(OFF)"
	@$(BUILD_CHANGELOG)
	@$(ECHO) "Next, review the staged changes, commit them and make a pull request."
	@$(WARN_BREAKING_CHANGES)

# Assemble Change Log.
# NOTE: We need to call Make recursively since _version-bump target updates
# Punch's version and hence we need Make to re-evaluate the PUNCH_VERSION
# variable.
changelog: _version-bump
	@$(MAKE) --no-print-directory _changelog

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

# Tag the next release.
release-tag: fetch-git
	@$(ECHO) "Checking if we can tag version $(PUNCH_VERSION) as the next release..."
	@$(ENSURE_VALID_RELEASE_BRANCH_NAME)
	@$(ENSURE_RELEASE_TAG_DOES_NOT_EXIST)
	@$(ENSURE_NO_CHANGELOG_FRAGMENTS)
	@$(ENSURE_NEXT_RELEASE_IN_CHANGELOG)
	@$(ECHO) "All checks have passed. Proceeding with tagging the $(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH)'s HEAD with tag '$(RELEASE_TAG)'."
	@$(CONFIRM_ACTION)
	@$(ECHO) "If this appears to be stuck, you might need to touch your security key for GPG sign operation."
	@git tag --sign --message="Version $(PUNCH_VERSION)" $(RELEASE_TAG) $(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH)
	@git push $(GIT_ORIGIN_REMOTE) $(RELEASE_TAG)
	@$(ECHO) "$(CYAN)*** Tag '$(RELEASE_TAG)' has been successfully pushed to $(GIT_ORIGIN_REMOTE) remote.$(OFF)"

# Create and push a stable branch for the current release.
release-stable-branch: fetch-git
	@$(ECHO) "Checking if we can create a stable release branch for version $(PUNCH_VERSION)...$(OFF)"
	@$(ENSURE_VALID_STABLE_BRANCH)
	@$(ENSURE_RELEASE_TAG_EXISTS)
	@$(ENSURE_STABLE_BRANCH_DOES_NOT_EXIST)
	@$(ECHO) "All checks have passed. Proceeding with creating the '$(STABLE_BRANCH)' branch on $(GIT_ORIGIN_REMOTE) remote."
	@$(CONFIRM_ACTION)
	@git branch $(STABLE_BRANCH) $(RELEASE_TAG)
	@git push $(GIT_ORIGIN_REMOTE) $(STABLE_BRANCH)
	@$(ECHO) "$(CYAN)*** Branch '$(STABLE_BRANCH)' has been sucessfully pushed to $(GIT_ORIGIN_REMOTE) remote.$(OFF)"

# Build and publish the next release.
release-build:
	@$(ENSURE_VALID_RELEASE_BRANCH_NAME)
ifeq ($(OASIS_CORE_ROSETTA_GATEWAY_REAL_RELEASE), true)
	@$(ENSURE_GIT_VERSION_EQUALS_PUNCH_VERSION)
endif
	@$(ECHO) "$(CYAN)*** Creating release for version $(PUNCH_VERSION)...$(OFF)"
	@goreleaser $(GORELEASER_ARGS)

# List of targets that are not actual files.
.PHONY: \
	all build build-tests \
	test \
	fmt \
	$(lint-targets) lint \
	fetch-git \
	_version-bump _changelog changelog \
	clean nuke \
	release-tag release-stable-branch release-build
