package importers

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	file := `package ast_test

import "Prefix/some/pkg/Name"

const c = Name.Constant
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
	if len(refs.Refs) != 1 {
		t.Fatalf("expected 1 reference; got %d", len(refs.Refs))
	}
	got := refs.Refs[0]
	if exp := (PkgRef{"some/pkg/Name", "Constant"}); exp != got {
		t.Errorf("expected ref %v; got %v", exp, got)
	}
	if _, exists := refs.Names["Constant"]; !exists {
		t.Errorf("expected \"Constant\" in the names set")
	}
}
