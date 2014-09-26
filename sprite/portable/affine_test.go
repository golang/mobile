// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package portable

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestAffine(t *testing.T) {
	red := color.RGBA{0xff, 0, 0, 0xff}
	dst := image.NewRGBA(image.Rect(0, 0, 30, 30))
	src := image.NewRGBA(dst.Bounds())
	draw.Draw(src, src.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)

	affine(dst, src, [3][3]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	})

	for y := 0; y < dst.Bounds().Dy(); y++ {
		for x := 0; x < dst.Bounds().Dx(); x++ {
			if c := dst.At(x, y); c != red {
				t.Fatalf("At(%d, %d) = %v, want red", x, y, c)
			}
		}
	}
}
