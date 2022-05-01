package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/minio/blake2b-simd"
	"go.uber.org/fx"
)

// PubSub provides a constructor for PubSub protocol with GossipSub routing.
func PubSub(cfg Config) func(pubSubParams) (*pubsub.PubSub, error) {
	return func(params pubSubParams) (*pubsub.PubSub, error) {
		opts := []pubsub.Option{
			pubsub.WithDiscovery(params.Discovery),
			pubsub.WithMessageIdFn(hashMsgID),
		}

		return pubsub.NewFloodSub(
			WithLifecycle(params.Ctx, params.Lc),
			params.Host,
			opts...,
		)
	}
}

func hashMsgID(m *pubsub_pb.Message) string {
	hash := blake2b.Sum256(m.Data)
	return string(hash[:])
}

type pubSubParams struct {
	fx.In

	Ctx       context.Context
	Lc        fx.Lifecycle
	Host      host.Host
	Discovery discovery.Discovery
}
