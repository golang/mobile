// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/mobile/internal/importers/java"
)

func TestClasses(t *testing.T) {
	if !java.IsAvailable() {
		t.Skipf("java importer is not available")
	}
	runTest(t, []string{
		"golang.org/x/mobile/bind/testdata/testpkg/javapkg",
	}, "", "ClassesTest")
}

func TestCustomPkg(t *testing.T) {
	runTest(t, []string{
		"golang.org/x/mobile/bind/testdata/testpkg",
	}, "org.golang.custompkg", "CustomPkgTest")
}

func TestJavaSeqTest(t *testing.T) {
	runTest(t, []string{
		"golang.org/x/mobile/bind/testdata/testpkg",
		"golang.org/x/mobile/bind/testdata/testpkg/secondpkg",
		"golang.org/x/mobile/bind/testdata/testpkg/simplepkg",
	}, "", "SeqTest")
}

// TestJavaSeqBench runs java test SeqBench.java, with the same
// environment requirements as TestJavaSeqTest.
//
// The benchmarks runs on the phone, so the benchmarkpkg implements
// rudimentary timing logic and outputs benchcmp compatible runtimes
// to logcat. Use
//
// adb logcat -v raw GoLog:* *:S
//
// while running the benchmark to see the results.
func TestJavaSeqBench(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark in short mode.")
	}
	runTest(t, []string{"golang.org/x/mobile/bind/testdata/benchmark"}, "", "SeqBench")
}

// runTest runs the Android java test class specified with javaCls. If javaPkg is
// set, it is passed with the -javapkg flag to gomobile. The pkgNames lists the Go
// packages to bind for the test.
// This requires the gradle command in PATH and
// the Android SDK whose path is available through ANDROID_HOME environment variable.
func runTest(t *testing.T, pkgNames []string, javaPkg, javaCls string) {
	gradle, err := exec.LookPath("gradle")
	if err != nil {
		t.Skip("command gradle not found, skipping")
	}
	if sdk := os.Getenv("ANDROID_HOME"); sdk == "" {
		t.Skip("ANDROID_HOME environment var not set, skipping")
	}
	gomobile, err := exec.LookPath("gomobile")
	if err != nil {
		t.Log("go install gomobile")
		if _, err := run("go install golang.org/x/mobile/cmd/gomobile"); err != nil {
			t.Fatalf("gomobile install failed: %v", err)
		}
		if gomobile, err = exec.LookPath("gomobile"); err != nil {
			t.Fatalf("gomobile install failed: %v", err)
		}
		t.Log("gomobile init")
		start := time.Now()
		if _, err := run(gomobile + " init"); err != nil {
			t.Fatalf("gomobile init failed: %v", err)
		}
		t.Logf("gomobile init took %v", time.Since(start))
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed pwd: %v", err)
	}
	tmpdir, err := ioutil.TempDir("", "bind-java-seq-test-")
	if err != nil {
		t.Fatalf("failed to prepare temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	t.Logf("tmpdir = %s", tmpdir)

	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("failed chdir: %v", err)
	}
	defer os.Chdir(cwd)

	for _, d := range []string{"src/main", "src/androidTest/java/go", "libs", "src/main/res/values"} {
		err = os.MkdirAll(filepath.Join(tmpdir, d), 0700)
		if err != nil {
			t.Fatal(err)
		}
	}

	args := []string{"bind", "-tags", "aaa bbb", "-o", "pkg.aar"}
	if javaPkg != "" {
		args = append(args, "-javapkg", javaPkg)
	}
	args = append(args, pkgNames...)
	buf, err := exec.Command(gomobile, args...).CombinedOutput()
	if err != nil {
		t.Logf("%s", buf)
		t.Fatalf("failed to run gomobile bind: %v", err)
	}

	fname := filepath.Join(tmpdir, "libs", "pkg.aar")
	err = cp(fname, filepath.Join(tmpdir, "pkg.aar"))
	if err != nil {
		t.Fatalf("failed to copy pkg.aar: %v", err)
	}

	fname = filepath.Join(tmpdir, "src/androidTest/java/go/"+javaCls+".java")
	err = cp(fname, filepath.Join(cwd, javaCls+".java"))
	if err != nil {
		t.Fatalf("failed to copy SeqTest.java: %v", err)
	}

	fname = filepath.Join(tmpdir, "src/main/AndroidManifest.xml")
	err = ioutil.WriteFile(fname, []byte(androidmanifest), 0700)
	if err != nil {
		t.Fatalf("failed to write android manifest file: %v", err)
	}

	// Add a dummy string resource to avoid errors from the Android build system.
	fname = filepath.Join(tmpdir, "src/main/res/values/strings.xml")
	err = ioutil.WriteFile(fname, []byte(stringsxml), 0700)
	if err != nil {
		t.Fatalf("failed to write strings.xml file: %v", err)
	}

	fname = filepath.Join(tmpdir, "build.gradle")
	err = ioutil.WriteFile(fname, []byte(buildgradle), 0700)
	if err != nil {
		t.Fatalf("failed to write build.gradle file: %v", err)
	}

	if buf, err := run(gradle + " connectedAndroidTest"); err != nil {
		t.Logf("%s", buf)
		t.Errorf("failed to run gradle test: %v", err)
	}
}

func run(cmd string) ([]byte, error) {
	c := strings.Split(cmd, " ")
	return exec.Command(c[0], c[1:]...).CombinedOutput()
}

func cp(dst, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %v", err)
	}
	defer r.Close()
	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to open destination: %v", err)
	}
	_, err = io.Copy(w, r)
	cerr := w.Close()
	if err != nil {
		return err
	}
	return cerr
}

const androidmanifest = `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
          package="com.example.BindTest"
	  android:versionCode="1"
	  android:versionName="1.0">
</manifest>`

const buildgradle = `buildscript {
    repositories {
        jcenter()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:1.5.0'
    }
}

allprojects {
    repositories { jcenter() }
}

apply plugin: 'com.android.library'

android {
    compileSdkVersion 'android-19'
    buildToolsVersion '21.1.2'
    defaultConfig { minSdkVersion 15 }
}

repositories {
    flatDir { dirs 'libs' }
}
dependencies {
    compile(name: "pkg", ext: "aar")
}
`

const stringsxml = `<?xml version="1.0" encoding="utf-8"?>
<resources>
	<string name="dummy">dummy</string>
</resources>`
