// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package portable

import "image"

func affine(dst *image.RGBA, src *image.RGBA, a [3][3]float32) {
	srcb := src.Bounds()
	b := dst.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sx, sy := pt(a, x, y)
			if !inBounds(srcb, sx, sy) {
				continue
			}
			c := bilinear(src, sx, sy)
			off := (y-dst.Rect.Min.Y)*dst.Stride + (x-dst.Rect.Min.X)*4
			dst.Pix[off+0] = c.R
			dst.Pix[off+1] = c.G
			dst.Pix[off+2] = c.B
			dst.Pix[off+3] = c.A
		}
	}
}

func inBounds(b image.Rectangle, x, y float32) bool {
	if x < float32(b.Min.X) || x >= float32(b.Max.X) {
		return false
	}
	if y < float32(b.Min.Y) || y >= float32(b.Max.Y) {
		return false
	}
	return true
}

func pt(a [3][3]float32, x0, y0 int) (x1, y1 float32) {
	fx := float32(x0) + 0.5
	fy := float32(y0) + 0.5
	x1 = fx*a[0][0] + fy*a[0][1] + a[0][2]
	y1 = fx*a[1][0] + fy*a[1][1] + a[1][2]
	return x1, y1
}
