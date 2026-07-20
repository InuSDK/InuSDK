package cmd

import (
	"fmt"
	"os"
	"runtime"
	"slices"

	"github.com/InuSDK/InuSDK/internal/bucket"
	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/InuSDK/InuSDK/internal/prompt"
	"github.com/InuSDK/InuSDK/internal/shim"
	versionUtil "github.com/InuSDK/InuSDK/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var useForce bool

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use <sdk> <version>",
	Short: "Use an specific SDK for a project",
	Long:  `Select the SDK and the version by running "inusdk <sdk name> <sdk version>" for an specific SDK's version. If no version set, it will use the latest installed version`,
	Run: func(cmd *cobra.Command, args []string) {
		sdk := args[0]
		version := ""

		_manifest, err := bucket.FetchManifest(sdk)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting manifest: %s\n", err)
			os.Exit(1)
		}

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
		} else if versionUtil.IsMajorOnly(version) {
			// check if version is installed
			latest, err := bucket.LatestVersionForMajor(_manifest, version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: no versions found for major %s: %s\n", version, err)
				os.Exit(1)
			}
			fmt.Printf("Only major version specified, using the latest patch: %s\n", latest)
			version = latest

			versions, _ := candidate.InstalledVersions(sdk)
			installed := slices.Contains(versions, version)

			// Not installed - offer to install it
			if !installed {
				if !prompt.Confirm(fmt.Sprintf("%s %s is not installed. Install it ?", sdk, version)) {
					fmt.Println("Cancelled.")
					return
				}

				goos := runtime.GOOS
				goarch := runtime.GOARCH
				build, err := _manifest.Resolve(version, goos, goarch)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: no build avaiable for %s/%s: %s\n", goos, goarch, err)
					os.Exit(1)
				}

				// Install directly
				fmt.Printf("Installing %s %s. . .\n", sdk, version)
				if err := candidate.Install(sdk, version, build.URL, build.Checksum, build.Bin); err != nil {
					fmt.Fprintf(os.Stderr, "Installation failed: %s\n", err)
					os.Exit(1)
				}

				fmt.Printf("Succesfully installed\nTo use run `inusdk use %s %s`", sdk, version)
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

		if sdk == "java" {
			baseDir := viper.GetString("base_dir")
			if err := shim.SetJavaHome(version, baseDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warn: Could not set JAVA_HOME: %s\n", err)
			} else {
				fmt.Println("JAVA_HOME set successfully")
			}
		}

		fmt.Printf("\nNow using %s %s\n", sdk, version)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)

	useCmd.Flags().BoolVar(&useForce, "force", false, "Skip confirmation prompt, useful for scripting ; by default is set to false.")
}
