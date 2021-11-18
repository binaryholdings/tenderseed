package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	"github.com/tendermint/tendermint/version"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

var (
	configDir = ".tinyseed"
	logger    = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

// Config defines the configuration format
type Config struct {
	ListenAddress       string `toml:"laddr" comment:"Address to listen for incoming connections"`
	HttpPort            string `toml:"http_port" comment:"Port for the http server"`
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
		HttpPort:            "3000",
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
	logger.Info("Starting Web Server...")
	StartWebServer(*seedConfig)
	logger.Info("Starting Seed Node...")
	Start(*seedConfig)
}

func StartWebServer(seedConfig Config) {

	// serve static assets
	fs := http.FileServer(http.Dir("./web/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// serve html files
	http.HandleFunc("/", serveTemplate)

	// start web server in non-blocking
	go func() {
		err := http.ListenAndServe(":"+seedConfig.HttpPort, nil)
		logger.Info("HTTP Server started", "port", seedConfig.HttpPort)
		if err != nil {
			panic(err)
		}
	}()
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	index := filepath.Join("./web/templates", "index.html")
	templates := filepath.Join("./web/templates", filepath.Clean(r.URL.Path))
	logger.Info("index", "i", index, "t", templates)

	// Return a 404 if the template doesn't exist
	fileInfo, err := os.Stat(templates)

	if err != nil || fileInfo.IsDir() {
		http.Redirect(w, r,"/index.html", 302)
		return
	}

	tmpl, err := template.ParseFiles(index, templates)
	if err != nil {
		// Log the detailed error
		logger.Error(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
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
		"listen", seedConfig.ListenAddress,
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
		SeedMode: true,
		Seeds:    tmstrings.SplitAndTrim(seedConfig.Seeds, ",", " "),
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
