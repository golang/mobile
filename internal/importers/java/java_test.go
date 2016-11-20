// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java

import (
	"reflect"
	"testing"

	"golang.org/x/mobile/internal/importers"
)

func TestImport(t *testing.T) {
	if !IsAvailable() {
		t.Skipf("javap not available")
	}
	tests := []struct {
		ref     importers.PkgRef
		name    string
		methods []*Func
	}{
		{
			ref:  importers.PkgRef{"java/lang/Object", "equals"},
			name: "java.lang.Object",
			methods: []*Func{
				&Func{FuncSig: FuncSig{Name: "equals", Desc: "(Ljava/lang/Object;)Z"}, ArgDesc: "Ljava/lang/Object;", GoName: "Equals", JNIName: "equals", Public: true, Params: []*Type{&Type{Kind: Object, Class: "java.lang.Object"}}, Ret: &Type{Kind: Boolean}},
			},
		},
		{
			ref:  importers.PkgRef{"java/lang/Runnable", "run"},
			name: "java.lang.Runnable",
			methods: []*Func{
				&Func{FuncSig: FuncSig{Name: "run", Desc: "()V"}, ArgDesc: "", GoName: "Run", JNIName: "run", Public: true, Abstract: true},
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
		classes, err := Import("", "", refs)
		if err != nil {
			t.Fatal(err)
		}
		if len(classes) != 1 {
			t.Fatalf("got %d classes, expected 1", len(classes))
		}
		cls := classes[0]
		if cls.Name != test.name {
			t.Errorf("got class name %s, expected %s", cls.Name, test.name)
		}
	loop:
		for _, exp := range test.methods {
			for _, got := range cls.Methods {
				if reflect.DeepEqual(exp, got) {
					continue loop
				}
			}
			t.Errorf("failed to find method: %+v", exp)
		}
	}
}
