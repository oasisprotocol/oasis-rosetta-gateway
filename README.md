# Oasis Rosetta Gateway

[![CI test status][github-ci-tests-badge]][github-ci-tests-link]
[![CI lint status][github-ci-lint-badge]][github-ci-lint-link]
[![Docker status][github-docker-badge]][github-docker-link]
[![Release status][github-release-badge]][github-release-link]

[github-ci-tests-badge]: https://github.com/oasisprotocol/oasis-rosetta-gateway/workflows/ci-tests/badge.svg
[github-ci-tests-link]: https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:ci-tests+branch:master
[github-ci-lint-badge]: https://github.com/oasisprotocol/oasis-rosetta-gateway/workflows/ci-lint/badge.svg
[github-ci-lint-link]: https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:ci-lint+branch:master
[github-docker-badge]: https://github.com/oasisprotocol/oasis-rosetta-gateway/workflows/docker/badge.svg
[github-docker-link]: https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:docker
[github-release-badge]: https://github.com/oasisprotocol/oasis-rosetta-gateway/workflows/release/badge.svg
[github-release-link]: https://github.com/oasisprotocol/oasis-rosetta-gateway/actions?query=workflow:release

This repository implements the [Rosetta] server for the [Oasis Network].
See the [Rosetta API] docs for information on how to use the API.

Oasis-specific Rosetta API information is given in the
[Oasis-specific Information] subsection below.

[Rosetta]: https://www.rosetta-api.org/
[Oasis Network]: https://docs.oasis.io/general/oasis-network/
[Rosetta API]: https://www.rosetta-api.org/docs/welcome.html
[Oasis-specific Information]: #oasis-specific-information

## Building and Testing

To build the server:

```
make
```

To run tests:

```
make test
```

To clean-up:

```
make clean
```

`make test` will automatically download the [Oasis Node] and
[Rosetta CLI], set up a test Oasis network, make some sample transactions,
then run the gateway and validate it using `rosetta-cli`.

[Oasis Node]: https://docs.oasis.io/node/run-your-node/prerequisites/oasis-node/
[Rosetta CLI]: https://github.com/coinbase/rosetta-cli

## Contributing

### Versioning

See our [Versioning] document.

[Versioning]: docs/versioning.md

### Release Process

See our [Release Process] document.

[Release Process]: docs/release-process.md

## Running the Gateway

The gateway connects to an Oasis Node, so make sure you have a running node
first. For more details, see the [Run a Non-validator Node] doc of the
[Run Node Oasis Docs].

Set the `OASIS_NODE_GRPC_ADDR` environment variable to the node's gRPC socket
address (e.g. `unix:/path/to/node/internal.sock`).

Optionally, set the `OASIS_ROSETTA_GATEWAY_PORT` environment variable to the
port that you want the gateway to listen on (default is 8080).

Start the gateway simply by running the executable `oasis-rosetta-gateway`.

[Run a Non-validator Node]:
  https://docs.oasis.io/node/run-your-node/non-validator-node/#configuration
[Run Node Oasis Docs]:
  https://docs.oasis.io/node/

## Offline Mode

The gateway supports an "offline" mode, which enables only a subset of the
[Construction API] and nothing else, but doesn't require a connection to an
Oasis Node.

