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
	var buildTemp string

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
	// Build static xcframework output directory.
	if err := removeAll(buildO); err != nil {
		return err
	}

	targets := allTargets("ios")

	// create separate framework for ios,simulator and catalyst
	// every target has at least one arch (arm64 and x86_64)
	for _, target := range targets {
		archs := allTargetArchs("ios", target)

		for index, arch := range archs {
			buildTemp = tmpdir + "/" + target + "/" + title + ".framework"

			fileBases := make([]string, len(pkgs)+1)
			for i, pkg := range pkgs {
				fileBases[i] = bindPrefix + strings.Title(pkg.Name)
			}
			fileBases[len(fileBases)-1] = "Universe"

			cmd = exec.Command("xcrun", "lipo", "-create")

			env := darwinEnv[target + "_" + arch]

			if err := writeGoMod("darwin", getenv(env, "GOARCH")); err != nil {
				return err
			}

			// Add the generated packages to GOPATH for reverse bindings.
			gopath := fmt.Sprintf("GOPATH=%s%c%s", tmpdir, filepath.ListSeparator, goEnv("GOPATH"))
			env = append(env, gopath)
			fmt.Printf("[debug] goenv:\n%s\n", env)

			path, err := goIOSBindArchive(name, env, filepath.Join(tmpdir, "src"))
			if err != nil {
				return fmt.Errorf("darwin-%s: %v", arch, err)
			}

			if index > 0 {
				// not the first static lib, attach to a fat library and skip create headers
				fatCmd := exec.Command("xcrun", "lipo", "-create",
					"-output", buildTemp+"/Versions/A/"+title,
					buildTemp+"/Versions/A/"+title, path)

				if err := runCmd(fatCmd); err != nil {
					return err
				}

				continue
			}

			cmd.Args = append(cmd.Args, "-arch", archClang(arch), path)

			headers := buildTemp + "/Versions/A/Headers"
			if err := mkdir(headers); err != nil {
				return err
			}
			if err := symlink("A", buildTemp+"/Versions/Current"); err != nil {
				return err
			}
			if err := symlink("Versions/Current/Headers", buildTemp+"/Headers"); err != nil {
				return err
			}
			if err := symlink("Versions/Current/"+title, buildTemp+"/"+title); err != nil {
				return err
			}

			cmd.Args = append(cmd.Args, "-o", buildTemp+"/Versions/A/"+title)
			if err := runCmd(cmd); err != nil {
				return err
			}

			//Copy header file next to output archive.
			headerFiles := make([]string, len(fileBases))
			if len(fileBases) == 1 {
				headerFiles[0] = title + ".h"
				err := copyFile(
					headers+"/"+title+".h",
					srcDir+"/"+bindPrefix+title+".objc.h",
				)
				if err != nil {
					return err
				}
			} else {
				for i, fileBase := range fileBases {
					headerFiles[i] = fileBase + ".objc.h"
					err := copyFile(
						headers+"/"+fileBase+".objc.h",
						srcDir+"/"+fileBase+".objc.h")
					if err != nil {
						return err
					}
				}
				err := copyFile(
					headers+"/ref.h",
					srcDir+"/ref.h")
				if err != nil {
					return err
				}
				headerFiles = append(headerFiles, title+".h")
				err = writeFile(headers+"/"+title+".h", func(w io.Writer) error {
					return iosBindHeaderTmpl.Execute(w, map[string]interface{}{
						"pkgs": pkgs, "title": title, "bases": fileBases,
					})
				})
				if err != nil {
					return err
				}
			}

			resources := buildTemp + "/Versions/A/Resources"
			if err := mkdir(resources); err != nil {
				return err
			}
			if err := symlink("Versions/Current/Resources", buildTemp+"/Resources"); err != nil {
				return err
			}
			err = writeFile(buildTemp+"/Resources/Info.plist", func(w io.Writer) error {
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
			err = writeFile(buildTemp+"/Versions/A/Modules/module.modulemap", func(w io.Writer) error {
				return iosModuleMapTmpl.Execute(w, mmVals)
			})
			if err != nil {
				return err
			}
			err = symlink("Versions/Current/Modules", buildTemp+"/Modules")
			if err != nil {
				return err
			}
		}
	}

	// Finally combine ios/simulator/catalyst framework to xcframework
	xcframeworkArgs := []string{"-create-xcframework"}

	for _, target := range allTargets("ios") {
		xcframeworkArgs = append(xcframeworkArgs, "-framework", tmpdir+"/"+target+"/"+title+".framework")
	}

	xcframeworkArgs = append(xcframeworkArgs, "-output", buildO)
	cmd = exec.Command("xcodebuild", xcframeworkArgs...)
	err := runCmd(cmd)
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
