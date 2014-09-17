// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package f32 implements some linear algebra and GL helpers for float32s.
//
// WARNING
//
// The interface to this package is not stable. It will change considerably.
// Only use functions that provide package documentation. Semantics are
// non-obvious. Be prepared for the package name to change.
//
// Oh, and it's slow. Sorry.
package f32

import "math"

type Radian float32

func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x))) // TODO(crawshaw): implement
}

func Tan(x float32) float32 {
	return float32(math.Tan(float64(x))) // TODO(crawshaw): fast version
}
