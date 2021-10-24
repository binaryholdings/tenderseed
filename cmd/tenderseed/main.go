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
	chainID := os.Getenv("ID")
	seeds := os.Getenv("SEEDS")
	// parse top level flags
	flag.Parse()



	// load from config file first
	configFilePath := filepath.Join(*homeDir, *configFile)
	tenderseed.MkdirAllPanic(filepath.Dir(configFilePath), os.ModePerm)

	seedConfig, err := tenderseed.LoadOrGenConfig(configFilePath)
	if err != nil {
		panic(err)
	}
	

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&cmd.StartArgs{
		HomeDir:    *homeDir,
		SeedConfig: *seedConfig,
	}, "")








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
