// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bind implements a code generator for gobind.
//
// See the documentation on the gobind command for usage details
// and the list of currently supported types.
// (http://godoc.org/golang.org/x/mobile/cmd/gobind)
package bind // import "golang.org/x/mobile/bind"

// TODO(crawshaw): slice support
// TODO(crawshaw): channel support

import (
	"bytes"
	"go/format"
	"go/token"
	"go/types"
	"io"
)

type fileType int

const (
	Java fileType = iota
	JavaC
	JavaH

	ObjcM
	ObjcH
	ObjcGoH
)

// GenJava generates a Java API from a Go package.
func GenJava(w io.Writer, fset *token.FileSet, pkg *types.Package, javaPkg string, ft fileType) error {
	if javaPkg == "" {
		javaPkg = javaPkgName(pkg.Name())
	}
	buf := new(bytes.Buffer)
	g := &javaGen{
		javaPkg: javaPkg,
		generator: &generator{
			printer: &printer{buf: buf, indentEach: []byte("    ")},
			fset:    fset,
			pkg:     pkg,
		},
	}
	g.init()
	var err error
	switch ft {
	case Java:
		err = g.genJava()
	case JavaC:
		err = g.genC()
	case JavaH:
		err = g.genH()
	default:
		panic("invalid fileType")
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(w, buf)
	return err
}

// GenGo generates a Go stub to support foreign language APIs.
func GenGo(w io.Writer, fset *token.FileSet, pkg *types.Package) error {
	buf := new(bytes.Buffer)
	g := &goGen{
		&generator{
			printer: &printer{buf: buf, indentEach: []byte("\t")},
			fset:    fset,
			pkg:     pkg,
		},
	}
	g.init()
	if err := g.gen(); err != nil {
		return err
	}
	src := buf.Bytes()
	srcf, err := format.Source(src)
	if err != nil {
		w.Write(src) // for debugging
		return err
	}
	_, err = w.Write(srcf)
	return err
}

// GenObjc generates the Objective-C API from a Go package.
func GenObjc(w io.Writer, fset *token.FileSet, pkg *types.Package, prefix string, ft fileType) error {
	if prefix == "" {
		prefix = "Go"
	}

	buf := new(bytes.Buffer)
	g := &objcGen{
		generator: &generator{
			printer: &printer{buf: buf, indentEach: []byte("\t")},
			fset:    fset,
			pkg:     pkg,
		},
		prefix: prefix,
	}
	g.init()
	var err error
	switch ft {
	case ObjcH:
		err = g.genH()
	case ObjcM:
		err = g.genM()
	case ObjcGoH:
		err = g.genGoH()
	default:
		panic("invalid fileType")
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(w, buf)
	return err
}
