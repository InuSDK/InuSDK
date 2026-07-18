package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/InuSDK/InuSDK/internal/bucket"
	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/InuSDK/InuSDK/internal/prompt"
	versionUtil "github.com/InuSDK/InuSDK/internal/version"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an SDK",
	Long:  `Command to install an SDK, by default it install the latest compatible version, but can specify the version using --sdkversion`,
	Run: func(cmd *cobra.Command, args []string) {
		sdk := args[0]
		version := ""

		if len(args) == 2 {
			version = args[1]
		}

		fmt.Printf("Fetching manifest for %s. . .\n", sdk)
		_manifest, err := bucket.FetchManifest(sdk)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Resolve the specified version
		if version == "" {
			latest, err := bucket.LatestVersion(_manifest)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("No version specified, installing latest version: %s\n", latest)
			version = latest
		} else if versionUtil.IsMajorOnly(version) {
			SpecifiedMajorVersion, err := bucket.LatestVersionForMajor(_manifest, version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: no versions found for major %s: %s\n", version, err)
				os.Exit(1)
			}
			fmt.Printf("Only major version specified, installing latest: %s\n", SpecifiedMajorVersion)
			version = SpecifiedMajorVersion
		}

		// Check if already installed
		versions, _ := candidate.InstalledVersions(sdk)
		for _, _version := range versions {
			if _version == version {
				if !prompt.Confirm(fmt.Sprintf("%s %s is already installed. Reinstall?", sdk, version)) {
					fmt.Println("Cancelled.")
					return
				}
				break
			}
		}

		// Resolve platform build
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		build, err := _manifest.Resolve(version, goos, goarch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: no build available for %s/%s: %s\n", goos, goarch, err)
			os.Exit(1)
		}

		if !prompt.Confirm(fmt.Sprintf("About to install %s %s. Continue?", sdk, version)) {
			fmt.Println("Cancelled.")
			return
		}

		// Install the SDK
		if err := candidate.Install(sdk, version, build.URL, build.Checksum, build.Bin); err != nil {
			fmt.Fprintf(os.Stderr, "Installation failed %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n%s %s installed successfully\n", sdk, version)
		fmt.Printf("Run `inusdk use %s` to activate it in any project", sdk)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
