package cored

import (
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/config"
)

func saveTendermintConfig(nodeConfig config.NodeConfig, homeDir string) {
	err := nodeConfig.SavePrivateKeys(homeDir)
	must.OK(err)
	cfg := nodeConfig.TendermintNodeConfig(nil)
	// set addr_book_strict to false so nodes connecting from non-routable hosts are added to address book
	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	cfg.P2P.MaxNumOutboundPeers = 100
	cfg.P2P.MaxNumInboundPeers = 100
	cfg.RPC.MaxSubscriptionClients = 10000
	cfg.RPC.MaxOpenConnections = 10000
	cfg.RPC.GRPCMaxOpenConnections = 10000
	cfg.RPC.MaxSubscriptionsPerClient = 10000
	cfg.Mempool.Size = 50000
	cfg.Mempool.MaxTxsBytes = 5368709120

	must.OK(config.WriteTendermintConfigToFile(filepath.Join(homeDir, config.DefaultNodeConfigPath), cfg))
}
