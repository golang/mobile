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

	// TODO(hyangah): test with go1_6.
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
		NDKARCH:   ndkarch,
		Xproj:     projPbxproj,
		Xcontents: contentsJSON,
		Xinfo:     infoplistTmplData{BundleID: "org.golang.todo.basic", Name: "Basic"},
	}
	if goos == "windows" {
		data.EXE = ".exe"
	}
	return data
}

var initTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
rm -r -f "$GOMOBILE"
mkdir -p $GOMOBILE
WORK={{.GOPATH}}/pkg/gomobile/work{{if eq .GOOS "darwin"}}
go install -x golang.org/x/mobile/gl
go install -x golang.org/x/mobile/app
go install -x golang.org/x/mobile/exp/app/debug{{end}}
GOOS=android GOARCH=arm CC=$NDK_PATH/toolchains/llvm/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/clang{{.EXE}} CXX=$NDK_PATH/toolchains/llvm/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/clang++{{.EXE}} CGO_CFLAGS=-target armv7a-none-linux-androideabi --sysroot $NDK_PATH/platforms/android-15/arch-arm -gcc-toolchain $NDK_PATH/toolchains/arm-linux-androideabi-4.9/prebuilt/{{.GOOS}}-{{.NDKARCH}} -I$GOMOBILE/include CGO_CPPFLAGS=-target armv7a-none-linux-androideabi --sysroot $NDK_PATH/platforms/android-15/arch-arm -gcc-toolchain $NDK_PATH/toolchains/arm-linux-androideabi-4.9/prebuilt/{{.GOOS}}-{{.NDKARCH}} -I$GOMOBILE/include CGO_LDFLAGS=-target armv7a-none-linux-androideabi --sysroot $NDK_PATH/platforms/android-15/arch-arm -gcc-toolchain $NDK_PATH/toolchains/arm-linux-androideabi-4.9/prebuilt/{{.GOOS}}-{{.NDKARCH}} -L$NDK_PATH/platforms/android-15/arch-arm/usr/lib -L$GOMOBILE/lib/arm CGO_ENABLED=1 GOARM=7 go install -pkgdir=$GOMOBILE/pkg_android_arm -x std
{{if eq .GOOS "darwin"}}GOOS=darwin GOARCH=arm GOARM=7 CC=clang-iphoneos CXX=clang-iphoneos CGO_CFLAGS=-isysroot=iphoneos -miphoneos-version-min=6.1 -arch armv7 CGO_LDFLAGS=-isysroot=iphoneos -miphoneos-version-min=6.1 -arch armv7 CGO_ENABLED=1 go install -pkgdir=$GOMOBILE/pkg_darwin_arm -x std
GOOS=darwin GOARCH=arm64 CC=clang-iphoneos CXX=clang-iphoneos CGO_CFLAGS=-isysroot=iphoneos -miphoneos-version-min=6.1 -arch arm64 CGO_LDFLAGS=-isysroot=iphoneos -miphoneos-version-min=6.1 -arch arm64 CGO_ENABLED=1 go install -pkgdir=$GOMOBILE/pkg_darwin_arm64 -x std
GOOS=darwin GOARCH=amd64 CC=clang-iphonesimulator CXX=clang-iphonesimulator CGO_CFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=6.1 -arch x86_64 CGO_LDFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=6.1 -arch x86_64 CGO_ENABLED=1 go install -tags=ios -pkgdir=$GOMOBILE/pkg_darwin_amd64 -x std
{{end}}cp $OPENAL_PATH/include/AL/al.h $GOMOBILE/include/AL/al.h
mkdir -p $GOMOBILE/include/AL
cp $OPENAL_PATH/include/AL/alc.h $GOMOBILE/include/AL/alc.h
mkdir -p $GOMOBILE/include/AL
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=arm --api=15 --install-dir=$WORK/build/armeabi/toolchain
PWD=$WORK/build/armeabi cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=arm-linux-androideabi
PWD=$WORK/build/armeabi $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/make
cp $WORK/build/armeabi/libopenal.so $GOMOBILE/lib/armeabi-v7a/libopenal.so
mkdir -p $GOMOBILE/lib/armeabi-v7a
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=arm64 --api=21 --install-dir=$WORK/build/arm64/toolchain
PWD=$WORK/build/arm64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=aarch64-linux-android
PWD=$WORK/build/arm64 $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/make
cp $WORK/build/arm64/libopenal.so $GOMOBILE/lib/arm64-v8a/libopenal.so
mkdir -p $GOMOBILE/lib/arm64-v8a
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=x86 --api=15 --install-dir=$WORK/build/x86/toolchain
PWD=$WORK/build/x86 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=i686-linux-android
PWD=$WORK/build/x86 $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/make
cp $WORK/build/x86/libopenal.so $GOMOBILE/lib/x86/libopenal.so
mkdir -p $GOMOBILE/lib/x86
PWD=$NDK_PATH $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/python2.7 build/tools/make_standalone_toolchain.py --arch=x86_64 --api=21 --install-dir=$WORK/build/x86_64/toolchain
PWD=$WORK/build/x86_64 cmake $OPENAL_PATH -DCMAKE_TOOLCHAIN_FILE=$OPENAL_PATH/XCompile-Android.txt -DHOST=x86_64-linux-android
PWD=$WORK/build/x86_64 $NDK_PATH/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin/make
cp $WORK/build/x86_64/libopenal.so $GOMOBILE/lib/x86_64/libopenal.so
mkdir -p $GOMOBILE/lib/x86_64
go version > $GOMOBILE/version
rm -r -f "$WORK"
`))
