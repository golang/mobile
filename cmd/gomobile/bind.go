// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/mobile/bind"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

// ctx, pkg, ndkccpath, tmpdir in build.go

var cmdBind = &command{
	run:   runBind,
	Name:  "bind",
	Usage: "[package]",
	Short: "build a shared library for android APK and/or iOS app",
	Long: `
Bind generates language bindings like gobind (golang.org/x/mobile/cmd/gobind)
for a package and builds a shared library for each platform from the go binding
code.

The -outdir flag specifies the output directory and is required.

For Android, the bind command will place the generated Java API stubs and the
compiled shared libraries in the android subdirectory of the following layout.

<outdir>/android
  libs/
     armeabi-v7a/libgojni.so
     ...
  src/main/java/go/
	Seq.java
	Go.java
        mypackage/Mypackage.java

The -v flag provides verbose output, including the list of packages built.

These build flags are shared by the build command.
For documentation, see 'go help build':
	-a
	-i
	-n
	-x
	-tags 'tag list'
`,
}

// TODO: -mobile

var bindOutdir *string // -outdir
func init() {
	bindOutdir = cmdBind.flag.String("outdir", "", "output directory. Default is the current directory.")
}

func runBind(cmd *command) error {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	args := cmd.flag.Args()

	var bindPkg *build.Package
	switch len(args) {
	case 0:
		bindPkg, err = ctx.ImportDir(cwd, build.ImportComment)
	case 1:
		bindPkg, err = ctx.Import(args[0], cwd, build.ImportComment)
	default:
		cmd.usage()
		os.Exit(1)
	}
	if err != nil {
		return err
	}

	if *bindOutdir == "" {
		return fmt.Errorf("-outdir is required")
	}

	if !buildN {
		sentinel := filepath.Join(*bindOutdir, "gomobile-bind-sentinel")
		if err := ioutil.WriteFile(sentinel, []byte("write test"), 0644); err != nil {
			return fmt.Errorf("output directory %q not writable", *bindOutdir)
		}
		os.Remove(sentinel)
	}

	if buildN {
		tmpdir = "$WORK"
	} else {
		tmpdir, err = ioutil.TempDir("", "gomobile-bind-work-")
		if err != nil {
			return err
		}
	}
	defer removeAll(tmpdir)
	if buildX {
		fmt.Fprintln(os.Stderr, "WORK="+tmpdir)
	}

	binder, err := newBinder(bindPkg)
	if err != nil {
		return err
	}

	if err := binder.GenGo(tmpdir); err != nil {
		return err
	}

	pathJoin := func(a, b string) string {
		return filepath.Join(a, filepath.FromSlash(b))
	}

	mainFile := pathJoin(tmpdir, "androidlib/main.go")
	err = writeFile(mainFile, func(w io.Writer) error {
		return androidMainTmpl.Execute(w, "../go_"+binder.pkg.Name())
	})
	if err != nil {
		return fmt.Errorf("failed to create the main package for android: %v", err)
	}

	androidOutdir := pathJoin(*bindOutdir, "android")

	err = gobuild(mainFile, pathJoin(androidOutdir, "libs/armeabi-v7a/libgojni.so"))
	if err != nil {
		return err
	}
	p, err := ctx.Import("golang.org/x/mobile/app", cwd, build.ImportComment)
	if err != nil {
		return fmt.Errorf(`"golang.org/x/mobile/app" is not found; run go get golang.org/x/mobile/app`)
	}
	repo := filepath.Clean(filepath.Join(p.Dir, "..")) // golang.org/x/mobile directory.

	// TODO(crawshaw): use a better package path derived from the go package.
	if err := binder.GenJava(pathJoin(androidOutdir, "src/main/java/go/"+binder.pkg.Name())); err != nil {
		return err
	}

	src := pathJoin(repo, "app/Go.java")
	dst := pathJoin(androidOutdir, "src/main/java/go/Go.java")
	rm(dst)
	if err := symlink(src, dst); err != nil {
		return err
	}

	src = pathJoin(repo, "bind/java/Seq.java")
	dst = pathJoin(androidOutdir, "src/main/java/go/Seq.java")
	rm(dst)
	if err := symlink(src, dst); err != nil {
		return err
	}

	return nil
}

type binder struct {
	files []*ast.File
	fset  *token.FileSet
	pkg   *types.Package
}

func (b *binder) GenJava(outdir string) error {
	firstRune, size := utf8.DecodeRuneInString(b.pkg.Name())
	className := string(unicode.ToUpper(firstRune)) + b.pkg.Name()[size:]
	javaFile := filepath.Join(outdir, className+".java")

	if buildX {
		printcmd("gobind -lang=java %s > %s", b.pkg.Path(), javaFile)
	}

	generate := func(w io.Writer) error {
		return bind.GenJava(w, b.fset, b.pkg)
	}
	if err := writeFile(javaFile, generate); err != nil {
		return err
	}
	return nil
}

func (b *binder) GenGo(outdir string) error {
	pkgName := "go_" + b.pkg.Name()
	goFile := filepath.Join(outdir, pkgName, pkgName+".go")

	if buildX {
		printcmd("gobind -lang=go %s > %s", b.pkg.Path(), goFile)
	}

	generate := func(w io.Writer) error {
		return bind.GenGo(w, b.fset, b.pkg)
	}
	if err := writeFile(goFile, generate); err != nil {
		return err
	}
	return nil
}

func writeFile(filename string, generate func(io.Writer) error) error {
	if buildV {
		fmt.Fprintf(os.Stderr, "write %s\n", filename)
	}

	err := mkdir(filepath.Dir(filename))
	if err != nil {
		return err
	}

	if buildN {
		return generate(ioutil.Discard)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	return generate(f)
}

func newBinder(bindPkg *build.Package) (*binder, error) {
	if bindPkg.Name == "main" {
		return nil, fmt.Errorf("package %q: can only bind a library package", bindPkg.Name)
	}

	if len(bindPkg.CgoFiles) > 0 {
		return nil, fmt.Errorf("cannot use cgo-dependent package as service definition: %s", bindPkg.CgoFiles[0])
	}

	fset := token.NewFileSet()

	hasErr := false
	var files []*ast.File
	for _, filename := range bindPkg.GoFiles {
		p := filepath.Join(bindPkg.Dir, filename)
		file, err := parser.ParseFile(fset, p, nil, parser.AllErrors)
		if err != nil {
			hasErr = true
			if list, _ := err.(scanner.ErrorList); len(list) > 0 {
				for _, err := range list {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		files = append(files, file)
	}

	if hasErr {
		return nil, errors.New("package parsing failed.")
	}

	conf := loader.Config{
		Fset: fset,
	}
	conf.TypeChecker.Error = func(err error) {
		fmt.Fprintln(os.Stderr, err)
	}

	conf.CreateFromFiles(bindPkg.ImportPath, files...)
	program, err := conf.Load()
	if err != nil {
		return nil, err
	}
	b := &binder{
		files: files,
		fset:  fset,
		pkg:   program.Created[0].Pkg,
	}
	return b, nil
}

var androidMainTmpl = template.Must(template.New("android.go").Parse(`
package main

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/bind/java"

	_ "{{.}}"
)

func main() {
	app.Run(app.Callbacks{Start: java.Init})
}
`))
