package node

import (
	"context"

	"github.com/p2p-games/wordle/node/p2p"
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
	)
}
