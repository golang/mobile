// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mobile/internal/sdkpath"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

func goAndroidBind(gobind string, pkgs []*packages.Package, targets []targetInfo) error {
	if _, err := sdkpath.AndroidHome(); err != nil {
		return fmt.Errorf("this command requires the Android SDK to be installed: %w", err)
	}

	// Run gobind to generate the bindings
	cmd := exec.Command(
		gobind,
		"-lang=go,java",
		"-outdir="+tmpdir,
	)
	cmd.Env = append(cmd.Env, "GOOS=android")
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

	androidDir := filepath.Join(tmpdir, "android")

	// Generate binding code and java source code only when processing the first package.
	var wg errgroup.Group
	for _, t := range targets {
		t := t
		wg.Go(func() error {
			return buildAndroidSO(androidDir, t.arch)
		})
	}
	if err := wg.Wait(); err != nil {
		return err
	}

	jsrc := filepath.Join(tmpdir, "java")
	if err := buildAAR(jsrc, androidDir, pkgs, targets); err != nil {
		return err
	}
	return buildSrcJar(jsrc)
}

func buildSrcJar(src string) error {
	var out io.Writer = ioutil.Discard
	if !buildN {
		ext := filepath.Ext(buildO)
		f, err := os.Create(buildO[:len(buildO)-len(ext)] + "-sources.jar")
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}()
		out = f
	}

	return writeJar(out, src)
}

// AAR is the format for the binary distribution of an Android Library Project
// and it is a ZIP archive with extension .aar.
// http://tools.android.com/tech-docs/new-build-system/aar-format
//
// These entries are directly at the root of the archive.
//
//	AndroidManifest.xml (mandatory)
//	classes.jar (mandatory)
//	assets/ (optional)
//	jni/<abi>/libgojni.so
//	R.txt (mandatory)
//	res/ (mandatory)
//	libs/*.jar (optional, not relevant)
//	proguard.txt (optional)
//	lint.jar (optional, not relevant)
//	aidl (optional, not relevant)
//
// javac and jar commands are needed to build classes.jar.
func buildAAR(srcDir, androidDir string, pkgs []*packages.Package, targets []targetInfo) (err error) {
	var out io.Writer = ioutil.Discard
	if buildO == "" {
		buildO = pkgs[0].Name + ".aar"
	}
	if !strings.HasSuffix(buildO, ".aar") {
		return fmt.Errorf("output file name %q does not end in '.aar'", buildO)
	}
	if !buildN {
		f, err := os.Create(buildO)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}()
		out = f
	}

	aarw := zip.NewWriter(out)
	aarwcreate := func(name string) (io.Writer, error) {
		if buildV {
			fmt.Fprintf(os.Stderr, "aar: %s\n", name)
		}
		return aarw.Create(name)
	}
	w, err := aarwcreate("AndroidManifest.xml")
	if err != nil {
		return err
	}
	const manifestFmt = `<manifest xmlns:android="http://schemas.android.com/apk/res/android" package=%q>
<uses-sdk android:minSdkVersion="%d"/></manifest>`
	fmt.Fprintf(w, manifestFmt, "go."+pkgs[0].Name+".gojni", buildAndroidAPI)

	w, err = aarwcreate("proguard.txt")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, `-keep class go.** { *; }`)
	if bindJavaPkg != "" {
		fmt.Fprintln(w, `-keep class `+bindJavaPkg+`.** { *; }`)
	} else {
		for _, p := range pkgs {
			fmt.Fprintln(w, `-keep class `+p.Name+`.** { *; }`)
		}
	}

	w, err = aarwcreate("classes.jar")
	if err != nil {
		return err
	}
	if err := buildJar(w, srcDir); err != nil {
		return err
	}

	files := map[string]string{}
	for _, pkg := range pkgs {
		// TODO(hajimehoshi): This works only with Go tools that assume all source files are in one directory.
		// Fix this to work with other Go tools.
		assetsDir := filepath.Join(filepath.Dir(pkg.GoFiles[0]), "assets")
		assetsDirExists := false
		if fi, err := os.Stat(assetsDir); err == nil {
			assetsDirExists = fi.IsDir()
		} else if !os.IsNotExist(err) {
			return err
		}

		if assetsDirExists {
			err := filepath.Walk(
				assetsDir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()
					name := "assets/" + path[len(assetsDir)+1:]
					if orig, exists := files[name]; exists {
						return fmt.Errorf("package %s asset name conflict: %s already added from package %s",
							pkg.PkgPath, name, orig)
					}
					files[name] = pkg.PkgPath
					w, err := aarwcreate(name)
					if err != nil {
						return nil
					}
					_, err = io.Copy(w, f)
					return err
				})
			if err != nil {
				return err
			}
		}
	}

	for _, t := range targets {
		toolchain := ndk.Toolchain(t.arch)
		lib := toolchain.abi + "/libgojni.so"
		w, err = aarwcreate("jni/" + lib)
		if err != nil {
			return err
		}
		if !buildN {
			r, err := os.Open(filepath.Join(androidDir, "src/main/jniLibs/"+lib))
			if err != nil {
				return err
			}
			defer r.Close()
			if _, err := io.Copy(w, r); err != nil {
				return err
			}
		}
	}

	// TODO(hyangah): do we need to use aapt to create R.txt?
	w, err = aarwcreate("R.txt")
	if err != nil {
		return err
	}

	w, err = aarwcreate("res/")
	if err != nil {
		return err
	}

	return aarw.Close()
}

