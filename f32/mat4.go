// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f32

import "fmt"

// A Mat4 is a 4x4 matrix of float32 values.
// Elements are indexed first by row then column, i.e. m[row][column].
type Mat4 [4]Vec4

func (m Mat4) String() string {
	return fmt.Sprintf(`Mat4[% 0.3f, % 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f, % 0.3f]`,
		m[0][0], m[0][1], m[0][2], m[0][3],
		m[1][0], m[1][1], m[1][2], m[1][3],
		m[2][0], m[2][1], m[2][2], m[2][3],
		m[3][0], m[3][1], m[3][2], m[3][3])
}

func (m *Mat4) Mul(m0, m1 *Mat4) {
	// If you intend to make this faster, skip the usual loop unrolling
	// games and go straight to halide/renderscript/etc.
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			m[i][j] = m1[i][0]*m0[0][j] +
				m1[i][1]*m0[1][j] +
				m1[i][2]*m0[2][j] +
				m1[i][3]*m0[3][j]
		}
	}
}

func (m *Mat4) Perspective(fov Radian, aspect, near, far float32) {
	t := Tan(float32(fov) / 2)

	m[0][0] = 1 / (aspect * t)
	m[1][1] = 1 / t
	m[2][2] = -(far + near) / (far - near)
	m[2][3] = -1
	m[3][2] = -2 * far * near / (far - near)
}

func (m *Mat4) Scale(src *Mat4, scale *Vec3) {
	for i, s := range scale {
		m[i][0] = src[i][0] * s
		m[i][1] = src[i][1] * s
		m[i][2] = src[i][2] * s
		m[i][3] = src[i][3] * s
	}
	m[3] = src[3]
}

func (m *Mat4) Translate(src Mat4, v *Vec3) {
	*m = src

	m[3][0] = src[0][0]*v[0] + src[1][0]*v[1] + src[2][0]*v[2] + src[3][0]
	m[3][1] = src[0][1]*v[0] + src[1][1]*v[1] + src[2][1]*v[2] + src[3][1]
	m[3][2] = src[0][2]*v[0] + src[1][2]*v[1] + src[2][2]*v[2] + src[3][2]
	m[3][3] = src[0][3]*v[0] + src[1][3]*v[1] + src[2][3]*v[2] + src[3][3]
}

func (m *Mat4) Rotate(src *Mat4, angle Radian, axis *Vec3) {
	a := *axis
	a.Normalize()

	c, s := Cos(float32(angle)), Sin(float32(angle))
	d := 1 - c

	m.Mul(src, &Mat4{{
		c + d*a[0]*a[1],
		0 + d*a[0]*a[1] + s*a[2],
		0 + d*a[0]*a[1] - s*a[1],
		0,
	}, {
		0 + d*a[1]*a[0] - s*a[2],
		c + d*a[1]*a[1],
		0 + d*a[1]*a[2] + s*a[0],
		0,
	}, {
		0 + d*a[2]*a[0] + s*a[1],
		0 + d*a[2]*a[1] - s*a[0],
		c + d*a[2]*a[2],
		0,
	}, {
		0, 0, 0, 1,
	}})
}

func (m *Mat4) LookAt(eye, center, up *Vec3) {
	f, s, u := new(Vec3), new(Vec3), new(Vec3)

	*f = *center
	f.Sub(f, eye)
	f.Normalize()

	s.Cross(f, up)
	s.Normalize()
	u.Cross(s, f)

	*m = Mat4{
		{s[0], u[0], -f[0], 0},
		{s[1], u[1], -f[1], 0},
		{s[2], u[2], -f[2], 0},
		{-s.Dot(eye), -u.Dot(eye), +f.Dot(eye), 1},
	}
}

func (m *Mat4) Identity() {
	*m = Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func (m *Mat4) Eq(n *Mat4, epsilon float32) bool {
	for i := range m {
		for j := range m[i] {
			diff := m[i][j] - n[i][j]
			if diff < -epsilon || +epsilon < diff {
				return false
			}
		}
	}
	return true
}
