// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/tools/go/packages"
)

type archBuildResult struct {
	libraryPath   string
	headerPath    string
	titlePath     string
	frameworkPath string

	targetInfo
}

type xcframeworkPayload struct {
	libraryPath string
	headerPath  string
}

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

	modulesUsed, err := areGoModulesUsed()
	if err != nil {
		return err
	}

	var buildResults []archBuildResult
	var platformBuildResults = make(map[string][]archBuildResult)
	targetBuildResultMutex := &sync.Mutex{}

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(targets))

	for _, target := range targets {
		go func(target targetInfo, targetBuildResultMutex *sync.Mutex) {
			defer waitGroup.Done()

			log.Println("start building for target", target)

			err, buildResult := buildTargetArch(target, gobind, pkgs, title, name, modulesUsed)
			if err != nil {
				log.Fatalf("%v", err)
				return
			}

			targetBuildResultMutex.Lock()
			buildResults = append(buildResults, *buildResult)
			if platformBuildResults[target.platform] == nil {
				platformBuildResults[target.platform] = []archBuildResult{}
			}
			platformBuildResults[target.platform] = append(platformBuildResults[target.platform], *buildResult)
			targetBuildResultMutex.Unlock()

			log.Println("finish building for target", target)

		}(target, targetBuildResultMutex)
	}

	waitGroup.Wait()

	// Finally combine all frameworks to an XCFramework
	xcframeworkArgs := []string{"-create-xcframework"}

	refinedBuildResults := []archBuildResult{}

	// Merge binary for single target
	for _, buildResults := range platformBuildResults {
		if len(buildResults) == 2 {
			mergeArchsForSinglePlatform(buildResults[0].titlePath, buildResults[1].titlePath)
			refinedBuildResults = append(refinedBuildResults, buildResults[0])
		} else if len(buildResults) == 1 {
			refinedBuildResults = append(refinedBuildResults, buildResults[0])
		} else {
			log.Fatalln("unexpected number of build results", len(buildResults))
		}
	}

	for _, result := range refinedBuildResults {
		xcframeworkArgs = append(xcframeworkArgs, "-framework", result.frameworkPath)
	}

	xcframeworkArgs = append(xcframeworkArgs, "-output", buildO)
	cmd := exec.Command("xcodebuild", xcframeworkArgs...)
	log.Println("running: xcodebuild", strings.Join(xcframeworkArgs, " "))
	err = runCmd(cmd)
	return err
}

const appleBindInfoPlist = `<?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
      <dict>
      </dict>
    </plist>
`

var appleModuleMapTmpl = template.Must(template.New("iosmmap").Parse(`framework module "{{.Module}}" {
	header "ref.h"
{{range .Headers}}    header "{{.}}"
{{end}}
    export *
}`))

func goAppleBindArchive(name string, env []string, gosrc string) (string, error) {
	archive := filepath.Join(tmpdir, name+".a")
	err := goBuildAt(gosrc, "./gobind", env, "-buildmode=c-archive", "-o", archive)
	if err != nil {
		return "", err
	}
	return archive, nil
}

func mergeArchsForSinglePlatform(from string, to string) error {
	fatCmd := exec.Command(
		"xcrun",
		"lipo", from, to, "-create", "-output", to,
	)
	if err := runCmd(fatCmd); err != nil {
		return err
	}
	return nil
}

