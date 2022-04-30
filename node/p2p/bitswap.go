package p2p

import (
	"context"

	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-bitswap/network"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	"go.uber.org/fx"
)

const (
	// default size of bloom filter in blockStore
	defaultBloomFilterSize = 512 << 10
	// default amount of hash functions defined for bloom filter
	defaultBloomFilterHashes = 7
	// default size of arc cache in blockStore
	defaultARCCacheSize = 64 << 10
)

// DataExchange provides a constructor for IPFS block's DataExchange over BitSwap.
func DataExchange(cfg Config) func(bitSwapParams) (exchange.Interface, blockstore.Blockstore, error) {
	return func(params bitSwapParams) (exchange.Interface, blockstore.Blockstore, error) {
		ctx := WithLifecycle(params.Ctx, params.Lc)
		bs, err := blockstore.CachedBlockstore(
			ctx,
			blockstore.NewBlockstore(params.Ds),
			blockstore.CacheOpts{
				HasBloomFilterSize:   defaultBloomFilterSize,
				HasBloomFilterHashes: defaultBloomFilterHashes,
				HasARCCacheSize:      defaultARCCacheSize,
			},
		)
		if err != nil {
			return nil, nil, err
		}
		prefix := protocol.ID("/wordle")
		return bitswap.New(
			ctx,
			network.NewFromIpfsHost(params.Host, params.Cr, network.Prefix(prefix)),
			bs,
			bitswap.ProvideEnabled(false),
		), bs, nil
	}
}

type bitSwapParams struct {
	fx.In

	Ctx  context.Context
	Lc   fx.Lifecycle
	Host host.Host
	Cr   routing.ContentRouting
	Ds   datastore.Batching
}

// WithLifecycle wraps a context to be canceled when the lifecycle stops.
func WithLifecycle(ctx context.Context, lc fx.Lifecycle) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
	return ctx
}
