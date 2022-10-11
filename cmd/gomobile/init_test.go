// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

var gopath string

func TestInit(t *testing.T) {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skipf("not available on %s", runtime.GOOS)
	}

	if _, err := exec.LookPath("diff"); err != nil {
		t.Skip("command diff not found, skipping")
	}
	buf := new(bytes.Buffer)
	gopathorig := os.Getenv("GOPATH")
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		os.Setenv("GOPATH", gopathorig)
	}()
	xout = buf
	buildN = true
	buildX = true

	// Test that first GOPATH element is chosen correctly.
	var err error
	gopath, err = ioutil.TempDir("", "gomobile-test")
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{gopath, "/path2", "/path3"}
	if goos == "windows" {
		gopath = filepath.ToSlash(`C:\GOPATH1`)
		paths = []string{gopath, `C:\PATH2`, `C:\PATH3`}
	}
	os.Setenv("GOPATH", strings.Join(paths, string(os.PathListSeparator)))
	os.Setenv("GOROOT_BOOTSTRAP", "go1.4")
	if goos == "windows" {
		os.Setenv("HOMEDRIVE", "C:")
	}

	emptymod, err := ioutil.TempDir("", "gomobile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(emptymod)

	// Create go.mod, but without Go files.
	f, err := os.Create(filepath.Join(emptymod, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("module example.com/m\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}

	dirs := []struct {
		dir  string
		name string
	}{
		{
			dir:  ".",
			name: "current",
		},
		{
			dir:  emptymod,
			name: "emptymod",
		},
	}
	for _, dir := range dirs {
		dir := dir
		t.Run(dir.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(dir.dir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(wd)

			if err := runInit(cmdInit); err != nil {
				t.Log(buf.String())
				t.Fatal(err)
			}

			if dir.name == "emptymod" {
				return
			}

			diff, err := diffOutput(buf.String(), initTmpl)
			if err != nil {
				t.Fatalf("computing diff failed: %v", err)
			}
			if diff != "" {
				t.Errorf("unexpected output:\n%s", diff)
			}
		})
	}
}

func diffOutput(got string, wantTmpl *template.Template) (string, error) {
	got = filepath.ToSlash(got)

	wantBuf := new(bytes.Buffer)
	data, err := defaultOutputData("")
	if err != nil {
		return "", err
	}
	if err := wantTmpl.Execute(wantBuf, data); err != nil {
		return "", err
	}
	want := wantBuf.String()
	if got != want {
		return diff(got, want)
	}
	return "", nil
}

type outputData struct {
	GOOS      string
	GOARCH    string
	GOPATH    string
	NDKARCH   string
	EXE       string // .extension for executables. (ex. ".exe" for windows)
	Xproj     string
	Xcontents string
	Xinfo     infoplistTmplData
}

func defaultOutputData(teamID string) (outputData, error) {
	projPbxproj := new(bytes.Buffer)
	if err := projPbxprojTmpl.Execute(projPbxproj, projPbxprojTmplData{
		TeamID: teamID,
	}); err != nil {
		return outputData{}, err
	}

	data := outputData{
		GOOS:      goos,
		GOARCH:    goarch,
		GOPATH:    gopath,
		NDKARCH:   archNDK(),
		Xproj:     projPbxproj.String(),
		Xcontents: contentsJSON,
		Xinfo:     infoplistTmplData{BundleID: "org.golang.todo.basic", Name: "Basic"},
	}
	if goos == "windows" {
		data.EXE = ".exe"
	}
	return data, nil
}

var initTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
rm -r -f "$GOMOBILE"
mkdir -p $GOMOBILE
WORK={{.GOPATH}}/pkg/gomobile/work
GOMODCACHE={{.GOPATH}}/pkg/mod go install -x golang.org/x/mobile/cmd/gobind@latest
cp $OPENAL_PATH/include/AL/al.h $GOMOBILE/include/AL/al.h
mkdir -p $GOMOBILE/include/AL
cp $OPENAL_PATH/include/AL/alc.h $GOMOBILE/include/AL/alc.h
mkdir -p $GOMOBILE/include/AL
mkdir -p $WORK/build/armeabi
PWD=$WORK/build/armeabi cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=armv7a-linux-androideabi16
PWD=$WORK/build/armeabi $NDK_PATH/prebuilt/{{.NDKARCH}}/bin/make
cp $WORK/build/armeabi/libopenal.so $GOMOBILE/lib/armeabi-v7a/libopenal.so
mkdir -p $GOMOBILE/lib/armeabi-v7a
mkdir -p $WORK/build/arm64
PWD=$WORK/build/arm64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=aarch64-linux-android21
PWD=$WORK/build/arm64 $NDK_PATH/prebuilt/{{.NDKARCH}}/bin/make
cp $WORK/build/arm64/libopenal.so $GOMOBILE/lib/arm64-v8a/libopenal.so
mkdir -p $GOMOBILE/lib/arm64-v8a
mkdir -p $WORK/build/x86
PWD=$WORK/build/x86 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=i686-linux-android16
PWD=$WORK/build/x86 $NDK_PATH/prebuilt/{{.NDKARCH}}/bin/make
cp $WORK/build/x86/libopenal.so $GOMOBILE/lib/x86/libopenal.so
mkdir -p $GOMOBILE/lib/x86
mkdir -p $WORK/build/x86_64
PWD=$WORK/build/x86_64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=x86_64-linux-android21
PWD=$WORK/build/x86_64 $NDK_PATH/prebuilt/{{.NDKARCH}}/bin/make
cp $WORK/build/x86_64/libopenal.so $GOMOBILE/lib/x86_64/libopenal.so
mkdir -p $GOMOBILE/lib/x86_64
rm -r -f "$WORK"
`))
