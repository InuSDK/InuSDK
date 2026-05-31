package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

func Create(sdk, version string) error {
	baseDir := viper.GetString("base_dir")
	shimsDir := filepath.Join(baseDir, "shims")

	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return fmt.Errorf("Could not create shims dir: %w", err)
	}

	if runtime.GOOS == "windows" {
		return createWindowsShim(sdk, version, baseDir, shimsDir)
	}
	return createUnixShim(sdk, version, baseDir, shimsDir)
}

func createWindowsShim(sdk, version, baseDir, shimsDir string) error {
	shimPath := filepath.Join(shimsDir, sdk+".cmd")
	realBin := filepath.Join(baseDir, "candidates", sdk, version, getBinPath(sdk))
	javaHome := filepath.Join(baseDir, "candidates", sdk, version)
	content := fmt.Sprintf("@echo off\nset JAVA_HOME=%s\n\"%s\" %%*\n", javaHome, realBin)

	return os.WriteFile(shimPath, []byte(content), 0644)
}

func createUnixShim(sdk, version, baseDir, shimsDir string) error {
	shimPath := filepath.Join(shimsDir, sdk)
	realBin := filepath.Join(baseDir, "candidates", sdk, version, getBinPath(sdk))
	content := fmt.Sprintf("#!/bin/sh\nexec \"%s\" \"$@\"\n", realBin)

	if err := os.WriteFile(shimPath, []byte(content), 0755); err != nil {
		return err
	}

	return os.Chmod(shimPath, 0755)
}

func getBinPath(sdk string) string {
	// Read from .active manifest bin path
	// For now, common defaults
	switch sdk {
	case "java":
		if runtime.GOOS == "windows" {
			return filepath.Join("bin", "java.exe")
		}
		return filepath.Join("bin", "java")
	default:
		if runtime.GOOS == "windows" {
			return filepath.Join("bin", sdk+".exe")
		}
		return filepath.Join("bin", sdk)
	}
}
