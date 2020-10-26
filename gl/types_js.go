// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package gl

import "syscall/js"

type Enum = uint32

// we use a pointer to js.Value as we need to be able to compare with the zero value of the struct

type Attrib struct {
	Value *js.Value
}

type Program struct {
	Init  bool
	Value *js.Value
}

type Shader struct {
	Value *js.Value
}

type Buffer struct {
	Value *js.Value
}

type Framebuffer struct {
	Value *js.Value
}

type Renderbuffer struct {
	Value *js.Value
}

type Texture struct {
	Value *js.Value
}

type Uniform struct {
	Value *js.Value
}

type VertexArray struct {
	Value *js.Value
}

func (v Attrib) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Program) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Shader) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Buffer) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Framebuffer) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Renderbuffer) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Texture) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v Uniform) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
func (v VertexArray) c() js.Value {
	if v.Value == nil {
		return js.Undefined()
	}

	return *v.Value
}
