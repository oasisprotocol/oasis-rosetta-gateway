# Use Bash shell.
# NOTE: You can control which Bash version is used by setting the PATH
# appropriately.
SHELL := bash

# Path to the directory of this Makefile.
# NOTE: Prepend all relative paths in this Makefile with this variable to ensure
# they are properly resolved when this Makefile is included from Makefiles in
# other directories.
SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

# Function for comparing two strings for equality.
# NOTE: This will also return false if both strings are empty.
eq = $(and $(findstring $(1),$(2)), $(findstring $(2),$(1)))

# Check if we're running in an interactive terminal.
ISATTY := $(shell [ -t 0 ] && echo 1)

# If running interactively, use terminal colors.
ifdef ISATTY
	MAGENTA := \e[35;1m
	CYAN := \e[36;1m
	RED := \e[0;31m
	OFF := \e[0m
	ECHO_CMD := echo -e
else
	MAGENTA := ""
	CYAN := ""
	RED := ""
	OFF := ""
	ECHO_CMD := echo
endif

# Output messages to stderr instead stdout.
ECHO := $(ECHO_CMD) 1>&2

# Boolean indicating whether to assume the 'yes' answer when confirming actions.
ASSUME_YES ?= 0

# Helper that asks the user to confirm the action.
define CONFIRM_ACTION =
	if [[ $(ASSUME_YES) != 1 ]]; then \
		$(ECHO) -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]; \
	fi
endef

# Name of git remote pointing to the canonical upstream git repository, i.e.
# git@github.com:oasisprotocol/<repo-name>.git.
GIT_ORIGIN_REMOTE ?= origin

# Name of the branch where to tag the next release.
RELEASE_BRANCH ?= master

# Determine project's version from git.
# NOTE: This computes the project's version from the latest version tag
# reachable from the given $(RELEASE_BRANCH) and does not search for version
# tags across the whole $(GIT_ORIGIN_REMOTE) repository.
GIT_VERSION := $(subst v,,$(shell \
	git describe --tags --match 'v*' --abbrev=0 2>/dev/null $(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH) || \
	echo undefined \
))

PUNCH_CONFIG_FILE := $(abspath $(SELF_DIR).punch_config.py)
PUNCH_VERSION_FILE := $(abspath $(SELF_DIR).punch_version.py)
# Obtain project's version as tracked by the Punch tool.
# NOTE: The Punch tool doesn't have the ability fo print project's version to
# stdout yet.
# For more details, see: https://github.com/lgiordani/punch/issues/42.
PUNCH_VERSION := $(shell \
	python3 -c "exec(open('$(PUNCH_VERSION_FILE)').read()); \
		print(f'{major}.{minor}.{patch}')" \
)

# Determine project's version.
# If the current git commit is exactly a tag and it equals the Punch version,
# then the project's version is that.
# Else, the project version is the Punch version appended with git commit and
# dirty state info.
GIT_COMMIT_EXACT_TAG := $(shell \
	git describe --tags --match 'v*' --exact-match &>/dev/null $(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH) && echo YES || echo NO \
)
VERSION := $(or \
	$(and $(call eq,$(GIT_COMMIT_EXACT_TAG),YES), $(call eq,$(GIT_VERSION),$(PUNCH_VERSION))), \
	$(PUNCH_VERSION)-git$(shell git describe --always --match '' --dirty=+dirty 2>/dev/null) \
)

# Helper that bumps project's version with the Punch tool.
define PUNCH_BUMP_VERSION =
	if [[ "$(RELEASE_BRANCH)" == master ]]; then \
		if [[ -n "$(CHANGELOG_FRAGMENTS_BREAKING)" ]]; then \
			PART=major; \
		else \
			PART=minor; \
		fi; \
	elif [[ "$(RELEASE_BRANCH)" == stable/* ]]; then \
		if [[ -n "$(CHANGELOG_FRAGMENTS_BREAKING)" ]]; then \
	        $(ECHO) "$(RED)Error: There shouldn't be breaking changes in a release on a stable branch.$(OFF)"; \
			$(ECHO) "List of detected breaking changes:"; \
			for fragment in "$(CHANGELOG_FRAGMENTS_BREAKING)"; do \
				$(ECHO) "- $$fragment"; \
			done; \
			exit 1; \
		else \
			PART=patch; \
		fi; \
    else \
	    $(ECHO) "$(RED)Error: Unsupported release branch: '$(RELEASE_BRANCH)'.$(OFF)"; \
		exit 1; \
	fi; \
	punch --config-file $(PUNCH_CONFIG_FILE) --version-file $(PUNCH_VERSION_FILE) --part $$PART --quiet
endef

# Helper that ensures project's version determined from git equals project's
# version as tracked by the Punch tool.
define ENSURE_GIT_VERSION_EQUALS_PUNCH_VERSION =
	if [[ "$(GIT_VERSION)" != "$(PUNCH_VERSION)" ]]; then \
		$(ECHO) "$(RED)Error: Project's version for $(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH) \
			determined from git ($(GIT_VERSION)) doesn't equal project's version in \
			$(PUNCH_VERSION_FILE) ($(PUNCH_VERSION)).$(OFF)"; \
		exit 1; \
	fi
endef

# Go binary to use for all Go commands.
export OASIS_GO ?= go

# Go command prefix to use in all Go commands.
GO := env -u GOPATH $(OASIS_GO)

# Helper that ensures the git workspace is clean.
define ENSURE_GIT_CLEAN =
	if [[ ! -z `git status --porcelain` ]]; then \
		$(ECHO) "$(RED)Error: Git workspace is dirty.$(OFF)"; \
		exit 1; \
	fi
endef

# Helper that checks if the go mod tidy command was run.
# NOTE: go mod tidy doesn't implement a check mode yet.
# For more details, see: https://github.com/golang/go/issues/27005.
define CHECK_GO_MOD_TIDY =
    $(GO) mod tidy; \
    if [[ ! -z `git status --porcelain go.mod go.sum` ]]; then \
        $(ECHO) "$(RED)Error: The following changes detected after running 'go mod tidy':$(OFF)"; \
        git diff go.mod go.sum; \
        exit 1; \
    fi
endef

# Helper that checks commits with gitlilnt.
# NOTE: gitlint internally uses git rev-list, where A..B is asymmetric
# difference, which is kind of the opposite of how git diff interprets
# A..B vs A...B.
define CHECK_GITLINT =
	BRANCH=$(GIT_ORIGIN_REMOTE)/$(RELEASE_BRANCH); \
	COMMIT_SHA=`git rev-parse $$BRANCH` && \
	$(ECHO) "$(CYAN)*** Running gitlint for commits from $$BRANCH ($${COMMIT_SHA:0:7})... $(OFF)"; \
	gitlint --commits $$BRANCH..HEAD
endef

# List of non-trivial Change Log fragments.
CHANGELOG_FRAGMENTS_NON_TRIVIAL := $(filter-out $(wildcard .changelog/*trivial*.md),$(wildcard .changelog/[0-9]*.md))

# List of breaking Change Log fragments.
CHANGELOG_FRAGMENTS_BREAKING := $(wildcard .changelog/*breaking*.md)

# Helper that checks Change Log fragments with markdownlint-cli and gitlint.
# NOTE: Non-zero exit status is recorded but only set at the end so that all
# markdownlint or gitlint errors can be seen at once.
define CHECK_CHANGELOG_FRAGMENTS =
	exit_status=0; \
	$(ECHO) "$(CYAN)*** Running markdownlint-cli for Change Log fragments... $(OFF)"; \
	npx markdownlint-cli --config .changelog/.markdownlint.yml .changelog/ || exit_status=$$?; \
	$(ECHO) "$(CYAN)*** Running gitlint for Change Log fragments: $(OFF)"; \
	for fragment in $(CHANGELOG_FRAGMENTS_NON_TRIVIAL); do \
		$(ECHO) "- $$fragment"; \
		gitlint --msg-filename $$fragment -c title-max-length.line-length=78 || exit_status=$$?; \
	done; \
	exit $$exit_status
endef
