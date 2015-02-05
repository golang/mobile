// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// TODO(crawshaw): build darwin/arm cross compiler on darwin/{386,amd64}
// TODO(crawshaw): android/{386,arm64}

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ndkVersion = "ndk-r10d"

var cmdInit = &command{
	run:   runInit,
	Name:  "init",
	Short: "install android compiler toolchain",
	Long: `
Init downloads and installs the Android C++ compiler toolchain.

The toolchain is installed in $GOPATH/pkg/gomobile.
`,
}

func runInit(cmd *command) error {
	if err := checkGoVersion(); err != nil {
		return err
	}

	// Provide an early error message if Go was installed system-wide.
	goroot := goEnv("GOROOT")
	sentinel := filepath.Join(goroot, "gomobile-sentinel")
	if err := ioutil.WriteFile(sentinel, []byte("write test"), 0664); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("GOROOT %q is not writable. Run:\n\tsudo gomobile init", goroot)
		}
		return fmt.Errorf("GOROOT not writable: %v", err)
	}
	os.Remove(sentinel)

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	dst := filepath.Join(goEnv("GOPATH"), filepath.FromSlash("pkg/gomobile/android-"+ndkVersion+"/arm"))
	if err := os.RemoveAll(dst); err != nil && !os.IsExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Join(dst, "sysroot", "usr"), 0755); err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir(dst, "gomobile-init-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	ndkName := "android-" + ndkVersion + "-" + runtime.GOOS + "-" + arch + "."
	if runtime.GOOS == "windows" {
		ndkName += "exe"
	} else {
		ndkName += "bin"
	}
	f, err := os.OpenFile(filepath.Join(tmpdir, ndkName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	// TODO(crawshaw): The arm compiler toolchain compresses to 33 MB, less than a tenth of the NDK. Provide an alternative binary download.
	// TODO(crawshaw): extra logging with -v
	resp, err := http.Get("http://dl.google.com/android/ndk/" + ndkName)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	err2 := resp.Body.Close()
	err3 := f.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}

	inflate := exec.Command(filepath.Join(tmpdir, ndkName))
	inflate.Dir = tmpdir
	out, err := inflate.CombinedOutput()
	if err != nil {
		os.Stderr.Write(out) // TODO(crawshaw): only in verbose mode?
		return err
	}
	srcSysroot := filepath.Join(tmpdir, "android-ndk-r10d", "platforms", "android-15", "arch-arm", "usr")
	dstSysroot := filepath.Join(dst, "sysroot", "usr")
	if err := move(dstSysroot, srcSysroot, "include", "lib"); err != nil {
		return err
	}

	ndkpath := filepath.Join(tmpdir, "android-ndk-r10d", "toolchains", "arm-linux-androideabi-4.8", "prebuilt", runtime.GOOS+"-"+arch)
	if err := move(dst, ndkpath, "bin", "lib", "libexec"); err != nil {
		return err
	}

	linkpath := filepath.Join(dst, "arm-linux-androideabi", "bin")
	if err := os.MkdirAll(linkpath, 0755); err != nil {
		return err
	}
	for _, name := range []string{"ld", "ld.gold", "as", "gcc", "g++"} {
		if err := os.Symlink(filepath.Join(dst, "bin", "arm-linux-androideabi-"+name), filepath.Join(linkpath, name)); err != nil {
			return err
		}
	}

	// TODO(crawshaw): make.bat on windows
	ccpath := filepath.Join(dst, "bin")
	make := exec.Command(filepath.Join(goroot, "src", "make.bash"), "--no-clean")
	make.Dir = filepath.Join(goroot, "src")
	make.Env = []string{
		`PATH=` + os.Getenv("PATH"),
		`TMPDIR=` + tmpdir,
		`HOME=` + os.Getenv("HOME"), // for default the go1.4 bootstrap
		`GOOS=android`,
		`GOARCH=arm`,
		`GOARM=7`,
		`CGO_ENABLED=1`,
		`CC_FOR_TARGET=` + filepath.Join(ccpath, "arm-linux-androideabi-gcc"),
		`CXX_FOR_TARGET=` + filepath.Join(ccpath, "arm-linux-androideabi-g++"),
		`GOBIN=` + tmpdir, // avoid overwriting current Go tool
	}
	if v := goEnv("GOROOT_BOOTSTRAP"); v != "" {
		make.Env = append(make.Env, `GOROOT_BOOTSTRAP=`+v)
	}
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return err
	}

	return nil
}

func move(dst, src string, names ...string) error {
	for _, name := range names {
		if err := os.Rename(filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
			return err
		}
	}
	return nil
}

func checkGoVersion() error {
	if err := exec.Command("which", "go").Run(); err != nil {
		return fmt.Errorf(`no Go tool on $PATH`)
	}
	_, err := exec.Command("go", "version").Output()
	if err != nil {
		return fmt.Errorf("bad Go tool: %v", err)
	}
	return nil
}

func goEnv(name string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	val, err := exec.Command("go", "env", name).Output()
	if err != nil {
		panic(err) // the Go tool was tested to work earlier
	}
	return strings.TrimSpace(string(val))
}
