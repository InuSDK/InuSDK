package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/InuSDK/InuSDK/internal/prompt"
	"github.com/InuSDK/InuSDK/internal/setup"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove InuSDK from your machine",
	Long: `Removes InuSDK manager completely; BUT, the installed SDKs will remain intact.
			You can use --fullremoval to delete every SDK installed along with the sdk manager.`,
	Run: func(cmd *cobra.Command, args []string) {
		baseDir := viper.GetString("base_dir")
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".inusdk")

		fmt.Println("Caution: this will remove InuSDK from your machine")
		fmt.Println("The following will be deleted: ")
		fmt.Printf("- %s (shims + config)\n", configDir)
		fmt.Println("- InuSDK PATH entries")
		fmt.Println("NOTE: Installed SDKs will NOT be deleted unless you use --fullremoval")

		if !prompt.Confirm("\n Continue?") {
			fmt.Println("Cancelled")
			return
		}

		hasErrors := false

		if err := setup.RemoveFromPath(baseDir); err != nil {
			fmt.Fprintf(os.Stderr, "Caution: Could not remove PATH entry: %s\n", err)
			hasErrors = true
		}

		// Remove shims dir
		shimsDir := filepath.Join(baseDir, "shims")
		if err := os.RemoveAll(shimsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Caution: Could not remove shims: %s\n", err)
			hasErrors = true
		}

		// Remove config
		configFile := filepath.Join(configDir, "config.yaml")
		if err := os.Remove(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Caution: could not remove config: %s\n", err)
			hasErrors = true
		}

		// Remove the binary itself when everything else works perfectly.
		if !hasErrors {
			exe, _ := os.Executable()
			fmt.Printf("\nInuSDK has been removed")
			fmt.Printf("You can manually delete binary at: %s\n", exe)
			fmt.Println("Restart your terminal to finish cleanup")
		} else {
			fmt.Println("\nRemoval completed with errors, Some components may need manual cleanup.")
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
