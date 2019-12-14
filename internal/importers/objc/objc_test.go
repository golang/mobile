// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objc

import (
	"reflect"
	"runtime"
	"testing"

	"golang.org/x/mobile/internal/importers"
)

func TestImport(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skipf("can only parse objc types on darwin")
	}
	tests := []struct {
		ref     importers.PkgRef
		name    string
		methods []*Func
	}{
		{
			ref:  importers.PkgRef{Pkg: "Foundation/NSObjectP", Name: "Hash"},
			name: "NSObject",
			methods: []*Func{
				&Func{Sig: "hash", GoName: "Hash", Ret: &Type{Kind: Uint, Decl: "NSUInteger"}},
			},
		},
		{
			ref:  importers.PkgRef{Pkg: "Foundation/NSString", Name: "StringWithContentsOfFileEncodingError"},
			name: "NSString",
			methods: []*Func{
				&Func{
					Sig:    "stringWithContentsOfFile:encoding:error:",
					GoName: "StringWithContentsOfFileEncodingError",
					Params: []*Param{
						&Param{Name: "path", Type: &Type{Kind: String, Decl: "NSString *"}},
						&Param{Name: "enc", Type: &Type{Kind: Uint, Decl: "NSStringEncoding"}},
						&Param{Name: "error", Type: &Type{Kind: Class, Name: "NSError", Decl: "NSError *   * _Nullable", Indirect: true}},
					},
					Ret:    &Type{Kind: 3, Decl: "NSString *"},
					Static: true,
				},
			},
		},
	}
	for _, test := range tests {
		refs := &importers.References{
			Refs:  []importers.PkgRef{test.ref},
			Names: make(map[string]struct{}),
		}
		for _, m := range test.methods {
			refs.Names[m.GoName] = struct{}{}
		}
		types, err := Import(refs)
		if err != nil {
			t.Fatal(err)
		}
		if len(types) == 0 {
			t.Fatalf("got no types, expected at least 1")
		}
		n := types[0]
		if n.Name != test.name {
			t.Errorf("got class name %s, expected %s", n.Name, test.name)
		}
	loop:
		for _, exp := range test.methods {
			for _, got := range n.AllMethods {
				if reflect.DeepEqual(exp, got) {
					continue loop
				}
			}
			for _, got := range n.Funcs {
				if reflect.DeepEqual(exp, got) {
					continue loop
				}
			}
			t.Errorf("failed to find method: %+v", exp)
		}
	}
}
