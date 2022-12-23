package cmd

import (
	"github.com/aminsalami/repartido/internal/node"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node/cache server",
	Long:  "Go away!",
}
var startNode = &cobra.Command{
	Use:   "start",
	Short: "start node/cache server",
	Long:  "...",
	Run: func(cmd *cobra.Command, args []string) {
		node.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(startNode)
}
