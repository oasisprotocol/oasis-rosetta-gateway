# Running Oasis Node and Rosetta Gateway in Docker

_NOTE: Oasis recommends using binaries from the [official GitHub Releases].
The versions of [Rosetta API] and [Oasis Core] that are supported for a given
release are indicated in a release's Change Log (e.g. [Change Log for 1.1.1]
release)._

This directory contains a Dockerfile that builds an image containing both
the [Oasis Node] and the [Oasis Rosetta gateway], as instructed by the
[Rosetta Docker Deployment] doc.

The node should be configured as described in [Run a Non-validator Node] doc
of the [Run Node Oasis Docs].

The `/node/data` directory in the instructions is equivalent to the `/data`
directory of the Docker image.
The `/node/etc` directory in the instructions is equivalent to the `/data/etc`
directory of the Docker image.
We recommend creating a volume mount for `/data` so that you can manage the
directory across version upgrades.

When using a volume mount for `/data`, you must provide a `config.yml` file
and a `genesis.json` file.
You can use the `config.yml` file included in this directory and get the
`genesis.json` file from the instructions above.

Don't forget to set the proper permissions on the directory you're using as
the `/data` mountpoint, as well as its subdirectories and files!

To build the Docker image, run:

```bash
docker build -t oasis-rosetta-gateway .
```

To run the Docker image, run:

```bash
docker run -v /path/to/your/data:/data -p 8080:8080 -it oasis-rosetta-gateway
```

When run, the Docker image starts the Oasis node and the Oasis Rosetta gateway
and exposes TCP ports 8080 (gateway) and 26656 (node).

The final image is around 156MB in size.  You can remove the intermediary
build  image to save disk space.

<!-- markdownlint-disable line-length -->
[official GitHub Releases]:
  https://github.com/oasisprotocol/oasis-rosetta-gateway/releases/
[Change Log for 1.1.1]:
  https://github.com/oasisprotocol/oasis-rosetta-gateway/blob/v1.1.1/CHANGELOG.md
[Rosetta API]: https://www.rosetta-api.org/docs/welcome.html
[Oasis Core]: https://github.com/oasisprotocol/oasis-core
[Oasis Node]:
  https://docs.oasis.io/node/run-your-node/prerequisites/oasis-node/
[Oasis Rosetta Gateway]:
  https://github.com/oasisprotocol/oasis-rosetta-gateway
[Rosetta Docker Deployment]:
  https://www.rosetta-api.org/docs/node_deployment.html
[Run a Non-validator Node]:
  https://docs.oasis.io/node/run-your-node/non-validator-node/#configuration
[Run Node Oasis Docs]:
  https://docs.oasis.io/node/
<!-- markdownlint-enable line-length -->
