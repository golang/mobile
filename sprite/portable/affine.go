// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package portable

import (
	"image"
	"image/draw"

	"code.google.com/p/go.mobile/f32"
)

// affine draws each pixel of dst using bilinear interpolation of the
// affine-transformed position in src. This is equivalent to:
//
//	for each (x,y) in dst:
//		dst(x,y) = bilinear interpolation of src(a*(x,y))
//
// While this is the simpler implementation, it can be counter-
// intuitive as an affine transformation is usually described in terms
// of the source, not the destination. For example, a scale transform
//
//	Affine{{2, 0, 0}, {0, 2, 0}}
//
// will produce a dst that is half the size of src. To perform a
// traditional affine transform, use the inverse of the affine matrix.
func affine(dst *image.RGBA, src *image.RGBA, a *f32.Affine, op draw.Op) {
	srcb := src.Bounds()
	b := dst.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// Interpolate from the bounds of the src sub-image
			// to the bounds of the dst sub-image.
			sx, sy := pt(a, x-b.Min.X, y-b.Min.Y)
			sx += float32(srcb.Min.X)
			sy += float32(srcb.Min.Y)
			if !inBounds(srcb, sx, sy) {
				continue
			}

			c := bilinear(src, sx, sy)
			off := (y-dst.Rect.Min.Y)*dst.Stride + (x-dst.Rect.Min.X)*4
			if op == draw.Over {
				// This logic comes from drawCopyOver in the image/draw package.
				sr := uint32(c.R) * 0x101
				sg := uint32(c.G) * 0x101
				sb := uint32(c.B) * 0x101
				sa := uint32(c.A) * 0x101

				dr := uint32(dst.Pix[off+0])
				dg := uint32(dst.Pix[off+1])
				db := uint32(dst.Pix[off+2])
				da := uint32(dst.Pix[off+3])

				// m is the maximum color value returned by image.Color.RGBA.
				const m = 1<<16 - 1

				// dr, dg, db and da are all 8-bit color at the moment, ranging in [0,255].
				// We work in 16-bit color, and so would normally do:
				// dr |= dr << 8
				// and similarly for dg, db and da, but instead we multiply a
				// (which is a 16-bit color, ranging in [0,65535]) by 0x101.
				// This yields the same result, but is fewer arithmetic operations.
				a := (m - sa) * 0x101

				dst.Pix[off+0] = uint8((dr*a/m + sr) >> 8)
				dst.Pix[off+1] = uint8((dg*a/m + sg) >> 8)
				dst.Pix[off+2] = uint8((db*a/m + sb) >> 8)
				dst.Pix[off+3] = uint8((da*a/m + sa) >> 8)
			} else {
				dst.Pix[off+0] = c.R
				dst.Pix[off+1] = c.G
				dst.Pix[off+2] = c.B
				dst.Pix[off+3] = c.A
			}
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

func pt(a *f32.Affine, x0, y0 int) (x1, y1 float32) {
	fx := float32(x0) + 0.5
	fy := float32(y0) + 0.5
	x1 = fx*a[0][0] + fy*a[0][1] + a[0][2]
	y1 = fx*a[1][0] + fy*a[1][1] + a[1][2]
	return x1, y1
}
