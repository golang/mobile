// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/mobile/internal/sdkpath"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

var cmdBind = &command{
	run:   runBind,
	Name:  "bind",
	Usage: "[-target android|" + strings.Join(applePlatforms, "|") + "] [-bootclasspath <path>] [-classpath <path>] [-o output] [build flags] [package]",
	Short: "build a library for Android and iOS",
	Long: `
Bind generates language bindings for the package named by the import
path, and compiles a library for the named target system.

The -target flag takes either android (the default), or one or more
comma-delimited Apple platforms (` + strings.Join(applePlatforms, ", ") + `).

For -target android, the bind command produces an AAR (Android ARchive)
file that archives the precompiled Java API stub classes, the compiled
shared libraries, and all asset files in the /assets subdirectory under
the package directory. The output is named '<package_name>.aar' by
default. This AAR file is commonly used for binary distribution of an
Android library project and most Android IDEs support AAR import. For
example, in Android Studio (1.2+), an AAR file can be imported using
the module import wizard (File > New > New Module > Import .JAR or
.AAR package), and setting it as a new dependency
(File > Project Structure > Dependencies).  This requires 'javac'
(version 1.8+) and Android SDK (API level 16 or newer) to build the
library for Android. The ANDROID_HOME and ANDROID_NDK_HOME environment
variables can be used to specify the Android SDK and NDK if they are
not in the default locations. Use the -javapkg flag to specify the Java
package prefix for the generated classes.

By default, -target=android builds shared libraries for all supported
instruction sets (arm, arm64, 386, amd64). A subset of instruction sets
can be selected by specifying target type with the architecture name. E.g.,
-target=android/arm,android/386.

For Apple -target platforms, gomobile must be run on an OS X machine with
Xcode installed. The generated Objective-C types can be prefixed with the
-prefix flag.

For -target android, the -bootclasspath and -classpath flags are used to
control the bootstrap classpath and the classpath for Go wrappers to Java
classes.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -n, -x, -gcflags, -ldflags, -tags, -trimpath, and -work
are shared with the build command. For documentation, see 'go help build'.
`,
}

func runBind(cmd *command) error {
	cleanup, err := buildEnvInit()
	if err != nil {
		return err
	}
	defer cleanup()

	args := cmd.flag.Args()

	targets, err := parseBuildTarget(buildTarget)
	if err != nil {
		return fmt.Errorf(`invalid -target=%q: %v`, buildTarget, err)
	}

	if isAndroidPlatform(targets[0].platform) {
		if bindPrefix != "" {
			return fmt.Errorf("-prefix is supported only for Apple targets")
		}
		if _, err := ndkRoot(targets[0]); err != nil {
			return err
		}
	} else {
		if bindJavaPkg != "" {
			return fmt.Errorf("-javapkg is supported only for android target")
		}
	}

	var gobind string
	if !buildN {
		gobind, err = exec.LookPath("gobind")
		if err != nil {
			return errors.New("gobind was not found. Please run gomobile init before trying again")
		}
	} else {
		gobind = "gobind"
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	// TODO(ydnar): this should work, unless build tags affect loading a single package.
	// Should we try to import packages with different build tags per platform?
	pkgs, err := packages.Load(packagesConfig(targets[0]), args...)
	if err != nil {
		return err
	}

	// check if any of the package is main
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			return fmt.Errorf(`binding "main" package (%s) is not supported`, pkg.PkgPath)
		}
	}

	switch {
	case isAndroidPlatform(targets[0].platform):
		return goAndroidBind(gobind, pkgs, targets)
	case isApplePlatform(targets[0].platform):
		if !xcodeAvailable() {
			return fmt.Errorf("-target=%q requires Xcode", buildTarget)
		}
		return goAppleBind(gobind, pkgs, targets)
	default:
		return fmt.Errorf(`invalid -target=%q`, buildTarget)
	}
}

var (
	bindPrefix        string // -prefix
	bindJavaPkg       string // -javapkg
	bindClasspath     string // -classpath
	bindBootClasspath string // -bootclasspath
)

func init() {
	// bind command specific commands.
	cmdBind.flag.StringVar(&bindJavaPkg, "javapkg", "",
		"specifies custom Java package path prefix. Valid only with -target=android.")
	cmdBind.flag.StringVar(&bindPrefix, "prefix", "",
		"custom Objective-C name prefix. Valid only with -target=ios.")
	cmdBind.flag.StringVar(&bindClasspath, "classpath", "", "The classpath for imported Java classes. Valid only with -target=android.")
	cmdBind.flag.StringVar(&bindBootClasspath, "bootclasspath", "", "The bootstrap classpath for imported Java classes. Valid only with -target=android.")
}

