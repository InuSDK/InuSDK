package version

import "strings"

func IsMajorOnly(version string) bool {
	parts := strings.Split(version, ".")
	return len(parts) <= 2
}
