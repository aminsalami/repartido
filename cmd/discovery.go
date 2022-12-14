package cmd

import (
	"fmt"
	"github.com/aminsalami/repartido/internal/discovery/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(discoveryCommand)
	discoveryCommand.AddCommand(discoveryRun)
}

var discoveryCommand = &cobra.Command{
	Use:   "discovery",
	Short: "manage the \"discovery server\" in repartido",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DiscoveryCommand is running!")
	},
}

var discoveryRun = &cobra.Command{
	Use:   "run",
	Short: "Run the discovery server on port 7100 (default)",
	Run: func(cmd *cobra.Command, args []string) {
		core.StartServer()
	},
}
