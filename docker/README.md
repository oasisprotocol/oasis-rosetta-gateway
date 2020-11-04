# Running Oasis Node and Rosetta Gateway in Docker

This directory contains a Dockerfile that builds an image containing both
the [Oasis Node] and the [Oasis Rosetta gateway], as instructed by the
[Rosetta Docker Deployment] doc.

The node should be configured as described in [Run a Non-validator Node] doc
of the general [Oasis Docs].

The `/node` directory in the instructions is equivalent to the `/data`
mountpoint of the Docker image.

You can use the `config.yml` file included in this directory and get the
`genesis.json` file from the instructions above.

Don't forget to set the proper permissions on the directory you're using as
the `/data` mountpoint, as well as its subdirectories and files!

To build the Docker image, run:

```bash
docker build -t oasis-core-rosetta-gateway .
```

To run the Docker image, run:

```bash
docker run -v /path/to/your/data:/data -it oasis-core-rosetta-gateway
```

When run, the Docker image starts the Oasis node and the Oasis Rosetta gateway
and exposes TCP ports 8080 (gateway) and 26656 (node).

The final image is around 87MB in size.  You can remove the intermediary build
image to save disk space.

<!-- markdownlint-disable line-length -->
[Oasis Node]:
  https://docs.oasis.dev/general/run-a-node/prerequisites/oasis-node
[Oasis Rosetta Gateway]:
  https://github.com/oasisprotocol/oasis-core-rosetta-gateway
[Rosetta Docker Deployment]:
  https://www.rosetta-api.org/docs/node_deployment.html
[Run a Non-validator Node]:
  https://docs.oasis.dev/general/run-a-node/set-up-your-node/run-non-validator#configuration
[Oasis Docs]:
  https://docs.oasis.dev/
<!-- markdownlint-enable line-length -->
