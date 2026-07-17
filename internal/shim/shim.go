package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

var jdkBinaries = []string{
	"java", "javac", "jar", "javap", "javadoc",
	"jshell", "jlink", "jmod", "jimage", "jdeps",
}

func Create(sdk, version string) error {
	baseDir := viper.GetString("base_dir")
	shimsDir := filepath.Join(baseDir, "shims")

	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return fmt.Errorf("Warn: Could not create shims dir: %w", err)
	}

	binaries := resolveBinaries(sdk)
	for _, bin := range binaries {
		if runtime.GOOS == "windows" {
			if err := createWindowsShim(bin, sdk, version, baseDir, shimsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warn: Could not create shims for %s: %s\n", bin, err)
			}
		} else {
			if err := createUnixShim(bin, sdk, version, baseDir, shimsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warn: Could not create shim for %s: %s\n", bin, err)
			}
		}
	}

	return nil
}

func resolveBinaries(sdk string) []string {
	switch sdk {
	case "java":
		return jdkBinaries
	default:
		return []string{sdk}
	}
}

func DetectConflicts(sdk string) []string {
	var conflicts []string

	pathEnv := os.Getenv("PATH")
	entries := filepath.SplitList(pathEnv)

	baseDir := viper.GetString("base_dir")
	shimsDir := filepath.Join(baseDir, "shims")

	shimFound := false
	for _, entry := range entries {
		if entry == shimsDir {
			shimFound = true
			continue
		}

		if shimFound {
			continue
		}

		var binName string
		if runtime.GOOS == "windows" {
			binName = sdk + ".exe"
		} else {
			binName = sdk
		}

		binPath := filepath.Join(entry, binName)
		if _, err := os.Stat(binPath); err == nil {
			conflicts = append(conflicts, entry)
		}
	}

	return conflicts
}

func createWindowsShim(binary, sdk, version, baseDir, shimsDir string) error {
	shimPath := filepath.Join(shimsDir, binary+".cmd")
	javaHome := filepath.Join(baseDir, "candidates", sdk, version)

	var realBin string
	if binary == "java" || binary == "javaw" {
		realBin = filepath.Join(javaHome, "bin", binary+".exe")
	} else {
		realBin = filepath.Join(javaHome, "bin", binary+".exe")
	}

	content := fmt.Sprintf(
		"@echo off\r\nset JAVA_HOME=%s\r\nset PATH=%%JAVA_HOME%%\\bin;%%PATH%%\r\n\"%s\" %%*\r\n",
		javaHome, realBin,
	)

	return os.WriteFile(shimPath, []byte(content), 0644)
}

func createUnixShim(binary, sdk, version, baseDir, shimsDir string) error {
	shimPath := filepath.Join(shimsDir, binary)
	javaHome := filepath.Join(baseDir, "candidates", sdk, version)
	realBin := filepath.Join(javaHome, "bin", binary)

	content := fmt.Sprintf("#!/bin/sh\nexport JAVA_HOME=\"%s\"\nexec \"%s\" \"$@\"\n", javaHome, realBin)

	if err := os.WriteFile(shimPath, []byte(content), 0755); err != nil {
		return err
	}

	return os.Chmod(shimPath, 0755)
}

func SetJavaHome(version, baseDir string) error {
	javaHome := filepath.Join(baseDir, "candidates", "java", version)
	return setEnv("JAVA_HOME", javaHome)
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
