# Oasis Gateway for Rosetta

This repository implements the [Rosetta][1] server for the [Oasis][0] network.

To build the server:

	make

To run tests:

	make test

To clean-up:

	make clean


`make test` will automatically download the [Oasis node][0] and [rosetta-cli][2],
set up a test Oasis network, make some sample transactions, then run the
gateway and validate it using `rosetta-cli`.

[0]: https://github.com/oasisprotocol/oasis-core
[1]: https://github.com/coinbase/rosetta-sdk-go
[2]: https://github.com/coinbase/rosetta-cli
