package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "short_worker",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(workerCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.worker.yaml)")

	//viper.SetConfigFile("/configs/cron-worker.yaml")
	//if err := viper.ReadInConfig(); err != nil {
	//	log.Fatalln(err)
	//}
}

func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("cron-worker") // name of config file (without extension)
	viper.SetConfigType("yaml")
	//viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AddConfigPath("../../configs/") // adding home directory as first search path
	//viper.AddConfigPath("/Users/apple/go/src/github.com/kackerx/crontab/configs") // adding home directory as first search path
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
