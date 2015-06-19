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

func TestBuild(t *testing.T) {
	buf := new(bytes.Buffer)
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
	}()
	xout = buf
	buildN = true
	buildX = true
	gopath = filepath.SplitList(os.Getenv("GOPATH"))[0]
	if goos == "windows" {
		os.Setenv("HOMEDRIVE", "C:")
	}
	cmdBuild.flag.Parse([]string{"golang.org/x/mobile/example/basic"})
	if *buildTarget != "android" {
		t.Fatalf("-target=%q, want android", *buildTarget)
	}
	err := runBuild(cmdBuild)
	if err != nil {
		t.Log(buf.String())
		t.Fatal(err)
	}

	diff, err := diffOutput(buf.String(), buildTmpl)
	if err != nil {
		t.Fatalf("computing diff failed: %v", err)
	}
	if diff != "" {
		t.Errorf("unexpected output:\n%s", diff)
	}

}

var buildTmpl = template.Must(template.New("output").Parse(`WORK=$WORK
GOMOBILE={{.GOPATH}}/pkg/gomobile
GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-gcc{{.EXE}} CXX=$GOMOBILE/android-{{.NDK}}/arm/bin/arm-linux-androideabi-g++{{.EXE}} GOGCCFLAGS="-fPIC -marm -pthread -fmessage-length=0" GOROOT=$GOROOT GOPATH=$HOME GOMOBILEPATH=$GOMOBILE/android-{{.NDK}}/arm/bin go build -tags="" -toolexec=$GOMOBILE/android-{{.NDK}}/arm/bin/toolexec -x -buildmode=c-shared -o $WORK/libbasic.so golang.org/x/mobile/example/basic
rm -r -f "$WORK"
`))
