// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(crawshaw): A GL texture backed by an *image.RGBA is a common
// concept we will want to export. Find a generally useful interface.
// Think about the tricky things: mipmaps, a varying Rect, just backed
// by texture, not 2D drawer, etc. Maybe skip all of these issues and
// just solve the smaller problem: an unscaled sprite.

package debug

import (
	"bytes"
	"encoding/binary"
	"image"
	"log"
	"runtime"
	"sync"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/gl"
	"code.google.com/p/go.mobile/gl/glutil"
)

var glimage struct {
	sync.Once
	square        gl.Buffer
	squareUV      gl.Buffer
	program       gl.Program
	pos           gl.Attrib
	mvp           gl.Uniform
	inUV          gl.Attrib
	textureSample gl.Uniform
}

func glInit() {
	var err error
	glimage.program, err = glutil.CreateProgram(vertexShader, fragmentShader)
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
	glimage.inUV = gl.GetAttribLocation(glimage.program, "inUV")
	glimage.textureSample = gl.GetUniformLocation(glimage.program, "textureSample")
}

type rgba struct {
	Image        *image.RGBA
	Texture      gl.Texture
	sizeX, sizeY float32
}

func roundToPower2(x int) int {
	x2 := 1
	for x2 < x {
		x2 *= 2
	}
	return x2
}

func newRGBA(size geom.Point) *rgba {
	dx := roundToPower2(int(size.X.Px()))
	dy := roundToPower2(int(size.Y.Px()))

	imgRGBA := image.NewRGBA(image.Rect(0, 0, dx, dy))

	glimage.Do(glInit)

	w, h := imgRGBA.Rect.Dx(), imgRGBA.Rect.Dy()

	img := &rgba{
		Image:   imgRGBA, // TODO: embed?
		Texture: gl.GenTexture(),
	}
	runtime.SetFinalizer(img, func(img *rgba) {
		gl.DeleteTexture(img.Texture)
	})
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	img.sizeX = float32(size.X / geom.Width)
	img.sizeY = float32(size.Y / geom.Height)

	return img
}

func (img *rgba) Upload() {
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	w, h := img.Image.Rect.Dx(), img.Image.Rect.Dy()
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, img.Image.Pix)
}

// TODO(crawshaw): Clip/Scale options. Introduce geom.Rectangle?
// TODO(crawshaw): Adjust viewport for the top bar on android?
func (img *rgba) Draw(at geom.Point) {
	gl.UseProgram(glimage.program)

	// Screen plane
	//	(-1, 1) ( 1, 1)
	//	(-1,-1) ( 1,-1)
	// Unscaled image corners
	//	(1,0) (1,1)
	//	(0,0) (0,1)
	// Orig size.X / geom.Width is a [0,1] fraction of screen.
	// The at position is referenced from the top of the image.
	var m f32.Mat4
	m.Identity()
	m.Translate(m, &f32.Vec3{
		-1 + float32(2*at.X/geom.Width),
		+1 - 2*img.sizeY - float32(2*at.Y/geom.Height),
		0,
	})
	m.Scale(&m, &f32.Vec3{
		2 * img.sizeX,
		2 * img.sizeY,
		1,
	})
	glimage.mvp.WriteMat4(&m)

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
	0, 0,
	1, 0,
	1, 1,

	0, 0,
	1, 1,
	0, 1,
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
