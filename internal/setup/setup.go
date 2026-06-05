package setup

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	BaseDir  string `mapstructure:"base_dir"`
	Language string `mapstructure:"language"`
}

type BucketConfig struct {
	Name string `mapstructure:"name" yaml:"name"`
	URL  string `mapstructure:"url" yaml:"url"`
}

var reader = bufio.NewReader(os.Stdin)

// function to execute the setup wizard
func Run(isReset bool) error {
	if isReset {
		fmt.Println("Warning: This will reset all InuSDK configuraiton")
		fmt.Println("Installed SDKs will not be deleted, only the configuration will be resetted to default")
		home, _ := os.UserHomeDir()
		oldBaseDir := viper.GetString("base_dir")

		fmt.Println("Proceed with reset? [y/N]")

		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" {
			fmt.Println("Reset cancelled")
			return nil
		}

		if oldBaseDir != "" && answer == "y" {
			fmt.Printf("Current base directory: %s\n\n", oldBaseDir)
			fmt.Println("Want to keep the files [1] or delete them [2]? Recommended [2] for clean state")
			fmt.Println("(Default 1): ")

			sdkChoice, _ := reader.ReadString('\n')
			sdkChoice = strings.TrimSpace(sdkChoice)

			if sdkChoice == "2" {
				fmt.Printf("\nDeleting %s...\n", oldBaseDir)

				entries, err := os.ReadDir(oldBaseDir)
				if err != nil {
					return fmt.Errorf("Could not delete old base directory: %s\n", err)
				}

				hasErrors := false
				for _, entry := range entries {
					path := filepath.Join(oldBaseDir, entry.Name())

					if entry.Name() == "shims" && entry.IsDir() {
						shimEntries, err := os.ReadDir(path)
						if err != nil {
							continue
						}
						for _, shimEntry := range shimEntries {
							if shimEntry.Name() == "inusdk.exe" || shimEntry.Name() == "inusdk" {
								continue
							}
							shimPath := filepath.Join(path, shimEntry.Name())
							if err := os.RemoveAll(shimPath); err != nil {
								fmt.Fprintf(os.Stderr, "Caution: could not delete %s: %s\n", shimEntry.Name(), err)
								hasErrors = true
							}
						}
						continue
					}

					// Skip binary at root level
					if entry.Name() == "inusdk.exe" || entry.Name() == "inusdk" {
						continue
					}

					if err := os.RemoveAll(path); err != nil {
						fmt.Fprintf(os.Stderr, "Caution: could not delete %s: %s\n", entry.Name(), err)
						hasErrors = true
					}
				}

				configFile := filepath.Join(home, ".inusdk", "config.yaml")
				os.Remove(configFile)

				if !hasErrors {
					fmt.Println("Old directory cleaned.")
				} else {
					fmt.Println("Some files could not be deleted, continuing anyway.")
				}

				fmt.Println("Old directory removed")
			} else {
				fmt.Printf("Keeping existing SDKs at %s\n", oldBaseDir)
			}

			if err := RemoveFromPath(oldBaseDir); err != nil {
				fmt.Printf("Could not remove old PATH entry: %s\n", err)
			}
		}
	}

	fmt.Println("\nWelcome to InuSDK")
	fmt.Print("Let's get your set up\n")

	// First, we going to resolve the candiate paths.
	home, err := os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("could not resolve home directory: %w", err)
	}

	options := buildDirOptions(home)

	baseDir, err := promptBaseDir(options)
	if err != nil {
		return fmt.Errorf("could not resolve base directory: %w", err)
	}

	if err := validateDir(baseDir); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	if err := createStructure(baseDir); err != nil {
		return fmt.Errorf("Could not create directory structure: %w", err)
	}

	if err := writeConfig(baseDir, home); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	fmt.Println("\nInuSDK is ready :3")
	fmt.Printf("Base directory: %s\n", baseDir)
	fmt.Println("\nRun `inusdk install <sdk> <version>` to get started")

	return nil
}

func IsFirstRun() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}

	configPath := filepath.Join(home, ".inusdk", "config.yaml")
	_, err = os.Stat(configPath)
	return os.IsNotExist(err)
}

func buildDirOptions(home string) []string {
	primary := filepath.Join(home, ".inusdk")
	options := []string{primary}

	docsDir := resolveDocumentsDir(home)
	if docsDir != "" {
		options = append(options, filepath.Join(docsDir, "InuSDK"))
	}

	options = append(options, "Custom path...")
	return options
}

func promptBaseDir(options []string) (string, error) {
	fmt.Print("Where should InuSDK store SDKs?\n")

	for index, options := range options {
		fmt.Printf(" [%d] %s\n", index+1, options)
	}

	fmt.Print("\nEnter choice (default 1): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// We going to set default option 1
	if input == "" {
		input = "1"
	}

	switch input {
	case "1":
		return options[0], nil
	case "2":
		if len(options) >= 3 {
			return options[1], nil
		}
		return options[0], nil
	default:
		fmt.Print("Enter a custom path: ")
		custom, _ := reader.ReadString('\n')
		custom = strings.TrimSpace(custom)

		if custom == "" {
			return "", fmt.Errorf("No path provided")
		}

		return custom, nil
	}
}

func validateDir(path string) error {
	// if it exists, check it's writable
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}

		// Write test
		testFile := filepath.Join(path, ".inusdk_write_test")

		file_, err := os.Create(testFile)

		if err != nil {
			return fmt.Errorf("Directory is not writable: %w", err)
		}

		file_.Close()

		os.Remove(testFile)
		return nil
	}
	// Even if it doesn't exist, we create it.
	return nil
}

func createStructure(baseDir string) error {
	dirs := []string{
		filepath.Join(baseDir, "candidates"),
		filepath.Join(baseDir, "shims"),
		filepath.Join(baseDir, "buckets"),
		filepath.Join(baseDir, "downloads"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %w", dir, err)
		}
	}

	shimsDir := filepath.Join(baseDir, "shims")
	if err := addToPath(shimsDir); err != nil {
		fmt.Printf("Could not add to PATH automatically: %s\n", err)
		fmt.Printf("Add this manually: %s", shimsDir)
	}

	return nil
}

func writeConfig(BaseDir, home string) error {
	configDir := filepath.Join(home, ".inusdk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.Set("base_dir", BaseDir)
	viper.Set("language", detectLanguage())
	viper.Set("os", runtime.GOOS)

	viper.Set("buckets", []map[string]string{
		{
			"name": "main",
			"url":  "https://raw.githubusercontent.com/InuSDK/inusdk-bucket/master/manifests",
		},
	})
	viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))

	return viper.WriteConfig()
}

func detectLanguage() string {
	lang := os.Getenv("LANG")

	if lang == "" {
		lang = os.Getenv("LANGUAGE")
	}

	if lang == "" {
		return "en"
	}

	return strings.Split(lang, "_")[0]
}
