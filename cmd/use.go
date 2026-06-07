package cmd

import (
	"fmt"
	"os"

	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/InuSDK/InuSDK/internal/prompt"
	"github.com/InuSDK/InuSDK/internal/shim"
	IsMajorOnly "github.com/InuSDK/InuSDK/internal/version" // FUCKING GOPLS WHATEVER THE FUCK SERVER NAME! I DON'T WANT TO FUCKING RENAME MY IMPORT! FUCKING PIEACE OF SHIT
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
		} else if IsMajorOnly.IsMajorOnly(version) {
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

		conflicts := shim.DetectConflicts(sdk)
		if len(conflicts) > 0 {
			fmt.Println("Warning: Found another installation ahead of InuSDK of PATH: ")
			for _, conflict := range conflicts {
				fmt.Printf(" - %s\n", conflict)
			}

			fmt.Println("These may override InuSDK's managed version")
			fmt.Println("To fix ; remove them from your system path or uninstall conflicting versions")
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
