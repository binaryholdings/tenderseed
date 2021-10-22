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

func main() {
	userHomeDir, err := homedir.Dir()
	var capChainID string
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

	capChainID = strings.Title(*chainID)
	// overwrite homedir and configfile with env var if they're set
	if *chainID != "" {
		os_homeDir := os.Getenv(capChainID + "_" + "homeDir")
		os_configFile := os.Getenv(capChainID + "_" + "configFile")
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
	capChainID = strings.Title(*chainID)

	// overwrite config with os evironment var

	os_seeds := os.Getenv("TENDERSEED" + capChainID + "_" + "SEEDS")
	os_maxNumOutBoundPeers := os.Getenv("TENDERSEED" + capChainID + "_" + "MAXNUMOUTBOUNDPEERS")
	os_maxNumInBoundPeers := os.Getenv("TENDERSEED" + capChainID + "_" + "MAXNUMINBOUNDPEERS")
	os_addrBookStrict := os.Getenv("TENDERSEED" + capChainID + "_" + "ADDRBOOKSTRICT")
	os_addrBookFile := os.Getenv("TENDERSEED" + capChainID + "_" + "ADDRBOOKFILE")
	os_listenAddress := os.Getenv("TENDERSEED" + capChainID + "_" + "LISTENADDRESS")
	os_nodeKeyFile := os.Getenv("TENDERSEED" + capChainID + "_" + "NODEKEYFILE")
	os_homeDir := os.Getenv("TENDERSEED" + capChainID + "_" + "HOMEDIR")

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
			fmt.Println("env var " + "TENDERSEED " + capChainID + "_" + "MAXNUMINBOUNDPEERS set to invalid value")
		}
	}
	if os_maxNumOutBoundPeers != "" {
		seedConfig.MaxNumOutboundPeers, err = strconv.Atoi(os_maxNumOutBoundPeers)
		if err != nil {
			fmt.Println("env var " + "TENDERSEED " + capChainID + "_" + "MAXNUMOUTBOUNDPEERS set to invalid value")
		}
	}
	if os_homeDir != "" {
		seedConfig.HomeDir = os_homeDir
	}
	subcommands.ImportantFlag("chain-id")

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&cmd.StartArgs{
		SeedConfig: *seedConfig,
	}, "")
	subcommands.Register(&cmd.ShowNodeIDArgs{
		SeedConfig: *seedConfig,
	}, "")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
