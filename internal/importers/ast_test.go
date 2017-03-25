package importers

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	file := `package ast_test

import "Prefix/some/pkg/Name"
import "Prefix/some/pkg/Name2"

const c = Name.Constant

type T struct {
	Name.Type
	unexported Name.Type2
}

func f() {
	Name2.Func().Func().Func()
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "ast_test.go", file, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := AnalyzeFile(f, "Prefix/")
	if err != nil {
		t.Fatal(err)
	}
	exps := []PkgRef{
		{Pkg: "some/pkg/Name", Name: "Constant"},
		{Pkg: "some/pkg/Name", Name: "Type"},
		{Pkg: "some/pkg/Name2", Name: "Func"},
		{Pkg: "some/pkg/Name", Name: "Type2"},
	}
	if len(refs.Refs) != len(exps) {
		t.Fatalf("expected %d references; got %d", len(exps), len(refs.Refs))
	}
	for i, exp := range exps {
		if got := refs.Refs[i]; exp != got {
			t.Errorf("expected ref %v; got %v", exp, got)
		}
	}
	if _, exists := refs.Names["Constant"]; !exists {
		t.Errorf("expected \"Constant\" in the names set")
	}
	if len(refs.Embedders) != 1 {
		t.Fatalf("expected 1 struct; got %d", len(refs.Embedders))
	}
	s := refs.Embedders[0]
	exp := Struct{
		Name: "T",
		Pkg:  "ast_test",
		Refs: []PkgRef{{Pkg: "some/pkg/Name", Name: "Type"}},
	}
	if !reflect.DeepEqual(exp, s) {
		t.Errorf("expected struct %v; got %v", exp, s)
	}
}