func buildTargetArch(t targetInfo, gobindCommandPath string, pkgs []*packages.Package, title string, name string, modulesUsed bool) (err error, buildResult *archBuildResult) {
	log.Println("build for target:", t)
	// Catalyst support requires iOS 13+
	v, _ := strconv.ParseFloat(buildIOSVersion, 64)
	if t.platform == "maccatalyst" && v < 13.0 {
		return errors.New("catalyst requires -iosversion=13 or higher"), nil
	}

	outDir := filepath.Join(tmpdir, t.platform, t.arch) // adding arch
	outSrcDir := filepath.Join(outDir, "src")
	gobindDir := filepath.Join(outSrcDir, "gobind")

	// Run gobind once per platform to generate the bindings
	cmd := exec.Command(
		gobindCommandPath,
		"-lang=go,objc",
		"-outdir="+outDir,
	)
	cmd.Env = append(cmd.Env, "GOOS="+platformOS(t.platform))
	cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	tags := append(buildTags[:], platformTags(t.platform)...)
	cmd.Args = append(cmd.Args, "-tags="+strings.Join(tags, ","))
	if bindPrefix != "" {
		cmd.Args = append(cmd.Args, "-prefix="+bindPrefix)
	}
	for _, p := range pkgs {
		cmd.Args = append(cmd.Args, p.PkgPath)
	}
	if err := runCmd(cmd); err != nil {
		log.Fatalln(err)
		os.Exit(1)
		return err, nil
	}

	env := appleEnv[t.String()][:]
	sdk := getenv(env, "DARWIN_SDK")

	frameworkDir := filepath.Join(tmpdir, t.platform, sdk, title+"_"+t.arch+".framework")

	fileBases := make([]string, len(pkgs)+1)
	for i, pkg := range pkgs {
		fileBases[i] = bindPrefix + strings.Title(pkg.Name)
	}
	fileBases[len(fileBases)-1] = "Universe"

	// Add the generated packages to GOPATH for reverse bindings.
	gopath := fmt.Sprintf("GOPATH=%s%c%s", outDir, filepath.ListSeparator, goEnv("GOPATH"))
	env = append(env, gopath)

	if err := writeGoMod(outDir, t.platform, t.arch); err != nil {
		return err, nil
	}

	// Run `go mod tidy` to force to create go.sum.
	// Without go.sum, `go build` fails as of Go 1.16.
	if modulesUsed {
		if err := goModTidyAt(outSrcDir, env); err != nil {
			return err, nil
		}
	}

	// staticLibPath, err := goAppleBindArchive(name+"-"+t.platform+"-"+t.arch, env, outSrcDir)
	staticLibPath, err := goAppleBindArchive(name+"-"+t.platform+"-"+t.arch, env, outSrcDir)
	if err != nil {
		return fmt.Errorf("%s/%s: %v", t.platform, t.arch, err), nil
	}

	versionsDir := filepath.Join(frameworkDir, "Versions")
	versionsADir := filepath.Join(versionsDir, "A")
	titlePath := filepath.Join(versionsADir, title)
	versionsAHeadersDir := filepath.Join(versionsADir, "Headers")
	if err := mkdir(versionsAHeadersDir); err != nil {
		return err, nil
	}
	if err := symlink("A", filepath.Join(versionsDir, "Current")); err != nil {
		return err, nil
	}
	if err := symlink("Versions/Current/Headers", filepath.Join(frameworkDir, "Headers")); err != nil {
		return err, nil
	}
	if err := symlink(filepath.Join("Versions/Current", title), filepath.Join(frameworkDir, title)); err != nil {
		return err, nil
	}

	log.Println("staticLibPath:" + staticLibPath)
	lipoCmd := exec.Command(
		"xcrun",
		"lipo", staticLibPath, "-create", "-o", titlePath,
	)
	if err := runCmd(lipoCmd); err != nil {
		return err, nil
	}

	// Copy header file next to output archive.
	var headerFiles []string
	if len(fileBases) == 1 {
		headerFiles = append(headerFiles, title+".h")
		err := copyFile(
			filepath.Join(versionsAHeadersDir, title+".h"),
			filepath.Join(gobindDir, bindPrefix+title+".objc.h"),
		)
		if err != nil {
			return err, nil
		}
	} else {
		for _, fileBase := range fileBases {
			headerFiles = append(headerFiles, fileBase+".objc.h")
			err := copyFile(
				filepath.Join(versionsAHeadersDir, fileBase+".objc.h"),
				filepath.Join(gobindDir, fileBase+".objc.h"),
			)
			if err != nil {
				return err, nil
			}
		}
		err := copyFile(
			filepath.Join(versionsAHeadersDir, "ref.h"),
			filepath.Join(gobindDir, "ref.h"),
		)
		if err != nil {
			return err, nil
		}
		headerFiles = append(headerFiles, title+".h")
		err = writeFile(filepath.Join(versionsAHeadersDir, title+".h"), func(w io.Writer) error {
			return appleBindHeaderTmpl.Execute(w, map[string]interface{}{
				"pkgs": pkgs, "title": title, "bases": fileBases,
			})
		})
		if err != nil {
			return err, nil
		}
	}

	if err := mkdir(filepath.Join(versionsADir, "Resources")); err != nil {
		return err, nil
	}

	if err := symlink("Versions/Current/Resources", filepath.Join(frameworkDir, "Resources")); err != nil {
		return err, nil
	}

	err = writeFile(filepath.Join(frameworkDir, "Resources", "Info.plist"), func(w io.Writer) error {
		_, err := w.Write([]byte(appleBindInfoPlist))
		return err
	})
	if err != nil {
		return err, nil
	}

	var mmVals = struct {
		Module  string
		Headers []string
	}{
		Module:  title,
		Headers: headerFiles,
	}
	err = writeFile(filepath.Join(versionsADir, "Modules", "module.modulemap"), func(w io.Writer) error {
		return appleModuleMapTmpl.Execute(w, mmVals)
	})
	if err != nil {
		return err, nil
	}
	err = symlink(filepath.Join("Versions/Current/Modules"), filepath.Join(frameworkDir, "Modules"))
	if err != nil {
		return err, nil
	}
	frameworkHeaderPath := filepath.Join(frameworkDir, "Versions", "A", "Headers")
	return err, &archBuildResult{
		headerPath:    frameworkHeaderPath,
		libraryPath:   staticLibPath,
		titlePath:     titlePath,
		targetInfo:    t,
		frameworkPath: frameworkDir,
	}
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
