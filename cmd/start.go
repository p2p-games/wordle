package cmd

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/p2p-games/wordle/node"
)

// Start constructs a CLI command to start Node daemon of any type with the given flags.
func Start(tp node.Type) *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		Short: `Starts Node daemon. First stopping signal gracefully stops the Node and second terminates it.
Options passed on start override configuration options only on start and are not persisted in config.`,
		Aliases:      []string{"run", "daemon"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := node.OpenStore("~/.wordle")
			if err != nil {
				return err
			}

			nd, err := node.New(tp, store)
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			err = nd.Start(ctx)
			if err != nil {
				return err
			}

			<-ctx.Done()
			cancel() // ensure we stop reading more signals for start context

			ctx, cancel = signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			err = nd.Stop(ctx)
			if err != nil {
				return err
			}

			return store.Close()
		},
	}
	return cmd
}
