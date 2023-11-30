package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmstrings "github.com/cometbft/cometbft/libs/strings"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/pex"
	"github.com/cometbft/cometbft/version"
	"github.com/google/subcommands"

	"github.com/binaryholdings/tenderseed/internal/tenderseed"
)

// StartArgs for the start command
type StartArgs struct {
	HomeDir    string
	SeedConfig tenderseed.Config
}

// Name returns the command name
func (*StartArgs) Name() string { return "start" }

// Synopsis returns a ummary for the command
func (*StartArgs) Synopsis() string { return "start tenderseed" }

// Usage returns full usage for the command
func (*StartArgs) Usage() string {
	return `start

start the tenderseed
`
}

// SetFlags initializes any command flags
func (args *StartArgs) SetFlags(flagSet *flag.FlagSet) {
}

// Execute runs the command
func (args *StartArgs) Execute(_ context.Context, flagSet *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	)

	chainID := args.SeedConfig.ChainID
	nodeKeyFilePath := args.SeedConfig.NodeKeyFile
	addrBookFilePath := args.SeedConfig.AddrBookFile

	if !filepath.IsAbs(nodeKeyFilePath) {
		nodeKeyFilePath = filepath.Join(args.HomeDir, nodeKeyFilePath)
	}
	if !filepath.IsAbs(addrBookFilePath) {
		addrBookFilePath = filepath.Join(args.HomeDir, addrBookFilePath)
	}

	tenderseed.MkdirAllPanic(filepath.Dir(nodeKeyFilePath), os.ModePerm)
	tenderseed.MkdirAllPanic(filepath.Dir(addrBookFilePath), os.ModePerm)

	cfg := config.DefaultP2PConfig()
	cfg.AllowDuplicateIP = true

	// allow a lot of inbound peers since we disconnect from them quickly in seed mode
	cfg.MaxNumInboundPeers = args.SeedConfig.MaxNumInboundPeers

	// keep trying to make outbound connections to exchange peering info
	cfg.MaxNumOutboundPeers = args.SeedConfig.MaxNumOutboundPeers

	// allow increasing maximum size of a message packet payload
	// because there are some chains that override this and result in larger payloads
	cfg.MaxPacketMsgPayloadSize = args.SeedConfig.MaxPacketMsgPayloadSize

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
	if err != nil {
		panic(err)
	}

	logger.Info("tenderseed",
		"key", nodeKey.ID(),
		"listen", args.SeedConfig.ListenAddress,
		"chain", chainID,
		"log-level", args.SeedConfig.LogLevel,
		"strict-routing", args.SeedConfig.AddrBookStrict,
		"max-inbound", args.SeedConfig.MaxNumInboundPeers,
		"max-outbound", args.SeedConfig.MaxNumOutboundPeers,
		"max-packet-msg-payload-size", args.SeedConfig.MaxPacketMsgPayloadSize,
	)

	// TODO(roman) expose per-module log levels in the config
	logOption, err := log.AllowLevel(args.SeedConfig.LogLevel)
	if err != nil {
		panic(err)
	}
	filteredLogger := log.NewFilter(logger, logOption)

	protocolVersion :=
		p2p.NewProtocolVersion(
			version.P2PProtocol,
			version.BlockProtocol,
			0,
		)

	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: protocolVersion,
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      args.SeedConfig.ListenAddress,
		Network:         chainID,
		Version:         "0.0.1",
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

	book := pex.NewAddrBook(addrBookFilePath, args.SeedConfig.AddrBookStrict)
	book.SetLogger(filteredLogger.With("module", "book"))

	pexReactor := pex.NewReactor(book, &pex.ReactorConfig{
		SeedMode: true,
		Seeds:    tmstrings.SplitAndTrim(args.SeedConfig.Seeds, ",", " "),
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
	return subcommands.ExitSuccess
}
