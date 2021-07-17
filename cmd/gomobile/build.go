// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gendex.go -o dex.go

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var tmpdir string

var cmdBuild = &command{
	run:   runBuild,
	Name:  "build",
	Usage: "[-target android|" + strings.Join(darwinPlatforms, "|") + "] [-o output] [-bundleid bundleID] [build flags] [package]",
	Short: "compile android APK and iOS app",
	Long: `
Build compiles and encodes the app named by the import path.

The named package must define a main function.

The -target flag takes a target platform, either android (the
default) or ` + strings.Join(darwinPlatforms, ", ") + `.

For -target android, if an AndroidManifest.xml is defined in the
package directory, it is added to the APK output. Otherwise, a default
manifest is generated. By default, this builds a fat APK for all supported
instruction sets (arm, 386, amd64, arm64). A subset of instruction sets can
be selected by specifying target type with the architecture name. E.g.
-target=android/arm,android/386.

For Apple -target platforms, gomobile must be run on an OS X machine with
Xcode installed.

If the package directory contains an assets subdirectory, its contents
are copied into the output.

Flag -iosversion sets the minimal version of the iOS SDK to compile against.
The default version is 13.0.

Flag -androidapi sets the Android API version to compile against.
The default and minimum is 15.

The -bundleid flag is required for -target ios and sets the bundle ID to use
with the app.

The -o flag specifies the output file name. If not specified, the
output file name depends on the package built.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -i, -n, -x, -gcflags, -ldflags, -tags, -trimpath, and -work are
shared with the build command. For documentation, see 'go help build'.
`,
}

func runBuild(cmd *command) (err error) {
	_, err = runBuildImpl(cmd)
	return
}

// runBuildImpl builds a package for mobiles based on the given commands.
// runBuildImpl returns a built package information and an error if exists.
func runBuildImpl(cmd *command) (*packages.Package, error) {
	cleanup, err := buildEnvInit()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	args := cmd.flag.Args()

	targetPlatforms, targetArchs, err := parseBuildTarget(buildTarget)
	if err != nil {
		return nil, fmt.Errorf(`invalid -target=%q: %v`, buildTarget, err)
	}

	var buildPath string
	switch len(args) {
	case 0:
		buildPath = "."
	case 1:
		buildPath = args[0]
	default:
		cmd.usage()
		os.Exit(1)
	}

	// TODO(ydnar): this should work, unless build tags affect loading a single package.
	// Should we try to import packages with different build tags per platform?
	pkgs, err := packages.Load(packagesConfig(targetPlatforms[0]), buildPath)
	if err != nil {
		return nil, err
	}

	// len(pkgs) can be more than 1 e.g., when the specified path includes `...`.
	if len(pkgs) != 1 {
		cmd.usage()
		os.Exit(1)
	}

	pkg := pkgs[0]

	if pkg.Name != "main" && buildO != "" {
		return nil, fmt.Errorf("cannot set -o when building non-main package")
	}

	var nmpkgs map[string]bool
	switch {
	case isAndroidPlatform(targetPlatforms[0]):
		if pkg.Name != "main" {
			for _, arch := range targetArchs {
				if err := goBuild(pkg.PkgPath, androidEnv[arch]); err != nil {
					return nil, err
				}
			}
			return pkg, nil
		}
		nmpkgs, err = goAndroidBuild(pkg, targetArchs)
		if err != nil {
			return nil, err
		}
	case isDarwinPlatform(targetPlatforms[0]):
		if !xcodeAvailable() {
			return nil, fmt.Errorf("-target=%s requires XCode", buildTarget)
		}
		if pkg.Name != "main" {
			for _, platform := range targetPlatforms {
				// Catalyst support requires iOS 13+
				v, _ := strconv.ParseFloat(buildIOSVersion, 64)
				if platform == "maccatalyst" && v < 13.0 {
					return nil, errors.New("catalyst requires -iosversion=13 or higher")
				}

				for _, arch := range targetArchs {
					// Skip unrequested architectures
					if !isSupportedArch(platform, arch) {
						continue
					}
					if err := goBuild(pkg.PkgPath, darwinEnv[platform+"/"+arch]); err != nil {
						return nil, err
					}
				}
			}
			return pkg, nil
		}
		if buildBundleID == "" {
			return nil, fmt.Errorf("-target=ios requires -bundleid set")
		}
		nmpkgs, err = goDarwinBuild(pkg, buildBundleID, targetPlatforms, targetArchs)
		if err != nil {
			return nil, err
		}
	}

	if !nmpkgs["golang.org/x/mobile/app"] {
		return nil, fmt.Errorf(`%s does not import "golang.org/x/mobile/app"`, pkg.PkgPath)
	}

	return pkg, nil
}

