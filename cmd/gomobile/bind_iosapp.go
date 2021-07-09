// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func goIOSBind(gobind string, pkgs []*packages.Package, archs []string) error {
	// Run gobind to generate the bindings
	cmd := exec.Command(
		gobind,
		"-lang=go,objc",
		"-outdir="+tmpdir,
	)
	cmd.Env = append(cmd.Env, "GOOS=darwin")
	cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	tags := append(buildTags, "ios")
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

	srcDir := filepath.Join(tmpdir, "src", "gobind")

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

	modulesUsed, err := areGoModulesUsed()
	if err != nil {
		return err
	}

	// create separate framework for ios,simulator and catalyst
	// every target has at least one arch (arm64 and x86_64)
	var frameworkDirs []string
	for _, target := range iOSTargets {
		frameworkDir := filepath.Join(tmpdir, target, title+".framework")
		frameworkDirs = append(frameworkDirs, frameworkDir)

		for index, arch := range iOSTargetArchs(target) {
			fileBases := make([]string, len(pkgs)+1)
			for i, pkg := range pkgs {
				fileBases[i] = bindPrefix + strings.Title(pkg.Name)
			}
			fileBases[len(fileBases)-1] = "Universe"

			env := darwinEnv[target+"_"+arch]

			if err := writeGoMod("darwin", getenv(env, "GOARCH")); err != nil {
				return err
			}

			// Add the generated packages to GOPATH for reverse bindings.
			gopath := fmt.Sprintf("GOPATH=%s%c%s", tmpdir, filepath.ListSeparator, goEnv("GOPATH"))
			env = append(env, gopath)

			// Run `go mod tidy` to force to create go.sum.
			// Without go.sum, `go build` fails as of Go 1.16.
			if modulesUsed {
				if err := goModTidyAt(filepath.Join(tmpdir, "src"), env); err != nil {
					return err
				}
			}

			path, err := goIOSBindArchive(name, env, filepath.Join(tmpdir, "src"))
			if err != nil {
				return fmt.Errorf("darwin-%s: %v", arch, err)
			}

			versionsDir := filepath.Join(frameworkDir, "Versions")
			versionsADir := filepath.Join(versionsDir, "A")
			titlePath := filepath.Join(versionsADir, title)
			if index > 0 {
				// not the first static lib, attach to a fat library and skip create headers
				fatCmd := exec.Command(
					"xcrun",
					"lipo", "-create", "-output", titlePath, titlePath, path,
				)
				if err := runCmd(fatCmd); err != nil {
					return err
				}
				continue
			}

			versionsAHeadersDir := filepath.Join(versionsADir, "Headers")
			if err := mkdir(versionsAHeadersDir); err != nil {
				return err
			}
			if err := symlink("A", filepath.Join(versionsDir, "Current")); err != nil {
				return err
			}
			if err := symlink("Versions/Current/Headers", filepath.Join(frameworkDir, "Headers")); err != nil {
				return err
			}
			if err := symlink(filepath.Join("Versions/Current", title), filepath.Join(frameworkDir, title)); err != nil {
				return err
			}

			lipoCmd := exec.Command(
				"xcrun",
				"lipo", "-create", "-arch", archClang(arch), path, "-o", titlePath,
			)
			if err := runCmd(lipoCmd); err != nil {
				return err
			}

			// Copy header file next to output archive.
			var headerFiles []string
			if len(fileBases) == 1 {
				headerFiles = append(headerFiles, title+".h")
				err := copyFile(
					filepath.Join(versionsAHeadersDir, title+".h"),
					filepath.Join(srcDir, bindPrefix+title+".objc.h"),
				)
				if err != nil {
					return err
				}
			} else {
				for _, fileBase := range fileBases {
					headerFiles = append(headerFiles, fileBase+".objc.h")
					err := copyFile(
						filepath.Join(versionsAHeadersDir, fileBase+".objc.h"),
						filepath.Join(srcDir, fileBase+".objc.h"),
					)
					if err != nil {
						return err
					}
				}
				err := copyFile(
					filepath.Join(versionsAHeadersDir, "ref.h"),
					filepath.Join(srcDir, "ref.h"),
				)
				if err != nil {
					return err
				}
				headerFiles = append(headerFiles, title+".h")
				err = writeFile(filepath.Join(versionsAHeadersDir, title+".h"), func(w io.Writer) error {
					return iosBindHeaderTmpl.Execute(w, map[string]interface{}{
						"pkgs": pkgs, "title": title, "bases": fileBases,
					})
				})
				if err != nil {
					return err
				}
			}

			if err := mkdir(filepath.Join(versionsADir, "Resources")); err != nil {
				return err
			}
			if err := symlink("Versions/Current/Resources", filepath.Join(frameworkDir, "Resources")); err != nil {
				return err
			}
			err = writeFile(filepath.Join(frameworkDir, "Resources", "Info.plist"), func(w io.Writer) error {
				_, err := w.Write([]byte(iosBindInfoPlist))
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
			err = writeFile(filepath.Join(versionsADir, "Modules", "module.modulemap"), func(w io.Writer) error {
				return iosModuleMapTmpl.Execute(w, mmVals)
			})
			if err != nil {
				return err
			}
			err = symlink(filepath.Join("Versions/Current/Modules"), filepath.Join(frameworkDir, "Modules"))
			if err != nil {
				return err
			}
		}
	}

	// Finally combine all frameworks to an XCFramework
	xcframeworkArgs := []string{"-create-xcframework"}

	for _, dir := range frameworkDirs {
		xcframeworkArgs = append(xcframeworkArgs, "-framework", dir)
	}

	xcframeworkArgs = append(xcframeworkArgs, "-output", buildO)
	cmd = exec.Command("xcodebuild", xcframeworkArgs...)
	err = runCmd(cmd)
	return err
}

const iosBindInfoPlist = `<?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
      <dict>
      </dict>
    </plist>
`

var iosModuleMapTmpl = template.Must(template.New("iosmmap").Parse(`framework module "{{.Module}}" {
	header "ref.h"
{{range .Headers}}    header "{{.}}"
{{end}}
    export *
}`))

func goIOSBindArchive(name string, env []string, gosrc string) (string, error) {
	arch := getenv(env, "GOARCH")
	archive := filepath.Join(tmpdir, name+"-"+arch+".a")
	err := goBuildAt(gosrc, "./gobind", env, "-buildmode=c-archive", "-o", archive)
	if err != nil {
		return "", err
	}
	return archive, nil
}

var iosBindHeaderTmpl = template.Must(template.New("ios.h").Parse(`
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
