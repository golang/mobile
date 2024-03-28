// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

func goAppleBind(gobind string, pkgs []*packages.Package, targets []targetInfo) error {
	var name string
	var title string

	if buildO == "" {
		name = pkgs[0].Name
		title = strings.Title(name)
		buildO = title + ".xcframework"
	} else {
		if !strings.HasSuffix(buildO, ".xcframework") {
			return fmt.Errorf("static framework name %q missing .xcframework suffix", buildO)
		}
		base := filepath.Base(buildO)
		name = base[:len(base)-len(".xcframework")]
		title = strings.Title(name)
	}

	if err := removeAll(buildO); err != nil {
		return err
	}

	outDirsForPlatform := map[string]string{}
	for _, t := range targets {
		outDirsForPlatform[t.platform] = filepath.Join(tmpdir, t.platform)
	}

	// Run the gobind command for each platform
	var gobindWG errgroup.Group
	for platform, outDir := range outDirsForPlatform {
		platform := platform
		outDir := outDir
		gobindWG.Go(func() error {
			// Catalyst support requires iOS 13+
			v, _ := strconv.ParseFloat(buildIOSVersion, 64)
			if platform == "maccatalyst" && v < 13.0 {
				return errors.New("catalyst requires -iosversion=13 or higher")
			}

			// Run gobind once per platform to generate the bindings
			cmd := exec.Command(
				gobind,
				"-lang=go,objc",
				"-outdir="+outDir,
			)
			cmd.Env = append(cmd.Env, "GOOS="+platformOS(platform))
			cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
			tags := append(buildTags[:], platformTags(platform)...)
			cmd.Args = append(cmd.Args, "-tags="+strings.Join(tags, ","))
			if bindPrefix != "" {
				cmd.Args = append(cmd.Args, "-prefix="+bindPrefix)
			}
			for _, p := range pkgs {
				cmd.Args = append(cmd.Args, p.PkgPath)
			}
			if err := runCmd(cmd); err != nil {
				return err
			}
			return nil
		})
	}
	if err := gobindWG.Wait(); err != nil {
		return err
	}

	modulesUsed, err := areGoModulesUsed()
	if err != nil {
		return err
	}

	// Build archive files.
	var buildWG errgroup.Group
	for _, t := range targets {
		t := t
		buildWG.Go(func() error {
			outDir := outDirsForPlatform[t.platform]
			outSrcDir := filepath.Join(outDir, "src")

			if modulesUsed {
				// Copy the source directory for each architecture for concurrent building.
				newOutSrcDir := filepath.Join(outDir, "src-"+t.arch)
				if !buildN {
					if err := doCopyAll(newOutSrcDir, outSrcDir); err != nil {
						return err
					}
				}
				outSrcDir = newOutSrcDir
			}

			// Copy the environment variables to make this function concurrent-safe.
			env := make([]string, len(appleEnv[t.String()]))
			copy(env, appleEnv[t.String()])

			// Add the generated packages to GOPATH for reverse bindings.
			gopath := fmt.Sprintf("GOPATH=%s%c%s", outDir, filepath.ListSeparator, goEnv("GOPATH"))
			env = append(env, gopath)

			// Run `go mod tidy` to force to create go.sum.
			// Without go.sum, `go build` fails as of Go 1.16.
			if modulesUsed {
				if err := writeGoMod(outSrcDir, t.platform, t.arch); err != nil {
					return err
				}
				if err := goModTidyAt(outSrcDir, env); err != nil {
					return err
				}
			}

			if err := goAppleBindArchive(appleArchiveFilepath(name, t), env, outSrcDir); err != nil {
				return fmt.Errorf("%s/%s: %v", t.platform, t.arch, err)
			}

			return nil
		})
	}
	if err := buildWG.Wait(); err != nil {
		return err
	}

	var frameworkDirs []string
	frameworkArchCount := map[string]int{}
	for _, t := range targets {
		outDir := outDirsForPlatform[t.platform]
		gobindDir := filepath.Join(outDir, "src", "gobind")

		env := appleEnv[t.String()][:]
		sdk := getenv(env, "DARWIN_SDK")

		frameworkDir := filepath.Join(tmpdir, t.platform, sdk, title+".framework")
		frameworkDirs = append(frameworkDirs, frameworkDir)
		frameworkArchCount[frameworkDir] = frameworkArchCount[frameworkDir] + 1

		frameworkLayout, err := frameworkLayoutForTarget(t, title)
		if err != nil {
			return err
		}

		titlePath := filepath.Join(frameworkDir, frameworkLayout.binaryPath, title)
		if frameworkArchCount[frameworkDir] > 1 {
			// Not the first static lib, attach to a fat library and skip create headers
			fatCmd := exec.Command(
				"xcrun",
				"lipo", appleArchiveFilepath(name, t), titlePath, "-create", "-output", titlePath,
			)
			if err := runCmd(fatCmd); err != nil {
				return err
			}
			continue
		}

		headersDir := filepath.Join(frameworkDir, frameworkLayout.headerPath)
		if err := mkdir(headersDir); err != nil {
			return err
		}

		lipoCmd := exec.Command(
			"xcrun",
			"lipo", appleArchiveFilepath(name, t), "-create", "-o", titlePath,
		)
		if err := runCmd(lipoCmd); err != nil {
			return err
		}

		fileBases := make([]string, len(pkgs)+1)
		for i, pkg := range pkgs {
			fileBases[i] = bindPrefix + strings.Title(pkg.Name)
		}
		fileBases[len(fileBases)-1] = "Universe"

		// Copy header file next to output archive.
		var headerFiles []string
		if len(fileBases) == 1 {
			headerFiles = append(headerFiles, title+".h")
			err := copyFile(
				filepath.Join(headersDir, title+".h"),
				filepath.Join(gobindDir, bindPrefix+title+".objc.h"),
			)
			if err != nil {
				return err
			}
		} else {
			for _, fileBase := range fileBases {
				headerFiles = append(headerFiles, fileBase+".objc.h")
				err := copyFile(
					filepath.Join(headersDir, fileBase+".objc.h"),
					filepath.Join(gobindDir, fileBase+".objc.h"),
				)
				if err != nil {
					return err
				}
			}
			err := copyFile(
				filepath.Join(headersDir, "ref.h"),
				filepath.Join(gobindDir, "ref.h"),
			)
			if err != nil {
				return err
			}
			headerFiles = append(headerFiles, title+".h")
			err = writeFile(filepath.Join(headersDir, title+".h"), func(w io.Writer) error {
				return appleBindHeaderTmpl.Execute(w, map[string]interface{}{
					"pkgs": pkgs, "title": title, "bases": fileBases,
				})
			})
			if err != nil {
				return err
			}
		}

		frameworkInfoPlistDir := filepath.Join(frameworkDir, frameworkLayout.infoPlistPath)
		if err := mkdir(frameworkInfoPlistDir); err != nil {
			return err
		}
		err = writeFile(filepath.Join(frameworkInfoPlistDir, "Info.plist"), func(w io.Writer) error {
			fmVersion := fmt.Sprintf("0.0.%d", time.Now().Unix())
			infoFrameworkPlistlData := infoFrameworkPlistlData{
				BundleID:       escapePlistValue(rfc1034Label(title)),
				ExecutableName: escapePlistValue(title),
				Version:        escapePlistValue(fmVersion),
			}
			infoplist := new(bytes.Buffer)
			if err := infoFrameworkPlistTmpl.Execute(infoplist, infoFrameworkPlistlData); err != nil {
				return err
			}
			_, err := w.Write(infoplist.Bytes())
			return err
		})
		if err != nil {
			return err
		}

		var mmVals = struct {
			Module  string
			Headers []string
		}{
			Module:  title,
			Headers: headerFiles,
		}
		modulesDir := filepath.Join(frameworkDir, frameworkLayout.modulePath)
		err = writeFile(filepath.Join(modulesDir, "module.modulemap"), func(w io.Writer) error {
			return appleModuleMapTmpl.Execute(w, mmVals)
		})
		if err != nil {
			return err
		}

		for src, dst := range frameworkLayout.symlinks {
			if err := symlink(src, filepath.Join(frameworkDir, dst)); err != nil {
				return err
			}
		}
	}

	// Finally combine all frameworks to an XCFramework
	xcframeworkArgs := []string{"-create-xcframework"}

	for _, dir := range frameworkDirs {
		// On macOS, a temporary directory starts with /var, which is a symbolic link to /private/var.
		// And in gomobile, a temporary directory is usually used as a working directly.
		// Unfortunately, xcodebuild in Xcode 15 seems to have a bug and might not be able to understand fullpaths with symbolic links.
		// As a workaround, resolve the path with symbolic links by filepath.EvalSymlinks.
		dir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return err
		}
		xcframeworkArgs = append(xcframeworkArgs, "-framework", dir)
	}

	xcframeworkArgs = append(xcframeworkArgs, "-output", buildO)
	cmd := exec.Command("xcodebuild", xcframeworkArgs...)
	err = runCmd(cmd)
	return err
}

