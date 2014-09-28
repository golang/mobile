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

// String returns a string representation of p like "3.2pt".
func (p Pt) String() string { return fmt.Sprintf("%.2fpt", p) }

// Point is a point in a two-dimensional plane.
type Point struct {
	X, Y Pt
}

// String returns a string representation of p like "(1.2,3.4)".
func (p Point) String() string { return fmt.Sprintf("(%.2f,%.2f)", p.X, p.Y) }

// A Rectangle is region of points.
// The top-left point is Min, and the bottom-right point is Max.
type Rectangle struct {
	Min, Max Point
}

// String returns a string representation of r like "(3,4)-(6,5)".
func (r Rectangle) String() string { return r.Min.String() + "-" + r.Max.String() }

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