To enable it, set the environment variable `OASIS_ROSETTA_GATEWAY_OFFLINE_MODE`
to a non-empty value.  You must also set the environment variable
`OASIS_ROSETTA_GATEWAY_OFFLINE_MODE_CHAIN_ID` to the [genesis document's hash]
of the network that you wish to construct transactions for.
In online mode, the genesis document's hash is fetched from the Oasis Node, but
in offline mode there is no connection to an Oasis Node, so it has to be
specified manually.

The only supported endpoints in offline mode are:

```
/construction/{combine,derive,hash,parse,payloads,preprocess}
```

[Construction API]:
  https://www.rosetta-api.org/docs/construction_api_introduction.html
[genesis document's hash]:
  https://docs.oasis.io/core/consensus/genesis#genesis-documents-hash

## Oasis-specific Information

This section describes how Oasis fits into the Rosetta APIs.

### Network Identifier

[Rosetta API documentation][api-network-identifier]

[api-network-identifier]: https://www.rosetta-api.org/docs/api_identifiers.html#network-identifier

For Amber (at time of writing):

```js
{
    "blockchain": "Oasis",
    "network": "c014bda208f670539e8f03016b0dcfe16e0c2ad9a060419d1aad580f5c7ff447"
    /* no sub_network_identifier */
}
```

In general (e.g., for other testnets), the `.network` string is the lowercase
hex encoded SHA-512/256 hash of the CBOR encoded genesis document.

### Account Identifier

[Rosetta API documentation][api-account-identifier]

[api-account-identifier]: https://www.rosetta-api.org/docs/api_identifiers.html#account-identifier

#### General Account

For an account `account_addr`'s (e.g.
`oasis1qzzd6khm3acqskpxlk9vd5044cmmcce78y5l6000`) general account:

```js
{
    "address": account_addr
    /* no sub_account */
    /* no metadata */
}
```

#### Escrow Account

For an account `account_addr`'s (e.g.
`oasis1qzzd6khm3acqskpxlk9vd5044cmmcce78y5l6000`) escrow account:

```js
{
    "address": account_addr,
    "sub_account": {
        "address": "escrow"
        /* no metadata */
    }
    /* no metadata */
}
```

#### Common Pool

For the common pool:

```js
{
    "address": "oasis1qrmufhkkyyf79s5za2r8yga9gnk4t446dcy3a5zm"
    /* no sub_account */
    /* no metadata */
}
```

#### Fee Accumulator

For the fee accumulator:

```js
{
    "address": "oasis1qqnv3peudzvekhulf8v3ht29z4cthkhy7gkxmph5"
    /* no sub_account */
    /* no metadata */
}
```

### Currency

[Rosetta API documentation][api-currency]

[api-currency]: https://www.rosetta-api.org/docs/api_objects.html#currency

#### ROSE

For ROSE:

```js
{
    "symbol": "ROSE",
    "decimals": 9
    /* no metadata */
}
```

### Transaction Intents

Rosetta API documentation on
[/construction/preprocess][api-constructionpreprocess] and
[/construction/payloads][api-constructionpayloads].

[api-constructionpreprocess]:
    https://www.rosetta-api.org/docs/ConstructionApi.html#constructionpreprocess
[api-constructionpayloads]:
    https://www.rosetta-api.org/docs/ConstructionApi.html#constructionpayloads

The first two operations in the listings are the gas fee payment.
For zero-fee transactions, omit them and decrease the remaining operation
identifier indices.

#### Staking Transfer

For transfer, `amount_bu` base units from `signer_addr` to `to_addr` with gas
limit `gas_limit` and fee `fee_bu` base units:

```js
[
    {
        "operation_identifier": {
            "index": 0
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        },
        /* no coin_change */
        "metadata": {
            "fee_gas": gas_limit
        }
    },
    {
        "operation_identifier": {
            "index": 1
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": "oasis1qqnv3peudzvekhulf8v3ht29z4cthkhy7gkxmph5" /* fee accumulator */
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 2
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + amount_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 3
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": to_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": amount_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    }
]
```

#### Staking Burn

For burn, `amount_bu` base units from `signer_addr` with gas limit `gas_limit`
and fee `fee_bu` base units:

```js
[
    {
        "operation_identifier": {
            "index": 0
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        },
        /* no coin_change */
        "metadata": {
            "fee_gas": gas_limit
        }
    },
    {
        "operation_identifier": {
            "index": 1
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": "oasis1qqnv3peudzvekhulf8v3ht29z4cthkhy7gkxmph5" /* fee accumulator */
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 2
            /* no network_index */
        },
        /* no related_operations */
        "type": "Burn",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + amount_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    }
]
```

