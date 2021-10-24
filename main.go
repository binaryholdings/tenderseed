package main

import (
	"fmt"
	"path/filepath"

	"os"

	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	"github.com/tendermint/tendermint/version"

	"github.com/mitchellh/go-homedir"
)

type Config struct {
	ListenAddress       string `toml:"laddr" comment:"Address to listen for incoming connections"`
	ChainID             string `toml:"chain_id" comment:"network identifier (todo move to cli flag argument? keeps the config network agnostic)"`
	NodeKeyFile         string `toml:"node_key_file" comment:"path to node_key (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookFile        string `toml:"addr_book_file" comment:"path to address book (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookStrict      bool   `toml:"addr_book_strict" comment:"Set true for strict routability rules\n Set false for private or local networks"`
	MaxNumInboundPeers  int    `toml:"max_num_inbound_peers" comment:"maximum number of inbound connections"`
	MaxNumOutboundPeers int    `toml:"max_num_outbound_peers" comment:"maximum number of outbound connections"`
	Seeds               string `toml:"seeds" comment:"seed nodes we can use to discover peers"`
}

// DefaultConfig returns a seed config initialized with default values
func DefaultConfig() *Config {
	return &Config{
		ListenAddress:       "tcp://0.0.0.0:26656",
		ChainID:             "cosmoshub-4",
		NodeKeyFile:         "config/node_key.json",
		AddrBookFile:        "data/addrbook.json",
		AddrBookStrict:      true,
		MaxNumInboundPeers:  1000,
		MaxNumOutboundPeers: 1000,
		Seeds:               "bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656,366ac852255c3ac8de17e11ae9ec814b8c68bddb@51.15.94.196:26656",
	}
}

func main() {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	homeDir := filepath.Join(userHomeDir, ".tenderseed")
	configFile := "config/config.toml"
	configFilePath := filepath.Join(homeDir, configFile)
	MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	SeedConfig := DefaultConfig()

	Start(*SeedConfig)

}

// MkdirAllPanic invokes os.MkdirAll but panics if there is an error
func MkdirAllPanic(path string, perm os.FileMode) {
	err := os.MkdirAll(path, perm)
	if err != nil {
		panic(err)
	}
}

// Start starts a Tenderseed
func Start(SeedConfig Config) {
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	)

	chainID := SeedConfig.ChainID
	nodeKeyFilePath := SeedConfig.NodeKeyFile
	addrBookFilePath := SeedConfig.AddrBookFile

	MkdirAllPanic(filepath.Dir(nodeKeyFilePath), os.ModePerm)
	MkdirAllPanic(filepath.Dir(addrBookFilePath), os.ModePerm)

	cfg := config.DefaultP2PConfig()
	cfg.AllowDuplicateIP = true

	// allow a lot of inbound peers since we disconnect from them quickly in seed mode
	cfg.MaxNumInboundPeers = 3000

	// keep trying to make outbound connections to exchange peering info
	cfg.MaxNumOutboundPeers = 400

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
	if err != nil {
		panic(err)
	}

	logger.Info("tenderseed",
		"key", nodeKey.ID(),
		"listen", SeedConfig.ListenAddress,
		"chain", chainID,
		"strict-routing", SeedConfig.AddrBookStrict,
		"max-inbound", SeedConfig.MaxNumInboundPeers,
		"max-outbound", SeedConfig.MaxNumOutboundPeers,
	)

	// TODO(roman) expose per-module log levels in the config
	filteredLogger := log.NewFilter(logger, log.AllowInfo())

	protocolVersion :=
		p2p.NewProtocolVersion(
			version.P2PProtocol,
			version.BlockProtocol,
			0,
		)

	// NodeInfo gets info on yhour node
	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: protocolVersion,
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      SeedConfig.ListenAddress,
		Network:         chainID,
		Version:         "0.6.9",
		Channels:        []byte{pex.PexChannel},
		Moniker:         fmt.Sprintf("%s-seed", chainID),
	}

	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeInfo.DefaultNodeID, nodeInfo.ListenAddr))
	if err != nil {
		panic(err)
	}

	transport := p2p.NewMultiplexTransport(nodeInfo, *nodeKey, p2p.MConnConfig(cfg))
	if err := transport.Listen(*addr); err != nil {
		panic(err)
	}

	book := pex.NewAddrBook(addrBookFilePath, SeedConfig.AddrBookStrict)
	book.SetLogger(filteredLogger.With("module", "book"))

	pexReactor := pex.NewReactor(book, &pex.ReactorConfig{
		SeedMode: true,
		Seeds:    tmstrings.SplitAndTrim(SeedConfig.Seeds, ",", " "),
	})
	pexReactor.SetLogger(filteredLogger.With("module", "pex"))

	sw := p2p.NewSwitch(cfg, transport)
	sw.SetLogger(filteredLogger.With("module", "switch"))
	sw.SetNodeKey(nodeKey)
	sw.SetAddrBook(book)
	sw.AddReactor("pex", pexReactor)

	// last
	sw.SetNodeInfo(nodeInfo)

	tmos.TrapSignal(logger, func() {
		logger.Info("shutting down...")
		book.Save()
		err := sw.Stop()
		if err != nil {
			panic(err)
		}
	})

	err = sw.Start()
	if err != nil {
		panic(err)
	}

	sw.Wait()
}
