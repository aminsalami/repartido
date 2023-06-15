package cmd

import (
	"github.com/aminsalami/repartido/internal/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.SetConfigName("node.conf")
			viper.AddConfigPath("/etc/repartido")
			viper.AddConfigPath("./")
		}
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
		err := viper.BindEnv("initCluster", "INIT_CLUSTER", "INITCLUSTER", "initCluster")
		if err != nil {
			logger.Fatal(err.Error())
		}
		if err := viper.ReadInConfig(); err != nil {
			logger.Fatal(err.Error())
		}

		conf := &node.Config{}
		if err := viper.Unmarshal(conf); err != nil {
			logger.Fatal(err.Error())
		}
		err, validConf := conf.Validate()
		if err != nil {
			logger.Fatal(err.Error())
		}

		node.StartService(validConf)
		node.StartServer(validConf)
	},
}

func init() {
	nodeCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file. default paths './node.conf', '/etc/repartido/node.conf'")
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(startNode)
}