type frameworkLayout struct {
	headerPath    string
	binaryPath    string
	modulePath    string
	infoPlistPath string
	// symlinks to create in the framework. Maps src (relative to dst) -> dst (relative to framework bundle root)
	symlinks map[string]string
}

// frameworkLayoutForTarget generates the filestructure for a framework for the given target platform (macos, ios, etc),
// according to Apple's spec https://developer.apple.com/documentation/bundleresources/placing_content_in_a_bundle
func frameworkLayoutForTarget(t targetInfo, title string) (*frameworkLayout, error) {
	switch t.platform {
	case "macos", "maccatalyst":
		return &frameworkLayout{
			headerPath:    "Versions/A/Headers",
			binaryPath:    "Versions/A",
			modulePath:    "Versions/A/Modules",
			infoPlistPath: "Versions/A/Resources",
			symlinks: map[string]string{
				"A":                                      "Versions/Current",
				"Versions/Current/Resources":             "Resources",
				"Versions/Current/Headers":               "Headers",
				"Versions/Current/Modules":               "Modules",
				filepath.Join("Versions/Current", title): title,
			},
		}, nil
	case "ios", "iossimulator":
		return &frameworkLayout{
			headerPath:    "Headers",
			binaryPath:    ".",
			modulePath:    "Modules",
			infoPlistPath: ".",
		}, nil
	}

	return nil, fmt.Errorf("unsupported platform %q", t.platform)
}

