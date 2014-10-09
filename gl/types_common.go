// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl

// This file contains GL Types and their methods that are independent of the
// "gldebug" build tag.

import "code.google.com/p/go.mobile/f32"

// WriteAffine writes the contents of an Affine to a 3x3 matrix GL uniform.
func (u Uniform) WriteAffine(a *f32.Affine) {
	var m [9]float32
	m[0*3+0] = a[0][0]
	m[0*3+1] = a[1][0]
	m[0*3+2] = 0
	m[1*3+0] = a[0][1]
	m[1*3+1] = a[1][1]
	m[1*3+2] = 0
	m[2*3+0] = a[0][2]
	m[2*3+1] = a[1][2]
	m[2*3+2] = 1
	UniformMatrix3fv(u, m[:])
}

// WriteMat4 writes the contents of a 4x4 matrix to a GL uniform.
func (u Uniform) WriteMat4(p *f32.Mat4) {
	var m [16]float32
	m[0*3+0] = p[0][0]
	m[0*3+1] = p[1][0]
	m[0*3+2] = p[2][0]
	m[0*3+3] = p[3][0]
	m[1*3+0] = p[0][1]
	m[1*3+1] = p[1][1]
	m[1*3+2] = p[2][1]
	m[1*3+3] = p[3][1]
	m[2*3+0] = p[0][2]
	m[2*3+1] = p[1][2]
	m[2*3+2] = p[2][2]
	m[2*3+3] = p[3][2]
	m[3*3+0] = p[0][3]
	m[3*3+1] = p[1][3]
	m[3*3+2] = p[2][3]
	m[3*3+3] = p[3][3]
	UniformMatrix4fv(u, m[:])
}

// WriteVec4 writes the contents of a 4-element vector to a GL uniform.
func (u Uniform) WriteVec4(v *f32.Vec4) {
	Uniform4f(u, v[0], v[1], v[2], v[3])
}
