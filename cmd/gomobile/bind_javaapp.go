// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

var osname = runtime.GOOS

func goJavaBind(gobind string, pkgs []*packages.Package, targets []targetInfo) error {
	// Run gobind to generate the bindings
	cmd := exec.Command(
		gobind,
		"-lang=go,java",
		"-outdir="+tmpdir,
	)
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
	if err := buildJavaJar(w, jsrc); err != nil {
		return err
	}
	_ = os.RemoveAll(filepath.Join(outputDir, "src", "main", "jniLibs"))
	return buildSrcJar(jsrc)
}

func buildJavaJar(w io.Writer, srcDir string) error {
	var srcFiles []string
	if buildN {
		srcFiles = []string{"*.java"}
	} else {
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".java" {
				srcFiles = append(srcFiles, filepath.Join(".", path[len(srcDir):]))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	dst := filepath.Join(tmpdir, "javac-output")
	if !buildN {
		if err := os.MkdirAll(dst, 0700); err != nil {
			return err
		}
	}

	cp := exec.Command("cp", "-r", filepath.Join(tmpdir, "java", "src", "main", "jniLibs"), dst)
	if err := runCmd(cp); err != nil {
		return err
	}

	args := []string{
		"-d", dst,
		"-source", javacTargetVer,
		"-target", javacTargetVer,
	}
	if bindClasspath != "" {
		args = append(args, "-classpath", bindClasspath)
	}

	args = append(args, srcFiles...)

	javac := exec.Command("javac", args...)
	javac.Dir = srcDir
	if err := runCmd(javac); err != nil {
		return err
	}

	if buildX {
		printcmd("jar c -C %s .", dst)
	}
	return writeJar(w, dst)
}

// buildJavaSO generates a libgojni.so file to outputDir (regardless of OS library file extension).
// buildJavaSO is concurrent-safe.
func buildJavaSO(outputDir string, arch string) error {
	// Copy the environment variables to make this function concurrent-safe.
	env := make([]string, len(javaEnv[arch]))
	copy(env, javaEnv[arch])

	// Add the generated packages to GOPATH for reverse bindings.
	gopath := fmt.Sprintf("GOPATH=%s%c%s", tmpdir, filepath.ListSeparator, goEnv("GOPATH"))
	env = append(env, gopath)

	modulesUsed, err := areGoModulesUsed()
	if err != nil {
		return err
	}

	srcDir := filepath.Join(tmpdir, "src")

	if modulesUsed {
		// Copy the source directory for each architecture for concurrent building.
		newSrcDir := filepath.Join(tmpdir, "src-"+osname+"-"+arch)
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

	// Javaify the arch descriptors
	if arch == "arm64" {
		if osname == "linux" {
			env = append(env, "CC=aarch64-linux-gnu-gcc")
			// env = append(env, "CC=clang")
		}
	}
	if arch == "amd64" {
		if osname == "linux" {
			env = append(env, "CC=x86_64-linux-gnu-gcc")
			// env = append(env, "CC=clang")
		}
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
