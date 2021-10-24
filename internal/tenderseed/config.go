package tenderseed

import (
	"io/ioutil"
	"os"

	toml "github.com/pelletier/go-toml"
)

// Config is a tenderseed configuration
type Config struct {
	ListenAddress       string   `toml:"laddr" comment:"Address to listen for incoming connections"`
	ChainID             string   `toml:"chain_id" comment:"network identifier (todo move to cli flag argument? keeps the config network agnostic)"`
	NodeKeyFile         string   `toml:"node_key_file" comment:"path to node_key (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookFile        string   `toml:"addr_book_file" comment:"path to address book (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookStrict      bool     `toml:"addr_book_strict" comment:"Set true for strict routability rules\n Set false for private or local networks"`
	MaxNumInboundPeers  int      `toml:"max_num_inbound_peers" comment:"maximum number of inbound connections"`
	MaxNumOutboundPeers int      `toml:"max_num_outbound_peers" comment:"maximum number of outbound connections"`
	Seeds               string   `toml:"seeds" comment:"seed nodes we can use to discover peers"`
}

// LoadOrGenConfig loads a seed config from file if the file exists
// If the file does not exist, make a default config, write it to the file
// Return either the loaded config or a default config
func LoadOrGenConfig(filePath string) (*Config, error) {
	config, err := LoadConfigFromFile(filePath)
	if err == nil {
		return config, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// file did not exist
	config = DefaultConfig()
	err = WriteConfigToFile(filePath, *config)
	return config, err
}

// LoadConfigFromFile loads a seed config from a file
func LoadConfigFromFile(file string) (*Config, error) {
	var config Config
	reader, err := os.Open(file)
	if err != nil {
		return &config, err
	}
	decoder := toml.NewDecoder(reader)
	if err := decoder.Decode(&config); err != nil {
		return &config, err
	}

	return &config, nil
}

// WriteConfigToFile writes the seed config to file
func WriteConfigToFile(file string, config Config) error {
	bytes, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bytes, 0600)
	return err
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
