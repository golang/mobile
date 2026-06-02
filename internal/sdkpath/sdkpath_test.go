// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sdkpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAndroidAPIPathAcceptsMinorVersionPlatforms(t *testing.T) {
	sdk := t.TempDir()
	t.Setenv("ANDROID_HOME", sdk)

	for _, name := range []string{"android-36.1", "android-37.0"} {
		writeAndroidPlatform(t, sdk, name)
	}

	got, err := AndroidAPIPath(24)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(sdk, "platforms", "android-37.0")
	if got != want {
		t.Fatalf("AndroidAPIPath(24) = %q, want %q", got, want)
	}
}

func TestAndroidAPIPathAcceptsMajorVersionPlatforms(t *testing.T) {
	sdk := t.TempDir()
	t.Setenv("ANDROID_HOME", sdk)
	writeAndroidPlatform(t, sdk, "android-35")

	got, err := AndroidAPIPath(24)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(sdk, "platforms", "android-35")
	if got != want {
		t.Fatalf("AndroidAPIPath(24) = %q, want %q", got, want)
	}
}

func writeAndroidPlatform(t *testing.T, sdk, name string) {
	t.Helper()

	platform := filepath.Join(sdk, "platforms", name)
	if err := os.MkdirAll(platform, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(platform, "android.jar"), nil, 0666); err != nil {
		t.Fatal(err)
	}
}
