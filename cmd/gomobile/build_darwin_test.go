// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"
)

func TestIOSBuild(t *testing.T) {
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
	}()
	buildN = true
	buildX = true
	buildTarget = "ios"
	buildBundleID = "org.golang.todo"
	gopath = filepath.SplitList(goEnv("GOPATH"))[0]
	oldTags := ctx.BuildTags
	ctx.BuildTags = []string{"tag1"}
	defer func() {
		ctx.BuildTags = oldTags
	}()
	tests := []struct {
		pkg  string
		main bool
	}{
		{"golang.org/x/mobile/example/basic", true},
		{"golang.org/x/mobile/bind/testdata/testpkg", false},
	}
	for _, test := range tests {
		buf := new(bytes.Buffer)
		xout = buf
		if test.main {
			buildO = "basic.app"
		} else {
			buildO = ""
		}
		cmdBuild.flag.Parse([]string{test.pkg})
		err := runBuild(cmdBuild)
		if err != nil {
			t.Log(buf.String())
			t.Fatal(err)
		}

		teamID, err := detectTeamID()
		if err != nil {
			t.Fatalf("detecting team ID failed: %v", err)
		}

		data := struct {
			outputData
			TeamID string
			Pkg    string
			Main   bool
		}{
			outputData: defaultOutputData(),
			TeamID:     teamID,
			Pkg:        test.pkg,
			Main:       test.main,
		}

		got := filepath.ToSlash(buf.String())

		wantBuf := new(bytes.Buffer)

		if err := iosBuildTmpl.Execute(wantBuf, data); err != nil {
			t.Fatalf("computing diff failed: %v", err)
		}

		diff, err := diff(got, wantBuf.String())

		if err != nil {
			t.Fatalf("computing diff failed: %v", err)
		}
		if diff != "" {
			t.Errorf("unexpected output:\n%s", diff)
		}
	}
}

var iosBuildTmpl = template.Must(infoplistTmpl.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK{{if .Main}}
mkdir -p $WORK/main.xcodeproj
echo "{{.Xproj}}" > $WORK/main.xcodeproj/project.pbxproj
mkdir -p $WORK/main
echo "{{template "infoplist" .Xinfo}}" > $WORK/main/Info.plist
mkdir -p $WORK/main/Images.xcassets/AppIcon.appiconset
echo "{{.Xcontents}}" > $WORK/main/Images.xcassets/AppIcon.appiconset/Contents.json{{end}}
GOARM=7 GOOS=darwin GOARCH=arm CC=iphoneos-clang CXX=iphoneos-clang++ CGO_CFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_CXXFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_LDFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch armv7 CGO_ENABLED=1 GO111MODULE=off go build -tags tag1 ios -x {{if .Main}}-ldflags=-w -o=$WORK/arm {{end}}{{.Pkg}}
GOOS=darwin GOARCH=arm64 CC=iphoneos-clang CXX=iphoneos-clang++ CGO_CFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch arm64 CGO_CXXFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch arm64 CGO_LDFLAGS=-isysroot=iphoneos -miphoneos-version-min=7.0 -fembed-bitcode -arch arm64 CGO_ENABLED=1 GO111MODULE=off go build -tags tag1 ios -x {{if .Main}}-ldflags=-w -o=$WORK/arm64 {{end}}{{.Pkg}}
GOOS=darwin GOARCH=386 CC=iphonesimulator-clang CXX=iphonesimulator-clang++ CGO_CFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch i386 CGO_CXXFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch i386 CGO_LDFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch i386 CGO_ENABLED=1 GO111MODULE=off go build -tags tag1 ios -x {{if .Main}}-ldflags=-w -o=$WORK/386 {{end}}{{.Pkg}}
GOOS=darwin GOARCH=amd64 CC=iphonesimulator-clang CXX=iphonesimulator-clang++ CGO_CFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch x86_64 CGO_CXXFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch x86_64 CGO_LDFLAGS=-isysroot=iphonesimulator -mios-simulator-version-min=7.0 -fembed-bitcode -arch x86_64 CGO_ENABLED=1 GO111MODULE=off go build -tags tag1 ios -x {{if .Main}}-ldflags=-w -o=$WORK/amd64 {{end}}{{.Pkg}}{{if .Main}}
xcrun lipo -o $WORK/main/main -create $WORK/arm $WORK/arm64 $WORK/386 $WORK/amd64
mkdir -p $WORK/main/assets
xcrun xcodebuild -configuration Release -project $WORK/main.xcodeproj -allowProvisioningUpdates DEVELOPMENT_TEAM={{.TeamID}}
mv $WORK/build/Release-iphoneos/main.app basic.app{{end}}
`))
