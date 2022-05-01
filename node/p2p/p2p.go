package p2p

import (
	"go.uber.org/fx"
)

// Config combines all configuration fields for P2P subsystem.
type Config struct {
	// ListenAddresses - Addresses to listen to on local NIC.
	ListenAddresses []string
	// AnnounceAddresses - Addresses to be announced/advertised for peers to connect to
	AnnounceAddresses []string
	// NoAnnounceAddresses - Addresses the P2P subsystem may know about, but that should not be announced/advertised,
	// as undialable from WAN
	NoAnnounceAddresses []string
	// Bootstrapper is flag telling this node is a bootstrapper.
	Bootstrapper bool
	// ConnManager is a configuration tuple for ConnectionManager.
	ConnManager ConnManagerConfig
}

// DefaultConfig returns default configuration for P2P subsystem.
func DefaultConfig() Config {
	return Config{
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		AnnounceAddresses: []string{},
		NoAnnounceAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip4/127.0.0.1/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		Bootstrapper: false,
		ConnManager:  DefaultConnManagerConfig(),
	}
}

// Components collects all the components and services related to p2p.
func Components(cfg Config) fx.Option {
	return fx.Options(
		fx.Provide(Key),
		fx.Provide(ID),
		fx.Provide(PeerStore),
		fx.Provide(ConnectionManager(cfg)),
		fx.Provide(ConnectionGater),
		fx.Provide(Host(cfg)),
		fx.Provide(RoutedHost),
		fx.Provide(PubSub(cfg)),
		fx.Provide(DataExchange(cfg)),
		fx.Provide(DAG),
		fx.Provide(PeerRouting(cfg)),
		fx.Provide(ContentRouting),
		fx.Provide(Discovery),
		fx.Provide(AddrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Invoke(Listen(cfg.ListenAddresses)),
	)
}
