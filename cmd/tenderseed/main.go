package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

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
	logLevel := flag.String("log-level", "", "logging level to filter output (\"info\", \"debug\", \"error\" or \"none\")")
			     
	// parse top level flags
	flag.Parse()
	
	configFilePath := filepath.Join(*homeDir, *configFile)
	tenderseed.MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	seedConfig, err := tenderseed.LoadOrGenConfig(configFilePath)
	if err != nil {
		panic(err)
	}
	
	// Get chain-id, seeds, log-level from ENV
        env_chainid, env_chainid_ok := os.LookupEnv("TENDERSEED_CHAIN_ID")
        env_seeds, env_seeds_ok := os.LookupEnv("TENDERSEED_SEEDS")
        env_loglevel, env_loglevel_ok := os.LookupEnv("TENDERSEED_LOG_LEVEL")

        // Set chain-id, seeds, log-level from ARGS or ENV
        if *chainID != ""  {
            seedConfig.ChainID = *chainID
        } else if env_chainid_ok {
             seedConfig.ChainID = env_chainid
        }
        if *seeds != "" {
            seedConfig.Seeds = *seeds
	} else if env_seeds_ok {
             seedConfig.Seeds = env_seeds
        }
        if *logLevel != "" {
            seedConfig.LogLevel = *logLevel
	} else if env_loglevel_ok {
             seedConfig.LogLevel = env_loglevel
        }

        if seedConfig.ChainID == "" || seedConfig.Seeds == "" {
            panic("Not set chain-id/seeds")
        }
        
	
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
