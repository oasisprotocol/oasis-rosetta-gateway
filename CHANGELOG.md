# Change Log

All notables changes to this project are documented in this file.

The format is inspired by [Keep a Changelog].

[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/

<!-- Custom markdownlint configuration for the Change Log. -->
<!-- markdownlint-configure-file
{
  "extends": ".markdownlint.yml",
  "line-length": {
    "stern": "true"
  },
  "no-duplicate-heading": false
}
-->

<!-- NOTE: towncrier will not alter content above the TOWNCRIER line below. -->

<!-- TOWNCRIER -->

## 1.2.0 (2021-08-11)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |
| Oasis Core   | 21.2      |

### Features

- Bump Oasis Core to 21.2.8
  ([#217](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/217),
   [#232](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/232))

### Bug Fixes

- Fix possible `nil` pointer dereference in `GetStatus`
  ([#196](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/196))

### Documentation Improvements

- Note that `RELEASE_BRANCH` variable needs to be exported in [Release Process]
  ([#198](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/198),
   [#200](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/200))

  [Release Process]: docs/release-process.md

### Internal Changes

- changelog: Automatically add important versions table
  ([#198](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/198))

- go: bump github.com/vmihailenco/msgpack/v5 from 5.0.0-beta.1 to 5.3.4
  ([#210](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/210))

- docker: Improve build steps and ignore everything in .dockerignore
  ([#222](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/222))

- github: Add [docker workflow] for testing building Docker images
  ([#222](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/222))

  [docker workflow]:
    https://github.com/oasisprotocol/oasis-core/actions?query=workflow:docker

- ci: bump docker/build-push-action from 2.5.0 to 2.6.1
  ([#223](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/223))

- ci: bump actions/setup-node from 2.1.5 to 2.4.0
  ([#229](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/229))

- go: bump google.golang.org/grpc from 1.38.0 to 1.39.1
  ([#230](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/230))

## 1.1.0 (2021-04-22)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |
| Oasis Core   | 21.1      |

### Features

- Add epoch number to block metadata
  ([#188](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/188))

### Documentation Improvements

- Add Oasis Core version to important versions listed in the Change Log
  ([#191](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/191))

### Internal Changes

- ci: bump actions/setup-node from v2.1.2 to v2.1.3
  ([#139](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/139))

- ci: bump actions/setup-node from v2.1.3 to v2.1.5
  ([#166](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/166))

- oasis: Use GetChainContext method instead of fetching genesis document
  ([#180](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/180))

- ci: bump golangci/golangci-lint-action from v2.3.0 to v2.5.2
  ([#183](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/183))

- go: bump google.golang.org/grpc from 1.36.0 to 1.37.0
  ([#185](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/185))

- ci: bump actions/setup-python from v2.1.4 to v2.2.2
  ([#186](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/186))

- Bump Oasis Core version to 21.1
  ([#175](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/175),
   [#187](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/187))

- ci: Bump golangci-lint version in _ci-lint_ GitHub Actions workflow to 1.39
  ([#190](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/190))

## 1.0.0 (2020-12-14)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |

### Process

- Define project's release process
  ([#55](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/55))

  For more details, see [Release Process]

  [Release Process]: docs/release-process.md

- Add Change Log and the Change Log fragments process for assembling it
  ([#120](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/120),
   [#140](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/140))

  This follows the same Change Log fragments process as is used by [Oasis Core].

  For more details, see [Change Log fragments].

  [Oasis Core]: https://github.com/oasisprotocol/oasis-core
  [Change Log fragments]: .changelog/README.md

- Define project's versioning
  ([#122](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/122))

  Adopt a [Semantic Versioning 2.0.0].

  For more details, see [Versioning] docs.

  [Semantic Versioning 2.0.0]: https://semver.org/spec/v2.0.0.html
  [Versioning]: docs/versioning.md

### Features

- common: Add package implementing common things for Oasis Core Rosetta Gateway
  ([#128](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/128))

  Initially, it stores the versions of the Rosetta API, Go toolchain and the
  Oasis Core Rosetta Gateway itself.

- cli: Add `-version` flag to `oasis-core-rosetta-gateway` binary
  ([#128](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/128),
   [#134](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/134))

- common: Add `GetOasisCoreVersion()` helper for obtaining Oasis Core's version
  ([#134](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/134))

### Internal Changes

- ci: bump golangci/golangci-lint-action from v2.2.0 to v2.3.0
  ([#90](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/90))

- Add linting for Change Log fragments
  ([#120](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/120))

  Add `lint-changelog` Make target and *Lint Change Log fragments* step to the
  *ci-lint* GitHub Actions workflow.

- Use [Punch] tool for tracking and bumping project's version
  ([#122](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/122))

  [Punch]: https://github.com/lgiordani/punch

- Make: Add `changelog` target for assembling the Change Log
  ([#122](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/122))

- Make: Add `fetch-git` target for fetching changes from the canonical git repo
  ([#122](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/122))

- go: bump google.golang.org/grpc from 1.32.0 to 1.34.0
  ([#126](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/126))

- cli: Extract port setting steps to `getPortOrExit()` function
  ([#128](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/128))

- Make: Add reproducibility and version info flags to Go builds
  ([#128](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/128))

- go: Bump Oasis Core dependency to 20.12.3
  ([#131](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/131))

- github: Add [*release* GitHub Actions workflow]
  ([#138](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/138))

  [*release* GitHub Actions workflow]:
    https://github.com/oasisprotocol/oasis-core-rosetta-gateway/actions?query=workflow:release

- Make: Add `release-tag`, `release-stable-branch` and `release-build` targets
  ([#138](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/138))

  They can be used for:

  - `release-tag`: tagging the next release,
  - `release-stable-branch`: creating and pushing a stable branch for the
    current release,
  - `release-build`: building and publishing a release.

- Use [GoReleaser] tool for building and publishing releases
  ([#138](https://github.com/oasisprotocol/oasis-core-rosetta-gateway/issues/138))

  [GoReleaser]: https://goreleaser.com/
