// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f32

import "fmt"

// A Mat3 is a 3x3 matrix of float32 values.
// Elements are indexed first by row then column, i.e. m[row][column].
type Mat3 [3]Vec3

func (m Mat3) String() string {
	return fmt.Sprintf(`Mat3[% 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f,
     % 0.3f, % 0.3f, % 0.3f]`,
		m[0][0], m[0][1], m[0][2],
		m[1][0], m[1][1], m[1][2],
		m[2][0], m[2][1], m[2][2])
}

func (m *Mat3) Identity() {
	*m = Mat3{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
}

func (m *Mat3) Eq(n *Mat3, epsilon float32) bool {
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

// Mul stores m0 × m1 in m.
func (m *Mat3) Mul(m0, m1 *Mat3) {
	// If you intend to make this faster, skip the usual loop unrolling
	// games and go straight to halide/renderscript/etc.
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = m0[i][0]*m1[0][j] +
				m0[i][1]*m1[1][j] +
				m0[i][2]*m1[2][j]
		}
	}
}
