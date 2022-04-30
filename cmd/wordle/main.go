package main

import (
	"context"
	"os"

	"github.com/p2p-games/wordle/cmd"
	"github.com/p2p-games/wordle/node"

	"github.com/spf13/cobra"
)

func init() {
	lightCmd.AddCommand(
		cmd.Start(node.Light),
		cmd.Init(node.Light),
	)
	fullCmd.AddCommand(
		cmd.Start(node.Full),
		cmd.Init(node.Full),
	)
	rootCmd.AddCommand(
		lightCmd,
		fullCmd,
	)
}

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}

func run() error {
	return rootCmd.ExecuteContext(context.Background())
}

var rootCmd = &cobra.Command{
	Use:  "wordle [light|full]",
	Args: cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var lightCmd = &cobra.Command{
	Use:  "light",
	Args: cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var fullCmd = &cobra.Command{
	Use:  "full",
	Args: cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}
