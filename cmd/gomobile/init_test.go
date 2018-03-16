// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

var gopath string

func TestInit(t *testing.T) {
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
	gopath = "/GOPATH1"
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

	err := runInit(cmdInit)
	if err != nil {
		t.Log(buf.String())
		t.Fatal(err)
	}

	diff, err := diffOutput(buf.String(), initTmpl)
	if err != nil {
		t.Fatalf("computing diff failed: %v", err)
	}
	if diff != "" {
		t.Errorf("unexpected output:\n%s", diff)
	}
}

func diffOutput(got string, wantTmpl *template.Template) (string, error) {
	got = filepath.ToSlash(got)

	wantBuf := new(bytes.Buffer)
	data := defaultOutputData()
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

func defaultOutputData() outputData {
	data := outputData{
		GOOS:      goos,
		GOARCH:    goarch,
		GOPATH:    gopath,
		NDKARCH:   ndkarch(),
		Xproj:     projPbxproj,
		Xcontents: contentsJSON,
		Xinfo:     infoplistTmplData{BundleID: "org.golang.todo.basic", Name: "Basic"},
	}
	if goos == "windows" {
		data.EXE = ".exe"
	}
	return data
}

func ndkarch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "x86"
	default:
		return runtime.GOARCH
	}
}

var initTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
rm -r -f "$GOMOBILE"
mkdir -p $GOMOBILE
WORK={{.GOPATH}}/pkg/gomobile/work
go install -x golang.org/x/mobile/cmd/gobind
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=arm --api=15 --install-dir=$GOMOBILE/ndk-toolchains/arm
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=arm64 --api=21 --install-dir=$GOMOBILE/ndk-toolchains/arm64
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=x86 --api=15 --install-dir=$GOMOBILE/ndk-toolchains/x86
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=x86_64 --api=21 --install-dir=$GOMOBILE/ndk-toolchains/x86_64
cp $OPENAL_PATH/include/AL/al.h $GOMOBILE/include/AL/al.h
mkdir -p $GOMOBILE/include/AL
cp $OPENAL_PATH/include/AL/alc.h $GOMOBILE/include/AL/alc.h
mkdir -p $GOMOBILE/include/AL
mkdir -p $WORK/build/armeabi
PWD=$WORK/build/armeabi cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=arm-linux-androideabi
PWD=$WORK/build/armeabi $GOMOBILE/ndk-toolchains/arm/bin/make
cp $WORK/build/armeabi/libopenal.so $GOMOBILE/lib/armeabi-v7a/libopenal.so
mkdir -p $GOMOBILE/lib/armeabi-v7a
mkdir -p $WORK/build/arm64
PWD=$WORK/build/arm64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=aarch64-linux-android
PWD=$WORK/build/arm64 $GOMOBILE/ndk-toolchains/arm64/bin/make
cp $WORK/build/arm64/libopenal.so $GOMOBILE/lib/arm64-v8a/libopenal.so
mkdir -p $GOMOBILE/lib/arm64-v8a
mkdir -p $WORK/build/x86
PWD=$WORK/build/x86 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=i686-linux-android
PWD=$WORK/build/x86 $GOMOBILE/ndk-toolchains/x86/bin/make
cp $WORK/build/x86/libopenal.so $GOMOBILE/lib/x86/libopenal.so
mkdir -p $GOMOBILE/lib/x86
mkdir -p $WORK/build/x86_64
PWD=$WORK/build/x86_64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=x86_64-linux-android
PWD=$WORK/build/x86_64 $GOMOBILE/ndk-toolchains/x86_64/bin/make
cp $WORK/build/x86_64/libopenal.so $GOMOBILE/lib/x86_64/libopenal.so
mkdir -p $GOMOBILE/lib/x86_64
rm -r -f "$WORK"
`))
