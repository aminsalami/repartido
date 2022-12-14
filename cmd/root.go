package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}

var rootCmd = &cobra.Command{
	Use:   "repartido",
	Short: "repartido is a fun distributed cache!",
	Long:  "-Whats your name?\n+What?\n-Whats yourrr name?\n+Tony\n-Fuck you tony!",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("wtf?")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
