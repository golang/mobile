// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"

	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"

	"code.google.com/p/go.mobile/bind"
	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
)

func genPkg(pkg *build.Package) {
	if len(pkg.CgoFiles) > 0 {
		errorf("gobind: cannot use cgo-dependent package as service definition: %s", pkg.CgoFiles[0])
		return
	}

	files := parseFiles(pkg.Dir, pkg.GoFiles)
	if len(files) == 0 {
		return // some error has been reported
	}

	conf := types.Config{
		Error: func(err error) {
			errorf("%v", err)
		},
	}
	p, err := conf.Check(pkg.ImportPath, fset, files, nil)
	if err != nil {
		return // printed above
	}

	switch *lang {
	case "java":
		err = bind.GenJava(os.Stdout, fset, p)
	case "go":
		err = bind.GenGo(os.Stdout, fset, p)
	default:
		errorf("unknown target language: %q", *lang)
	}

	if err != nil {
		if list, _ := err.(bind.ErrorList); len(list) > 0 {
			for _, err := range list {
				errorf("%v", err)
			}
		} else {
			errorf("%v", err)
		}
	}
}

var fset = token.NewFileSet()

func parseFiles(dir string, filenames []string) []*ast.File {
	var files []*ast.File
	hasErr := false
	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			hasErr = true
			if list, _ := err.(scanner.ErrorList); len(list) > 0 {
				for _, err := range list {
					errorf("%v", err)
				}
			} else {
				errorf("%v", err)
			}
		}
		files = append(files, file)
	}
	if hasErr {
		return nil
	}
	return files
}
