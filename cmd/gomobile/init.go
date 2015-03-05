// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// TODO(crawshaw): build darwin/arm cross compiler on darwin/{386,amd64}
// TODO(crawshaw): android/{386,arm64}

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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

// useStrippedNDK determines whether the init subcommand fetches the GCC
// toolchain from the original Android NDK, or from the stripped-down NDK
// hosted specifically for the gomobile tool.
//
// There is a significant size different (400MB compared to 30MB).
var useStrippedNDK = true

const ndkVersion = "ndk-r10d"
const openALVersion = "openal-soft-1.16.0.1"

var (
	goos    = runtime.GOOS
	goarch  = runtime.GOARCH
	ndkarch string
)

func init() {
	if runtime.GOARCH == "amd64" {
		ndkarch = "x86_64"
	} else {
		ndkarch = runtime.GOARCH
	}
}

var cmdInit = &command{
	run:   runInit,
	Name:  "init",
	Short: "install android compiler toolchain",
	Long: `
Init downloads and installs the Android C++ compiler toolchain.

The toolchain is installed in $GOPATH/pkg/gomobile.
If the Android C++ compiler toolchain already exists in the path,
it skips download and uses the existing toolchain.

The -u option forces download and installation of the new toolchain
even when the toolchain exists.
`,
}

var initU bool // -u

func init() {
	cmdInit.flag.BoolVar(&initU, "u", false, "force toolchain download")
}

func runInit(cmd *command) error {
	version, err := goVersion()
	if err != nil {
		return err
	}

	gopaths := filepath.SplitList(goEnv("GOPATH"))
	if len(gopaths) == 0 {
		return fmt.Errorf("GOPATH is not set")
	}
	ndkccpath = filepath.Join(gopaths[0], "pkg/gomobile/android-"+ndkVersion)
	ndkccdl := filepath.Join(ndkccpath, "downloaded")
	verpath := filepath.Join(gopaths[0], "pkg/gomobile/version")
	if buildX {
		fmt.Fprintln(xout, "NDKCCPATH="+ndkccpath)
	}

	needNDK := initU
	if !needNDK {
		if _, err := os.Stat(ndkccdl); err != nil {
			needNDK = true
		}
	}

	if needNDK {
		if err := removeAll(ndkccpath); err != nil && !os.IsExist(err) {
			return err
		}
		if err := mkdir(ndkccpath); err != nil {
			return err
		}
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
		fmt.Fprintln(xout, "WORK="+tmpdir)
	}
	defer removeAll(tmpdir)

	goroot := goEnv("GOROOT")
	tmpGoroot := filepath.Join(tmpdir, "go")
	if err := copyGoroot(tmpGoroot, goroot); err != nil {
		return err
	}

	if needNDK {
		if err := fetchNDK(); err != nil {
			return err
		}
		if err := fetchOpenAL(); err != nil {
			return err
		}
		if !buildN {
			if err := ioutil.WriteFile(ndkccdl, []byte("done"), 0644); err != nil {
				return err
			}
		}
	}

	dst := filepath.Join(ndkccpath, "arm")

	ndkccbin := filepath.Join(dst, "bin")
	envpath := os.Getenv("PATH")
	if buildN {
		envpath = "$PATH"
	}
	makeScript := filepath.Join(tmpGoroot, "src/make")
	if goos == "windows" {
		makeScript += ".bat"
	} else {
		makeScript += ".bash"
	}

	bin := func(name string) string {
		if goos == "windows" {
			return name + ".exe"
		}
		return name
	}

	make := exec.Command(makeScript, "--no-clean")
	make.Dir = filepath.Join(tmpGoroot, "src")
	make.Env = []string{
		`PATH=` + envpath,
		`GOOS=android`,
		`GOARCH=arm`,
		`GOARM=7`,
		`CGO_ENABLED=1`,
		`CC_FOR_TARGET=` + filepath.Join(ndkccbin, bin("arm-linux-androideabi-gcc")),
		`CXX_FOR_TARGET=` + filepath.Join(ndkccbin, bin("arm-linux-androideabi-g++")),
	}
	if goos == "windows" {
		make.Env = append(make.Env, `TEMP=`+tmpdir)
		make.Env = append(make.Env, `TMP=`+tmpdir)
		make.Env = append(make.Env, `HOMEDRIVE=`+os.Getenv("HOMEDRIVE"))
		make.Env = append(make.Env, `HOMEPATH=`+os.Getenv("HOMEPATH"))
	} else {
		make.Env = append(make.Env, `TMPDIR=`+tmpdir)
		// for default the go1.4 bootstrap
		make.Env = append(make.Env, `HOME=`+os.Getenv("HOME"))
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
	if !buildN {
		make.Env = environ(make.Env)
		if err := make.Run(); err != nil {
			return err
		}
	}

	// Move the Go cross compiler toolchain into GOPATH.
	gotoolsrc := filepath.Join(tmpGoroot, "pkg/tool", goos+"_"+goarch)
	tools := []string{"5l", "5g", "asm", "cgo", "nm", "pack", "link"}
	for i, name := range tools {
		tools[i] = bin(name)
	}
	if err := move(ndkccbin, gotoolsrc, tools...); err != nil {
		return err
	}

	// Build toolexec command.
	toolexecSrc := filepath.Join(tmpdir, "toolexec.go")
	if !buildN {
		if err := ioutil.WriteFile(toolexecSrc, []byte(toolexec), 0644); err != nil {
			return err
		}
	}
	make = exec.Command("go", "build", "-o", filepath.Join(ndkccbin, bin("toolexec")), toolexecSrc)
	if buildV {
		fmt.Fprintf(os.Stderr, "building gomobile toolexec\n")
		make.Stdout = os.Stdout
		make.Stderr = os.Stderr
	}
	if buildX {
		printcmd("%s", strings.Join(make.Args, " "))
	}
	if !buildN {
		if err := make.Run(); err != nil {
			return err
		}
	}

	// Move pre-compiled stdlib for android into GOROOT. This is
	// the only time we modify the user's GOROOT.
	cannotRemove := false
	if err := removeAll(filepath.Join(goroot, "pkg/android_arm")); err != nil {
		cannotRemove = true
	}
	if err := move(filepath.Join(goroot, "pkg"), filepath.Join(tmpGoroot, "pkg"), "android_arm"); err != nil {
		// Move android_arm into a temp directory that outlives
		// this process and give the user installation instructions.
		dir, err := ioutil.TempDir("", "gomobile-")
		if err != nil {
			return err
		}
		if err := move(dir, filepath.Join(tmpGoroot, "pkg"), "android_arm"); err != nil {
			return err
		}
		remove := ""
		if cannotRemove {
			if goos == "windows" {
				remove = "\trd /s /q %s\\pkg\\android_arm\n"
			} else {
				remove = "\trm -r -f %s/pkg/android_arm\n"
			}
		}
		return fmt.Errorf(
			`Cannot install android/arm in GOROOT.
Make GOROOT writable (possibly by becoming the super user, using sudo) and run:
%s	mv %s %s`,
			remove,
			filepath.Join(dir, "android_arm"),
			filepath.Join(goroot, "pkg"),
		)
	}

	if buildX {
		printcmd("go version > %s", verpath)
	}
	if !buildN {
		if err := ioutil.WriteFile(verpath, version, 0644); err != nil {
			return err
		}
	}

	return nil
}

// toolexec is the source of a small program designed to be passed to
// the -toolexec flag of go build.
const toolexec = `package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	args := append([]string{}, os.Args[1:]...)
	args[0] = filepath.Join(os.Getenv("GOMOBILEPATH"), filepath.Base(args[0]))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
`

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
		if goos == "windows" {
			// os.Rename fails if dstf already exists.
			os.Remove(dstf)
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
	if goos == "windows" {
		return doCopyAll(dst, src)
	}
	return os.Symlink(src, dst)
}

func rm(name string) error {
	if buildX {
		printcmd("rm %s", name)
	}
	if buildN {
		return nil
	}
	return os.Remove(name)
}

func goVersion() ([]byte, error) {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf(`no Go tool on $PATH`)
	}
	buildHelp, err := exec.Command(gobin, "help", "build").Output()
	if err != nil {
		return nil, fmt.Errorf("bad Go tool: %v", err)
	}
	if !bytes.Contains(buildHelp, []byte("-toolexec")) {
		return nil, fmt.Errorf("installed Go tool does not support -toolexec")
	}
	return exec.Command(gobin, "version").Output()
}

