package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/binaryholdings/tenderseed/internal/tenderseed"

	"github.com/cometbft/cometbft/p2p"
	"github.com/google/subcommands"
)

// ShowNodeIDArgs for the show-node-id command
type ShowNodeIDArgs struct {
	HomeDir    string
	SeedConfig tenderseed.Config
}

// Name returns the command name
func (*ShowNodeIDArgs) Name() string { return "show-node-id" }

// Synopsis returns a ummary for the command
func (*ShowNodeIDArgs) Synopsis() string { return "show the node id" }

// Usage returns full usage for the command
func (*ShowNodeIDArgs) Usage() string {
	return `show-node-id

Show the node id (public part of the node key).

If a node key does not exist, it will be created and the id shown.
`
}

// SetFlags initializes any command flags
func (args *ShowNodeIDArgs) SetFlags(flagSet *flag.FlagSet) {
}

// Execute runs the command
func (args *ShowNodeIDArgs) Execute(_ context.Context, flagSet *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	nodeKeyFilePath := args.SeedConfig.NodeKeyFile
	if !filepath.IsAbs(nodeKeyFilePath) {
		nodeKeyFilePath = filepath.Join(args.HomeDir, nodeKeyFilePath)
	}

	tenderseed.MkdirAllPanic(filepath.Dir(nodeKeyFilePath), os.ModePerm)

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
	if err != nil {
		panic(err)
	}

	fmt.Println(nodeKey.ID())
	return subcommands.ExitSuccess
}
