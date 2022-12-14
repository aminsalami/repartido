package cmd

import (
	"fmt"
	"github.com/aminsalami/repartido/internal/discovery/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		viper.SetConfigName("discovery.conf")
		viper.AddConfigPath("./")
		viper.AddConfigPath("/etc/repartido")
		viper.SetConfigType("yaml")

		err := viper.ReadInConfig()
		if err != nil {
			logger.Fatal(err)
		}
		core.StartServer()
	},
}
