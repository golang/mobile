// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The importers package uses go/ast to analyze Go packages or Go files
// and collect references to types whose package has a package prefix.
// It is used by the language specific importers to determine the set of
// wrapper types to be generated.
//
// For example, in the Go file
//
// package javaprogram
//
// import "Java/java/lang"
//
// func F() {
//     o := lang.Object.New()
//     ...
// }
//
// the java importer uses this package to determine that the "java/lang"
// package and the wrapper interface, lang.Object, needs to be generated.
// After calling AnalyzeFile or AnalyzePackages, the References result
// contains the reference to lang.Object and the names set will contain
// "New".
package importers

import (
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// References is the result of analyzing a Go file or set of Go packages.
//
// For example, the Go file
//
// package pkg
//
// import "Prefix/some/Package"
//
// var A = Package.Identifier
//
// Will result in a single PkgRef with the "some/Package" package and
// the Identifier name. The Names set will contain the single name,
// "Identifier".
type References struct {
	// The list of references to identifiers in packages that are
	// identified by a package prefix.
	Refs []PkgRef
	// The list of names used in at least one selector expression.
	// Useful as a conservative upper bound on the set of identifiers
	// referenced from a set of packages.
	Names map[string]struct{}
}

// PkgRef is a reference to an identifier in a package.
type PkgRef struct {
	Pkg  string
	Name string
}

type refsSaver struct {
	pkgPrefix string
	References
	refMap map[PkgRef]struct{}
}

// AnalyzeFile scans the provided file for references to packages with the given
// package prefix. The list of unique (package, identifier) pairs is returned
func AnalyzeFile(file *ast.File, pkgPrefix string) (*References, error) {
	visitor := newRefsSaver(pkgPrefix)
	fset := token.NewFileSet()
	files := map[string]*ast.File{file.Name.Name: file}
	// Ignore errors (from unknown packages)
	pkg, _ := ast.NewPackage(fset, files, visitor.importer(), nil)
	ast.Walk(visitor, pkg)
	return &visitor.References, nil
}

// AnalyzePackages scans the provided packages for references to packages with the given
// package prefix. The list of unique (package, identifier) pairs is returned
func AnalyzePackages(pkgs []*build.Package, pkgPrefix string) (*References, error) {
	visitor := newRefsSaver(pkgPrefix)
	imp := visitor.importer()
	fset := token.NewFileSet()
	for _, pkg := range pkgs {
		fileNames := append(append([]string{}, pkg.GoFiles...), pkg.CgoFiles...)
		files := make(map[string]*ast.File)
		for _, name := range fileNames {
			f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, 0)
			if err != nil {
				return nil, err
			}
			files[name] = f
		}
		// Ignore errors (from unknown packages)
		astpkg, _ := ast.NewPackage(fset, files, imp, nil)
		ast.Walk(visitor, astpkg)
	}
	return &visitor.References, nil
}

func newRefsSaver(pkgPrefix string) *refsSaver {
	s := &refsSaver{
		pkgPrefix: pkgPrefix,
		refMap:    make(map[PkgRef]struct{}),
	}
	s.Names = make(map[string]struct{})
	return s
}

func (v *refsSaver) importer() ast.Importer {
	return func(imports map[string]*ast.Object, pkgPath string) (*ast.Object, error) {
		if pkg, exists := imports[pkgPath]; exists {
			return pkg, nil
		}
		if !strings.HasPrefix(pkgPath, v.pkgPrefix) {
			return nil, errors.New("ignored")
		}
		pkg := ast.NewObj(ast.Pkg, path.Base(pkgPath))
		imports[pkgPath] = pkg
		return pkg, nil
	}
}

func (v *refsSaver) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		v.Names[n.Sel.Name] = struct{}{}
		if x, ok := n.X.(*ast.Ident); ok && x.Obj != nil {
			if imp, ok := x.Obj.Decl.(*ast.ImportSpec); ok {
				pkgPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return nil
				}
				if strings.HasPrefix(pkgPath, v.pkgPrefix) {
					pkgPath = pkgPath[len(v.pkgPrefix):]
					ref := PkgRef{Pkg: pkgPath, Name: n.Sel.Name}
					if _, exists := v.refMap[ref]; !exists {
						v.refMap[ref] = struct{}{}
						v.Refs = append(v.Refs, ref)
					}
				}
				return nil
			}
		}
	case *ast.FuncDecl:
		if n.Recv != nil { // Methods
			v.Names[n.Name.Name] = struct{}{}
		}
	}
	return v
}
