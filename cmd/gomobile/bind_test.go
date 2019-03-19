// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

func TestImportPackagesPathCleaning(t *testing.T) {
	if runtime.GOOS == "android" {
		t.Skip("not available on Android")
	}
	slashPath := "golang.org/x/mobile/example/bind/hello/"
	pkgs, err := importPackages([]string{slashPath})
	if err != nil {
		t.Fatal(err)
	}
	p := pkgs[0]
	if c := path.Clean(slashPath); p.ImportPath != c {
		t.Errorf("expected %s; got %s", c, p.ImportPath)
	}
}

func TestBindAndroid(t *testing.T) {
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		t.Skip("ANDROID_HOME not found, skipping bind")
	}
	platform, err := androidAPIPath()
	if err != nil {
		t.Skip("No android API platform found in $ANDROID_HOME, skipping bind")
	}
	platform = strings.Replace(platform, androidHome, "$ANDROID_HOME", -1)

	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		buildO = ""
		buildTarget = ""
		bindJavaPkg = ""
	}()
	buildN = true
	buildX = true
	buildO = "asset.aar"
	buildTarget = "android/arm"

	tests := []struct {
		javaPkg string
	}{
		{
			// Empty javaPkg
		},
		{
			javaPkg: "com.example.foo",
		},
	}
	for _, tc := range tests {
		bindJavaPkg = tc.javaPkg

		buf := new(bytes.Buffer)
		xout = buf
		gopath = filepath.SplitList(goEnv("GOPATH"))[0]
		if goos == "windows" {
			os.Setenv("HOMEDRIVE", "C:")
		}
		cmdBind.flag.Parse([]string{"golang.org/x/mobile/asset"})
		err := runBind(cmdBind)
		if err != nil {
			t.Log(buf.String())
			t.Fatal(err)
		}
		got := filepath.ToSlash(buf.String())

		data := struct {
			outputData
			AndroidPlatform string
			JavaPkg         string
		}{
			outputData:      defaultOutputData(),
			AndroidPlatform: platform,
			JavaPkg:         tc.javaPkg,
		}

		wantBuf := new(bytes.Buffer)
		if err := bindAndroidTmpl.Execute(wantBuf, data); err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}

		diff, err := diff(got, wantBuf.String())
		if err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}
		if diff != "" {
			t.Errorf("%+v: unexpected output:\n%s", tc, diff)
		}
	}
}

func TestBindIOS(t *testing.T) {
	if !xcodeAvailable() {
		t.Skip("Xcode is missing")
	}
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		buildO = ""
		buildTarget = ""
		bindPrefix = ""
	}()
	buildN = true
	buildX = true
	buildO = "Asset.framework"
	buildTarget = "ios/arm"

	tests := []struct {
		prefix string
	}{
		{
			// empty prefix
		},
		{
			prefix: "Foo",
		},
	}
	for _, tc := range tests {
		bindPrefix = tc.prefix

		buf := new(bytes.Buffer)
		xout = buf
		gopath = filepath.SplitList(goEnv("GOPATH"))[0]
		if goos == "windows" {
			os.Setenv("HOMEDRIVE", "C:")
		}
		cmdBind.flag.Parse([]string{"golang.org/x/mobile/asset"})
		err := runBind(cmdBind)
		if err != nil {
			t.Log(buf.String())
			t.Fatal(err)
		}
		got := filepath.ToSlash(buf.String())

		data := struct {
			outputData
			Prefix string
		}{
			outputData: defaultOutputData(),
			Prefix:     tc.prefix,
		}

		wantBuf := new(bytes.Buffer)
		if err := bindIOSTmpl.Execute(wantBuf, data); err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}

		diff, err := diff(got, wantBuf.String())
		if err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}
		if diff != "" {
			t.Errorf("%+v: unexpected output:\n%s", tc, diff)
		}
	}
}

var bindAndroidTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
GOOS=android CGO_ENABLED=1 gobind -lang=go,java -outdir=$WORK{{if .JavaPkg}} -javapkg={{.JavaPkg}}{{end}} golang.org/x/mobile/asset
GOOS=android GOARCH=arm CC=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang CXX=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang++ CGO_ENABLED=1 GOARM=7 GOPATH=$WORK:$GOPATH GO111MODULE=off go build -x -buildmode=c-shared -o=$WORK/android/src/main/jniLibs/armeabi-v7a/libgojni.so gobind
PWD=$WORK/java javac -d $WORK/javac-output -source 1.7 -target 1.7 -bootclasspath {{.AndroidPlatform}}/android.jar *.java
jar c -C $WORK/javac-output .
`))

var bindIOSTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
GOOS=darwin CGO_ENABLED=1 gobind -lang=go,objc -outdir=$WORK -tags=ios{{if .Prefix}} -prefix={{.Prefix}}{{end}} golang.org/x/mobile/asset
GOARM=7 GOOS=darwin GOARCH=arm CC=iphoneos-clang CXX=iphoneos-clang++ CGO_CFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_CXXFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_LDFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_ENABLED=1 GOPATH=$WORK:$GOPATH GO111MODULE=off go build -tags ios -x -buildmode=c-archive -o $WORK/asset-arm.a gobind
rm -r -f "Asset.framework"
mkdir -p Asset.framework/Versions/A/Headers
ln -s A Asset.framework/Versions/Current
ln -s Versions/Current/Headers Asset.framework/Headers
ln -s Versions/Current/Asset Asset.framework/Asset
xcrun lipo -create -arch armv7 $WORK/asset-arm.a -o Asset.framework/Versions/A/Asset
cp $WORK/src/gobind/{{.Prefix}}Asset.objc.h Asset.framework/Versions/A/Headers/{{.Prefix}}Asset.objc.h
mkdir -p Asset.framework/Versions/A/Headers
cp $WORK/src/gobind/Universe.objc.h Asset.framework/Versions/A/Headers/Universe.objc.h
mkdir -p Asset.framework/Versions/A/Headers
cp $WORK/src/gobind/ref.h Asset.framework/Versions/A/Headers/ref.h
mkdir -p Asset.framework/Versions/A/Headers
mkdir -p Asset.framework/Versions/A/Headers
mkdir -p Asset.framework/Versions/A/Resources
ln -s Versions/Current/Resources Asset.framework/Resources
mkdir -p Asset.framework/Resources
mkdir -p Asset.framework/Versions/A/Modules
ln -s Versions/Current/Modules Asset.framework/Modules
`))
