# TenderSeed

A lightweight seed node for a tendermint p2p network.

Seed nodes maintain an address book of active peers on a tendermint p2p network. New nodes can dial known seeds and request lists of active peers for establishing p2p connections.

This project implementes a lightweight seed node. The lightweight node maintains an address book of active peers, but **does not** relay or store blocks or transactions.

Familiarity with [tendermint network operation](https://tendermint.com/docs/tendermint-core/using-tendermint.html) is a pre-requisite to understanding how to use TenderSeed.

## Quickstart

Build with `make` and start a seed node with the `start` command.

```shell
$ tenderseed start
```

To view your node id (you will need this for other nodes to connect), invoke the `show-node-id` command.

> The first run of tenderseed will generate a node key if one does not exist.

```shell
$ tenderseed show-node-id
```

## Home Dir

All TenderSeed configuration and address book data is stored in the TenderSeed home directory.

The default path is `$HOME/.tenderseed` but you can specify your own path via the `--home` command line argument.

```shell
tenderseed --home /some/path/to/home/dir
```

> The default configuration stores the node key in a `config` folder and the address book in a `data` folder within the home folder.

## Configuration

TenderSeed is configured by a [toml](https://github.com/toml-lang/toml) config file found in the tenderseed [home dir](#Home-Dir) as `config/config.toml`

The seed is configured via a [toml](https://github.com/toml-lang/toml) config file. The default configuration file is shown below.

> A first run of tenderseed will generate a default configuration if one does not exist.

```toml
# path to address book (relative to tendermint-seed home directory or an absolute path)
addr_book_file = "data/addrbook.json"

# Set true for strict routability rules
# Set false for private or local networks
addr_book_strict = true

# network identifier (todo move to cli flag argument? keeps the config network agnostic)
chain_id = "some-chain-id"

# Address to listen for incoming connections
laddr = "tcp://0.0.0.0:26656"

# maximum number of inbound connections
max_num_inbound_peers = 1000

# maximum number of outbound connections
max_num_outbound_peers = 10

# path to node_key (relative to tendermint-seed home directory or an absolute path)
node_key_file = "config/node_key.json"
```

## License

[Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0)
