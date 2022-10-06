package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
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
	ListenAddress       string   `toml:"laddr" comment:"Address to listen for incoming connections"`
	ChainID             string   `toml:"chain_id" comment:"network identifier (todo move to cli flag argument? keeps the config network agnostic)"`
	NodeKeyFile         string   `toml:"node_key_file" comment:"path to node_key (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookFile        string   `toml:"addr_book_file" comment:"path to address book (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookStrict      bool     `toml:"addr_book_strict" comment:"Set true for strict routability rules\n Set false for private or local networks"`
	MaxNumInboundPeers  int      `toml:"max_num_inbound_peers" comment:"maximum number of inbound connections"`
	MaxNumOutboundPeers int      `toml:"max_num_outbound_peers" comment:"maximum number of outbound connections"`
	Seeds               []string `toml:"seeds" comment:"seed nodes we can use to discover peers"`
	Peers               []string `toml:"persistent_peers" comment:"persistent peers we will always keep connected to"`
}

// DefaultConfig returns a seed config initialized with default values
func DefaultConfig() *Config {
	return &Config{
		ListenAddress:       "tcp://0.0.0.0:6969",
		NodeKeyFile:         "node_key.json",
		AddrBookFile:        "addrbook.json",
		AddrBookStrict:      true,
		MaxNumInboundPeers:  3000,
		MaxNumOutboundPeers: 100,
	}
}

func main() {
	userHomeDir, err := homedir.Dir()
	seedConfig := DefaultConfig()
	if err != nil {
		panic(err)
	}

	chains := getchains()

	var allchains []Chain
	// Get all chains that seeds
	for _, chain := range chains.Chains {
		current := getchain(chain)
		allchains = append(allchains, current)
		if err != nil {
			panic(err)
		}
	}

	port := 6969

	// Seed each chain
	for _, chain := range allchains {
		// increment the port number
		port++
		address := "tcp://0.0.0.0:" + fmt.Sprint(port)

		peers := chain.Peers.PersistentPeers
		seeds := chain.Peers.Seeds
		// make the struct of seeds into a string
		var allseeds []string
		for _, seed := range seeds {
			allseeds = append(allseeds, seed.ID+"@"+seed.Address)
		}

		// allpeers is a slice of peers
		var allpeers []string
		// make the struct of peers into a string
		for _, peer := range peers {
			allpeers = append(allpeers, peer.ID+"@"+peer.Address)
		}

		// set the configuration
		seedConfig.ChainID = chain.ChainID
		seedConfig.Seeds = append(seedConfig.Peers, seedConfig.Seeds...)
		seedConfig.ListenAddress = address

		// init config directory & files
		homeDir := filepath.Join(userHomeDir, configDir+"/"+chain.ChainID, "config")
		configFilePath := filepath.Join(homeDir, "config.toml")
		nodeKeyFilePath := filepath.Join(homeDir, seedConfig.NodeKeyFile)
		addrBookFilePath := filepath.Join(homeDir, seedConfig.AddrBookFile)

		// Make folders
		os.MkdirAll(filepath.Dir(nodeKeyFilePath), os.ModePerm)
		os.MkdirAll(filepath.Dir(addrBookFilePath), os.ModePerm)
		os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm)

		logger.Info("Starting Seed Node for" + chain.ChainID)
		defer Start(*seedConfig)
	}
}

// Start starts a Tenderseed
func Start(seedConfig Config) {
	chainID := seedConfig.ChainID
	cfg := config.DefaultP2PConfig()
	cfg.AllowDuplicateIP = true

	userHomeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	nodeKeyFilePath := filepath.Join(userHomeDir, configDir, "config", seedConfig.NodeKeyFile)
	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
	if err != nil {
		panic(err)
	}

	filteredLogger := log.NewFilter(logger, log.AllowInfo())

	protocolVersion := p2p.NewProtocolVersion(
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
		Seeds:                        seedConfig.Seeds,
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

// getchains() gets the list of chains from the chain registry
func getchains() Chains {
	resp, err := http.Get("https://cosmos-chain.directory/chains")
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var chains Chains

	json.Unmarshal([]byte(body), &chains)
	return chains
}

// getchain() gets one chain's records from the chain registry
func getchain(chainid string) Chain {
	resp, err := http.Get("https://cosmos-chain.directory/chains/" + chainid)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var chain Chain

	json.Unmarshal([]byte(body), &chain)
	return chain
}