const (
	javacTargetVer = "1.8"
	minAndroidAPI  = 16
)

func buildJar(w io.Writer, srcDir string) error {
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

	bClspath, err := bootClasspath()

	if err != nil {
		return err
	}

	args := []string{
		"-d", dst,
		"-source", javacTargetVer,
		"-target", javacTargetVer,
		"-bootclasspath", bClspath,
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

func writeJar(w io.Writer, dir string) error {
	if buildN {
		return nil
	}
	jarw := zip.NewWriter(w)
	jarwcreate := func(name string) (io.Writer, error) {
		if buildV {
			fmt.Fprintf(os.Stderr, "jar: %s\n", name)
		}
		return jarw.Create(name)
	}
	f, err := jarwcreate("META-INF/MANIFEST.MF")
	if err != nil {
		return err
	}
	fmt.Fprintf(f, manifestHeader)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out, err := jarwcreate(filepath.ToSlash(path[len(dir)+1:]))
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(out, in)
		return err
	})
	if err != nil {
		return err
	}
	return jarw.Close()
}

// buildAndroidSO generates an Android libgojni.so file to outputDir.
// buildAndroidSO is concurrent-safe.
func buildAndroidSO(outputDir string, arch string) error {
	// Copy the environment variables to make this function concurrent-safe.
	env := make([]string, len(androidEnv[arch]))
	copy(env, androidEnv[arch])

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
		newSrcDir := filepath.Join(tmpdir, "src-android-"+arch)
		if !buildN {
			if err := doCopyAll(newSrcDir, srcDir); err != nil {
				return err
			}
		}
		srcDir = newSrcDir

		if err := writeGoMod(srcDir, "android", arch); err != nil {
			return err
		}

		// Run `go mod tidy` to force to create go.sum.
		// Without go.sum, `go build` fails as of Go 1.16.
		if err := goModTidyAt(srcDir, env); err != nil {
			return err
		}
	}

	toolchain := ndk.Toolchain(arch)
	if err := goBuildAt(
		srcDir,
		"./gobind",
		env,
		"-buildmode=c-shared",
		"-o="+filepath.Join(outputDir, "src", "main", "jniLibs", toolchain.abi, "libgojni.so"),
	); err != nil {
		return err
	}

	return nil
}