#### Staking Add Escrow

For add escrow, `amount_bu` base units from `signer_addr` to `escrow_addr` with
gas limit `gas_limit` and fee `fee_bu` base units:

```js
[
    {
        "operation_identifier": {
            "index": 0
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        },
        /* no coin_change */
        "metadata": {
            "fee_gas": gas_limit
        }
    },
    {
        "operation_identifier": {
            "index": 1
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": "oasis1qqnv3peudzvekhulf8v3ht29z4cthkhy7gkxmph5" /* fee accumulator */
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 2
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + amount_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 3
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": escrow_addr,
            "sub_account": {
                "address": "escrow"
                /* no metadata */
            }
            /* no metadata */
        },
        "amount": {
            "value": amount_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    }
]
```

#### Staking Reclaim Escrow

For transfer, `amount_sh` shares to `signer_addr` from `escrow_addr` with gas
limit `gas_limit` and fee `fee_bu` base units:

```js
[
    {
        "operation_identifier": {
            "index": 0
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": "-" + fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        },
        /* no coin_change */
        "metadata": {
            "fee_gas": gas_limit
        }
    },
    {
        "operation_identifier": {
            "index": 1
            /* no network_index */
        },
        /* no related_operations */
        "type": "Transfer",
        /* no status */
        "account": {
            "address": "oasis1qqnv3peudzvekhulf8v3ht29z4cthkhy7gkxmph5" /* fee accumulator */
            /* no sub_account */
            /* no metadata */
        },
        "amount": {
            "value": fee_bu.toString(),
            "currency": {
                "symbol": "ROSE",
                "decimals": 9
                /* no metadata */
            }
            /* no metadata */
        }
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 2
            /* no network_index */
        },
        /* no related_operations */
        "type": "ReclaimEscrow",
        /* no status */
        "account": {
            "address": signer_addr
            /* no sub_account */
            /* no metadata */
        },
        /* no amount */
        /* no coin_change */
        /* no metadata */
    },
    {
        "operation_identifier": {
            "index": 3
            /* no network_index */
        },
        /* no related_operations */
        "type": "ReclaimEscrow",
        /* no status */
        "account": {
            "address": escrow_addr,
            "sub_account": {
                "address": "escrow"
                /* no metadata */
            }
            /* no metadata */
        },
        /* no amount */
        /* no coin_change */
        "metadata": {
            "reclaim_escrow_shares": amount_sh.toString()
        }
    }
]
```

### Block API

[Rosetta API documentation][api-block]

In a [partial block identifier]:

* Set only the `index` field to the block height.

In a [block response]:

* The `other_transactions` field is absent.

In a [block]:

* Block identifier `index` fields contain the height of the block.
* Block identifier `hash` fields are lowercase hex encoded.
* The `parent_block_identifier` field contains the block identifier of the
  previous block, except when querying for first block.
* The `metadata` field is absent.

In a [transaction]:

* The transaction identifier `hash` field is lowercase hex encoded.
* The `operations` field contains the transaction intent with some
  modifications.
* The `metadata` field is absent.

In an [operation] as compared to the corresponding operation from the
transaction's intent:

* The `related_operations` field may be set.
* The `status` field is set to `OK` for successful transactions and `Failed` for
  failed transactions.
* The `metadata` field is absent.

[api-block]:
  https://www.rosetta-api.org/docs/BlockApi.html#block
[partial block identifier]:
  https://www.rosetta-api.org/docs/models/PartialBlockIdentifier.html
[block response]:
  https://www.rosetta-api.org/docs/models/BlockResponse.html
[block]:
  https://www.rosetta-api.org/docs/models/Block.html
[transaction]:
  https://www.rosetta-api.org/docs/models/Transaction.html
[operation]:
  https://www.rosetta-api.org/docs/models/Operation.html
