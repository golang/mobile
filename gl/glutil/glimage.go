// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import (
	"bytes"
	"encoding/binary"
	"image"
	"log"
	"math"
	"sync"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/gl"
)

var glimage struct {
	sync.Once
	square        gl.Buffer
	squareUV      gl.Buffer
	program       gl.Program
	pos           gl.Attrib
	mvp           gl.Uniform
	uvp           gl.Uniform
	inUV          gl.Attrib
	textureSample gl.Uniform
}

func glInit() {
	var err error
	glimage.program, err = CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	glimage.square = gl.GenBuffer()
	glimage.squareUV = gl.GenBuffer()

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.square)
	gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, squareCoords)
	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.squareUV)
	gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, squareUVCoords)

	glimage.pos = gl.GetAttribLocation(glimage.program, "pos")
	glimage.mvp = gl.GetUniformLocation(glimage.program, "mvp")
	glimage.uvp = gl.GetUniformLocation(glimage.program, "uvp")
	glimage.inUV = gl.GetAttribLocation(glimage.program, "inUV")
	glimage.textureSample = gl.GetUniformLocation(glimage.program, "textureSample")
}

// Image bridges between an *image.RGBA and an OpenGL texture.
//
// The contents of the embedded *image.RGBA can be uploaded as a
// texture and drawn as a 2D quad.
//
// The number of active Images must fit in the system's OpenGL texture
// limit. The typical use of an Image is as a texture atlas.
type Image struct {
	*image.RGBA

	Texture   gl.Texture
	texWidth  int
	texHeight int
}

// NewImage creates an Image of the given size.
//
// Both a host-memory *image.RGBA and a GL texture are created.
func NewImage(size geom.Point) *Image {
	realx := int(math.Ceil(float64(size.X.Px())))
	realy := int(math.Ceil(float64(size.Y.Px())))
	dx := roundToPower2(realx)
	dy := roundToPower2(realy)

	// TODO(crawshaw): Using VertexAttribPointer we can pass texture
	// data with a stride, which would let us use the exact number of
	// pixels on the host instead of the rounded up power 2 size.
	m := image.NewRGBA(image.Rect(0, 0, dx, dy))

	glimage.Do(glInit)

	img := &Image{
		RGBA:      m.SubImage(image.Rect(0, 0, realx, realy)).(*image.RGBA),
		Texture:   gl.GenTexture(),
		texWidth:  dx,
		texHeight: dy,
	}
	// TODO(crawshaw): We don't have the context on a finalizer. Find a way.
	// runtime.SetFinalizer(img, func(img *Image) { gl.DeleteTexture(img.Texture) })
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return img
}

func roundToPower2(x int) int {
	x2 := 1
	for x2 < x {
		x2 *= 2
	}
	return x2
}

// Upload copies the host image data to the GL device.
func (img *Image) Upload() {
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, img.texWidth, img.texHeight, gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)
}

