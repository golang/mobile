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

const c = Name.Constant

type T struct {
	Name.Type
	hidden Name.Type2
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
	if len(refs.Refs) != 2 {
		t.Fatalf("expected 2 references; got %d", len(refs.Refs))
	}
	got := refs.Refs[0]
	if exp := (PkgRef{"some/pkg/Name", "Constant"}); exp != got {
		t.Errorf("expected ref %v; got %v", exp, got)
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
		Refs: []PkgRef{{"some/pkg/Name", "Type"}},
	}
	if !reflect.DeepEqual(exp, s) {
		t.Errorf("expected struct %v; got %v", exp, got)
	}
}