func fetchOpenAL() error {
	name := "gomobile-" + openALVersion + ".tar.gz"
	url := "https://dl.google.com/go/mobile/" + name
	if err := fetch(filepath.Join(tmpdir, name), url); err != nil {
		return err
	}
	if err := extract(name, "openal"); err != nil {
		return err
	}
	dst := filepath.Join(ndkccpath, "arm", "sysroot", "usr", "include")
	src := filepath.Join(tmpdir, "openal", "include")
	if err := move(dst, src, "AL"); err != nil {
		return err
	}
	libDst := filepath.Join(ndkccpath, "openal")
	libSrc := filepath.Join(tmpdir, "openal")
	if err := mkdir(libDst); err != nil {
		return nil
	}
	if err := move(libDst, libSrc, "lib"); err != nil {
		return err
	}
	return nil
}

func extract(name, dst string) error {
	if buildX {
		printcmd("tar xfz %s", name)
	}
	if buildN {
		return nil
	}
	tf, err := os.Open(filepath.Join(tmpdir, name))
	if err != nil {
		return err
	}
	defer tf.Close()
	zr, err := gzip.NewReader(tf)
	if err != nil {
		return err
	}
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		dst := filepath.Join(tmpdir, dst+"/"+hdr.Name)
		if hdr.Typeflag == tar.TypeSymlink {
			if err := symlink(hdr.Linkname, dst); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(hdr.Mode)&0777)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func fetchNDK() error {
	if useStrippedNDK {
		if err := fetchStrippedNDK(); err != nil {
			return err
		}
	} else {
		if err := fetchFullNDK(); err != nil {
			return err
		}
	}

	dst := filepath.Join(ndkccpath, "arm")
	dstSysroot := filepath.Join(dst, "sysroot/usr")
	if err := mkdir(dstSysroot); err != nil {
		return err
	}

	srcSysroot := filepath.Join(tmpdir, "android-ndk-r10d/platforms/android-15/arch-arm/usr")
	if err := move(dstSysroot, srcSysroot, "include", "lib"); err != nil {
		return err
	}

	ndkpath := filepath.Join(tmpdir, "android-ndk-r10d/toolchains/arm-linux-androideabi-4.8/prebuilt")
	if goos == "windows" && ndkarch == "x86" {
		ndkpath = filepath.Join(ndkpath, "windows")
	} else {
		ndkpath = filepath.Join(ndkpath, goos+"-"+ndkarch)
	}
	if err := move(dst, ndkpath, "bin", "lib", "libexec"); err != nil {
		return err
	}

	linkpath := filepath.Join(dst, "arm-linux-androideabi/bin")
	if err := mkdir(linkpath); err != nil {
		return err
	}
	for _, name := range []string{"ld", "as", "gcc", "g++"} {
		if goos == "windows" {
			name += ".exe"
		}
		if err := symlink(filepath.Join(dst, "bin", "arm-linux-androideabi-"+name), filepath.Join(linkpath, name)); err != nil {
			return err
		}
	}
	return nil
}

func fetchStrippedNDK() error {
	name := "gomobile-ndk-r10d-" + goos + "-" + ndkarch + ".tar.gz"
	url := "https://dl.google.com/go/mobile/" + name
	if err := fetch(filepath.Join(tmpdir, name), url); err != nil {
		return err
	}
	return extract(name, "")
}

func fetchFullNDK() error {
	ndkName := "android-" + ndkVersion + "-" + goos + "-" + ndkarch + "."
	if goos == "windows" {
		ndkName += "exe"
	} else {
		ndkName += "bin"
	}
	archive := filepath.Join(tmpdir, ndkName)
	url := "https://dl.google.com/android/ndk/" + ndkName
	if err := fetch(archive, url); err != nil {
		return err
	}

	// The self-extracting ndk dist file for Windows
	// terminates with an error (error code 2 - corrupted or incomplete file)
	// but there are no details on what caused this.
	//
	// Strangely, if the file is launched from file
	// browser or unzipped with 7z.exe no error is reported.
	// For now, we require 7z.exe on windows.
	//
	// Once we start using the stripped NDK, this code path
	// will not matter.
	//
	// TODO(hyangah): don't depend on 7z.exe for windows.
	var inflate *exec.Cmd
	if goos != "windows" {
		inflate = exec.Command(archive)
	} else {
		inflate = exec.Command("7z.exe", "x", archive)
	}
	inflate.Dir = tmpdir
	if buildX {
		printcmd("%s", archive)
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

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	var err2 error
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error fetching %v, status: %v", url, resp.Status)
	} else {
		_, err2 = io.Copy(f, resp.Body)
	}
	err3 := resp.Body.Close()
	err4 := f.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	return err4
}

// copyGoroot copies GOROOT from src to dst.
//
// It skips the pkg directory, which is not necessary for make.bash,
// and symlinks .git to avoid a 70MB copy.
func copyGoroot(dst, src string) error {
	if err := mkdir(filepath.Join(dst, "pkg")); err != nil {
		return err
	}
	for _, dir := range []string{"lib", "src"} {
		if err := copyAll(filepath.Join(dst, dir), filepath.Join(src, dir)); err != nil {
			return err
		}
	}
	return symlink(filepath.Join(src, ".git"), filepath.Join(dst, ".git"))
}

func copyAll(dst, src string) error {
	if buildX {
		printcmd("cp -a %s %s", src, dst)
	}
	if buildN {
		return nil
	}
	return doCopyAll(dst, src)
}

func doCopyAll(dst, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, errin error) (err error) {
		if errin != nil {
			return errin
		}
		prefixLen := len(src)
		if len(path) > prefixLen {
			prefixLen++ // file separator
		}
		outpath := filepath.Join(dst, path[prefixLen:])
		if info.IsDir() {
			return os.Mkdir(outpath, 0755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(outpath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if errc := out.Close(); err == nil {
				err = errc
			}
		}()
		_, err = io.Copy(out, in)
		return err
	})
}

func removeAll(path string) error {
	if buildX {
		printcmd("rm -r -f %q", path)
	}
	if buildN {
		return nil
	}
	// os.RemoveAll behaves differently in windows.
	// http://golang.org/issues/9606
	if goos == "windows" {
		resetReadOnlyFlagAll(path)
	}

	return os.RemoveAll(path)
}

func resetReadOnlyFlagAll(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return os.Chmod(path, 0666)
	}
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	names, _ := fd.Readdirnames(-1)
	for _, name := range names {
		resetReadOnlyFlagAll(path + string(filepath.Separator) + name)
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
