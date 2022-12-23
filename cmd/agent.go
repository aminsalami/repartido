package cmd

import (
	"github.com/aminsalami/repartido/internal/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "?",
	Long:  "???",
}

var startAgent = &cobra.Command{
	Use:   "start",
	Short: "start agent",
	Long:  "start agent ...",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigName("agent.conf")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/repartido")
		viper.AddConfigPath("./")

		if err := viper.ReadInConfig(); err != nil {
			logger.Fatal("Cannot read the config file. Make sure agent.conf is available.")
		}

		a := agent.NewDefaultAgent()
		a.Start()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(startAgent)
}
