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
		methods []*FuncSet
	}{
		{
			ref:  importers.PkgRef{Pkg: "java/lang/Object", Name: "equals"},
			name: "java.lang.Object",
			methods: []*FuncSet{
				&FuncSet{
					Name:   "equals",
					GoName: "Equals",
					CommonSig: CommonSig{
						Params: []*Type{&Type{Kind: Object, Class: "java.lang.Object"}}, Ret: &Type{Kind: Boolean}, HasRet: true,
					},
					Funcs: []*Func{&Func{FuncSig: FuncSig{Name: "equals", Desc: "(Ljava/lang/Object;)Z"}, ArgDesc: "Ljava/lang/Object;", JNIName: "equals", Public: true, Params: []*Type{&Type{Kind: Object, Class: "java.lang.Object"}}, Ret: &Type{Kind: Boolean}}},
				},
			},
		},
		{
			ref:  importers.PkgRef{Pkg: "java/lang/Runnable", Name: "run"},
			name: "java.lang.Runnable",
			methods: []*FuncSet{
				&FuncSet{
					Name:      "run",
					GoName:    "Run",
					CommonSig: CommonSig{},
					Funcs:     []*Func{&Func{FuncSig: FuncSig{Name: "run", Desc: "()V"}, ArgDesc: "", JNIName: "run", Public: true, Abstract: true}},
				},
			},
		},
	}
	toString := &FuncSet{
		Name:   "toString",
		GoName: "ToString",
		CommonSig: CommonSig{
			Ret: &Type{Kind: String}, HasRet: true,
		},
		Funcs: []*Func{&Func{FuncSig: FuncSig{Name: "toString", Desc: "()Ljava/lang/String;"}, ArgDesc: "", JNIName: "toString", Public: true, Ret: &Type{Kind: String}}},
	}
	for _, test := range tests {
		refs := &importers.References{
			Refs:  []importers.PkgRef{test.ref},
			Names: make(map[string]struct{}),
		}
		for _, m := range test.methods {
			refs.Names[m.GoName] = struct{}{}
		}
		classes, err := (&Importer{}).Import(refs)
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
		methods := test.methods
		if !cls.Interface {
			methods = append(methods, toString)
		}
	loop:
		for _, exp := range methods {
			for _, got := range cls.AllMethods {
				if reflect.DeepEqual(exp, got) {
					continue loop
				}
			}
			t.Errorf("failed to find method: %+v", exp)
		}
	}
}

func testClsMap() map[string]*Class {
	//
	//    A--
	//   / \  \
	//  B   C  \
	//   \ / \  \
	//    D   E  F
	//
	return map[string]*Class{
		"A": &Class{},
		"B": &Class{
			Supers: []string{"A"},
		},
		"C": &Class{
			Supers: []string{"A"},
		},
		"D": &Class{
			Supers: []string{"B", "C"},
		},
		"E": &Class{
			Supers: []string{"C"},
		},
		"F": &Class{
			Supers: []string{"A"},
		},
	}
}

func TestCommonTypes(t *testing.T) {
	clsMap := testClsMap()
	tests := [][3]*Type{
		{nil, nil, nil},
		{&Type{Kind: Int}, nil, nil},
		{&Type{Kind: Int}, &Type{Kind: Float}, nil},
		{&Type{Kind: Int}, &Type{Kind: Int}, &Type{Kind: Int}},
		{&Type{Kind: Object, Class: "D"}, &Type{Kind: Object, Class: "D"}, &Type{Kind: Object, Class: "D"}},
		{&Type{Kind: Object, Class: "D"}, &Type{Kind: Object, Class: "E"}, &Type{Kind: Object, Class: "C"}},
		{&Type{Kind: Object, Class: "D"}, &Type{Kind: Object, Class: "F"}, &Type{Kind: Object, Class: "A"}},
		{&Type{Kind: Object, Class: "B"}, &Type{Kind: Object, Class: "E"}, &Type{Kind: Object, Class: "A"}},
	}
	for _, test := range tests {
		t1, t2, exp := test[0], test[1], test[2]
		got := commonType(clsMap, t1, t2)
		if !reflect.DeepEqual(got, exp) {
			t.Errorf("commonType(%+v, %+v) = %+v, expected %+v", t1, t2, got, exp)
		}
	}
}

func TestCommonSig(t *testing.T) {
	tests := []struct {
		Sigs []CommonSig
		CommonSig
	}{
		{
			Sigs: []CommonSig{
				CommonSig{}, // f()
			},
			CommonSig: CommonSig{}, // f()
		},
		{
			Sigs: []CommonSig{
				CommonSig{Throws: true, HasRet: true, Ret: &Type{Kind: Int}}, // int f() throws
			},
			// int f() throws
			CommonSig: CommonSig{Throws: true, HasRet: true, Ret: &Type{Kind: Int}},
		},
		{
			Sigs: []CommonSig{
				CommonSig{}, // f()
				CommonSig{Params: []*Type{&Type{Kind: Int}}}, // f(int)
			},
			CommonSig: CommonSig{ // f(int...)
				Variadic: true,
				Params:   []*Type{&Type{Kind: Int}},
			},
		},
		{
			Sigs: []CommonSig{
				CommonSig{Params: []*Type{&Type{Kind: Int}}},   // f(int)
				CommonSig{Params: []*Type{&Type{Kind: Float}}}, // f(float)
			},
			CommonSig: CommonSig{ // f(interface{})
				Params: []*Type{nil},
			},
		},
		{
			Sigs: []CommonSig{
				CommonSig{Params: []*Type{&Type{Kind: Int}}},                   // f(int)
				CommonSig{Params: []*Type{&Type{Kind: Int}, &Type{Kind: Int}}}, // f(int, int)
			},
			CommonSig: CommonSig{ // f(int, int...)
				Variadic: true,
				Params:   []*Type{&Type{Kind: Int}, &Type{Kind: Int}},
			},
		},
		{
			Sigs: []CommonSig{
				CommonSig{Params: []*Type{&Type{Kind: Object, Class: "A"}}}, // f(A)
				CommonSig{Params: []*Type{&Type{Kind: Object, Class: "B"}}}, // f(B)
			},
			CommonSig: CommonSig{ // f(A)
				Params: []*Type{&Type{Kind: Object, Class: "A"}},
			},
		},
	}
	clsMap := testClsMap()
	for _, test := range tests {
		got := combineSigs(clsMap, test.Sigs...)
		if !reflect.DeepEqual(got, test.CommonSig) {
			t.Errorf("commonSig(%+v) = %+v, expected %+v", test.Sigs, got, test.CommonSig)
		}
	}
}
