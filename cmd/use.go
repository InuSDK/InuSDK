package cmd

import (
	"fmt"
	"os"

	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/InuSDK/InuSDK/internal/prompt"
	"github.com/InuSDK/InuSDK/internal/shim"
	"github.com/spf13/cobra"
)

var useForce bool

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use <sdk> [version]",
	Short: "Use an specific SDK for a project",
	Long:  `Select the SDK and the version [--sdkversion] for a specific project. Can set a default SDK using [--default <default SDK>]`,
	Run: func(cmd *cobra.Command, args []string) {
		sdk := args[0]
		version := ""

		if len(args) == 2 {
			version = args[1]
		}

		// Resolve version
		if version == "" {
			latest, err := candidate.LatestInstalled(sdk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
			version = latest
			fmt.Printf("Caution: No version specified, activating latest: %s\n", version)
		} else {
			// check if version is installed
			versions, _ := candidate.InstalledVersions(sdk)
			installed := false
			for _, _version := range versions {
				if _version == version {
					installed = true
					break
				}
			}

			// Not installed - offer to install it
			if !installed {
				if !prompt.Confirm(fmt.Sprintf("%s %s is not installed. Install it ?", sdk, version)) {
					fmt.Println("Cancelled.")
					return
				}

				// Redirect to install flow
				fmt.Printf("Run `inusdk install %s %s` to install it first.\n", sdk, version)
				return
			}
		}

		// confirm if the user wants to actually use it, if --force enabled, no confirmation required.
		if !useForce && !prompt.Confirm(fmt.Sprintf("Activate %s %s", sdk, version)) {
			fmt.Println("Cancelled")
			return
		}

		// Set active
		if err := candidate.SetActive(sdk, version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Create shim
		if err := shim.Create(sdk, version); err != nil {
			fmt.Fprintf(os.Stderr, "Could not create shim: %s\n", err)
		}

		fmt.Printf("\nNow using %s %s\n", sdk, version)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)

	useCmd.Flags().BoolVar(&useForce, "force", false, "Skip confirmation prompt, useful for scripting ; by default is set to false.")
}
