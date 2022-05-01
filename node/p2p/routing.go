package p2p

import (
	"context"

	"github.com/ipfs/go-datastore"
	idiscovery "github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"go.uber.org/fx"
)

func ContentRouting(router routing.PeerRouting) routing.ContentRouting {
	return router.(*dht.IpfsDHT)
}

// PeerRouting provides constructor for PeerRouting over DHT.
// Basically, this provides a way to discover peer addresses by respecting public keys.
func PeerRouting(cfg Config) func(routingParams) (routing.PeerRouting, error) {
	return func(params routingParams) (routing.PeerRouting, error) {
		opts := []dht.Option{
			dht.Mode(dht.ModeAuto),
			dht.Datastore(params.DataStore),
			dht.QueryFilter(dht.PublicQueryFilter),
			dht.RoutingTableFilter(dht.PublicRoutingTableFilter),
			dht.BootstrapPeers(dht.GetDefaultBootstrapPeerAddrInfos()...),
		}

		if cfg.Bootstrapper {
			// override options for bootstrapper
			opts = append(opts,
				dht.Mode(dht.ModeServer), // it must accept incoming connections
			)
		}

		d, err := dht.New(WithLifecycle(params.Ctx, params.Lc), params.Host, opts...)
		if err != nil {
			return nil, err
		}
		params.Lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return d.Bootstrap(ctx)
			},
			OnStop: func(context.Context) error {
				return d.Close()
			},
		})

		return d, nil
	}
}

func Discovery(r routing.ContentRouting) idiscovery.Discovery {
	return discovery.NewRoutingDiscovery(r)
}

type routingParams struct {
	fx.In

	Ctx       context.Context
	Lc        fx.Lifecycle
	Host      HostBase
	DataStore datastore.Batching
}
