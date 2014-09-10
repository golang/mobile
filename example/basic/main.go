// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// An android app that draws a green triangle on a red background.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"code.google.com/p/go.mobile/app"
	"code.google.com/p/go.mobile/gl"
)

var (
	program   gl.Program
	vPosition gl.Attrib
	fColor    gl.Uniform
	buf       gl.Buffer
	green     float32
)

func main() {
	app.Run(app.Callbacks{
		Draw: renderFrame,
	})
}

// TODO(crawshaw): Need an easier way to do GL-dependent initialization.

func initGL() {
	var err error
	program, err = createProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.GenBuffers(1)[0]
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.STATIC_DRAW, triangleData)

	vPosition = gl.GetAttribLocation(program, "vPosition")
	fColor = gl.GetUniformLocation(program, "fColor")
}

func renderFrame() {
	if program == 0 {
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
	gl.Uniform4f(fColor, 0, green, 0, 1)

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.EnableVertexAttribArray(vPosition)
	gl.VertexAttribPointer(vPosition, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(vPosition)
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
	// in counterclockwise order:
	0.0, 0.622008459, 0.0, // top
	-0.5, -0.311004243, 0.0, // bottom left
	0.5, -0.311004243, 0.0, // bottom right
}
var vertexCount = len(triangleCoords) / coordsPerVertex

const vertexShader = `
attribute vec4 vPosition;
void main() {
  gl_Position = vPosition;
}`

const fragmentShader = `
precision mediump float;
uniform vec4 fColor;
void main() {
  gl_FragColor = fColor;
}`

func loadShader(shaderType gl.Enum, src string) (gl.Shader, error) {
	shader := gl.CreateShader(shaderType)
	if shader == 0 {
		return 0, fmt.Errorf("could not create shader (type %v)", shaderType)
	}
	gl.ShaderSource(shader, src)
	gl.CompileShader(shader)
	if gl.GetShaderi(shader, gl.COMPILE_STATUS) == 0 {
		defer gl.DeleteShader(shader)
		return 0, fmt.Errorf("shader compile: %s", gl.GetShaderInfoLog(shader))
	}
	return shader, nil
}

// TODO(crawshaw): Consider moving createProgram to a utility package.

func createProgram(vertexSrc, fragmentSrc string) (gl.Program, error) {
	program := gl.CreateProgram()
	if program == 0 {
		return 0, fmt.Errorf("cannot create program")
	}

	vertexShader, err := loadShader(gl.VERTEX_SHADER, vertexSrc)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, fragmentSrc)
	if err != nil {
		gl.DeleteShader(vertexShader)
		return 0, err
	}

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Flag shaders for deletion when program is unlinked.
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	if gl.GetProgrami(program, gl.LINK_STATUS) == 0 {
		defer gl.DeleteProgram(program)
		return 0, fmt.Errorf("program link: %s", gl.GetProgramInfoLog(program))
	}
	return program, nil
}
