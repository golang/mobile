// Copyright 2019 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNdkRoot(t *testing.T) {
	home, err := ioutil.TempDir("", "gomobile-test-")
	if err != nil {
		t.Fatal(err)
	}

	homeorig := os.Getenv("ANDROID_HOME")
	os.Unsetenv("ANDROID_HOME")
	ndkhomeorig := os.Getenv("ANDROID_NDK_HOME")
	os.Unsetenv("ANDROID_NDK_HOME")

	defer func() {
		os.Setenv("ANDROID_HOME", homeorig)
		os.Setenv("ANDROID_NDK_HOME", ndkhomeorig)
		os.RemoveAll(home)
	}()

	makeMockNDK := func(path, version, platforms, abis string) string {
		dir := filepath.Join(home, path)
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("couldn't mkdir %q", dir)
		}
		propertiesPath := filepath.Join(dir, "source.properties")
		propertiesData := []byte("Pkg.Revision = " + version)
		if err := os.WriteFile(propertiesPath, propertiesData, 0644); err != nil {
			t.Fatalf("couldn't write source.properties: %v", err)
		}
		metaDir := filepath.Join(dir, "meta")
		if err := os.Mkdir(metaDir, 0755); err != nil {
			t.Fatalf("couldn't mkdir %q", metaDir)
		}
		platformsPath := filepath.Join(metaDir, "platforms.json")
		platformsData := []byte(platforms)
		if err := os.WriteFile(platformsPath, platformsData, 0644); err != nil {
			t.Fatalf("couldn't write platforms.json: %v", err)
		}
		abisPath := filepath.Join(metaDir, "abis.json")
		abisData := []byte(abis)
		if err := os.WriteFile(abisPath, abisData, 0644); err != nil {
			t.Fatalf("couldn't populate abis.json: %v", err)
		}
		return dir
	}

	t.Run("no NDK in the default location", func(t *testing.T) {
		os.Setenv("ANDROID_HOME", home)
		defer os.Unsetenv("ANDROID_HOME")
		if ndk, err := ndkRoot(); err == nil {
			t.Errorf("expected error but got %q", ndk)
		}
	})

	t.Run("NDK location is set but is wrong", func(t *testing.T) {
		os.Setenv("ANDROID_NDK_HOME", filepath.Join(home, "no-such-path"))
		defer os.Unsetenv("ANDROID_NDK_HOME")
		if ndk, err := ndkRoot(); err == nil {
			t.Errorf("expected error but got %q", ndk)
		}
	})

	t.Run("Two NDKs installed", func(t *testing.T) {
		// Default path for pre-side-by-side NDKs.
		sdkNDK := makeMockNDK("ndk-bundle", "fake-version", `{"min":16,"max":32}`, "{}")
		defer os.RemoveAll(sdkNDK)
		// Arbitrary location for testing ANDROID_NDK_HOME.
		envNDK := makeMockNDK("custom-location", "fake-version", `{"min":16,"max":32}`, "{}")

		// ANDROID_NDK_HOME is sufficient.
		os.Setenv("ANDROID_NDK_HOME", envNDK)
		if ndk, err := ndkRoot(); ndk != envNDK {
			t.Errorf("got (%q, %v) want (%q, nil)", ndk, err, envNDK)
		}

		// ANDROID_NDK_HOME takes precedence over ANDROID_HOME.
		os.Setenv("ANDROID_HOME", home)
		if ndk, err := ndkRoot(); ndk != envNDK {
			t.Errorf("got (%q, %v) want (%q, nil)", ndk, err, envNDK)
		}

		// ANDROID_NDK_HOME is respected even if there is no NDK there.
		os.RemoveAll(envNDK)
		if ndk, err := ndkRoot(); err == nil {
			t.Errorf("expected error but got %q", ndk)
		}

		// ANDROID_HOME is used if ANDROID_NDK_HOME is not set.
		os.Unsetenv("ANDROID_NDK_HOME")
		if ndk, err := ndkRoot(); ndk != sdkNDK {
			t.Errorf("got (%q, %v) want (%q, nil)", ndk, err, envNDK)
		}
	})

	t.Run("Modern 'side-by-side' NDK selection", func(t *testing.T) {
		defer func() {
			buildAndroidAPI = minAndroidAPI
		}()

		ndkForest := filepath.Join(home, "ndk")
		if err := os.Mkdir(ndkForest, 0755); err != nil {
			t.Fatalf("couldn't mkdir %q", ndkForest)
		}

		path := filepath.Join("ndk", "newer")
		platforms := `{"min":19,"max":32}`
		abis := `{"arm64-v8a": {}, "armeabi-v7a": {}, "x86_64": {}}`
		version := "17.2.0"
		newerNDK := makeMockNDK(path, version, platforms, abis)

		path = filepath.Join("ndk", "older")
		platforms = `{"min":16,"max":31}`
		abis = `{"arm64-v8a": {}, "armeabi-v7a": {}, "x86": {}}`
		version = "17.1.0"
		olderNDK := makeMockNDK(path, version, platforms, abis)

		testCases := []struct {
			api         int
			targets     []targetInfo
			wantNDKRoot string
		}{
			{15, nil, ""},
			{16, nil, olderNDK},
			{16, []targetInfo{{"android", "arm"}}, olderNDK},
			{16, []targetInfo{{"android", "arm"}, {"android", "arm64"}}, olderNDK},
			{16, []targetInfo{{"android", "x86_64"}}, ""},
			{19, nil, newerNDK},
			{19, []targetInfo{{"android", "arm"}}, newerNDK},
			{19, []targetInfo{{"android", "arm"}, {"android", "arm64"}, {"android", "386"}}, olderNDK},
			{32, nil, newerNDK},
			{32, []targetInfo{{"android", "arm"}}, newerNDK},
			{32, []targetInfo{{"android", "386"}}, ""},
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				buildAndroidAPI = tc.api
				ndk, err := ndkRoot(tc.targets...)
				if len(tc.wantNDKRoot) != 0 {
					if ndk != tc.wantNDKRoot || err != nil {
						t.Errorf("got (%q, %v), want (%q, nil)", ndk, err, tc.wantNDKRoot)
					}
				} else if err == nil {
					t.Error("expected error")
				}
			})
		}
	})
}
