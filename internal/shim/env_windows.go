//go:build windows

package shim

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func setEnv(key, value string) error {
	_key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err == nil {
		defer _key.Close()
		return _key.SetExpandStringValue(key, value)
	}

	_key, err = registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer _key.Close()
	return _key.SetExpandStringValue(key, value)
}

func GetSystemJavaHome() string {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "CRIT-WARN: Unknown error reading registry"
	}

	defer key.Close()

	val, _, err := key.GetStringValue("JAVA_HOME")
	if err != nil {
		return ""
	}

	return val
}

func CheckJavaHomeConflict(version, baseDir string) {
	systemJavaHome := GetSystemJavaHome()
	javaHome := filepath.Join(baseDir, "candidates", "java", version)

	if systemJavaHome != "" && systemJavaHome != javaHome {
		fmt.Println("Warn: JAVA_HOME is set at system level by another program.")
		fmt.Println("      Run the terminal as administrator and run `inusdk use java <version>` to override it")
	} else {
		fmt.Println("JAVA_HOME updated succesfully")
	}
}
