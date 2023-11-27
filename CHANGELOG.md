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

## 2.6.0 (2023-10-11)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   |  23.0   |

### Features

- Bump Oasis Core to 23.0
  ([#474](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/474))

### Internal Changes

- ci: bump golangci/golangci-lint-action from 3.6.0 to 3.7.0
  ([#457](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/457))

- ci: bump actions/setup-node from 3.7.0 to 3.8.1
  ([#459](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/459))

- ci: bump actions/checkout from 3 to 4
  ([#460](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/460))

- ci: bump actions/cache from 3.3.1 to 3.3.2
  ([#463](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/463))

- ci: bump docker/build-push-action from 4.1.1 to 5.0.0
  ([#465](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/465))

- ci: bump docker/setup-buildx-action from 2 to 3
  ([#470](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/470))

- ci: bump actions/setup-python from 4.7.0 to 4.7.1
  ([#471](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/471))

- go: bump golang.org/x/net from 0.13.0 to 0.17.0
  ([#472](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/472))

## 2.5.0 (2023-08-11)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   | 22.2.11 |

### Features

- Bump Oasis Core to 22.2.11
  ([#455](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/455))

### Internal Changes

- ci: bump golangci/golangci-lint-action from 3.4.0 to 3.6.0
  ([#444](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/444))

- ci: bump docker/build-push-action from 4.0.0 to 4.1.1
  ([#445](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/445))

- ci: bump actions/setup-node from 3.6.0 to 3.7.0
  ([#449](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/449))

- ci: bump actions/setup-python from 4.6.0 to 4.7.0
  ([#451](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/451))

- go: bump google.golang.org/grpc from 1.55.0 to 1.57.0
  ([#452](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/452))

## 2.4.0 (2023-05-18)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   | 22.2.8  |

### Features

- Bump Oasis Core to 22.2.8
  ([#433](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/433))

- network: add gateway version as middleware_version
  ([#438](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/438))

### Documentation Improvements

- fixed error on Reclaim Escrow documentation
  ([#424](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/424))

### Internal Changes

- ci: bump actions/setup-node from 3.5.0 to 3.6.0
  ([#405](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/405))

- ci: bump actions/setup-python from 4.2.0 to 4.6.0
  ([#409](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/409),
   [#434](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/434))

- ci: bump golangci/golangci-lint-action from 3.2.0 to 3.4.0
  ([#412](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/412))

- ci: bump actions/cache from 3.0.8 to 3.3.1
  ([#413](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/413))

- ci: bump docker/build-push-action from 3.1.1 to 4.0.0
  ([#417](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/417))

- go: bump github.com/coinbase/rosetta-cli from 0.10.0 to 0.10.3
  ([#418](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/418))

- ci: bump Node.js
  ([#428](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/428))

- ci: bump actions/setup-go from 3 to 4
  ([#430](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/430))

- go: bump google.golang.org/grpc from 1.49.0 to 1.55.0
  ([#432](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/432),
   [#437](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/437))

- ci: update gitlint config
  ([#435](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/435))

- ci: Append Change Log fragments to Dependabot PRs
  ([#436](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/436))

## 2.3.0 (2022-11-07)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   | 22.2.1  |

### Features

- docker: Update mainnet genesis
  ([#318](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/318),
   [#325](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/325))

- Bump Oasis Core to 22.2.1
  ([#318](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/318),
   [#336](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/336),
   [#340](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/340),
   [#349](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/349),
   [#360](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/360),
   [#369](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/369),
   [#378](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/378),
   [#390](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/390))

- go: bump github.com/coinbase/rosetta-sdk-go from 0.7.3 to 0.8.1
  ([#329](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/329),
   [#334](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/334),
   [#359](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/359),
   [#380](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/380))

### Internal Changes

- go: bump github.com/coinbase/rosetta-cli from 0.7.3 to 0.10.0
  ([#313](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/313),
   [#335](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/335),
   [#346](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/346),
   [#366](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/366),
   [#375](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/375),
   [#381](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/381))

- ci: bump actions/cache from 3.0.1 to 3.0.8
  ([#317](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/317),
   [#350](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/350),
   [#363](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/363),
   [#374](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/374))

- ci: bump actions/setup-node from 3.1.0 to 3.5.0
  ([#326](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/326),
  [#344](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/344),
  [#365](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/365),
  [#382](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/382))

- ci: bump actions/setup-python from 3.1.0 to 4.2.0
  ([#327](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/327),
  [#361](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/361),
  [#370](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/370))

- go: bump google.golang.org/grpc from 1.45.0 to 1.49.0
  ([#332](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/332),
  [#343](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/343),
  [#351](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/351),
  [#364](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/364),
  [#376](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/376))

- ci: bump docker/setup-buildx-action from 1 to 2
  ([#338](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/338))

- ci: bump docker/build-push-action from 2.10.0 to 3.1.1
  ([#339](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/339),
  [#367](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/367),
  [#371](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/371))

- ci: bump golangci/golangci-lint-action from 3.1.0 to 3.2.0
  ([#341](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/341))

## 2.2.0 (2022-04-06)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   |  22.1   |

### Features

- Bump Oasis Core to 22.1.2
  ([#312](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/312))

### Internal Changes

- ci: bump golangci/golangci-lint-action from 2.5.2 to 3.1.0
  ([#286](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/286))

- ci: bump actions/setup-go from 2.1.3 to 3
  ([#287](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/287))

- go: bump google.golang.org/grpc from 1.44.0 to 1.45.0
  ([#293](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/293))

- ci: bump docker/build-push-action from 2.6.1 to 2.10.0
  ([#295](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/295))

- ci: bump actions/setup-python from 2.2.2 to 3.1.0
  ([#304](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/304))

- ci: bump actions/setup-node from 2.4.0 to 3.1.0
  ([#305](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/305))

- ci: bump actions/cache from 2 to 3.0.1
  ([#311](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/311))

- ci: bump actions/checkout from 2 to 3
  ([#314](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/314))

## 2.1.0 (2022-03-07)

| Name         | Version |
|:-------------|:-------:|
| Rosetta API  | 1.4.12  |
| Oasis Core   |  22.0   |

### Features

- Bump Oasis Core to 22.0
  ([#290](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/290))

## 2.0.0 (2022-02-16)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.12    |
| Oasis Core   | 21.3      |

### Removals and Breaking Changes

- Bring Rosetta dependencies up to date
  ([#261](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/261),
   [#277](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/277))

  We're updating our Rosetta dependencies, including the
  Go SDK, the CLI, and along with those, the Rosetta API
  specifications.

  - Rosetta CLI: 0.4.0 -> 0.7.3
  - Rosetta Go SDK: 0.3.3 -> 0.7.3
  - Rosetta API: 1.4.1 -> 1.4.12

  Updating the Rosetta API along with its Go SDK is a
  significant update and may have introduced breaking
  changes.

- Rename project to Oasis Rosetta Gateway
  ([#281](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/281))

  Previous name was quite long and we were already shortening it in some places.
  Take the opportunity of doing a breaking 2.0.0 release to shorten the name to
  Oasis Rosetta Gateway.

### Features

- docker: Change Dockerfile to Ubuntu base
  ([#257](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/257))

  <!-- markdownlint-disable line-length -->
  The Rosetta ecosystem prefers this:
  [Rosetta's guidance on Docker deployment](https://docs.cloud.coinbase.com/rosetta/docs/docker-deployment#ubuntu-image-compatibility).
  <!-- markdownlint-enable line-length -->

- docker: Skip running node if offline var is set
  ([#278](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/278))

  This is only an optimization.
  We recommend that users don't rely on software configuration to operate
  offline.

### Internal Changes

- common: Obtain `RosettaAPIVersion` from [rosetta-sdk-go/types] package
  ([#129](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/129))

  [rosetta-sdk-go/types]:
    https://pkg.go.dev/github.com/coinbase/rosetta-sdk-go/types#pkg-constants

- Bump Go to version 1.17
  ([#251](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/251))

- Make: Build `rosetta-cli` with the correct version of Go
  ([#251](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/251))

- go: bump google.golang.org/grpc from 1.41.0 to 1.44.0
  ([#270](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/270))

- github: For PRs, build docker image with the PR's branch of Rosetta Gateway
  ([#273](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/273))

## 1.3.0 (2021-11-03)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |
| Oasis Core   | 21.3      |

### Features

- Bump Oasis Core to 21.3.5
  ([#241](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/241))

### Internal Changes

- go: bump github.com/ethereum/go-ethereum to 1.10.9
  ([#241](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/241))

  It fixes a DoS issue via a maliciously crafted p2p message.
  For more details, see [GHSA-59hh-656j-3p7v].

  [GHSA-59hh-656j-3p7v]: https://github.com/advisories/GHSA-59hh-656j-3p7v

## 1.2.0 (2021-08-11)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |
| Oasis Core   | 21.2      |

### Features

- Bump Oasis Core to 21.2.8
  ([#217](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/217),
   [#232](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/232))

### Bug Fixes

- Fix possible `nil` pointer dereference in `GetStatus`
  ([#196](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/196))

### Documentation Improvements

- Note that `RELEASE_BRANCH` variable needs to be exported in [Release Process]
  ([#198](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/198),
   [#200](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/200))

### Internal Changes

- changelog: Automatically add important versions table
  ([#198](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/198))

- go: bump github.com/vmihailenco/msgpack/v5 from 5.0.0-beta.1 to 5.3.4
  ([#210](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/210))

- docker: Improve build steps and ignore everything in .dockerignore
  ([#222](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/222))

- github: Add [docker workflow] for testing building Docker images
  ([#222](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/222))

  [docker workflow]:
    https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:docker

- ci: bump docker/build-push-action from 2.5.0 to 2.6.1
  ([#223](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/223))

- ci: bump actions/setup-node from 2.1.5 to 2.4.0
  ([#229](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/229))

- go: bump google.golang.org/grpc from 1.38.0 to 1.39.1
  ([#230](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/230))

## 1.1.0 (2021-04-22)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |
| Oasis Core   | 21.1      |

### Features

- Add epoch number to block metadata
  ([#188](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/188))

### Documentation Improvements

- Add Oasis Core version to important versions listed in the Change Log
  ([#191](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/191))

### Internal Changes

- ci: bump actions/setup-node from v2.1.2 to v2.1.3
  ([#139](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/139))

- ci: bump actions/setup-node from v2.1.3 to v2.1.5
  ([#166](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/166))

- oasis: Use GetChainContext method instead of fetching genesis document
  ([#180](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/180))

- ci: bump golangci/golangci-lint-action from v2.3.0 to v2.5.2
  ([#183](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/183))

- go: bump google.golang.org/grpc from 1.36.0 to 1.37.0
  ([#185](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/185))

- ci: bump actions/setup-python from v2.1.4 to v2.2.2
  ([#186](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/186))

- Bump Oasis Core version to 21.1
  ([#175](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/175),
   [#187](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/187))

- ci: Bump golangci-lint version in *ci-lint* GitHub Actions workflow to 1.39
  ([#190](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/190))

## 1.0.0 (2020-12-14)

| Name         | Version   |
|:-------------|:---------:|
| Rosetta API  | 1.4.1     |

### Process

- Define project's release process
  ([#55](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/55))

  For more details, see [Release Process]

  [Release Process]: docs/release-process.md

- Add Change Log and the Change Log fragments process for assembling it
  ([#120](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/120),
   [#140](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/140))

  This follows the same Change Log fragments process as is used by [Oasis Core].

  For more details, see [Change Log fragments].

  [Oasis Core]: https://github.com/oasisprotocol/oasis-core
  [Change Log fragments]: .changelog/README.md

- Define project's versioning
  ([#122](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/122))

  Adopt a [Semantic Versioning 2.0.0].

  For more details, see [Versioning] docs.

  [Semantic Versioning 2.0.0]: https://semver.org/spec/v2.0.0.html
  [Versioning]: docs/versioning.md

### Features

- common: Add package implementing common things for Oasis Core Rosetta Gateway
  ([#128](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/128))

  Initially, it stores the versions of the Rosetta API, Go toolchain and the
  Oasis Core Rosetta Gateway itself.

- cli: Add `-version` flag to `oasis-core-rosetta-gateway` binary
  ([#128](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/128),
   [#134](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/134))

- common: Add `GetOasisCoreVersion()` helper for obtaining Oasis Core's version
  ([#134](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/134))

### Internal Changes

- ci: bump golangci/golangci-lint-action from v2.2.0 to v2.3.0
  ([#90](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/90))

- Add linting for Change Log fragments
  ([#120](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/120))

  Add `lint-changelog` Make target and *Lint Change Log fragments* step to the
  *ci-lint* GitHub Actions workflow.

- Use [Punch] tool for tracking and bumping project's version
  ([#122](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/122))

  [Punch]: https://github.com/lgiordani/punch

- Make: Add `changelog` target for assembling the Change Log
  ([#122](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/122))

- Make: Add `fetch-git` target for fetching changes from the canonical git repo
  ([#122](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/122))

- go: bump google.golang.org/grpc from 1.32.0 to 1.34.0
  ([#126](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/126))

- cli: Extract port setting steps to `getPortOrExit()` function
  ([#128](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/128))

- Make: Add reproducibility and version info flags to Go builds
  ([#128](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/128))

- go: Bump Oasis Core dependency to 20.12.3
  ([#131](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/131))

- github: Add [*release* GitHub Actions workflow]
  ([#138](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/138))

  [*release* GitHub Actions workflow]:
    https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:release

- Make: Add `release-tag`, `release-stable-branch` and `release-build` targets
  ([#138](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/138))

  They can be used for:

  - `release-tag`: tagging the next release,
  - `release-stable-branch`: creating and pushing a stable branch for the
    current release,
  - `release-build`: building and publishing a release.

- Use [GoReleaser] tool for building and publishing releases
  ([#138](https://github.com/oasisprotocol/oasis-rosetta-gateway/issues/138))

  [GoReleaser]: https://goreleaser.com/
