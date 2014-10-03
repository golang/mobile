// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f32

import "fmt"

// An Affine is a 3x3 matrix of float32 values for which the bottom row is
// implicitly always equal to [0 0 1].
// Elements are indexed first by row then column, i.e. m[row][column].
type Affine [2]Vec3

func (m Affine) String() string {
	return fmt.Sprintf(`Affine[% 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f]`,
		m[0][0], m[0][1], m[0][2],
		m[1][0], m[1][1], m[1][2])
}

func (m *Affine) Identity() {
	*m = Affine{
		{1, 0, 0},
		{0, 1, 0},
	}
}

func (m *Affine) Eq(n *Affine, epsilon float32) bool {
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

// Mul stores a × b in m.
func (m *Affine) Mul(a, b *Affine) {
	// Store the result in local variables, in case m == a || m == b.
	m00 := a[0][0]*b[0][0] + a[0][1]*b[1][0]
	m01 := a[0][0]*b[0][1] + a[0][1]*b[1][1]
	m02 := a[0][0]*b[0][2] + a[0][1]*b[1][2] + a[0][2]
	m10 := a[1][0]*b[0][0] + a[1][1]*b[1][0]
	m11 := a[1][0]*b[0][1] + a[1][1]*b[1][1]
	m12 := a[1][0]*b[0][2] + a[1][1]*b[1][2] + a[1][2]
	m[0][0] = m00
	m[0][1] = m01
	m[0][2] = m02
	m[1][0] = m10
	m[1][1] = m11
	m[1][2] = m12
}
