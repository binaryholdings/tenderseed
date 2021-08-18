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
tenderseed -seed=bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656,366ac852255c3ac8de17e11ae9ec814b8c68bddb@51.15.94.196:26656 -chain-id cosmoshub-4 start
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

Use flags.  Statefulness bad.



## License

[Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0)
