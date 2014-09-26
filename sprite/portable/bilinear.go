// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package portable

import (
	"image"
	"image/color"
	"math"
)

func bilinear(src *image.RGBA, x, y float32) color.RGBA {
	p := findLinearSrc(src.Bounds(), x, y)

	// Array offsets for the surrounding pixels.
	off00 := offRGBA(src, p.low.X, p.low.Y)
	off01 := offRGBA(src, p.high.X, p.low.Y)
	off10 := offRGBA(src, p.low.X, p.high.Y)
	off11 := offRGBA(src, p.high.X, p.high.Y)

	fr := float32(src.Pix[off00+0]) * p.frac00
	fg := float32(src.Pix[off00+1]) * p.frac00
	fb := float32(src.Pix[off00+2]) * p.frac00
	fa := float32(src.Pix[off00+3]) * p.frac00

	fr += float32(src.Pix[off01+0]) * p.frac01
	fg += float32(src.Pix[off01+1]) * p.frac01
	fb += float32(src.Pix[off01+2]) * p.frac01
	fa += float32(src.Pix[off01+3]) * p.frac01

	fr += float32(src.Pix[off10+0]) * p.frac10
	fg += float32(src.Pix[off10+1]) * p.frac10
	fb += float32(src.Pix[off10+2]) * p.frac10
	fa += float32(src.Pix[off10+3]) * p.frac10

	fr += float32(src.Pix[off11+0]) * p.frac11
	fg += float32(src.Pix[off11+1]) * p.frac11
	fb += float32(src.Pix[off11+2]) * p.frac11
	fa += float32(src.Pix[off11+3]) * p.frac11

	return color.RGBA{
		R: uint8(fr + 0.5),
		G: uint8(fg + 0.5),
		B: uint8(fb + 0.5),
		A: uint8(fa + 0.5),
	}
}

type bilinearSrc struct {
	// Top-left and bottom-right interpolation sources
	low, high image.Point
	// Fraction of each pixel to take. The 0 suffix indicates
	// top/left, and the 1 suffix indicates bottom/right.
	frac00, frac01, frac10, frac11 float32
}

func floor(x float32) float32 { return float32(math.Floor(float64(x))) }
func ceil(x float32) float32  { return float32(math.Ceil(float64(x))) }

func findLinearSrc(b image.Rectangle, sx, sy float32) bilinearSrc {
	maxX := float32(b.Max.X)
	maxY := float32(b.Max.Y)
	minX := float32(b.Min.X)
	minY := float32(b.Min.Y)
	lowX := floor(sx - 0.5)
	lowY := floor(sy - 0.5)
	if lowX < minX {
		lowX = minX
	}
	if lowY < minY {
		lowY = minY
	}

	highX := ceil(sx - 0.5)
	highY := ceil(sy - 0.5)
	if highX >= maxX {
		highX = maxX - 1
	}
	if highY >= maxY {
		highY = maxY - 1
	}

	// In the variables below, the 0 suffix indicates top/left, and the
	// 1 suffix indicates bottom/right.

	// Center of each surrounding pixel.
	x00 := lowX + 0.5
	y00 := lowY + 0.5
	x01 := highX + 0.5
	y01 := lowY + 0.5
	x10 := lowX + 0.5
	y10 := highY + 0.5
	x11 := highX + 0.5
	y11 := highY + 0.5

	p := bilinearSrc{
		low:  image.Pt(int(lowX), int(lowY)),
		high: image.Pt(int(highX), int(highY)),
	}

	// Literally, edge cases. If we are close enough to the edge of
	// the image, curtail the interpolation sources.
	if lowX == highX && lowY == highY {
		p.frac00 = 1.0
	} else if sy-minY <= 0.5 && sx-minX <= 0.5 {
		p.frac00 = 1.0
	} else if maxY-sy <= 0.5 && maxX-sx <= 0.5 {
		p.frac11 = 1.0
	} else if sy-minY <= 0.5 || lowY == highY {
		p.frac00 = x01 - sx
		p.frac01 = sx - x00
	} else if sx-minX <= 0.5 || lowX == highX {
		p.frac00 = y10 - sy
		p.frac10 = sy - y00
	} else if maxY-sy <= 0.5 {
		p.frac10 = x11 - sx
		p.frac11 = sx - x10
	} else if maxX-sx <= 0.5 {
		p.frac01 = y11 - sy
		p.frac11 = sy - y01
	} else {
		p.frac00 = (x01 - sx) * (y10 - sy)
		p.frac01 = (sx - x00) * (y11 - sy)
		p.frac10 = (x11 - sx) * (sy - y00)
		p.frac11 = (sx - x10) * (sy - y01)
	}

	return p
}

func offRGBA(src *image.RGBA, x, y int) int {
	return (y-src.Rect.Min.Y)*src.Stride + (x-src.Rect.Min.X)*4
}
