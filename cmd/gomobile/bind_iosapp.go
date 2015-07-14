// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func goIOSBind(pkg *build.Package) error {
	binder, err := newBinder(pkg)
	if err != nil {
		return err
	}
	name := binder.pkg.Name()

	if buildO != "" && !strings.HasSuffix(buildO, ".a") {
		return fmt.Errorf("archive name %q missing .a suffix", buildO)
	}
	if buildO == "" {
		buildO = name + ".a"
	}

	if err := binder.GenGo(filepath.Join(tmpdir, "src")); err != nil {
		return err
	}
	mainFile := filepath.Join(tmpdir, "src/iosbin/main.go")
	err = writeFile(mainFile, func(w io.Writer) error {
		return iosBindTmpl.Execute(w, "../go_"+name)
	})
	if err != nil {
		return fmt.Errorf("failed to create the binding package for iOS: %v", err)
	}
	if err := binder.GenObjc(filepath.Join(tmpdir, "objc")); err != nil {
		return err
	}

	armPath, err := goIOSBindArchive(name, mainFile, darwinArmEnv)
	if err != nil {
		return err
	}
	arm64Path, err := goIOSBindArchive(name, mainFile, darwinArm64Env)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"xcrun", "lipo",
		"-create",
		"-arch", "arm",
		armPath,
		"-arch", "arm64",
		arm64Path,
		"-o", buildO,
	)
	// TODO(crawshaw): arch i386/x86_64 for iOS simulator
	if buildX {
		printcmd(strings.Join(cmd.Args, " "))
	}
	if !buildN {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// TODO(crawshaw): seq.h

	// Copy header file next to output archive.
	return copyFile(
		filepath.Join(buildO[:len(buildO)-2]+".h"),
		filepath.Join(tmpdir, "objc/Go"+strings.Title(name)+".h"),
	)
}

func goIOSBindArchive(name, path string, env []string) (string, error) {
	arch := getenv(env, "GOARCH")
	archive := filepath.Join(tmpdir, name+"-"+arch+".a")
	err := goBuild(path, env, "-buildmode=c-archive", "-o", archive)
	if err != nil {
		return "", err
	}

	// Build env suitable for invoking $CC.
	cmd := exec.Command("go", "env", "GOGCCFLAGS")
	cmd.Env = environ(env)
	ccflags, err := cmd.Output()
	if err != nil {
		panic(err) // the Go tool must work by now
	}
	env = append([]string{fmt.Sprintf("CCFLAGS=%q", string(ccflags))}, env...)

	obj := "gobind-" + name + "-" + arch + ".o"
	cmd = exec.Command(
		getenv(env, "CC"),
		"-I", ".",
		"-g", "-O2",
		"-o", obj,
		"-c", "Go"+strings.Title(name)+".m",
	)
	cmd.Dir = filepath.Join(tmpdir, "objc")
	cmd.Env = env
	if buildX {
		printcmd("PWD=" + cmd.Dir + " " + strings.Join(cmd.Env, " ") + strings.Join(cmd.Args, " "))
	}
	if !buildN {
		cmd.Stderr = os.Stderr // nominally silent
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	cmd = exec.Command("ar", "-q", "-s", archive, obj)
	cmd.Dir = filepath.Join(tmpdir, "objc")
	if buildX {
		printcmd("PWD=" + cmd.Dir + " " + strings.Join(cmd.Args, " "))
	}
	if !buildN {
		out, err := cmd.CombinedOutput()
		if buildV {
			os.Stderr.Write(out)
		}
		if err != nil {
			return "", fmt.Errorf("ar: %v\n%s", err, out)
		}
	}

	return archive, nil
}

func getenv(env []string, key string) string {
	prefix := key + "="
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			return kv[len(prefix):]
		}
	}
	return ""
}

var iosBindTmpl = template.Must(template.New("ios.go").Parse(`
package main

import (
	_ "golang.org/x/mobile/bind/objc"
	_ "{{.}}"
)

import "C"

func main() {}
`))
