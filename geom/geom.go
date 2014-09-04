// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geom

import "fmt"

// Pt is a length.
//
// The unit Pt is a typographical point, 1/72 of an inch (0.3527 mm).
//
// It can be be converted to a length in current device pixels by
// multiplying with Scale after app initialization is complete.
type Pt float32

// Px converts the length to current device pixels.
func (p Pt) Px() float32 { return float32(p) * Scale }

func (p Pt) String() string { return fmt.Sprintf("%.2fpt", p) }

// Point is a point in a two-dimensional plane.
type Point struct {
	X, Y Pt
}

func (p Point) String() string { return fmt.Sprintf("Point(%.2f, %.2f)", p.X, p.Y) }

// Scale is the number of pixels in a single Pt on the current device.
//
// There are a wide variety of pixel densities in existing phones and
// tablets, so apps should be written to expect various non-integer
// Scale values. In general, work in Pt.
//
// Not valid until app initialization has completed.
var Scale float32

// Width is the width of the device screen.
// Not valid until app initialization has completed.
var Width Pt

// Height is the height of the device screen.
// Not valid until app initialization has completed.
var Height Pt
