package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/InuSDK/InuSDK/internal/prompt"
	versionUtil "github.com/InuSDK/InuSDK/internal/version"

	"github.com/spf13/cobra"
)

var uninstallAll bool
var uninstallForce bool

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall <sdk> [version]",
	Short: "Uninstall an SDK version",
	Long: `Remove an existing Software Development Kit, you can specify the version [--sdkversion <version>] or remove them all [--all]
			  If there any SDK in use, need to use [--all --force]`,
	Run: func(cmd *cobra.Command, args []string) {
		sdk := args[0]

		if len(args) == 1 && strings.Contains(args[0], ".") {
			fmt.Fprintf(os.Stderr, "Error: missing SDK name. Did you mean `inusdk uninstall <sdk> %s`?", args[0])
			os.Exit(1)
		}

		// --all flag
		if uninstallAll {
			versions, err := candidate.InstalledVersions(sdk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
			if len(versions) == 0 {
				fmt.Printf("no versions of %s installed.\n", sdk)
				return
			}

			if !uninstallForce && !prompt.Confirm(fmt.Sprintf("About to delete ALL versions of %s. Continue?", sdk)) {
				fmt.Println("Cancelled")
				return
			}

			for _, _version := range versions {
				if err := candidate.DeleteVersion(sdk, _version); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting %s: %s\n", _version, err)
					continue
				}
				fmt.Printf("Succesfully removed %s ; %s\n", sdk, _version)
			}
			return
		}

		// Specific version
		var version string
		if len(args) == 2 {
			version = args[1]
		}

		// Resolve local version
		if version == "" {
			LocalLatest, err := candidate.LatestInstalled(sdk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("No specified version, uninstalled latest local version: %s\n", LocalLatest)
			version = LocalLatest
		} else if versionUtil.IsMajorOnly(version) {
			versions, _ := candidate.InstalledVersions(sdk)
			var LocalLatest string
			for _, _version := range versions {
				if strings.HasPrefix(_version, version+".") {
					if LocalLatest == "" || compareSemver(_version, LocalLatest) > 0 {
						LocalLatest = _version
					}
				}
			}

			if LocalLatest == "" {
				fmt.Fprintf(os.Stderr, "Error: No installed versions founds for major %s\n", version)
				os.Exit(1)
			}
			fmt.Printf("Major version specified, uninstalled version: %s\n", LocalLatest)
			version = LocalLatest
		}

		// Warn the user if it is active
		active, _ := candidate.ActiveVersion(sdk)
		if active == version {
			fmt.Printf("Warning: %s ; %s is currently active.\n", sdk, version)
		}

		if !uninstallForce && !prompt.Confirm(fmt.Sprintf("About to delete %s ; %s. Continue?", sdk, version)) {
			fmt.Println("Cancelled")
			return
		}

		if err := candidate.DeleteVersion(sdk, version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Succesfully removed %s ; %s\n", sdk, version)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove all installed versions")
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "Skip confirmation prompt, useful for scripting ; by default is set to false")
}
