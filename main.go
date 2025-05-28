package main

import (
	"github.com/clems4ever/lgtm/internal/client"
	"github.com/clems4ever/lgtm/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lgtm",
		Short: "Approve GitHub PRs automatically",
	}

	rootCmd.AddCommand(client.BuildCommand())
	rootCmd.AddCommand(server.BuildCommand())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
