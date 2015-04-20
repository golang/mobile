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

func TestInit(t *testing.T) {
	buf := new(bytes.Buffer)
	gopath := os.Getenv("GOPATH")
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		os.Setenv("GOPATH", gopath)
	}()
	xout = buf
	buildN = true
	buildX = true
	// Test that first GOPATH element is chosen correctly.
	paths := []string{"GOPATH1", "/path2", "/path3"}
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
	data := outputData{
		NDK:         ndkVersion,
		GOOS:        goos,
		GOARCH:      goarch,
		NDKARCH:     ndkarch,
		BuildScript: unixBuildScript,
	}
	if goos == "windows" {
		data.EXE = ".exe"
		data.BuildScript = windowsBuildScript
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
	NDK         string
	GOOS        string
	GOARCH      string
	NDKARCH     string
	EXE         string // .extension for executables. (ex. ".exe" for windows)
	BuildScript string
}

const (
	unixBuildScript    = `PATH=$PATH GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 CC_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-gcc CXX_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-g++ TMPDIR=$WORK HOME=$HOME GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bash --no-clean`
	windowsBuildScript = `PATH=$PATH GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 CC_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-gcc.exe CXX_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-g++.exe TEMP=$WORK TMP=$WORK HOMEDRIVE=C: HOMEPATH=$HOMEPATH GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bat --no-clean`
)

var initTmpl = template.Must(template.New("output").Parse(`NDKCCPATH=GOPATH1/pkg/gomobile/android-{{.NDK}}
rm -r -f "$NDKCCPATH"
mkdir -p $NDKCCPATH
WORK=GOPATH1/pkg/gomobile/android-{{.NDK}}/work
mkdir -p $WORK/go/pkg
cp -a $GOROOT/lib $WORK/go/lib
cp -a $GOROOT/src $WORK/go/src
ln -s $GOROOT/.git $WORK/go/.git
curl -o$WORK/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz https://dl.google.com/go/mobile/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
tar xfz gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
mkdir -p $NDKCCPATH/arm/sysroot/usr
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/include $NDKCCPATH/arm/sysroot/usr/include
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/lib $NDKCCPATH/arm/sysroot/usr/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin $NDKCCPATH/arm/bin
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/lib $NDKCCPATH/arm/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/libexec $NDKCCPATH/arm/libexec
mkdir -p $NDKCCPATH/arm/arm-linux-androideabi/bin
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-ld{{.EXE}} $NDKCCPATH/arm/arm-linux-androideabi/bin/ld{{.EXE}}
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-as{{.EXE}} $NDKCCPATH/arm/arm-linux-androideabi/bin/as{{.EXE}}
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-gcc{{.EXE}} $NDKCCPATH/arm/arm-linux-androideabi/bin/gcc{{.EXE}}
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-g++{{.EXE}} $NDKCCPATH/arm/arm-linux-androideabi/bin/g++{{.EXE}}
curl -o$WORK/gomobile-openal-soft-1.16.0.1.tar.gz https://dl.google.com/go/mobile/gomobile-openal-soft-1.16.0.1.tar.gz
tar xfz gomobile-openal-soft-1.16.0.1.tar.gz
mv $WORK/openal/include/AL $NDKCCPATH/arm/sysroot/usr/include/AL
mkdir -p $NDKCCPATH/openal
mv $WORK/openal/lib $NDKCCPATH/openal/lib
{{.BuildScript}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/5l{{.EXE}} $NDKCCPATH/arm/bin/5l{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/5g{{.EXE}} $NDKCCPATH/arm/bin/5g{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/asm{{.EXE}} $NDKCCPATH/arm/bin/asm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/cgo{{.EXE}} $NDKCCPATH/arm/bin/cgo{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/nm{{.EXE}} $NDKCCPATH/arm/bin/nm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/old5a{{.EXE}} $NDKCCPATH/arm/bin/old5a{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/pack{{.EXE}} $NDKCCPATH/arm/bin/pack{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/link{{.EXE}} $NDKCCPATH/arm/bin/link{{.EXE}}
go build -o $NDKCCPATH/arm/bin/toolexec{{.EXE}} $WORK/toolexec.go
rm -r -f "$GOROOT/pkg/android_arm"
mv $WORK/go/pkg/android_arm $GOROOT/pkg/android_arm
go version > GOPATH1/pkg/gomobile/version
rm -r -f "$WORK"
`))