// Draw draws the image onto the current GL framebuffer.
func (img *Image) Draw(dstBounds geom.Rectangle, srcBounds image.Rectangle) {
	// TODO(crawshaw): Adjust viewport for the top bar on android?
	gl.UseProgram(glimage.program)

	// We are drawing a sub-image of dst, defined by dstBounds. Let ABCD
	// be the image, and PQRS be the sub-image. The two images may actually
	// be equal, but in the general case, PQRS can be smaller:
	//
	//	       M
	//	A +----+----------+ B
	//	  |               |
	//	N +  P +-----+ Q  |
	//	  |    |     |    |
	//	  |  S +-----+ R  |
	//	  |               |
	//	D +---------------+ C
	//
	// There are two co-ordinate spaces: geom space and framebuffer space.
	// In geom space, the ABCD rectangle is:
	//
	//	(0, 0)           (geom.Width, 0)
	//	(0, geom.Height) (geom.Width, geom.Height)
	//
	// and the PQRS rectangle is:
	//
	//	(dstBounds.Min.X, dstBounds.Min.Y) (dstBounds.Max.X, dstBounds.Min.Y)
	//	(dstBounds.Min.X, dstBounds.Max.Y) (dstBounds.Max.X, dstBounds.Max.Y)
	//
	// In framebuffer space, the ABCD rectangle is:
	//
	//	(-1, +1) (+1, +1)
	//	(-1, -1) (+1, -1)
	//
	// We need to solve for PQRS' co-ordinates in framebuffer space, and
	// calculate the MVP matrix that transforms the -1/+1 ABCD co-ordinates
	// to PQRS co-ordinates.
	//
	// To solve for PQRS, note that PQ / AB must match in both spaces. Call
	// this ratio fracX, and likewise for fracY.
	//
	//	[EQ1]   fracX = (dstBounds.Max.X - dstBounds.Min.X) / geom.Width
	//
	// Similarly, the AM / AB ratio must match:
	//
	//	(P.x - -1) / (1 - -1) = dstBounds.Min.X / geom.Width
	//
	// where the LHS is in framebuffer space and the RHS in geom space. This
	// equation is equivalent to:
	//
	//	[EQ2]   P.x = -1 + 2 * dstBounds.Min.X / geom.Width
	//
	// This MVP matrix is a scale followed by a translate. The scale is by
	// (fracX, fracY). After this, our corners have been transformed to:
	//
	//	(-fracX, +fracY) (+fracX, +fracY)
	//	(-fracX, -fracY) (+fracX, -fracY)
	//
	// so the translate is by (P.x + fracX) in the X direction, and
	// likewise for Y. Combining equations EQ1 and EQ2 simplifies the
	// translate to be:
	//
	//	-1 + (dstBounds.Max.X + dstBounds.Min.X) / geom.Width
	//	+1 - (dstBounds.Max.Y + dstBounds.Min.Y) / geom.Height
	var a f32.Affine
	a.Identity()
	a.Translate(
		&a,
		-1+float32((dstBounds.Max.X+dstBounds.Min.X)/geom.Width),
		+1-float32((dstBounds.Max.Y+dstBounds.Min.Y)/geom.Height),
	)
	a.Scale(
		&a,
		float32((dstBounds.Max.X-dstBounds.Min.X)/geom.Width),
		float32((dstBounds.Max.Y-dstBounds.Min.Y)/geom.Height),
	)
	glimage.mvp.WriteAffine(&a)

	// Texture UV co-ordinates start out as:
	//
	//	(0,0) (1,0)
	//	(0,1) (1,1)
	//
	// These co-ordinates need to be scaled to texWidth/Height,
	// which may be less than 1 as the source image may not have
	// power-of-2 dimensions. Then it is scaled and translated
	// to represent the srcBounds rectangle of the source texture.
	//
	// The math is simpler here because in both co-ordinate spaces,
	// the top-left corner is (0, 0).
	a.Identity()
	a.Translate(
		&a,
		float32(srcBounds.Min.X)/float32(img.Rect.Dx()),
		float32(srcBounds.Min.Y)/float32(img.Rect.Dy()),
	)
	a.Scale(
		&a,
		float32(srcBounds.Dx())/float32(img.texWidth),
		float32(srcBounds.Dy())/float32(img.texHeight),
	)
	glimage.uvp.WriteAffine(&a)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.Uniform1i(glimage.textureSample, 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.square)
	gl.EnableVertexAttribArray(glimage.pos)
	gl.VertexAttribPointer(glimage.pos, 2, gl.FLOAT, false, 0, 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.squareUV)
	gl.EnableVertexAttribArray(glimage.inUV)
	gl.VertexAttribPointer(glimage.inUV, 2, gl.FLOAT, false, 0, 0)

	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	gl.DisableVertexAttribArray(glimage.pos)
	gl.DisableVertexAttribArray(glimage.inUV)
}

// Vertices of two triangles.
var squareCoords = toBytes([]float32{
	-1, -1,
	+1, -1,
	+1, +1,

	-1, -1,
	+1, +1,
	-1, +1,
})

var squareUVCoords = toBytes([]float32{
	0, 1,
	1, 1,
	1, 0,

	0, 1,
	1, 0,
	0, 0,
})

func toBytes(v []float32) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
