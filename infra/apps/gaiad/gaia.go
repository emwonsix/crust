package gaiad

import (
	"bytes"
	"context"
	_ "embed"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/template"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/crust/infra"
	"github.com/CoreumFoundation/crust/infra/targets"
)

var (
	//go:embed run.tmpl
	tmpl              string
	runScriptTemplate = template.Must(template.New("").Parse(tmpl))
)

const (
	// AppType is the type of gaia application.
	AppType infra.AppType = "gaiad"
	// DefaultChainID is the gaia's default chain id.
	DefaultChainID = "gaia-localnet-1"
	// DefaultAccountPrefix is the account prefix used in the gaia.
	DefaultAccountPrefix = "cosmos"

	dockerEntrypoint = "run.sh"
)

// Ports defines ports used by application.
type Ports struct {
	RPC     int `json:"rpc"`
	P2P     int `json:"p2p"`
	GRPC    int `json:"grpc"`
	GRPCWeb int `json:"grpcWeb"`
	PProf   int `json:"pprof"`
}

// DefaultPorts are the default ports listens on.
var DefaultPorts = Ports{
	RPC:     26557,
	P2P:     26556,
	GRPC:    9080,
	GRPCWeb: 9081,
	PProf:   6050,
}

// Config stores gaia app config.
type Config struct {
	Name            string
	HomeDir         string
	ChainID         string
	AccountPrefix   string
	AppInfo         *infra.AppInfo
	Ports           Ports
	RelayerMnemonic string
}

// New creates new gaia app.
func New(config Config) Gaia {
	return Gaia{
		config: config,
	}
}

// Gaia represents gaia.
type Gaia struct {
	config Config
}

// Type returns type of application.
func (g Gaia) Type() infra.AppType {
	return AppType
}

// Name returns name of app.
func (g Gaia) Name() string {
	return g.config.Name
}

// Ports returns port used by the application.
func (g Gaia) Ports() Ports {
	return g.config.Ports
}

// Info returns deployment info.
func (g Gaia) Info() infra.DeploymentInfo {
	return g.config.AppInfo.Info()
}

// ClientContext creates new cored ClientContext.
func (g Gaia) ClientContext() client.Context {
	rpcClient, err := cosmosclient.NewClientFromNode(infra.JoinNetAddr("http", g.Info().HostFromHost, g.Config().Ports.RPC))
	must.OK(err)

	grpcClient, err := grpc.Dial(infra.JoinNetAddr("http", g.Info().HostFromHost, g.Config().Ports.GRPC), grpc.WithInsecure())
	must.OK(err)

	return client.NewContext(client.DefaultContextConfig(), newBasicManager()).
		WithChainID(g.config.ChainID).
		WithRPCClient(rpcClient).
		WithGRPCClient(grpcClient).
		WithKeyring(keyring.NewInMemory()).
		WithBroadcastMode(flags.BroadcastBlock)
}

// HealthCheck checks if chain is ready.
func (g Gaia) HealthCheck(ctx context.Context) error {
	return infra.CheckCosmosNodeHealth(ctx, g.ClientContext(), g.Info())
}

// Deployment returns deployment.
func (g Gaia) Deployment() infra.Deployment {
	return infra.Deployment{
		RunAsUser: true,
		Image:     "gaiad:znet",
		Name:      g.Name(),
		Info:      g.config.AppInfo,
		Volumes: []infra.Volume{
			{
				Source:      g.config.HomeDir,
				Destination: targets.AppHomeDir,
			},
		},
		Ports:       infra.PortsToMap(g.config.Ports),
		PrepareFunc: g.prepare,
		Entrypoint:  filepath.Join(targets.AppHomeDir, dockerEntrypoint),
	}
}

// Config returns the config.
func (g Gaia) Config() Config {
	return g.config
}

func (g Gaia) prepare() error {
	args := struct {
		HomePath        string
		ChainID         string
		RelayerMnemonic string
		RPCLaddr        string
		P2PLaddr        string
		GRPCAddress     string
		GRPCWebAddress  string
		RPCPprofLaddr   string
	}{
		HomePath:        targets.AppHomeDir,
		ChainID:         g.config.ChainID,
		RelayerMnemonic: RelayerMnemonic,
		RPCLaddr:        infra.JoinNetAddrIP("tcp", net.IPv4zero, g.config.Ports.RPC),
		P2PLaddr:        infra.JoinNetAddrIP("tcp", net.IPv4zero, g.config.Ports.P2P),
		GRPCAddress:     infra.JoinNetAddrIP("", net.IPv4zero, g.config.Ports.GRPC),
		GRPCWebAddress:  infra.JoinNetAddrIP("", net.IPv4zero, g.config.Ports.GRPCWeb),
		RPCPprofLaddr:   infra.JoinNetAddrIP("", net.IPv4zero, g.config.Ports.PProf),
	}

	buf := &bytes.Buffer{}
	if err := runScriptTemplate.Execute(buf, args); err != nil {
		return errors.WithStack(err)
	}

	err := os.WriteFile(path.Join(g.config.HomeDir, dockerEntrypoint), buf.Bytes(), 0o777)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func newBasicManager() module.BasicManager {
	return module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
	)
}
