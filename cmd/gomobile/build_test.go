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

func TestRFC1034Label(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"a", "a"},
		{"123", "-23"},
		{"a.b.c", "a-b-c"},
		{"a-b", "a-b"},
		{"a:b", "a-b"},
		{"a?b", "a-b"},
		{"αβγ", "---"},
		{"💩", "--"},
		{"My App", "My-App"},
		{"...", ""},
		{".-.", "--"},
	}

	for _, tc := range tests {
		if got := rfc1034Label(tc.in); got != tc.want {
			t.Errorf("rfc1034Label(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAndroidBuild(t *testing.T) {
	buf := new(bytes.Buffer)
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
	}()
	xout = buf
	buildN = true
	buildX = true
	buildO = "basic.apk"
	buildTarget = "android"
	gopath = filepath.ToSlash(filepath.SplitList(os.Getenv("GOPATH"))[0])
	if goos == "windows" {
		os.Setenv("HOMEDRIVE", "C:")
	}
	cmdBuild.flag.Parse([]string{"golang.org/x/mobile/example/basic"})
	err := runBuild(cmdBuild)
	if err != nil {
		t.Log(buf.String())
		t.Fatal(err)
	}

	diff, err := diffOutput(buf.String(), androidBuildTmpl)
	if err != nil {
		t.Fatalf("computing diff failed: %v", err)
	}
	if diff != "" {
		t.Errorf("unexpected output:\n%s", diff)
	}
}

var androidBuildTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
GOOS=android GOARCH=arm GOARM=7 CC=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} CXX=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} CGO_ENABLED=1 go build -p={{.NumCPU}} -pkgdir=$GOMOBILE/pkg_android_arm -tags="" -x -buildmode=c-shared -o $WORK/libbasic.so golang.org/x/mobile/example/basic
`))
