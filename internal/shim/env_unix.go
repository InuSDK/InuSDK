//go:build !windows

package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func setEnv(key, value string) error {
	shellConfig := resolveShellConfig()
	if shellConfig == "" {
		return fmt.Errorf("Warning: could not detect shell config")
	}

	content, _ := os.ReadFile(shellConfig)
	lines := strings.Split(string(content), "\n")

	// We remove old JAVA_HOME line if exists to avoid conflicts.
	filtered := []string{}
	for _, line := range lines {
		if !strings.HasPrefix(line, fmt.Sprintf("export %s=", key)) {
			filtered = append(filtered, line)
		}
	}

	filtered = append(filtered, fmt.Sprintf("export %s=\"%s\"", key, value))

	return os.WriteFile(shellConfig, []byte(strings.Join(filtered, "\n")), 0644)
}

func resolveShellConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "Warning: If you getting this ; you got undefined Home directory."
	}

	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".bashrc")
	}
}
