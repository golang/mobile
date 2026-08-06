// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sdkpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAndroidAPIPath(t *testing.T) {
	tests := []struct {
		name      string
		api       int
		platforms []string
		want      string
	}{
		{
			name:      "minor-only versions",
			api:       24,
			platforms: []string{"android-36.1", "android-37.0"},
			want:      "android-37.0",
		},
		{
			name:      "major-only version",
			api:       24,
			platforms: []string{"android-35"},
			want:      "android-35",
		},
		{
			name:      "minor version is newer than major-only version",
			api:       24,
			platforms: []string{"android-35", "android-36", "android-36.1"},
			want:      "android-36.1",
		},
		{
			name:      "higher major-only version is newer than lower minor version",
			api:       24,
			platforms: []string{"android-36.1", "android-37"},
			want:      "android-37",
		},
		{
			name:      "minor version below requested API is ignored",
			api:       37,
			platforms: []string{"android-36.1", "android-37.0"},
			want:      "android-37.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdk := t.TempDir()
			t.Setenv("ANDROID_HOME", sdk)
			for _, name := range tt.platforms {
				writeAndroidPlatform(t, sdk, name)
			}

			got, err := AndroidAPIPath(tt.api)
			if err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(sdk, "platforms", tt.want)
			if got != want {
				t.Fatalf("AndroidAPIPath(%d) = %q, want %q", tt.api, got, want)
			}
		})
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
