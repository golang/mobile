// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gen.go -output table.go

// Package f32 implements some linear algebra and GL helpers for float32s.
//
// Types defined in this package have methods implementing common
// mathematical operations. The common form for these functions is
//
//	func (dst *T) Op(lhs, rhs *T)
//
// which reads in traditional mathematical notation as
//
//	dst = lhs op rhs.
//
// It is safe to use the destination address as the left-hand side,
// that is, dst *= rhs is dst.Mul(dst, rhs).
//
// WARNING
//
// The interface to this package is not stable. It will change considerably.
// Only use functions that provide package documentation. Semantics are
// non-obvious. Be prepared for the package name to change.
package f32

import "math"

type Radian float32

func Cos(x float32) float32 {
	const n = sinTableLen
	i := uint32(int32(x * (n / math.Pi)))
	i += n / 2
	i &= 2*n - 1
	if i >= n {
		return -sinTable[i&(n-1)]
	}
	return sinTable[i&(n-1)]
}

func Sin(x float32) float32 {
	const n = sinTableLen
	i := uint32(int32(x * (n / math.Pi)))
	i &= 2*n - 1
	if i >= n {
		return -sinTable[i&(n-1)]
	}
	return sinTable[i&(n-1)]
}

func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x))) // TODO(crawshaw): implement
}

func Tan(x float32) float32 {
	return float32(math.Tan(float64(x))) // TODO(crawshaw): fast version
}
