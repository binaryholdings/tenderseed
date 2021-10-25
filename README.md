# TinySeed

## This is a fork of Binary Holding's Tenderseed, which is a fork of Polychain's Tenderseed

Familiarity with [Tendermint network operation](https://tendermint.com/docs/tendermint-core/using-tendermint.html) is **NOT** a pre-requisite to understanding how to use TinySeed.

To make it easier to use in Docker on Aakash, everything else has been given a default value.

If you do nothing, eg:

```bash
git clone https://github.com/notional-labs/tinyseed
go mod tidy
go install .
tenderseed
```

Theyn you'll become a seed node on Osmosis-1.  Let's do Cosmoshub-4, shall we?  We've made Osmosis zeroconf, but hey this thing here reads 2 env vars!

```bash
export ID=cosmoshub-4
export SEEDS=bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656,366ac852255c3ac8de17e11ae9ec814b8c68bddb@51.15.94.196:26656
git clone https://github.com/notional-labs/tinyseed
go mod tidy
go install .
tenderseed
```



## Quickstart

Build with `make` and start a seed node with the `start` command.

**This will run with defaults and seed/crawl Osmosis**
```bash
tenderseed start
```



## License

[Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0)
