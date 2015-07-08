// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gendex.go -o dex.go

package main

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var ctx = build.Default
var pkg *build.Package // TODO(crawshaw): remove global pkg variable
var tmpdir string

var cmdBuild = &command{
	run:   runBuild,
	Name:  "build",
	Usage: "[-target android|ios] [-o output] [build flags] [package]",
	Short: "compile android APK and iOS app",
	Long: `
Build compiles and encodes the app named by the import path.

The named package must define a main function.

The -target flag takes a target system name, either android (the
default) or ios.

For -target android, if an AndroidManifest.xml is defined in the
package directory, it is added to the APK output. Otherwise, a default
manifest is generated.

For -target ios, gomobile must be run on an OS X machine with Xcode
installed. Support is not complete.

If the package directory contains an assets subdirectory, its contents
are copied into the output.

The -o flag specifies the output file name. If not specified, the
output file name depends on the package built.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -i, -n, -x, and -tags are shared with the build command.
For documentation, see 'go help build'.
`,
}

func runBuild(cmd *command) (err error) {
	cleanup, err := envInit()
	if err != nil {
		return err
	}
	defer cleanup()

	args := cmd.flag.Args()

	switch len(args) {
	case 0:
		pkg, err = ctx.ImportDir(cwd, build.ImportComment)
	case 1:
		pkg, err = ctx.Import(args[0], cwd, build.ImportComment)
	default:
		cmd.usage()
		os.Exit(1)
	}
	if err != nil {
		return err
	}

	if pkg.Name != "main" && buildO != "" {
		return fmt.Errorf("cannot set -o when building non-main package")
	}

	switch buildTarget {
	case "android":
		if pkg.Name != "main" {
			return goBuild(pkg.ImportPath, androidArmEnv)
		}
		if err := goAndroidBuild(pkg); err != nil {
			return err
		}
	case "ios":
		if runtime.GOOS != "darwin" {
			return fmt.Errorf("-target=ios requires darwin host")
		}
		if pkg.Name != "main" {
			if err := goBuild(pkg.ImportPath, darwinArmEnv); err != nil {
				return err
			}
			return goBuild(pkg.ImportPath, darwinArm64Env)
		}
		if err := goIOSBuild(pkg); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`unknown -target, %q.`, buildTarget)
	}

	// TODO(crawshaw): This is an incomplete package scan.
	// A complete package scan would be too expensive. Instead,
	// fake it. After the binary is built, scan its symbols
	// with nm and look for the app and al packages.
	if err := importsApp(pkg); err != nil {
		return err
	}

	return nil
}

func importsApp(pkg *build.Package) error {
	// Building a program, make sure it is appropriate for mobile.
	for _, path := range pkg.Imports {
		if path == "golang.org/x/mobile/app" {
			return nil
		}
	}
	return fmt.Errorf(`%s does not import "golang.org/x/mobile/app"`, pkg.ImportPath)
}

var xout io.Writer = os.Stderr

func printcmd(format string, args ...interface{}) {
	cmd := fmt.Sprintf(format+"\n", args...)
	if tmpdir != "" {
		cmd = strings.Replace(cmd, tmpdir, "$WORK", -1)
	}
	if gomobilepath != "" {
		cmd = strings.Replace(cmd, gomobilepath, "$GOMOBILE", -1)
	}
	if goroot := goEnv("GOROOT"); goroot != "" {
		cmd = strings.Replace(cmd, goroot, "$GOROOT", -1)
	}
	if gopath := goEnv("GOPATH"); gopath != "" {
		cmd = strings.Replace(cmd, gopath, "$GOPATH", -1)
	}
	if env := os.Getenv("HOME"); env != "" {
		cmd = strings.Replace(cmd, env, "$HOME", -1)
	}
	if env := os.Getenv("HOMEPATH"); env != "" {
		cmd = strings.Replace(cmd, env, "$HOMEPATH", -1)
	}
	fmt.Fprint(xout, cmd)
}

// "Build flags", used by multiple commands.
var (
	buildA      bool   // -a
	buildI      bool   // -i
	buildN      bool   // -n
	buildV      bool   // -v
	buildX      bool   // -x
	buildO      string // -o
	buildTarget string // -target
)

func addBuildFlags(cmd *command) {
	cmd.flag.StringVar(&buildO, "o", "", "")
	cmd.flag.StringVar(&buildTarget, "target", "android", "")

	cmd.flag.BoolVar(&buildA, "a", false, "")
	cmd.flag.BoolVar(&buildI, "i", false, "")
	cmd.flag.Var((*stringsFlag)(&ctx.BuildTags), "tags", "")
}

func addBuildFlagsNVX(cmd *command) {
	cmd.flag.BoolVar(&buildN, "n", false, "")
	cmd.flag.BoolVar(&buildV, "v", false, "")
	cmd.flag.BoolVar(&buildX, "x", false, "")
}

type binInfo struct {
	hasPkgApp bool
	hasPkgAL  bool
}

func init() {
	addBuildFlags(cmdBuild)
	addBuildFlagsNVX(cmdBuild)

	addBuildFlags(cmdInstall)
	addBuildFlagsNVX(cmdInstall)

	addBuildFlagsNVX(cmdInit)

	addBuildFlags(cmdBind)
	addBuildFlagsNVX(cmdBind)
}

// environ merges os.Environ and the given "key=value" pairs.
func environ(kv []string) []string {
	envs := map[string]string{}

	cur := os.Environ()
	new := make([]string, 0, len(cur)+len(kv))
	for _, ev := range cur {
		elem := strings.SplitN(ev, "=", 2)
		if len(elem) != 2 || elem[0] == "" {
			// pass the env var of unusual form untouched.
			// e.g. Windows may have env var names starting with "=".
			new = append(new, ev)
			continue
		}
		if goos == "windows" {
			elem[0] = strings.ToUpper(elem[0])
		}
		envs[elem[0]] = elem[1]
	}
	for _, ev := range kv {
		elem := strings.SplitN(ev, "=", 2)
		if len(elem) != 2 || elem[0] == "" {
			panic(fmt.Sprintf("malformed env var %q from input", ev))
		}
		if goos == "windows" {
			elem[0] = strings.ToUpper(elem[0])
		}
		envs[elem[0]] = elem[1]
	}
	for k, v := range envs {
		new = append(new, k+"="+v)
	}
	return new
}

func goBuild(src string, env []string, args ...string) error {
	cmd := exec.Command(
		`go`,
		`build`,
		`-tags=`+strconv.Quote(strings.Join(ctx.BuildTags, ",")),
	)
	if buildV {
		cmd.Args = append(cmd.Args, "-v")
	}
	if buildI {
		cmd.Args = append(cmd.Args, "-i")
	}
	if buildX {
		cmd.Args = append(cmd.Args, "-x")
	}
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, src)
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env,
		`CGO_ENABLED=1`,
		`GOROOT=`+goEnv("GOROOT"),
		`GOPATH=`+goEnv("GOPATH"),
	)
	buf := new(bytes.Buffer)
	buf.WriteByte('\n')
	if buildV {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = buf
		cmd.Stderr = buf
	}

	if buildX {
		printcmd("%s", strings.Join(cmd.Env, " ")+" "+strings.Join(cmd.Args, " "))
	}
	if !buildN {
		cmd.Env = environ(cmd.Env)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build failed: %v%s", err, buf)
		}
	}
	return nil
}
