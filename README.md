# TenderSeed

## This is a fork of polychainlabs tenderseed repo

A lightweight seed node for a Tendermint p2p network.

Seed nodes maintain an address book of active peers on a Tendermint p2p network. New nodes can dial known seeds and request lists of active peers for establishing p2p connections.

This project implements a lightweight seed node. The lightweight node maintains an address book of active peers, but **does not** relay or store blocks or transactions.

Familiarity with [Tendermint network operation](https://tendermint.com/docs/tendermint-core/using-tendermint.html) is a pre-requisite to understanding how to use TenderSeed.

## Quickstart

Build with `make` and start a seed node with the `start` command.

**This will run with defaults and seed/crawl Osmosis**
```bash
tenderseed start
```

**This will seed/crawl cosmoshub-4**
```bash
tenderseed start --seeds "bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656,366ac852255c3ac8de17e11ae9ec814b8c68bddb@51.15.94.196:26656" --chain-id cosmoshub-4
```

To view your node id (you will need this for other nodes to connect), invoke the `show-node-id` command.

> The first run of Tenderseed will generate a node key if one does not exist.

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

> A first run of Tenderseed will generate a default configuration if one does not exist.

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
