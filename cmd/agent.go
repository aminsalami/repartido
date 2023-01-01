package cmd

import (
	"github.com/aminsalami/repartido/internal/agent"
	"github.com/aminsalami/repartido/internal/agent/adaptors/httpHandler"
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

		defaultAgent := agent.NewDefaultAgent()
		host := viper.GetString("agent.host")
		port := viper.GetString("agent.port")
		if host == "" {
			host = "0.0.0.0"
		}
		if port == "" {
			port = "6000"
		}
		h := httpHandler.HttpHandler{
			Addr:  host + ":" + port,
			Agent: &defaultAgent,
		}

		defaultAgent.Start()
		h.Run()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(startAgent)
}
