// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build gldebug

package gl

//#cgo darwin  LDFLAGS: -framework OpenGL
//#cgo linux   LDFLAGS: -lGLESv2
//#include "gl2.h"
import "C"
import (
	"fmt"
	"unsafe"

	"code.google.com/p/go.mobile/f32"
)

type Enum uint32

type Attrib struct {
	Value uint
	name  string
}

type Program struct {
	Value uint32
}

type Shader struct {
	Value uint32
}

type Buffer struct {
	Value uint32
}

type Framebuffer struct {
	Value uint32
}

type Renderbuffer struct {
	Value uint32
}

type Texture struct {
	Value uint32
}

type Uniform struct {
	Value int32
	name  string
}

func (u Uniform) WriteMat4(m *f32.Mat4) {
	UniformMatrix4fv(u, (*[16]float32)(unsafe.Pointer(m))[:])
}

func (u Uniform) WriteVec4(v *f32.Vec4) {
	Uniform4f(u, v[0], v[1], v[2], v[3])
}

func (v Attrib) c() C.GLuint       { return C.GLuint(v.Value) }
func (v Enum) c() C.GLenum         { return C.GLenum(v) }
func (v Program) c() C.GLuint      { return C.GLuint(v.Value) }
func (v Shader) c() C.GLuint       { return C.GLuint(v.Value) }
func (v Buffer) c() C.GLuint       { return C.GLuint(v.Value) }
func (v Framebuffer) c() C.GLuint  { return C.GLuint(v.Value) }
func (v Renderbuffer) c() C.GLuint { return C.GLuint(v.Value) }
func (v Texture) c() C.GLuint      { return C.GLuint(v.Value) }
func (v Uniform) c() C.GLint       { return C.GLint(v.Value) }

func (v Attrib) String() string       { return fmt.Sprintf("Attrib(%d:%s)", v.Value, v.name) }
func (v Program) String() string      { return fmt.Sprintf("Program(%d)", v.Value) }
func (v Shader) String() string       { return fmt.Sprintf("Shader(%d)", v.Value) }
func (v Buffer) String() string       { return fmt.Sprintf("Buffer(%d)", v.Value) }
func (v Framebuffer) String() string  { return fmt.Sprintf("Framebuffer(%d)", v.Value) }
func (v Renderbuffer) String() string { return fmt.Sprintf("Renderbuffer(%d)", v.Value) }
func (v Texture) String() string      { return fmt.Sprintf("Texture(%d)", v.Value) }
func (v Uniform) String() string      { return fmt.Sprintf("Uniform(%d:%s)", v.Value, v.name) }
