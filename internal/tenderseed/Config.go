package tenderseed

import (
	"io/ioutil"
	"os"

	toml "github.com/pelletier/go-toml"
)

// Config is a tenderseed configuration
//nolint:lll
type Config struct {
	ListenAddress           string   `toml:"laddr" comment:"Address to listen for incoming connections"`
	ChainID                 string   `toml:"chain_id" comment:"network identifier (todo move to cli flag argument? keeps the config network agnostic)"`
	LogLevel                string   `toml:"log_level" comment:"logging level to filter output (\"info\", \"debug\", \"error\" or \"none\")"`
	NodeKeyFile             string   `toml:"node_key_file" comment:"path to node_key (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookFile            string   `toml:"addr_book_file" comment:"path to address book (relative to tendermint-seed home directory or an absolute path)"`
	AddrBookStrict          bool     `toml:"addr_book_strict" comment:"Set true for strict routability rules\n Set false for private or local networks"`
	MaxNumInboundPeers      int      `toml:"max_num_inbound_peers" comment:"maximum number of inbound connections"`
	MaxNumOutboundPeers     int      `toml:"max_num_outbound_peers" comment:"maximum number of outbound connections"`
	MaxPacketMsgPayloadSize int      `toml:"max_packet_msg_payload_size" comment:"maximum size of a message packet payload, in bytes"`
	Seeds                   string   `toml:"seeds" comment:"seed nodes we can use to discover peers"`
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
	config := DefaultConfig()
	reader, err := os.Open(file)
	if err != nil {
		return config, err
	}
	decoder := toml.NewDecoder(reader)
	if err := decoder.Decode(config); err != nil {
		return config, err
	}

	return config, nil
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
		ListenAddress:           "tcp://0.0.0.0:26656",
		ChainID:                 "",
		LogLevel:                "info",
		NodeKeyFile:             "config/node_key.json",
		AddrBookFile:            "data/addrbook.json",
		AddrBookStrict:          true,
		MaxNumInboundPeers:      100,
		MaxNumOutboundPeers:     60,
		MaxPacketMsgPayloadSize: 1024,
		Seeds:                   "",
	}
}
