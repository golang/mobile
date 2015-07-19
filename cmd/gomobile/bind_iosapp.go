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

	cmd := exec.Command("xcrun", "lipo", "-create")

	// TODO(crawshaw): Build in parallel.
	for _, env := range [][]string{darwinArmEnv, darwinArm64Env, darwinAmd64Env} {
		arch := archClang(getenv(env, "GOARCH"))
		path, err := goIOSBindArchive(name, mainFile, env)
		if err != nil {
			return fmt.Errorf("darwin-%s: %v", arch, err)
		}
		cmd.Args = append(cmd.Args, "-arch", arch, path)
	}

	cmd.Args = append(cmd.Args, "-o", buildO)
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

	// Copy header file next to output archive.
	return copyFile(
		filepath.Join(buildO[:len(buildO)-2]+".h"),
		filepath.Join(tmpdir, "objc/Go"+strings.Title(name)+".h"),
	)
}

func goIOSBindArchive(name, path string, env []string) (string, error) {
	arch := getenv(env, "GOARCH")
	archive := filepath.Join(tmpdir, name+"-"+arch+".a")
	err := goBuild(path, env, "-buildmode=c-archive", "-tags=ios", "-o", archive)
	if err != nil {
		return "", err
	}

	obj := "gobind-" + name + "-" + arch + ".o"
	cmd := exec.Command(
		getenv(env, "CC"),
		"-I", ".",
		"-g", "-O2",
		"-o", obj,
		"-c", "Go"+strings.Title(name)+".m",
	)
	cmd.Args = append(cmd.Args, strings.Split(getenv(env, "CGO_CFLAGS"), " ")...)
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

var iosBindTmpl = template.Must(template.New("ios.go").Parse(`
package main

import (
	_ "golang.org/x/mobile/bind/objc"
	_ "{{.}}"
)

import "C"

func main() {}
`))
