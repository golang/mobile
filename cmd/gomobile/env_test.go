// Copyright 2019 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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

	homeorig := os.Getenv("ANDROID_SDK_ROOT")
	ndkhomeorig := os.Getenv("ANDROID_NDK_ROOT")
	defer func() {
		os.Setenv("ANDROID_SKD_ROOT", homeorig)
		os.Setenv("ANDROID_NDK_ROOT", ndkhomeorig)
		os.RemoveAll(home)
	}()

	os.Setenv("ANDROID_SDK_ROOT", home)
	os.Setenv("ANDROID_NDK_ROOT", "")

	if ndk, err := ndkRoot(); err == nil {
		t.Errorf("expected error but got %q", ndk)
	}

	sdkNDK := filepath.Join(home, "ndk/21.3.6528147")
	envNDK := filepath.Join(home, "android-ndk")

	for _, dir := range []string{sdkNDK, envNDK} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("couldn't mkdir %q", dir)
		}
	}

	os.Setenv("ANDROID_NDK_ROOT", envNDK)

	if ndk, _ := ndkRoot(); ndk != envNDK {
		t.Errorf("got %q want %q", ndk, envNDK)
	}

	os.Unsetenv("ANDROID_SDK_ROOT")

	if ndk, _ := ndkRoot(); ndk != envNDK {
		t.Errorf("got %q want %q", ndk, envNDK)
	}

	os.RemoveAll(envNDK)

	if ndk, err := ndkRoot(); err == nil {
		t.Errorf("expected error but got %q", ndk)
	}

	os.Setenv("ANDROID_SDK_ROOT", home)

	if ndk, _ := ndkRoot(); ndk != sdkNDK {
		t.Errorf("got %q want %q", ndk, sdkNDK)
	}
}
