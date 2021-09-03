package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/binaryholdings/tenderseed/internal/cmd"
	"github.com/binaryholdings/tenderseed/internal/tenderseed"

	"github.com/google/subcommands"
	"github.com/mitchellh/go-homedir"
)

func main() {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	homeDir := flag.String("home", filepath.Join(userHomeDir, ".tenderseed"), "path to tenderseed home directory")
	configFile := flag.String("config", "config/config.toml", "path to configuration file within home directory")
	chainID := flag.String("chain-id", "", "chain id")
	seeds := flag.String("seeds", "", "comma separated list of seeds")
			     
	// parse top level flags
	flag.Parse()
	
	configFilePath := filepath.Join(*homeDir, *configFile)
	tenderseed.MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	seedConfig, err := tenderseed.LoadOrGenConfig(configFilePath)
	if err != nil {
		panic(err)
	}
	
	// Get and set predefined chain-id, seeds-nodes from ENV or ARGS
	if os.getenv("TENDERSEED_CHAIN_ID"):
            seedConfig.ChainID = os.getenv("TENDERSEED_CHAIN_ID")
	if os.getenv("TENDERSEED_SEEDS"):
            seedConfig.Seeds = os.getenv("TENDERSEED_SEEDS")
        if seeds:
            seedConfig.Seeds = seeds
        if chainID:
            seedConfig.ChainID = chainID
	
	subcommands.ImportantFlag("home")
	subcommands.ImportantFlag("config")
	
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&cmd.StartArgs{
		HomeDir:    *homeDir,
		SeedConfig: *seedConfig,
	}, "")
	subcommands.Register(&cmd.ShowNodeIDArgs{
		HomeDir:    *homeDir,
		SeedConfig: *seedConfig,
	}, "")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
