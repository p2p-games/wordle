package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	core "github.com/libp2p/go-libp2p-core"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/p2p-games/wordle/node/p2p"
	"github.com/p2p-games/wordle/wordle"
	"go.uber.org/fx"
)

func baseComponents(cfg *Config, store Store) fx.Option {
	return fx.Options(
		fx.Provide(context.Background),
		fx.Supply(cfg),
		fx.Supply(store.Config),
		fx.Provide(store.Datastore),
		fx.Provide(store.Keystore),
		p2p.Components(cfg.P2P),
		fx.Provide(wordleService),
	)
}

func wordleService(lc fx.Lifecycle, host core.Host, ds datastore.Batching, pubsub *pubsub.PubSub) *wordle.Service {
	serv := wordle.NewService(host, ds, pubsub)
	lc.Append(fx.Hook{
		OnStart: serv.Start,
		OnStop: serv.Stop,
	})
	return serv
}