type infoFrameworkPlistlData struct {
	BundleID       string
	ExecutableName string
	Version        string
}

// infoFrameworkPlistTmpl is a template for the Info.plist file in a framework.
// Minimum OS version == 100.0 is a workaround for SPM issue
// https://github.com/firebase/firebase-ios-sdk/pull/12439/files#diff-f4eb4ff5ec89af999cbe8fa3ffe5647d7853ffbc9c1515b337ca043c684b6bb4R679
var infoFrameworkPlistTmpl = template.Must(template.New("infoFrameworkPlist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>{{.ExecutableName}}</string>
  <key>CFBundleIdentifier</key>
  <string>{{.BundleID}}</string>
  <key>MinimumOSVersion</key>
  <string>100.0</string>
  <key>CFBundleShortVersionString</key>
  <string>{{.Version}}</string>
  <key>CFBundleVersion</key>
  <string>{{.Version}}</string>
  <key>CFBundlePackageType</key>
  <string>FMWK</string>
</dict>
</plist>
`))

func escapePlistValue(value string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(value))
	return b.String()
}

var appleModuleMapTmpl = template.Must(template.New("iosmmap").Parse(`framework module "{{.Module}}" {
	header "ref.h"
{{range .Headers}}    header "{{.}}"
{{end}}
    export *
}`))

func appleArchiveFilepath(name string, t targetInfo) string {
	return filepath.Join(tmpdir, name+"-"+t.platform+"-"+t.arch+".a")
}

func goAppleBindArchive(out string, env []string, gosrc string) error {
	return goBuildAt(gosrc, "./gobind", env, "-buildmode=c-archive", "-o", out)
}

var appleBindHeaderTmpl = template.Must(template.New("apple.h").Parse(`
// Objective-C API for talking to the following Go packages
//
{{range .pkgs}}//	{{.PkgPath}}
{{end}}//
// File is generated by gomobile bind. Do not edit.
#ifndef __{{.title}}_FRAMEWORK_H__
#define __{{.title}}_FRAMEWORK_H__

{{range .bases}}#include "{{.}}.objc.h"
{{end}}
#endif
`))