var nmRE = regexp.MustCompile(`[0-9a-f]{8} t (?:.*/vendor/|_)?(golang.org/x.*/[^.]*)`)

func extractPkgs(nm string, path string) (map[string]bool, error) {
	if buildN {
		return map[string]bool{"golang.org/x/mobile/app": true}, nil
	}
	r, w := io.Pipe()
	cmd := exec.Command(nm, path)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	nmpkgs := make(map[string]bool)
	errc := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(r)
		for s.Scan() {
			if res := nmRE.FindStringSubmatch(s.Text()); res != nil {
				nmpkgs[res[1]] = true
			}
		}
		errc <- s.Err()
	}()

	err := cmd.Run()
	w.Close()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %v", nm, path, err)
	}
	if err := <-errc; err != nil {
		return nil, fmt.Errorf("%s %s: %v", nm, path, err)
	}
	return nmpkgs, nil
}

var xout io.Writer = os.Stderr

func printcmd(format string, args ...interface{}) {
	cmd := fmt.Sprintf(format+"\n", args...)
	if tmpdir != "" {
		cmd = strings.Replace(cmd, tmpdir, "$WORK", -1)
	}
	if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
		cmd = strings.Replace(cmd, androidHome, "$ANDROID_HOME", -1)
	}
	if gomobilepath != "" {
		cmd = strings.Replace(cmd, gomobilepath, "$GOMOBILE", -1)
	}
	if gopath := goEnv("GOPATH"); gopath != "" {
		cmd = strings.Replace(cmd, gopath, "$GOPATH", -1)
	}
	if env := os.Getenv("HOMEPATH"); env != "" {
		cmd = strings.Replace(cmd, env, "$HOMEPATH", -1)
	}
	fmt.Fprint(xout, cmd)
}

// "Build flags", used by multiple commands.
var (
	buildA          bool        // -a
	buildI          bool        // -i
	buildN          bool        // -n
	buildV          bool        // -v
	buildX          bool        // -x
	buildO          string      // -o
	buildGcflags    string      // -gcflags
	buildLdflags    string      // -ldflags
	buildTarget     string      // -target
	buildTrimpath   bool        // -trimpath
	buildWork       bool        // -work
	buildBundleID   string      // -bundleid
	buildIOSVersion string      // -iosversion
	buildAndroidAPI int         // -androidapi
	buildTags       stringsFlag // -tags
)

func addBuildFlags(cmd *command) {
	cmd.flag.StringVar(&buildO, "o", "", "")
	cmd.flag.StringVar(&buildGcflags, "gcflags", "", "")
	cmd.flag.StringVar(&buildLdflags, "ldflags", "", "")
	cmd.flag.StringVar(&buildTarget, "target", "android", "")
	cmd.flag.StringVar(&buildBundleID, "bundleid", "", "")
	cmd.flag.StringVar(&buildIOSVersion, "iosversion", "13.0", "")
	cmd.flag.IntVar(&buildAndroidAPI, "androidapi", minAndroidAPI, "")

	cmd.flag.BoolVar(&buildA, "a", false, "")
	cmd.flag.BoolVar(&buildI, "i", false, "")
	cmd.flag.BoolVar(&buildTrimpath, "trimpath", false, "")
	cmd.flag.Var(&buildTags, "tags", "")
}

func addBuildFlagsNVXWork(cmd *command) {
	cmd.flag.BoolVar(&buildN, "n", false, "")
	cmd.flag.BoolVar(&buildV, "v", false, "")
	cmd.flag.BoolVar(&buildX, "x", false, "")
	cmd.flag.BoolVar(&buildWork, "work", false, "")
}

type binInfo struct {
	hasPkgApp bool
	hasPkgAL  bool
}

