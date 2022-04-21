// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/mobile/internal/importers/java"
	"golang.org/x/mobile/internal/sdkpath"
)

var gomobileBin string

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	// Build gomobile and gobind and put them into PATH.
	binDir, err := ioutil.TempDir("", "bind-java-test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(binDir)
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	if runtime.GOOS != "android" {
		gocmd := filepath.Join(runtime.GOROOT(), "bin", "go")
		gomobileBin = filepath.Join(binDir, "gomobile"+exe)
		gobindBin := filepath.Join(binDir, "gobind"+exe)
		if out, err := exec.Command(gocmd, "build", "-o", gomobileBin, "golang.org/x/mobile/cmd/gomobile").CombinedOutput(); err != nil {
			log.Fatalf("gomobile build failed: %v: %s", err, out)
		}
		if out, err := exec.Command(gocmd, "build", "-o", gobindBin, "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
			log.Fatalf("gobind build failed: %v: %s", err, out)
		}
		path := binDir
		if oldPath := os.Getenv("PATH"); oldPath != "" {
			path += string(filepath.ListSeparator) + oldPath
		}
		os.Setenv("PATH", path)
	}
	return m.Run()
}

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
// This requires the gradle command to be in PATH and the Android SDK to be
// installed.
func runTest(t *testing.T, pkgNames []string, javaPkg, javaCls string) {
	if gomobileBin == "" {
		t.Skipf("no gomobile on %s", runtime.GOOS)
	}
	gradle, err := exec.LookPath("gradle")
	if err != nil {
		t.Skip("command gradle not found, skipping")
	}
	if _, err := sdkpath.AndroidHome(); err != nil {
		t.Skip("Android SDK not found, skipping")
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
	cmd := exec.Command(gomobileBin, args...)
	// Reverse binding doesn't work with Go module since imports starting with Java or ObjC are not valid FQDNs.
	// Disable Go module explicitly until this problem is solved. See golang/go#27234.
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	buf, err := cmd.CombinedOutput()
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
        google()
        jcenter()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:3.1.0'
    }
}

allprojects {
    repositories {
		google()
		jcenter()
	}
}

apply plugin: 'com.android.library'

android {
    compileSdkVersion 'android-19'
    defaultConfig { minSdkVersion 16 }
}

repositories {
    flatDir { dirs 'libs' }
}

dependencies {
    implementation(name: "pkg", ext: "aar")
}
`

const stringsxml = `<?xml version="1.0" encoding="utf-8"?>
<resources>
	<string name="dummy">dummy</string>
</resources>`
