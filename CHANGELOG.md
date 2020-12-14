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
