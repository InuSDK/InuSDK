package bucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/InuSDK/InuSDK/internal/manifest"

	"github.com/spf13/viper"
)

type Bucket struct {
	Name string
	URL  string
}

func ListSDKs(_bucket Bucket) ([]string, error) {
	// Fetch the bucke tindex
	url := fmt.Sprintf("%s/index.json", _bucket.URL)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Could not fetch SDK list from bucket '%s'", _bucket.Name)
	}

	var sdks []string
	if err := json.NewDecoder(resp.Body).Decode(&sdks); err != nil {
		return nil, err
	}

	return sdks, nil
}

func LatestVersionForMajor(_manifest *manifest.Manifest, major string) (string, error) {
	var latest string

	for version := range _manifest.Versions {
		if !strings.HasPrefix(version, major+".") {
			continue
		}
		if latest == "" {
			latest = version
			continue
		}
		if compareSemver(version, latest) > 0 {
			latest = version
		}
	}

	if latest == "" {
		return "", fmt.Errorf("No versions found for major %s", major)
	}

	return latest, nil
}

func GetBuckets() []Bucket {
	var buckets []Bucket

	viper.UnmarshalKey("buckets", &buckets)
	return buckets
}

func FetchManifest(sdkName string) (*manifest.Manifest, error) {
	buckets := GetBuckets()

	if len(buckets) == 0 {
		return nil, fmt.Errorf("No buckets configured, run `inusdk bucket add <url>`")
	}

	// try each bucket in order until one has the manifest.
	for _, eachBucket := range buckets {
		_manifest, err := fetchFromBucket(eachBucket, sdkName)
		if err != nil {
			continue
		}
		return _manifest, nil
	}

	return nil, fmt.Errorf("SDK '%s' not found in any bucket", sdkName)
}

func fetchFromBucket(_bucket Bucket, sdkName string) (*manifest.Manifest, error) {
	url := fmt.Sprintf("%s/%s.json", _bucket.URL, sdkName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Could not reach bucket %s: %w", _bucket.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("SDK '%s' not found in bucket '%s'", sdkName, _bucket.Name)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bucket '%s' returned %d", _bucket.Name, resp.StatusCode)
	}

	var _manifest manifest.Manifest
	if err := json.NewDecoder(resp.Body).Decode(&_manifest); err != nil {
		return nil, fmt.Errorf("Could not parse manifest: %w", err)
	}

	return &_manifest, nil
}

func LatestVersion(_manifest *manifest.Manifest) (string, error) {
	if len(_manifest.Versions) == 0 {
		return "", fmt.Errorf("No versions available")
	}

	var latest string
	for version := range _manifest.Versions {
		if latest == "" {
			latest = version
			continue
		}
		if compareSemver(version, latest) > 0 {
			latest = version
		}
	}

	return latest, nil
}

func compareSemver(Version_A, Version_B string) int {
	versionA, err1 := parseSemver(Version_A)
	versionB, err2 := parseSemver(Version_B)
	if err1 != nil || err2 != nil {
		if Version_A > Version_B {
			return 1
		}
		return -1
	}

	if versionA[0] != versionB[0] {
		return versionA[0] - versionB[0]
	}
	if versionA[1] != versionB[1] {
		return versionA[1] - versionB[1]
	}
	return versionA[2] - versionB[2]
}

func parseSemver(Version string) ([3]int, error) {
	var major, minor, patch int
	_, err := fmt.Sscanf(Version, "%d.%d.%d", &major, &minor, &patch)
	return [3]int{major, minor, patch}, err
}
