# Running Oasis node and gateway in Docker

This directory contains a Dockerfile that builds an image containing both
the [Oasis node][0] and the [Oasis Rosetta gateway][1], as instructed by the
[Rosetta requirements][2].

The node should be configured by following the instructions at:
https://docs.oasis.dev/operators/joining-the-testnet.html

The `/serverdir` in the instructions is equivalent to the `/data` mountpoint
of the Docker image.

Certain steps are not needed, e.g. provisioning entities and nodes.

You can use the `config.yml` file included in this directory and get the
`genesis.json` file from the instructions above.

Don't forget to set the proper permissions on the directory you're using as
the `/data` mountpoint, as well as its subdirectories and files!

To build the Docker image:

	docker build -t oasis-core-rosetta-gateway .

To run the Docker image:

	docker run -v /path/to/your/data:/data -it oasis-core-rosetta-gateway

When run, the Docker image starts the Oasis node and the Oasis Rosetta gateway
and exposes TCP ports 8080 (gateway) and 26656 (node).

The final image is around 87MB in size.  You can remove the intermediary build
image to save disk space.

[0]: https://github.com/oasisprotocol/oasis-core
[1]: https://github.com/oasisprotocol/oasis-core-rosetta-gateway
[2]: https://djr6hkgq2tjcs.cloudfront.net/docs/NodeRequirements.html
