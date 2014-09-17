// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// The gendebug program takes gl.go and generates a version of it
// where each function includes tracing code that writes its arguments
// to the standard log.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strconv"
)

var fset = new(token.FileSet)

func typeString(t ast.Expr) string {
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, t)
	return buf.String()
}

func typePrinter(t string) string {
	switch t {
	case "[]float32", "[]byte":
		return "len(%d)"
	}
	return "%v"
}

func typePrinterArg(t, name string) string {
	switch t {
	case "[]float32", "[]byte":
		return "len(" + name + ")"
	}
	return name
}

func main() {
	f, err := parser.ParseFile(fset, "consts.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	entries := enum(f)

	f, err = parser.ParseFile(fset, "gl.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	buf := new(bytes.Buffer)

	fmt.Fprint(buf, preamble)

	fmt.Fprintf(buf, "func (v Enum) String() string {\n")
	fmt.Fprintf(buf, "\tswitch v {\n")
	for _, e := range dedup(entries) {
		fmt.Fprintf(buf, "\tcase 0x%x: return %q\n", e.value, e.name)
	}
	fmt.Fprintf(buf, "\t%s\n", `default: return fmt.Sprintf("gl.Enum(0x%x)", uint32(v))`)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")

	for _, d := range f.Decls {
		// Before:
		// func StencilMask(mask uint32) {
		//	C.glStencilMask(C.GLuint(mask))
		// }
		//
		// After:
		// func StencilMask(mask uint32) {
		// 	defer func() {
		// 		errstr := errDrain()
		// 		log.Printf("gl.StencilMask(%v) %v", mask, errstr)
		//	}()
		//	C.glStencilMask(C.GLuint(mask))
		// }
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Recv != nil {
			continue
		}

		var (
			params      []string
			paramTypes  []string
			results     []string
			resultTypes []string
		)

		// Print function signature.
		fmt.Fprintf(buf, "func %s(", fn.Name.Name)
		for i, p := range fn.Type.Params.List {
			if i > 0 {
				fmt.Fprint(buf, ", ")
			}
			ty := typeString(p.Type)
			for i, n := range p.Names {
				if i > 0 {
					fmt.Fprint(buf, ", ")
				}
				fmt.Fprintf(buf, "%s ", n.Name)
				params = append(params, n.Name)
				paramTypes = append(paramTypes, ty)
			}
			fmt.Fprint(buf, ty)
		}
		fmt.Fprintf(buf, ") (")
		if fn.Type.Results != nil {
			for i, r := range fn.Type.Results.List {
				if i > 0 {
					fmt.Fprint(buf, ", ")
				}
				ty := typeString(r.Type)
				if len(r.Names) == 0 {
					name := fmt.Sprintf("r%d", i)
					fmt.Fprintf(buf, "%s ", name)
					results = append(results, name)
					resultTypes = append(resultTypes, ty)
				}
				for i, n := range r.Names {
					if i > 0 {
						fmt.Fprint(buf, ", ")
					}
					fmt.Fprintf(buf, "%s ", n.Name)
					results = append(results, n.Name)
					resultTypes = append(resultTypes, ty)
				}
				fmt.Fprint(buf, ty)
			}
		}
		fmt.Fprintf(buf, ") {\n")

		// Insert a defer block for tracing.
		fmt.Fprintf(buf, "defer func() {\n")
		fmt.Fprintf(buf, "\terrstr := errDrain()\n")
		switch fn.Name.Name {
		case "GetUniformLocation", "GetAttribLocation":
			fmt.Fprintf(buf, "\tr0.name = name\n")
		}
		fmt.Fprintf(buf, "\tlog.Printf(\"gl.%s(", fn.Name.Name)
		for i, p := range paramTypes {
			if i > 0 {
				fmt.Fprint(buf, ", ")
			}
			fmt.Fprint(buf, typePrinter(p))
		}
		fmt.Fprintf(buf, ") ")
		if len(resultTypes) > 1 {
			fmt.Fprint(buf, "(")
		}
		for i, r := range resultTypes {
			if i > 0 {
				fmt.Fprint(buf, ", ")
			}
			fmt.Fprint(buf, typePrinter(r))
		}
		if len(resultTypes) > 1 {
			fmt.Fprint(buf, ") ")
		}
		fmt.Fprintf(buf, "%%v\"")
		for i, p := range paramTypes {
			fmt.Fprintf(buf, ", %s", typePrinterArg(p, params[i]))
		}
		for i, r := range resultTypes {
			fmt.Fprintf(buf, ", %s", typePrinterArg(r, results[i]))
		}
		fmt.Fprintf(buf, ", errstr)\n")
		fmt.Fprintf(buf, "}()\n")

		// Print original body of function.
		for _, s := range fn.Body.List {
			printer.Fprint(buf, fset, s)
			fmt.Fprintf(buf, "\n")
		}
		fmt.Fprintf(buf, "}\n\n")
	}

	b, err := format.Source(buf.Bytes())
	if err != nil {
		os.Stdout.Write(buf.Bytes())
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Stdout.Write(b)
}

const preamble = `// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Generated from gl.go. DO NOT EDIT.
// See doc.org for details.

// +build gldebug

package gl

//#cgo darwin  LDFLAGS: -framework OpenGL
//#cgo linux   LDFLAGS: -lGLESv2
//#include <stdlib.h>
//#include "gl2.h"
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

/*
Partially generated from the Khronos OpenGL API specification in XML
format, which is covered by the license:

	Copyright (c) 2013-2014 The Khronos Group Inc.

	Permission is hereby granted, free of charge, to any person obtaining a
	copy of this software and/or associated documentation files (the
	"Materials"), to deal in the Materials without restriction, including
	without limitation the rights to use, copy, modify, merge, publish,
	distribute, sublicense, and/or sell copies of the Materials, and to
	permit persons to whom the Materials are furnished to do so, subject to
	the following conditions:

	The above copyright notice and this permission notice shall be included
	in all copies or substantial portions of the Materials.

	THE MATERIALS ARE PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
	MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
	CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
	MATERIALS OR THE USE OR OTHER DEALINGS IN THE MATERIALS.

*/

func errDrain() string {
	var errs []Enum
	for {
		e := Enum(C.glGetError())
		if e == 0 {
			break
		}
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		return fmt.Sprintf(" error: %v", errs)
	}
	return ""
}

`

type entry struct {
	name  string
	value int
}

// enum builds a list of all GL constants that make up the gl.Enum type.
func enum(f *ast.File) []entry {
	var entries []entry
	for _, d := range f.Decls {
		gendecl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gendecl.Tok != token.CONST {
			continue
		}
		for _, s := range gendecl.Specs {
			v, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if len(v.Names) != 1 || len(v.Values) != 1 {
				continue
			}
			val, err := strconv.ParseInt(v.Values[0].(*ast.BasicLit).Value, 0, 32)
			if err != nil {
				log.Fatalf("enum %s: %v", v.Names[0].Name, err)
			}
			entries = append(entries, entry{v.Names[0].Name, int(val)})
		}
	}
	return entries
}

func dedup(entries []entry) []entry {
	// Find all duplicates. Use "%d" as the name of any value with duplicates.
	seen := make(map[int]int)
	for _, e := range entries {
		seen[e.value]++
	}
	var dedup []entry
	for _, e := range entries {
		switch seen[e.value] {
		case 0: // skip, already here
		case 1:
			dedup = append(dedup, e)
		default:
			// value is duplicated
			dedup = append(dedup, entry{fmt.Sprintf("%d", e.value), e.value})
			seen[e.value] = 0
		}
	}
	return dedup
}
