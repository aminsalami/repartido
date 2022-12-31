package cmd

import (
	"github.com/aminsalami/repartido/internal/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node/cache server",
	Long:  "Go away!",
}
var startNode = &cobra.Command{
	Use:   "run",
	Short: "start node/cache server",
	Long:  "...",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigName("node.conf")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/repartido")
		viper.AddConfigPath("./")

		if err := viper.ReadInConfig(); err != nil {
			logger.Fatal(err.Error())
		}
		node.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(startNode)
}
