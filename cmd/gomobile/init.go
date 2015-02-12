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

	gopaths := filepath.SplitList(goEnv("GOPATH"))
	if len(gopaths) == 0 {
		return fmt.Errorf("GOPATH is not set")
	}
	ndkccpath = filepath.Join(gopaths[0], filepath.FromSlash("pkg/gomobile/android-"+ndkVersion))
	if buildX {
		fmt.Fprintln(os.Stderr, "NDKCCPATH="+ndkccpath)
	}

	if err := removeAll(ndkccpath); err != nil && !os.IsExist(err) {
		return err
	}
	dst := filepath.Join(ndkccpath, "arm")
	dstSysroot := filepath.Join(dst, "sysroot", "usr")
	if err := mkdir(dstSysroot); err != nil {
		return err
	}

	if buildN {
		tmpdir = filepath.Join(ndkccpath, "work")
	} else {
		var err error
		tmpdir, err = ioutil.TempDir(ndkccpath, "gomobile-init-")
		if err != nil {
			return err
		}
	}
	if buildX {
		fmt.Fprintln(os.Stderr, "WORK="+tmpdir)
	}
	defer removeAll(tmpdir)

	ndkName := "android-" + ndkVersion + "-" + runtime.GOOS + "-" + arch + "."
	if runtime.GOOS == "windows" {
		ndkName += "exe"
	} else {
		ndkName += "bin"
	}

	url := "http://dl.google.com/android/ndk/" + ndkName
	if err := fetch(filepath.Join(tmpdir, ndkName), url); err != nil {
		return err
	}

	inflate := exec.Command(filepath.Join(tmpdir, ndkName))
	inflate.Dir = tmpdir
	if buildX {
		printcmd("%s", inflate.Args[0])
	}
	if !buildN {
		out, err := inflate.CombinedOutput()
		if err != nil {
			if buildV {
				os.Stderr.Write(out)
			}
			return err
		}
	}
	srcSysroot := filepath.Join(tmpdir, "android-ndk-r10d", "platforms", "android-15", "arch-arm", "usr")
	if err := move(dstSysroot, srcSysroot, "include", "lib"); err != nil {
		return err
	}

	ndkpath := filepath.Join(tmpdir, "android-ndk-r10d", "toolchains", "arm-linux-androideabi-4.8", "prebuilt", runtime.GOOS+"-"+arch)
	if err := move(dst, ndkpath, "bin", "lib", "libexec"); err != nil {
		return err
	}

	linkpath := filepath.Join(dst, "arm-linux-androideabi", "bin")
	if err := mkdir(linkpath); err != nil {
		return err
	}
	for _, name := range []string{"ld", "ld.gold", "as", "gcc", "g++"} {
		if err := symlink(filepath.Join(dst, "bin", "arm-linux-androideabi-"+name), filepath.Join(linkpath, name)); err != nil {
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
	if buildV {
		fmt.Fprintf(os.Stderr, "building android/arm cross compiler\n")
		make.Stdout = os.Stdout
		make.Stderr = os.Stderr
	}
	if buildX {
		printcmd("%s", strings.Join(make.Env, " ")+" "+strings.Join(make.Args, " "))
	}
	if buildN {
		return nil
	}
	if err := make.Run(); err != nil {
		return err
	}
	return nil
}

func move(dst, src string, names ...string) error {
	for _, name := range names {
		srcf := filepath.Join(src, name)
		dstf := filepath.Join(dst, name)
		if buildX {
			printcmd("mv %s %s", srcf, dstf)
		}
		if buildN {
			continue
		}
		if err := os.Rename(srcf, dstf); err != nil {
			return err
		}
	}
	return nil
}

func mkdir(dir string) error {
	if buildX {
		printcmd("mkdir -p %s", dir)
	}
	if buildN {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func symlink(src, dst string) error {
	if buildX {
		printcmd("ln -s %s %s", src, dst)
	}
	if buildN {
		return nil
	}
	return os.Symlink(src, dst)
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

func fetch(dst, url string) error {
	if buildV {
		fmt.Fprintf(os.Stderr, "fetching %s\n", url)
	}
	if buildX {
		printcmd("curl -o%s %s", dst, url)
	}
	if buildN {
		return nil
	}

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	// TODO(crawshaw): The arm compiler toolchain compresses to 33 MB, less than a tenth of the NDK. Provide an alternative binary download.
	resp, err := http.Get(url)
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
	return err3
}

func removeAll(path string) error {
	if buildX {
		printcmd("rm -r -f %q", path)
	}
	if buildN {
		return nil
	}
	return os.RemoveAll(path)
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