func bootClasspath() (string, error) {
	if bindBootClasspath != "" {
		return bindBootClasspath, nil
	}
	apiPath, err := sdkpath.AndroidAPIPath(buildAndroidAPI)
	if err != nil {
		return "", err
	}
	return filepath.Join(apiPath, "android.jar"), nil
}

func copyFile(dst, src string) error {
	if buildX {
		printcmd("cp %s %s", src, dst)
	}
	return writeFile(dst, func(w io.Writer) error {
		if buildN {
			return nil
		}
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("cp %s %s failed: %v", src, dst, err)
		}
		return nil
	})
}

func writeFile(filename string, generate func(io.Writer) error) error {
	if buildV {
		fmt.Fprintf(os.Stderr, "write %s\n", filename)
	}

	if err := mkdir(filepath.Dir(filename)); err != nil {
		return err
	}

	if buildN {
		return generate(ioutil.Discard)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	return generate(f)
}

func packagesConfig(t targetInfo) *packages.Config {
	config := &packages.Config{}
	// Add CGO_ENABLED=1 explicitly since Cgo is disabled when GOOS is different from host OS.
	config.Env = append(os.Environ(), "GOARCH="+t.arch, "GOOS="+platformOS(t.platform), "CGO_ENABLED=1")
	tags := append(buildTags[:], platformTags(t.platform)...)

	if len(tags) > 0 {
		config.BuildFlags = []string{"-tags=" + strings.Join(tags, ",")}
	}
	return config
}

// getModuleVersions returns a module information at the directory src.
func getModuleVersions(targetPlatform string, targetArch string, src string) (*modfile.File, error) {
	cmd := exec.Command("go", "list")
	cmd.Env = append(os.Environ(), "GOOS="+platformOS(targetPlatform), "GOARCH="+targetArch)

	tags := append(buildTags[:], platformTags(targetPlatform)...)

	// TODO(hyangah): probably we don't need to add all the dependencies.
	cmd.Args = append(cmd.Args, "-m", "-json", "-tags="+strings.Join(tags, ","), "all")
	cmd.Dir = src

	output, err := cmd.Output()
	if err != nil {
		// Module information is not available at src.
		return nil, nil
	}

	type Module struct {
		Main    bool
		Path    string
		Version string
		Dir     string
		Replace *Module
	}

	f := &modfile.File{}
	if err := f.AddModuleStmt("gobind"); err != nil {
		return nil, err
	}
	e := json.NewDecoder(bytes.NewReader(output))
	for {
		var mod *Module
		err := e.Decode(&mod)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if mod != nil {
			if mod.Replace != nil {
				p, v := mod.Replace.Path, mod.Replace.Version
				if modfile.IsDirectoryPath(p) {
					// replaced by a local directory
					p = mod.Replace.Dir
				}
				if err := f.AddReplace(mod.Path, mod.Version, p, v); err != nil {
					return nil, err
				}
			} else {
				// When the version part is empty, the module is local and mod.Dir represents the location.
				if v := mod.Version; v == "" {
					if err := f.AddReplace(mod.Path, mod.Version, mod.Dir, ""); err != nil {
						return nil, err
					}
				} else {
					if err := f.AddRequire(mod.Path, v); err != nil {
						return nil, err
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
	}

	v, err := ensureGoVersion()
	if err != nil {
		return nil, err
	}
	// ensureGoVersion can return an empty string for a devel version. In this case, use the minimum version.
	if v == "" {
		v = fmt.Sprintf("go1.%d", minimumGoMinorVersion)
	}
	if err := f.AddGoStmt(strings.TrimPrefix(v, "go")); err != nil {
		return nil, err
	}

	return f, nil
}

// writeGoMod writes go.mod file at dir when Go modules are used.
func writeGoMod(dir, targetPlatform, targetArch string) error {
	m, err := areGoModulesUsed()
	if err != nil {
		return err
	}
	// If Go modules are not used, go.mod should not be created because the dependencies might not be compatible with Go modules.
	if !m {
		return nil
	}

	return writeFile(filepath.Join(dir, "go.mod"), func(w io.Writer) error {
		f, err := getModuleVersions(targetPlatform, targetArch, ".")
		if err != nil {
			return err
		}
		if f == nil {
			return nil
		}
		bs, err := f.Format()
		if err != nil {
			return err
		}
		if _, err := w.Write(bs); err != nil {
			return err
		}
		return nil
	})
}

var (
	areGoModulesUsedResult struct {
		used bool
		err  error
	}
	areGoModulesUsedOnce sync.Once
)

func areGoModulesUsed() (bool, error) {
	areGoModulesUsedOnce.Do(func() {
		out, err := exec.Command("go", "env", "GOMOD").Output()
		if err != nil {
			areGoModulesUsedResult.err = err
			return
		}
		outstr := strings.TrimSpace(string(out))
		areGoModulesUsedResult.used = outstr != ""
	})
	return areGoModulesUsedResult.used, areGoModulesUsedResult.err
}
