package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/InuSDK/InuSDK/internal/manifest"
)

type AdoptiumAsset struct {
	Binary struct {
		OS           string `json:"os"`
		Architecture string `json:"architecture"`
		Package      struct {
			Link     string `json:"link"`
			Checksum string `json:"checksum"`
		} `json:"package"`
	} `json:"binary"`
}

var archMap = map[string]string{
	"x64":     "amd64",
	"aarch64": "arm64",
	"x32":     "386",
}

var osMap = map[string]string{
	"windows": "windows",
	"linux":   "linux",
	"mac":     "darwin",
}

var binPath = map[string]map[string]string{
	"windows": {
		"amd64": "bin/java.exe",
		"arm64": "bin/java.exe",
	},
	"linux": {
		"amd64": "bin/java",
		"arm64": "bin/java",
	},
	"darwin": {
		"amd64": "Contents/Home/bin/java",
		"arm64": "Contents/Home/bin/java",
	},
}

func generateJava(version string) (*manifest.Manifest, error) {
	fmt.Fprintf(os.Stderr, "Fetching adoptium API for java %s . . .\n", version)

	platforms := []struct{ os, arch string }{
		{"windows", "x64"},
		{"windows", "aarch64"},
		{"linux", "x64"},
		{"linux", "aarch64"},
		{"mac", "x64"},
		{"mac", "aarch64"},
	}

	_manifest := &manifest.Manifest{
		Name:        "java",
		Description: "OpenJDK via Eclipse Temurin",
		Homepage:    "https://adoptium.net",
		Versions:    map[string]manifest.Version{},
	}

	_Version := manifest.Version{
		Windows: map[string]manifest.PlatformBuild{},
		Linux:   map[string]manifest.PlatformBuild{},
		Darwin:  map[string]manifest.PlatformBuild{},
	}

	for _, _platform := range platforms {
		build, err := fetchAdoptium(version, _platform.os, _platform.arch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping %s/%s: %s\n", _platform.os, _platform.arch, err)
			continue
		}

		goOS := osMap[_platform.os]
		goArch := archMap[_platform.arch]
		bin := binPath[goOS][goArch]

		_platformBuild := manifest.PlatformBuild{
			URL:      build.Binary.Package.Link,
			Checksum: "sha256:" + build.Binary.Package.Checksum,
			Bin:      bin,
		}

		switch goOS {
		case "windows":
			_Version.Windows[goArch] = _platformBuild
		case "linux":
			_Version.Linux[goArch] = _platformBuild
		case "darwin":
			_Version.Darwin[goArch] = _platformBuild
		}

		fmt.Fprintf(os.Stderr, "%s/%s\n", goOS, goArch)
	}

	_manifest.Versions[version] = _Version

	return _manifest, nil
}

func fetchAdoptium(version, goos, arch string) (*AdoptiumAsset, error) {
	major := strings.Split(version, ".")[0]

	url := fmt.Sprintf(
		"https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&image_type=jdk&os=%s&vendor=eclipse",
		major, arch, goos,
	)

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var assets []AdoptiumAsset
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, fmt.Errorf("Could not decode responde: %w", err)
	}

	if len(assets) == 0 {
		return nil, fmt.Errorf("No assets found")
	}

	return &assets[0], nil
}
