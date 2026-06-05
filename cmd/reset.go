package cmd

import (
	"fmt"
	"os"

	"github.com/InuSDK/InuSDK/internal/setup"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset InuSDK configuration",
	Long:  `Runs a setup wizard in the terminal to configure the PATH once again, also asking if to keep all the SDKs`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := setup.Run(true); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
