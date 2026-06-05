package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/InuSDK/InuSDK/internal/bucket"
	"github.com/InuSDK/InuSDK/internal/candidate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listAvailable bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the installed SDKs",
	Long: `This shows the SDKs installed, using the flag (--available) will show the SDKs that are available to install,
	specifying the name (e.g: inusdk list --available java) will show the list of that specific SDK, otherwise will show every available sdk`,
	Run: func(cmd *cobra.Command, args []string) {
		if listAvailable {
			if len(args) == 1 {
				// inusdk list --available java
				listAvailableVersions(args[0])
			} else {
				// inusdk list --available
				listAvailableSDKs()
			}
			return
		}

		if len(args) == 1 {
			// inusdk list java
			listInstalledVersions(args[0])
		} else {
			// inusdk list
			listAllInstalled()
		}
	},
}

func listAllInstalled() {
	baseDir := viper.GetString("base_dir")
	candidatesDir := filepath.Join(baseDir, "candidates")

	entries, err := os.ReadDir(candidatesDir)
	if err != nil {
		fmt.Println("No SDKs installed")
		return
	}

	if len(entries) == 0 {
		fmt.Println("No SDKs installed.")
		fmt.Println("Run `inusdk install <sdk>` to install one")

		return
	}

	fmt.Println("Installed SDKs:")
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sdk := entry.Name()
		versions, _ := candidate.InstalledVersions(sdk)
		active, _ := candidate.ActiveVersion(sdk)

		fmt.Printf("\n  %s\n", sdk)
		for _, _version := range versions {
			if _version == active {
				fmt.Printf("   %s (active)\n", _version)
			} else {
				fmt.Printf("   %s\n", _version)
			}
		}
	}
}

func listInstalledVersions(sdk string) {
	versions, err := candidate.InstalledVersions(sdk)

	if err != nil || len(versions) == 0 {
		fmt.Printf("No versions of %s installed\n", sdk)
		fmt.Printf("Run `inusdk install %s` to install one.\n", sdk)
		return
	}

	active, _ := candidate.ActiveVersion(sdk)

	fmt.Printf("%s\n", sdk)
	fmt.Println("  Installed:")

	for _, _version := range versions {
		if _version == active {
			fmt.Printf("   %s (active)\n", _version)
		} else {
			fmt.Printf("   %s\n", _version)
		}
	}

	fmt.Printf("\nRun `inusdk list --available %s` to see all available versions.\n", sdk)
}

func listAvailableVersions(sdk string) {
	_manifest, err := bucket.FetchManifest(sdk)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	installed, _ := candidate.InstalledVersions(sdk)
	active, _ := candidate.ActiveVersion(sdk)

	// build installed lookup
	installedMap := map[string]bool{}
	for _, _version := range installed {
		installedMap[_version] = true
	}

	fmt.Printf("%s - %s\n", sdk, _manifest.Description)
	fmt.Printf("Homepage: %s\n\n", _manifest.Homepage)
	fmt.Println("  Available versions:")

	// Sort versions
	versions := make([]string, 0, len(_manifest.Versions))
	for _version := range _manifest.Versions {
		versions = append(versions, _version)
	}

	// Sort descending
	sortVersionsDesc(versions)

	for _, _version := range versions {
		marker := ""
		if installedMap[_version] {
			marker = "installed"
		}

		if _version == active {
			marker = "active"
		}
		fmt.Printf("    %s%s\n", _version, marker)
	}
}

func listAvailableSDKs() {
	buckets := bucket.GetBuckets()
	if len(buckets) == 0 {
		fmt.Println("No buckets configured")
		return
	}

	fmt.Println("available SDKs:")
	for _, _bucket := range buckets {
		fmt.Printf("\n  Bucket: %s\n", _bucket.Name)
		sdks, err := bucket.ListSDKs(_bucket)
		if err != nil {
			fmt.Printf("Could not fetch SDK list: %s\n", err)
			continue
		}
		for _, sdk := range sdks {
			fmt.Printf("    %s\n", sdk)
		}
	}
}

func sortVersionsDesc(versions []string) {
	// Bubble sort by semver components
	for index := 0; index < len(versions); index++ {
		for j_index := index + 1; j_index < len(versions); j_index++ {
			if compareSemver(versions[j_index], versions[index]) > 0 {
				versions[index], versions[j_index] = versions[j_index], versions[index]
			}
		}
	}
}

func compareSemver(a, b string) int {
	var aMajor, aMinor, aPatch int
	var bMajor, bMinor, bPatch int
	fmt.Sscanf(a, "%d.%d.%d", &aMajor, &aMinor, &aPatch)
	fmt.Sscanf(b, "%d.%d.%d", &bMajor, &bMinor, &bPatch)

	if aMajor != bMajor {
		return aMajor - bMajor
	}

	if aMinor != bMinor {
		return aMinor - bMinor
	}

	return aPatch - bPatch
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listAvailable, "available", false, "Show available SDKs from bucket")
}
