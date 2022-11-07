// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

func goJavaBind(gobind string, pkgs []*packages.Package, targets []targetInfo) error {
	// if _, err := sdkpath.AndroidHome(); err != nil {
	// 	return fmt.Errorf("this command requires the Android SDK to be installed: %w", err)
	// }

	// Run gobind to generate the bindings
	cmd := exec.Command(
		gobind,
		"-lang=go,java",
		"-outdir="+tmpdir,
	)
	// cmd.Env = append(cmd.Env, "GOOS=android")
	cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	if len(buildTags) > 0 {
		cmd.Args = append(cmd.Args, "-tags="+strings.Join(buildTags, ","))
	}
	if bindJavaPkg != "" {
		cmd.Args = append(cmd.Args, "-javapkg="+bindJavaPkg)
	}
	if bindClasspath != "" {
		cmd.Args = append(cmd.Args, "-classpath="+bindClasspath)
	}
	if bindBootClasspath != "" {
		cmd.Args = append(cmd.Args, "-bootclasspath="+bindBootClasspath)
	}
	for _, p := range pkgs {
		cmd.Args = append(cmd.Args, p.PkgPath)
	}
	if err := runCmd(cmd); err != nil {
		return err
	}

	outputDir := filepath.Join(tmpdir, "java")

	// Generate binding code and java source code only when processing the first package.
	var wg errgroup.Group
	for _, t := range targets {
		t := t
		wg.Go(func() error {
			return buildJavaSO(outputDir, t.arch)
		})
	}
	if err := wg.Wait(); err != nil {
		return err
	}

	w, err := os.Create(buildO)
	if err != nil {
		return err
	}
	jsrc := filepath.Join(tmpdir, "java")
	if err := buildJar(w, jsrc); err != nil {
		return err
	}
	return buildSrcJar(jsrc)
}

// buildJavaSO generates an Android libgojni.so file to outputDir.
// buildJavaSO is concurrent-safe.
func buildJavaSO(outputDir string, arch string) error {
	// Copy the environment variables to make this function concurrent-safe.
	env := make([]string, len(androidEnv[arch]))
	copy(env, androidEnv[arch])

	// Add the generated packages to GOPATH for reverse bindings.
	gopath := fmt.Sprintf("GOPATH=%s%c%s", tmpdir, filepath.ListSeparator, goEnv("GOPATH"))
	env = append(env, gopath)
	javac, err := exec.LookPath("javac")
	if err != nil {
		return err
	}
	javaHome := strings.TrimSuffix(javac, "/bin/javac")
	env = append(env, "CGO_CFLAGS=\"-I"+javaHome+"/include/\"")

	modulesUsed, err := areGoModulesUsed()
	if err != nil {
		return err
	}

	srcDir := filepath.Join(tmpdir, "src")

	if modulesUsed {
		// Copy the source directory for each architecture for concurrent building.
		newSrcDir := filepath.Join(tmpdir, "src-"+runtime.GOOS+"-"+arch)
		if !buildN {
			if err := doCopyAll(newSrcDir, srcDir); err != nil {
				return err
			}
		}
		srcDir = newSrcDir

		if err := writeGoMod(srcDir, "java", arch); err != nil {
			return err
		}

		// Run `go mod tidy` to force to create go.sum.
		// Without go.sum, `go build` fails as of Go 1.16.
		if err := goModTidyAt(srcDir, env); err != nil {
			return err
		}
	}

	if arch == "arm64" {
		arch = "aarch64"
	}
	if err := goBuildAt(
		srcDir,
		"./gobind",
		env,
		"-buildmode=c-shared",
		"-o="+filepath.Join(outputDir, "src", "main", "jniLibs", arch, "libgojni.so"),
	); err != nil {
		return err
	}

	return nil
}
