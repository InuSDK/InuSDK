package candidate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
	"github.com/Masterminds/semver/v3"
)

func InstalledVersions(sdk string) ([]string, error) {
	baseDir := viper.GetString("base_dir")
	sdkDir := filepath.Join(baseDir, "candidates", sdk)

	entries, err := os.ReadDir(sdkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("Could not read candidates dir: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			versions = append(versions, entry.Name())
		}
	}

	sort.Slice(versions, func(A_index, B_index int) bool {
		v_index, err_index := semver.NewVersion(versions[A_index])
		v_jndex, err_jndex := semver.NewVersion(versions[B_index])

		if err_index != nil || err_jndex != nil {
			return versions[A_index] > versions[B_index]
		}

		return v_index.GreaterThan(v_jndex)
	})

	return versions, nil
}

// Function that returns the latest installed version
func LatestInstalled(sdk string) (string, error) {
	versions, err := InstalledVersions(sdk)

	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("No versions of %s installed", sdk)
	}

	return versions[0], nil
}

// Returns the active version from .active file
func ActiveVersion(sdk string) (string, error) {
	baseDir := viper.GetString("base_dir")
	activePath := filepath.Join(baseDir, "candidates", sdk, ".active")

	data, err := os.ReadFile(activePath)
	if err != nil {
		return "", nil // No active version set, return none
	}

	return string(data), nil
}

// Set the active version
func SetActive(sdk, version string) error {
	baseDir := viper.GetString("base_dir")
	activePath := filepath.Join(baseDir, "candidates", sdk, ".active")

	return os.WriteFile(activePath, []byte(version), 0644)
}

// Delete an specific version
func DeleteVersion(sdk, version string) error {
	baseDir := viper.GetString("base_dir")
	versionDir := filepath.Join(baseDir, "candidates", sdk, version)

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("%s %s is not installed", sdk, version)
	}

	return os.RemoveAll(versionDir)
}
