//go:build !windows

package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RemoveFromPath(oldBaseDir string) error {
	shellConfig := resolveShellConfig()
	if shellConfig == "" {
		return nil
	}

	content, err := os.ReadFile(shellConfig)

	if err != nil {
		return err
	}

	oldShims := filepath.Join(oldBaseDir, "shims")

	lines := strings.Split(string(content), "\n")
	filtered := []string{}
	skip := false

	for _, line := range lines {
		if line == "# InuSDK" {
			skip = true
		}
		if skip && strings.Contains(line, oldShims) {
			skip = false
			continue
		}
		if !skip {
			filtered = append(filtered, line)
		}
	}

	return os.WriteFile(shellConfig, []byte(strings.Join(filtered, "\n")), 0644)
}

func resolveDocumentsDir(home string) string {
	return filepath.Join(home, "Documents")
}

func addToPath(shimsDir string) error {
	shellConfig := resolveShellConfig()
	if shellConfig == "" {
		fmt.Println("Could not detect shell config file. Add this manually: ")
		fmt.Printf("Export path=\"%s:$PATH\"\n", shimsDir)

		return nil
	}

	content, err := os.ReadFile(shellConfig)
	if err == nil && strings.Contains(string(content), shimsDir) {
		fmt.Println("Shims directory already in PATH")
		return nil
	}

	file, err := os.OpenFile(shellConfig, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("Could not open %s: %w", shellConfig, err)
	}

	defer file.Close()

	line := fmt.Sprintf("\n# InuSDK\nexport PATH=\"%s:$PATH\"\n", shimsDir)

	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("Could not write to %s: %w", shellConfig, err)
	}

	fmt.Printf("Added to PATH via %s\n", shellConfig)
	fmt.Println("Restart to your terminal or run: ")
	fmt.Printf("source %s\n", shellConfig)

	return nil
}

func resolveShellConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Detect the current shell
	shell := os.Getenv("SHELL")

	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		return filepath.Join(shell, ".bashrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".bashrc")
	}
}
