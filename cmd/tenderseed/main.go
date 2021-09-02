package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/binaryholdings/tenderseed/internal/cmd"
	"github.com/binaryholdings/tenderseed/internal/tenderseed"

	"github.com/google/subcommands"
	"github.com/mitchellh/go-homedir"
)

func getSeedConfig() *tenderseed.Config {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	// Param that must not be in config.toml
	configFile := flag.String("config", "config/config.toml", "path to configuration file within home directory")
	homeDir := flag.String("home", filepath.Join(userHomeDir, ".tenderseed"), "path to tenderseed home directory")

	// Param that can't be set by env var
	chainID := flag.String("chain-id", "", "Chain ID")

	// Param that can be set by both env var and config.toml
	seeds := flag.String("seeds", "", "Comma separated list of seeds.")
	listenAddress := flag.String("listenAddress", "", "Address to listen for incoming connections")
	nodeKeyFile := flag.String("nodeKeyFile", "", "path to node_key (relative to tendermint-seed home directory or an absolute path)")
	addrBookFile := flag.String("addrBookFile", "", "path to address book (relative to tendermint-seed home directory or an absolute path)")
	addrBookStrict := flag.String("addrBookStrict", "", "Set true for strict routability rules\n Set false for private or local networks")
	maxNumInboundPeers := flag.Int("maxNumInboundPeers", -2, "maximum number of inbound connections")
	maxNumOutboundPeers := flag.Int("max_num_outbound_peers", -2, "seed nodes we can use to discover peers")

	flag.Parse()

	// overwrite homedir and configfile with env var if they're set
	if *chainID != "" {
		os_homeDir := os.Getenv(*chainID + "_" + "homeDir")
		os_configFile := os.Getenv(*chainID + "_" + "configFile")
		if os_homeDir != "" {
			homeDir = &os_homeDir
		}
		if os_configFile != "" {
			configFile = &os_configFile
		}
	}
	// load from config file first
	configFilePath := filepath.Join(*homeDir, *configFile)
	tenderseed.MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	seedConfig, err := tenderseed.LoadOrGenConfig(configFilePath)

	// overwrite config with flag
	seedConfig.HomeDir = *homeDir
	if *chainID != "" {
		seedConfig.ChainID = *chainID
	}
	if *seeds != "" {
		//split the seeds after parsing the flags.
		seedSlice := strings.Split(*seeds, ",")
		seedConfig.Seeds = seedSlice
	}
	if *listenAddress != "" {
		seedConfig.ListenAddress = *listenAddress
	}
	if *nodeKeyFile != "" {
		seedConfig.NodeKeyFile = *nodeKeyFile
	}
	if *addrBookFile != "" {
		seedConfig.AddrBookFile = *addrBookFile
	}
	if *addrBookStrict != "" {
		if *addrBookStrict == "true" {
			seedConfig.AddrBookStrict = true
		} else if *addrBookStrict == "false" {
			seedConfig.AddrBookStrict = false
		}
	}
	if *maxNumInboundPeers != -2 {
		seedConfig.MaxNumInboundPeers = *maxNumInboundPeers
	}
	if *maxNumOutboundPeers != -2 {
		seedConfig.MaxNumOutboundPeers = *maxNumOutboundPeers
	}

	if *chainID == "" {
		chainID = &seedConfig.ChainID
	}
	// overwrite config with os evironment var
	os_seeds := os.Getenv(*chainID + "_" + "seeds")
	os_maxNumOutBoundPeers := os.Getenv(*chainID + "_" + "maxNumOutBoundPeers")
	os_maxNumInBoundPeers := os.Getenv(*chainID + "_" + "maxNumInBoundPeers")
	os_addrBookStrict := os.Getenv(*chainID + "_" + "addrBookStrict")
	os_addrBookFile := os.Getenv(*chainID + "_" + "addrBookFile")
	os_listenAddress := os.Getenv(*chainID + "_" + "listenAddress")
	os_nodeKeyFile := os.Getenv(*chainID + "_" + "nodeKeyFile")
	os_homeDir := os.Getenv(*chainID + "_" + "homeDir")

	if os_seeds != "" {
		seedSlice := strings.Split(os_seeds, ",")
		seedConfig.Seeds = seedSlice
	}
	if os_listenAddress != "" {
		seedConfig.ListenAddress = *listenAddress
	}
	if os_nodeKeyFile != "" {
		seedConfig.NodeKeyFile = *nodeKeyFile
	}
	if os_addrBookFile != "" {
		seedConfig.AddrBookFile = *addrBookFile
	}
	if os_addrBookStrict != "" {
		if *addrBookStrict == "true" {
			seedConfig.AddrBookStrict = true
		} else if *addrBookStrict == "false" {
			seedConfig.AddrBookStrict = false
		}
	}
	if os_maxNumInBoundPeers != "" {
		seedConfig.MaxNumInboundPeers, err = strconv.Atoi(os_maxNumInBoundPeers)
		if err != nil {
			fmt.Println("env var " + *chainID + "_" + "maxNumInBoundPeers set to invalid value")
		}
	}
	if os_maxNumOutBoundPeers != "" {
		seedConfig.MaxNumOutboundPeers, err = strconv.Atoi(os_maxNumOutBoundPeers)
		if err != nil {
			fmt.Println("env var " + *chainID + "_" + "maxNumOutBoundPeers set to invalid value")
		}
	}
	if os_homeDir != "" {
		seedConfig.HomeDir = os_homeDir
	}

	return seedConfig
}

func main() {
	// userHomeDir, err := homedir.Dir()
	// if err != nil {
	// 	panic(err)
	// }

	// // Param to get config.toml file
	// homeDir := flag.String("home", filepath.Join(userHomeDir, ".tenderseed"), "path to tenderseed home directory")
	// configFile := flag.String("config", "config/config.toml", "path to configuration file within home directory")
	// chainID := flag.String("chain-id", "osmosis-1", "Chain ID")
	// seeds := flag.String("seeds", "2e3e3b7703a598024a2fb287587095bc4d14fe52@95.217.196.54:2000,f5be19f84deb843c18e9b612b7987138ba13ac02@5.9.106.185:2000,f9c49739f0641a0a673e7a1e8edc38054fefc840@144.76.183.180:2000,40aafcd9b6959d58dd1c567d9daf2a82a23311cf@162.55.132.230:2000", "Comma separated list of seeds.")
	// listenAddress := flag.String("listenAddress", "", "Address to listen for incoming connections")
	// nodeKeyFile := flag.String("nodeKeyFile", "", "path to node_key (relative to tendermint-seed home directory or an absolute path)")
	// addrBookFile := flag.String("addrBookFile", "", "path to address book (relative to tendermint-seed home directory or an absolute path)")
	// addrBookStrict := flag.Bool("addrBookStrict", true, "Set true for strict routability rules\n Set false for private or local networks")
	// maxNumInboundPeers := flag.Int("maxNumInboundPeers", 1000, "maximum number of inbound connections")
	// maxNumOutboundPeers := flag.Int("max_num_outbound_peers", 60, "seed nodes we can use to discover peers")

	subcommands.ImportantFlag("chain-id")

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&cmd.StartArgs{
		SeedConfig: *getSeedConfig(),
	}, "")
	subcommands.Register(&cmd.ShowNodeIDArgs{
		SeedConfig: *getSeedConfig(),
	}, "")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
