// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
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
	os.Setenv("GOPATH", "GOPATH1:/path2:/path3")

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
	wantBuf := new(bytes.Buffer)
	data := outputData{ndkVersion, goos, goarch, ndkarch}
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
	NDKARCH string
}

var initTmpl = template.Must(template.New("output").Parse(`NDKCCPATH=GOPATH1/pkg/gomobile/android-{{.NDK}}
rm -r -f "$NDKCCPATH"
mkdir -p $NDKCCPATH
WORK=GOPATH1/pkg/gomobile/android-{{.NDK}}/work
mkdir -p $WORK/go/pkg
cp -a $HOME/go/include $WORK/go/include
cp -a $HOME/go/lib $WORK/go/lib
cp -a $HOME/go/src $WORK/go/src
ln -s $HOME/go/.git $WORK/go/.git
curl -o$WORK/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz https://dl.google.com/go/mobile/gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
tar xfz gomobile-{{.NDK}}-{{.GOOS}}-{{.NDKARCH}}.tar.gz
mkdir -p $NDKCCPATH/arm/sysroot/usr
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/include $NDKCCPATH/arm/sysroot/usr/include
mv $WORK/android-{{.NDK}}/platforms/android-15/arch-arm/usr/lib $NDKCCPATH/arm/sysroot/usr/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/bin $NDKCCPATH/arm/bin
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/lib $NDKCCPATH/arm/lib
mv $WORK/android-{{.NDK}}/toolchains/arm-linux-androideabi-4.8/prebuilt/{{.GOOS}}-{{.NDKARCH}}/libexec $NDKCCPATH/arm/libexec
mkdir -p $NDKCCPATH/arm/arm-linux-androideabi/bin
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-ld $NDKCCPATH/arm/arm-linux-androideabi/bin/ld
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-as $NDKCCPATH/arm/arm-linux-androideabi/bin/as
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-gcc $NDKCCPATH/arm/arm-linux-androideabi/bin/gcc
ln -s $NDKCCPATH/arm/bin/arm-linux-androideabi-g++ $NDKCCPATH/arm/arm-linux-androideabi/bin/g++
PATH=$PATH TMPDIR=$WORK HOME=$HOME GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 CC_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-gcc CXX_FOR_TARGET=$NDKCCPATH/arm/bin/arm-linux-androideabi-g++ $WORK/go/src/make.bash --no-clean
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/5a $NDKCCPATH/arm/bin/5a
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/5l $NDKCCPATH/arm/bin/5l
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/5g $NDKCCPATH/arm/bin/5g
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/cgo $NDKCCPATH/arm/bin/cgo
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/nm $NDKCCPATH/arm/bin/nm
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/pack $NDKCCPATH/arm/bin/pack
mv $WORK/go/pkg/tool/{{.GOOS}}_{{.GOARCH}}/link $NDKCCPATH/arm/bin/link
go build -o $NDKCCPATH/arm/bin/toolexec $WORK/toolexec.go
rm -r -f "$HOME/go/pkg/android_arm"
mv $WORK/go/pkg/android_arm $HOME/go/pkg/android_arm
go version > GOPATH1/pkg/gomobile/version
rm -r -f "$WORK"
`))
