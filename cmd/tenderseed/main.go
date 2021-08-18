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
	chainID := flag.String("chain-id", "osmosis-1", "Chain ID.  Defaults to osmosis-1.")
	seeds := flag.String("seed", "2e3e3b7703a598024a2fb287587095bc4d14fe52@95.217.196.54:2000,f5be19f84deb843c18e9b612b7987138ba13ac02@5.9.106.185:2000,f9c49739f0641a0a673e7a1e8edc38054fefc840@144.76.183.180:2000,40aafcd9b6959d58dd1c567d9daf2a82a23311cf@162.55.132.230:2000", "")

	seedslice := strings.Split(*seeds, ",")
	// parse top level flags
	flag.Parse()

	configFilePath := filepath.Join(*homeDir, *configFile)
	tenderseed.MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	seedConfig, err := tenderseed.LoadOrGenConfig(configFilePath)
	if err != nil {
		panic(err)
	}

	subcommands.ImportantFlag("home")
	subcommands.ImportantFlag("config")
	subcommands.ImportantFlag("chain-id")
	subcommands.ImportantFlag("seed")

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&cmd.StartArgs{
		HomeDir:    *homeDir,
		SeedConfig: *seedConfig,
		ChainID:    *chainID,
		Seeds:      seedslice,
	}, "")
	subcommands.Register(&cmd.ShowNodeIDArgs{
		HomeDir:    *homeDir,
		SeedConfig: *seedConfig,
	}, "")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
