// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path/filepath"
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
	paths := []string{"/GOPATH1", "/path2", "/path3"}
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

	tmpl := initTmpl

	diff, err := diffOutput(buf.String(), tmpl)
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
	data := outputData{
		NDK:     ndkVersion,
		GOOS:    goos,
		GOARCH:  goarch,
		GOPATH:  gopath,
		NDKARCH: ndkarch,
	}
	if goos == "windows" {
		data.EXE = ".exe"
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
	NDK     string
	GOOS    string
	GOARCH  string
	GOPATH  string
	NDKARCH string
	EXE     string // .extension for executables. (ex. ".exe" for windows)
}

var initTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
mkdir -p $GOMOBILE/android-{{.NDK}}
WORK=/GOPATH1/pkg/gomobile/work
mkdir -p $GOMOBILE/dl
curl -o$GOMOBILE/dl/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz https://dl.google.com/go/mobile/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
tar xfz $GOMOBILE/dl/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
mkdir -p $GOMOBILE/android-{{.NDK}}/arm/sysroot/usr
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/include $GOMOBILE/android-{{.NDK}}/arm/sysroot/usr/include
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/lib $GOMOBILE/android-{{.NDK}}/arm/sysroot/usr/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin $GOMOBILE/android-{{.NDK}}/arm/bin
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/lib $GOMOBILE/android-{{.NDK}}/arm/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/libexec $GOMOBILE/android-{{.NDK}}/arm/libexec
mkdir -p $GOMOBILE/android-{{.NDK}}/arm/arm-linux-androideabi/bin
ln -s $GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-ld{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/arm-linux-androideabi/bin/ld{{.EXE}}
ln -s $GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-as{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/arm-linux-androideabi/bin/as{{.EXE}}
ln -s $GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/arm-linux-androideabi/bin/gcc{{.EXE}}
ln -s $GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/arm-linux-androideabi/bin/g++{{.EXE}}
mkdir -p $GOMOBILE/dl
curl -o$GOMOBILE/dl/gomobile-openal-soft-1.16.0.1.tar.gz https://dl.google.com/go/mobile/gomobile-openal-soft-1.16.0.1.tar.gz
tar xfz $GOMOBILE/dl/gomobile-openal-soft-1.16.0.1.tar.gz
mv $WORK/openal/include/AL $GOMOBILE/android-{{.NDK}}/arm/sysroot/usr/include/AL
mkdir -p $GOMOBILE/android-{{.NDK}}/openal
mv $WORK/openal/lib $GOMOBILE/android-{{.NDK}}/openal/lib
rm -r -f "$GOROOT/pkg/android_arm"
PATH=$PATH GOOS=android GOARCH=arm CGO_ENABLED=1 CC=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} CXX=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} GOARM=7 go install std
{{if eq .GOOS "darwin"}}rm -r -f "$GOROOT/pkg/darwin_arm"
PATH=$PATH GOOS=darwin GOARCH=arm CGO_ENABLED=1 CC=$GOROOT/misc/ios/clangwrap.sh CXX=$GOROOT/misc/ios/clangwrap.sh GOARM=7 go install std
rm -r -f "$GOROOT/pkg/darwin_arm64"
PATH=$PATH GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CC=$GOROOT/misc/ios/clangwrap.sh CXX=$GOROOT/misc/ios/clangwrap.sh go install std
{{end}}go version > $GOMOBILE/version
rm -r -f "$WORK"
`))
