package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logger = zap.NewExample().Sugar()

var rootCmd = &cobra.Command{
	Use:   "repartido",
	Short: "Repartido is a key-value data store",
	Long:  "Repartido is a key-value data store with high availability and replication.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
