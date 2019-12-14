// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package doc tests that Go documentation is transferred
// to the generated code.
package doc

// F is a function.
func F() {}

// C is a constant.
const C = true

// V is a var.
var V string

// A group of vars.
var (
	// A specific var.
	Specific string
	NoDocVar float64
)

// Before is a method.
func (_ *S) Before() {}

// S is a struct.
type S struct {
	// SF is a field.
	SF string
	// blank (unexported) field.
	_ string
	// Anonymous field.
	*S2
	// Multiple fields.
	F1, F2 string
}

// After is another method.
func (_ *S) After() {}

// A generic comment with <HTML>.
type (
	// S2 is a struct.
	S2    struct{}
	NoDoc struct{}
)

// NewS is a constructor.
func NewS() *S {
	return nil
}

// I is an interface.
type I interface {
	// IM is a method.
	IM()
}
