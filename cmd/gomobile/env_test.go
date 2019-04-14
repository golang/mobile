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

	homeorig := os.Getenv("ANDROID_HOME")
	ndkhomeorig := os.Getenv("ANDROID_NDK_HOME")
	defer func() {
		os.Setenv("ANDROID_HOME", homeorig)
		os.Setenv("ANDROID_NDK_HOME", ndkhomeorig)
		os.RemoveAll(home)
	}()

	os.Setenv("ANDROID_HOME", home)

	if ndk, err := ndkRoot(); err == nil {
		t.Errorf("expected error but got %q", ndk)
	}

	sdkNDK := filepath.Join(home, "ndk-bundle")
	envNDK := filepath.Join(home, "android-ndk")

	for _, dir := range []string{sdkNDK, envNDK} {
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("couldn't mkdir %q", dir)
		}
	}

	os.Setenv("ANDROID_NDK_HOME", envNDK)

	if ndk, _ := ndkRoot(); ndk != sdkNDK {
		t.Errorf("got %q want %q", ndk, sdkNDK)
	}

	os.Unsetenv("ANDROID_HOME")

	if ndk, _ := ndkRoot(); ndk != envNDK {
		t.Errorf("got %q want %q", ndk, envNDK)
	}

	os.RemoveAll(envNDK)

	if ndk, err := ndkRoot(); err == nil {
		t.Errorf("expected error but got %q", ndk)
	}

	os.Setenv("ANDROID_HOME", home)

	if ndk, _ := ndkRoot(); ndk != sdkNDK {
		t.Errorf("got %q want %q", ndk, sdkNDK)
	}
}
