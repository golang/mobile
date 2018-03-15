// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// TODO(crawshaw): TestBindIOS

func TestImportPackagesPathCleaning(t *testing.T) {
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
	}()
	buildN = true
	buildX = true
	buildO = "asset.aar"
	buildTarget = "android/arm"

	tests := []struct {
		javaPkg    string
		wantGobind string
	}{
		{
			wantGobind: "gobind -lang=java",
		},
		{
			javaPkg:    "com.example.foo",
			wantGobind: "gobind -lang=java -javapkg=com.example.foo",
		},
	}
	for _, tc := range tests {
		bindJavaPkg = tc.javaPkg

		buf := new(bytes.Buffer)
		xout = buf
		gopath = filepath.SplitList(os.Getenv("GOPATH"))[0]
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
			GobindJavaCmd   string
			JavaPkg         string
		}{
			outputData:      defaultOutputData(),
			AndroidPlatform: platform,
			GobindJavaCmd:   tc.wantGobind,
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

var bindAndroidTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
gobind -lang=go,java -outdir=$WORK{{if .JavaPkg}} -javapkg={{.JavaPkg}}{{end}} golang.org/x/mobile/asset
GOOS=android GOARCH=arm CC=$GOMOBILE/ndk-toolchains/arm/bin/arm-linux-androideabi-clang CXX=$GOMOBILE/ndk-toolchains/arm/bin/arm-linux-androideabi-clang++ CGO_ENABLED=1 GOARM=7 GOPATH=$WORK:$GOPATH go build -x -buildmode=c-shared -o=$WORK/android/src/main/jniLibs/armeabi-v7a/libgojni.so gobind
PWD=$WORK/java javac -d $WORK/javac-output -source 1.7 -target 1.7 -bootclasspath {{.AndroidPlatform}}/android.jar *.java
jar c -C $WORK/javac-output .
`))
