// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sdkpath provides functions for locating the Android SDK.
// These functions respect the ANDROID_HOME environment variable, and
// otherwise use the default SDK location.
package sdkpath

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// AndroidHome returns the absolute path of the selected Android SDK,
// if one can be found.
func AndroidHome() (string, error) {
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch runtime.GOOS {
		case "windows":
			// See https://android.googlesource.com/platform/tools/adt/idea/+/85b4bfb7a10ad858a30ffa4003085b54f9424087/native/installer/win/setup_android_studio.nsi#100
			androidHome = filepath.Join(home, "AppData", "Local", "Android", "sdk")
		case "darwin":
			// See https://android.googlesource.com/platform/tools/asuite/+/67e0cd9604379e9663df57f16a318d76423c0aa8/aidegen/lib/ide_util.py#88
			androidHome = filepath.Join(home, "Library", "Android", "sdk")
		default: // Linux, BSDs, etc.
			// See LINUX_ANDROID_SDK_PATH in ide_util.py above.
			androidHome = filepath.Join(home, "Android", "Sdk")
		}
	}
	if info, err := os.Stat(androidHome); err != nil {
		return "", fmt.Errorf("%w; Android SDK was not found at %s", err, androidHome)
	} else if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", androidHome)
	}
	return androidHome, nil
}

// AndroidAPIPath returns an android SDK platform directory within the configured SDK.
// If there are multiple platforms that satisfy the minimum version requirement,
// AndroidAPIPath returns the latest one among them.
func AndroidAPIPath(api int) (string, error) {
	sdk, err := AndroidHome()
	if err != nil {
		return "", err
	}
	sdkDir, err := os.Open(filepath.Join(sdk, "platforms"))
	if err != nil {
		return "", fmt.Errorf("failed to find android SDK platform: %w", err)
	}
	defer sdkDir.Close()
	fis, err := sdkDir.Readdir(-1)
	if err != nil {
		return "", fmt.Errorf("failed to find android SDK platform (API level: %d): %w", api, err)
	}

	var apiPath string
	var apiMajor int
	var apiMinor int
	for _, fi := range fis {
		name := fi.Name()
		major, minor, ok := parseAndroidPlatformVersion(name)
		if !ok || major < api {
			continue
		}
		p := filepath.Join(sdkDir.Name(), name)
		_, err = os.Stat(filepath.Join(p, "android.jar"))
		if err == nil && (major > apiMajor || major == apiMajor && minor > apiMinor) {
			apiPath = p
			apiMajor = major
			apiMinor = minor
		}
	}
	if apiMajor == 0 {
		return "", fmt.Errorf("failed to find android SDK platform (API level: %d) in %s",
			api, sdkDir.Name())
	}
	return apiPath, nil
}

func parseAndroidPlatformVersion(name string) (major int, minor int, ok bool) {
	version, ok := strings.CutPrefix(name, "android-")
	if !ok {
		return 0, 0, false
	}

	majorString, minorString, hasMinor := strings.Cut(version, ".")
	major, err := strconv.Atoi(majorString)
	if err != nil {
		return 0, 0, false
	}
	if !hasMinor {
		return major, 0, true
	}
	minor, err = strconv.Atoi(minorString)
	if err != nil {
		return 0, 0, false
	}
	return major, minor, true
}
