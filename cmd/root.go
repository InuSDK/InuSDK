package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/InuSDK/InuSDK/internal/setup"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "InuSDK",
	Short: "A Multiplatform Software Development Kit Installer",
	Long: `InuSDK is a multiplatform Software Development Kit installer

	Made for developers by a developer, this CLI provides the comfort of simplicity but also
	the freedom to configure your SDKs as you like. Providing full detail of each installation.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip setup check for reset itself
		if cmd.Name() == "reset" {
			return
		}

		if setup.IsFirstRun() {
			if err := setup.Run(false); err != nil {
				fmt.Fprintf(os.Stderr, "Setup failed: %s\n", err)
				os.Exit(1)
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (Default: $HOME/.InuSDK/config.yaml)")
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configPath := filepath.Join(home, ".inusdk", "config.yaml")

	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config read error: %v\n", err)
	}
}
