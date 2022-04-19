package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	"github.com/tendermint/tendermint/version"
)

var (
	configDir = ".tinyseed"
	logger    = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

// Config defines the configuration format
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
		ListenAddress:       "tcp://0.0.0.0:6969",
		ChainID:             "osmosis-1",
		NodeKeyFile:         "node_key.json",
		AddrBookFile:        "addrbook.json",
		AddrBookStrict:      true,
		MaxNumInboundPeers:  3000,
		MaxNumOutboundPeers: 1000,
		Seeds:               "1b077d96ceeba7ef503fb048f343a538b2dcdf1b@136.243.218.244:26656,2308bed9e096a8b96d2aa343acc1147813c59ed2@3.225.38.25:26656,085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656,f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656",
	}
}

func main() {
	idOverride := os.Getenv("ID")
	seedOverride := os.Getenv("SEEDS")
	userHomeDir, err := homedir.Dir()
	seedConfig := DefaultConfig()

	if err != nil {
		panic(err)
	}

	// init config directory & files
	homeDir := filepath.Join(userHomeDir, configDir, "config")
	configFilePath := filepath.Join(homeDir, "config.toml")
	nodeKeyFilePath := filepath.Join(homeDir, seedConfig.NodeKeyFile)
	addrBookFilePath := filepath.Join(homeDir, seedConfig.AddrBookFile)

	MkdirAllPanic(filepath.Dir(nodeKeyFilePath), os.ModePerm)
	MkdirAllPanic(filepath.Dir(addrBookFilePath), os.ModePerm)
	MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	if idOverride != "" {
		seedConfig.ChainID = idOverride
	}
	if seedOverride != "" {
		seedConfig.Seeds = seedOverride
	}
	logger.Info("Starting Seed Node...")
	Start(*seedConfig)
}

// MkdirAllPanic invokes os.MkdirAll but panics if there is an error
func MkdirAllPanic(path string, perm os.FileMode) {
	err := os.MkdirAll(path, perm)
	if err != nil {
		panic(err)
	}
}

// Start starts a Tenderseed
func Start(seedConfig Config) {

	chainID := seedConfig.ChainID

	cfg := config.DefaultP2PConfig()
	cfg.AllowDuplicateIP = true

	userHomeDir, err := homedir.Dir()
	nodeKeyFilePath := filepath.Join(userHomeDir, configDir, "config", seedConfig.NodeKeyFile)
	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
	if err != nil {
		panic(err)
	}

	logger.Info("Configuration",
		"key", nodeKey.ID(),
		"node listen", seedConfig.ListenAddress,
		"chain", chainID,
		"strict-routing", seedConfig.AddrBookStrict,
		"max-inbound", seedConfig.MaxNumInboundPeers,
		"max-outbound", seedConfig.MaxNumOutboundPeers,
	)

	filteredLogger := log.NewFilter(logger, log.AllowInfo())

	protocolVersion :=
		p2p.NewProtocolVersion(
			version.P2PProtocol,
			version.BlockProtocol,
			0,
		)

	// NodeInfo gets info on your node
	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: protocolVersion,
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      seedConfig.ListenAddress,
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

	addrBookFilePath := filepath.Join(userHomeDir, configDir, "config", seedConfig.AddrBookFile)
	book := pex.NewAddrBook(addrBookFilePath, seedConfig.AddrBookStrict)
	book.SetLogger(filteredLogger.With("module", "book"))

	pexReactor := pex.NewReactor(book, &pex.ReactorConfig{
		SeedMode:                     true,
		Seeds:                        tmstrings.SplitAndTrim(seedConfig.Seeds, ",", " "),
		SeedDisconnectWaitPeriod:     1 * time.Second, // default is 28 hours, we just want to harvest as many addresses as possible
		PersistentPeersMaxDialPeriod: 0,               // use exponential back-off
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

	go func() {
		// Fire periodically
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				logger.Info("Peers list", "peers", sw.Peers().List())
			}
		}
	}()

	sw.Wait()
}
