package cmd

import (
	"github.com/spf13/cobra"

	"github.com/p2p-games/wordle/node"
)

// Init constructs a CLI command to initialize Celestia Node of any type with the given flags.
func Init(tp node.Type) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialization for Node. Passed flags have persisted effect.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return node.Init("~/.wordle", tp)
		},
	}
	return cmd
}
