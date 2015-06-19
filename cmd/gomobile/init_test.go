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
	if runtime.GOOS == "darwin" {
		tmpl = initDarwinTmpl
	}

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
		NDK:         ndkVersion,
		GOOS:        goos,
		GOARCH:      goarch,
		GOPATH:      gopath,
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
	GOPATH      string
	NDKARCH     string
	EXE         string // .extension for executables. (ex. ".exe" for windows)
	BuildScript string
}

const (
	unixBuildScript    = `TMPDIR=$WORK HOME=$HOME GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bash --no-clean`
	windowsBuildScript = `TEMP=$WORK TMP=$WORK HOMEDRIVE=C: HOMEPATH=$HOMEPATH GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bat --no-clean`
)

var initTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
mkdir -p $GOMOBILE/android-{{.NDK}}
WORK=/GOPATH1/pkg/gomobile/work
mkdir -p $WORK/go/pkg
cp -a $GOROOT/lib $WORK/go/lib
cp -a $GOROOT/src $WORK/go/src
cp -a $GOROOT/misc $WORK/go/misc
ln -s $GOROOT/.git $WORK/go/.git
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
PATH=$PATH GOOS=android GOROOT=$WORK/go GOARCH=arm GOARM=7 CGO_ENABLED=1 CC_FOR_TARGET=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} CXX_FOR_TARGET=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} {{.BuildScript}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/compile{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/compile{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/asm{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/asm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/cgo{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/cgo{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/nm{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/nm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/old5a{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/old5a{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/pack{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/pack{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/link{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/link{{.EXE}}
go build -o $GOMOBILE/android-{{.NDK}}/arm/bin/toolexec{{.EXE}} $WORK/toolexec.go
rm -r -f "$GOROOT/pkg/android_arm"
mv $WORK/go/pkg/android_arm $GOROOT/pkg/android_arm
go version > $GOMOBILE/version
rm -r -f "$WORK"
`))

var initDarwinTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
mkdir -p $GOMOBILE/android-{{.NDK}}
WORK=/GOPATH1/pkg/gomobile/work
mkdir -p $WORK/go/pkg
cp -a $GOROOT/lib $WORK/go/lib
cp -a $GOROOT/src $WORK/go/src
cp -a $GOROOT/misc $WORK/go/misc
ln -s $GOROOT/.git $WORK/go/.git
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
PATH=$PATH GOOS=android GOROOT=$WORK/go GOARCH=arm GOARM=7 CGO_ENABLED=1 CC_FOR_TARGET=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} CXX_FOR_TARGET=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} {{.BuildScript}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/compile{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/compile{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/asm{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/asm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/cgo{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/cgo{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/nm{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/nm{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/old5a{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/old5a{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/pack{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/pack{{.EXE}}
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/link{{.EXE}} $GOMOBILE/android-{{.NDK}}/arm/bin/link{{.EXE}}
go build -o $GOMOBILE/android-{{.NDK}}/arm/bin/toolexec{{.EXE}} $WORK/toolexec.go
rm -r -f "$GOROOT/pkg/android_arm"
mv $WORK/go/pkg/android_arm $GOROOT/pkg/android_arm
PATH=$PATH GOOS=darwin GOROOT=$WORK/go GOARCH=arm CGO_ENABLED=1 CC_FOR_TARGET=$WORK/go/misc/ios/clangwrap.sh CXX_FOR_TARGET=$WORK/go/misc/ios/clangwrap.sh GOARM=7 TMPDIR=$WORK HOME=$HOME GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bash --no-clean
PATH=$PATH GOOS=darwin GOROOT=$WORK/go GOARCH=arm64 CGO_ENABLED=1 CC_FOR_TARGET=$WORK/go/misc/ios/clangwrap.sh CXX_FOR_TARGET=$WORK/go/misc/ios/clangwrap.sh TMPDIR=$WORK HOME=$HOME GOROOT_BOOTSTRAP=go1.4 $WORK/go/src/make.bash --no-clean
rm -r -f "$GOROOT/pkg/darwin_arm"
mv $WORK/go/pkg/darwin_arm $GOROOT/pkg/darwin_arm
rm -r -f "$GOROOT/pkg/darwin_arm64"
mv $WORK/go/pkg/darwin_arm64 $GOROOT/pkg/darwin_arm64
go version > $GOMOBILE/version
rm -r -f "$WORK"
`))
