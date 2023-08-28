package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"goland-ddns/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "ddns",
	Short: "Update cloudflare ddns for homelab",
}

var (
	cfgFile string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initDependency)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $APPLICATION_DIR/ddns.yaml)")
}

func initDependency() {
	initConfig()
	initLog()
}

func initConfig() {
	if err := config.LoadConfig(cfgFile); err != nil {
		panic("Can't load config from environment")
	}
}

func initLog() {
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}(logger)
	zap.ReplaceGlobals(logger)
}
