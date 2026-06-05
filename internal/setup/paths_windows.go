//go:build windows

package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func RemoveFromPath(oldBaseDir string) error {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)

	if err != nil {
		return fmt.Errorf("Could not open registry: %s", err)
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		return fmt.Errorf("could not read PATH: %s", err)
	}

	oldShims := filepath.Join(oldBaseDir, "shims")

	entries := strings.Split(currentPath, ";")
	filtered := []string{}

	for _, entry := range entries {
		if strings.TrimSpace(entry) != oldShims {
			filtered = append(filtered, entry)
		}
	}

	newPath := strings.Join(filtered, ";")
	if err := key.SetStringValue("PATH", newPath); err != nil {
		return fmt.Errorf("could not update PATH: %s", err)
	}
	return nil
}

func resolveDocumentsDir(_ string) string {
	path, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0)

	if err != nil {
		return ""
	}
	return path
}

func addToPath(shimsDir string) error {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)

	if err != nil {
		return fmt.Errorf("could not open registry: %w", err)
	}

	defer key.Close()

	currentPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		return fmt.Errorf("could not read PATH: %w", err)
	}

	// Check if already in addToPath
	if strings.Contains(currentPath, shimsDir) {
		fmt.Println("Shims directory already in PATH.")
		return nil
	}

	newPath := shimsDir + ";" + currentPath
	if err := key.SetStringValue("PATH", newPath); err != nil {
		return fmt.Errorf("Could not update PATH: %w", err)
	}

	fmt.Println("Added to path succesfully")
	fmt.Println("Restart your terminal for PATH changes to take effect.")

	return nil
}
