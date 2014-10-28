// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// An android app that draws a green triangle on a red background.
package main

import (
	"bytes"
	"encoding/binary"
	"log"

	"code.google.com/p/go.mobile/app"
	"code.google.com/p/go.mobile/app/debug"
	"code.google.com/p/go.mobile/event"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/gl"
	"code.google.com/p/go.mobile/gl/glutil"
)

var (
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	green    float32
	touchLoc geom.Point
)

func main() {
	log.Printf("basic example starting")
	app.Run(app.Callbacks{
		Draw:  draw,
		Touch: touch,
	})
}

// TODO(crawshaw): Need an easier way to do GL-dependent initialization.

func initGL() {
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.GenBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, triangleData)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")
	touchLoc = geom.Point{geom.Width / 2, geom.Height / 2}
}

func touch(t event.Touch) {
	touchLoc = t.Loc
}

func draw() {
	if program.Value == 0 {
		initGL()
		log.Printf("example/basic rendering initialized")
	}

	gl.ClearColor(1, 0, 0, 1)
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	green += 0.01
	if green > 1 {
		green = 0
	}
	gl.Uniform4f(color, 0, green, 0, 1)

	gl.Uniform2f(offset, float32(touchLoc.X/geom.Width), float32(touchLoc.Y/geom.Height))

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(position)

	debug.DrawFPS()
}

var triangleData []byte

func init() {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, triangleCoords); err != nil {
		log.Fatal(err)
	}
	triangleData = buf.Bytes()
}

const coordsPerVertex = 3

var triangleCoords = []float32{
	0, 0.4, 0, // top left
	0, 0, 0, // bottom left
	0.4, 0, 0, // bottom right
}
var vertexCount = len(triangleCoords) / coordsPerVertex

const vertexShader = `#version 100
uniform vec2 offset;

attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
