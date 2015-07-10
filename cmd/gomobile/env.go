package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// General mobile build environment. Initialized by envInit.
var (
	cwd          string
	gomobilepath string // $GOPATH/pkg/gomobile
	ndkccpath    string // $GOPATH/pkg/gomobile/android-{{.NDK}}

	darwinArmEnv   []string
	darwinArm64Env []string
	androidArmEnv  []string
)

func envInit() (cleanup func(), err error) {
	cwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	// Find gomobilepath.
	gopath := goEnv("GOPATH")
	for _, p := range filepath.SplitList(gopath) {
		gomobilepath = filepath.Join(p, "pkg", "gomobile")
		if _, err := os.Stat(gomobilepath); err == nil {
			break
		}
	}
	if buildX {
		fmt.Fprintln(xout, "GOMOBILE="+gomobilepath)
	}

	// Check the toolchain is in a good state.
	version, err := goVersion()
	if err != nil {
		return nil, err
	}
	if gomobilepath == "" {
		return nil, errors.New("toolchain not installed, run `gomobile init`")
	}
	verpath := filepath.Join(gomobilepath, "version")
	installedVersion, err := ioutil.ReadFile(verpath)
	if err != nil {
		return nil, errors.New("toolchain partially installed, run `gomobile init`")
	}
	if !bytes.Equal(installedVersion, version) {
		return nil, errors.New("toolchain out of date, run `gomobile init`")
	}

	// Setup the cross-compiler environments.

	// TODO(crawshaw): Remove ndkccpath global.
	ndkccpath = filepath.Join(gomobilepath, "android-"+ndkVersion)
	ndkccbin := filepath.Join(ndkccpath, "arm", "bin")

	androidEnv := []string{
		"CC=" + filepath.Join(ndkccbin, "arm-linux-androideabi-gcc"),
		"CXX=" + filepath.Join(ndkccbin, "arm-linux-androideabi-g++"),
		`GOGCCFLAGS="-fPIC -marm -pthread -fmessage-length=0"`,
	}
	androidArmEnv = append([]string{
		"GOOS=android",
		"GOARCH=arm",
		"GOARM=7",
	}, androidEnv...)

	// TODO(jbd): Remove clangwrap.sh dependency by implementing clangwrap.sh
	// in Go in this package.
	goroot := goEnv("GOROOT")
	iosEnv := []string{
		"CC=" + filepath.Join(goroot, "misc/ios/clangwrap.sh"),
		"CCX=" + filepath.Join(goroot, "misc/ios/clangwrap.sh"),
	}
	darwinArmEnv = append([]string{
		"GOOS=darwin",
		"GOARCH=arm",
	}, iosEnv...)
	darwinArm64Env = append([]string{
		"GOOS=darwin",
		"GOARCH=arm64",
	}, iosEnv...)

	// We need a temporary directory when assembling an apk/app.
	if buildN {
		tmpdir = "$WORK"
	} else {
		tmpdir, err = ioutil.TempDir("", "gomobile-work-")
		if err != nil {
			return nil, err
		}
	}
	if buildX {
		fmt.Fprintln(xout, "WORK="+tmpdir)
	}

	return func() { removeAll(tmpdir) }, nil
}