func init() {
	addBuildFlags(cmdBuild)
	addBuildFlagsNVXWork(cmdBuild)

	addBuildFlags(cmdInstall)
	addBuildFlagsNVXWork(cmdInstall)

	addBuildFlagsNVXWork(cmdInit)

	addBuildFlags(cmdBind)
	addBuildFlagsNVXWork(cmdBind)

	addBuildFlagsNVXWork(cmdClean)
}

func goBuild(src string, env []string, args ...string) error {
	return goCmd("build", []string{src}, env, args...)
}

func goBuildAt(at string, src string, env []string, args ...string) error {
	return goCmdAt(at, "build", []string{src}, env, args...)
}

func goInstall(srcs []string, env []string, args ...string) error {
	return goCmd("install", srcs, env, args...)
}

func goCmd(subcmd string, srcs []string, env []string, args ...string) error {
	return goCmdAt("", subcmd, srcs, env, args...)
}

func goCmdAt(at string, subcmd string, srcs []string, env []string, args ...string) error {
	cmd := exec.Command("go", subcmd)
	tags := buildTags
	if len(tags) > 0 {
		cmd.Args = append(cmd.Args, "-tags", strings.Join(tags, ","))
	}
	if buildV {
		cmd.Args = append(cmd.Args, "-v")
	}
	if subcmd != "install" && buildI {
		cmd.Args = append(cmd.Args, "-i")
	}
	if buildX {
		cmd.Args = append(cmd.Args, "-x")
	}
	if buildGcflags != "" {
		cmd.Args = append(cmd.Args, "-gcflags", buildGcflags)
	}
	if buildLdflags != "" {
		cmd.Args = append(cmd.Args, "-ldflags", buildLdflags)
	}
	if buildTrimpath {
		cmd.Args = append(cmd.Args, "-trimpath")
	}
	if buildWork {
		cmd.Args = append(cmd.Args, "-work")
	}
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, srcs...)
	cmd.Env = append([]string{}, env...)
	cmd.Dir = at
	return runCmd(cmd)
}

func goModTidyAt(at string, env []string) error {
	cmd := exec.Command("go", "mod", "tidy")
	if buildV {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Env = append([]string{}, env...)
	cmd.Dir = at
	return runCmd(cmd)
}

// parseBuildTarget parses buildTarget into 1 or more platforms and architectures.
// Returns an error if buildTarget contains invalid input.
// Example valid target strings:
//    android
//    android/arm64,android/386,android/amd64
//    ios,iossimulator,maccatalyst
//    macos/amd64
func parseBuildTarget(buildTarget string) (targetPlatforms, targetArchs []string, _ error) {
	if buildTarget == "" {
		return nil, nil, fmt.Errorf(`invalid target ""`)
	}

	var platforms, archs orderedSet

	var isAndroid, isDarwin bool
	for _, target := range strings.Split(buildTarget, ",") {
		tuple := strings.SplitN(target, "/", 2)
		platform := tuple[0]
		hasArch := len(tuple) == 2

		if isAndroidPlatform(platform) {
			isAndroid = true
		} else if isDarwinPlatform(platform) {
			isDarwin = true
		} else {
			return nil, nil, fmt.Errorf("unsupported platform: %q", platform)
		}
		if isAndroid && isDarwin {
			return nil, nil, fmt.Errorf(`cannot mix android and darwin platforms`)
		}

		platforms.Add(platform)
		if platform == "ios" {
			platforms.Add("iossimulator")
		}

		if hasArch {
			arch := tuple[1]
			if !isSupportedArch(platform, arch) {
				return nil, nil, fmt.Errorf(`unsupported platform/arch: %q`, target)
			}
			archs.Add(arch)
		}
	}

	if len(archs.Slice()) == 0 {
		for _, platform := range platforms.Slice() {
			for _, arch := range platformArchs(platform) {
				archs.Add(arch)
			}
		}
	}

	return platforms.Slice(), archs.Slice(), nil
}

type orderedSet struct {
	strs []string
	m    map[string]bool
}

func (s *orderedSet) Has(str string) bool {
	return s.m[str]
}

func (s *orderedSet) Add(strs ...string) {
	for _, str := range strs {
		if !s.m[str] {
			if s.m == nil {
				s.m = make(map[string]bool)
			}
			s.m[str] = true
			s.strs = append(s.strs, str)
		}
	}
}

func (s *orderedSet) Slice() []string {
	return s.strs
}
