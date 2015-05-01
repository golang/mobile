// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(jbd): Refer to http://golang.org/doc/install/source
// once Go 1.5 instructions are complete.

// An app that draws a green triangle on a red background.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool, you need to install
// Go 1.5 from the source.
//
// Clone the source from the tip under $HOME/go directory. On Windows,
// you may like to clone the repo to your user folder, %USERPROFILE%\go.
//
//   $ git clone https://go.googlesource.com/go $HOME/go
//
// Go 1.5 requires Go 1.4. Read more about this requirement at
// http://golang.org/s/go15bootstrap.
// Set GOROOT_BOOTSTRAP to the GOROOT of your existing 1.4 installation or
// follow the steps below to checkout go1.4 from the source and build.
//
//   $ git clone https://go.googlesource.com/go $HOME/go1.4
//   $ cd $HOME/go1.4
//   $ git checkout go1.4.1
//   $ cd src && ./make.bash
//
// If you clone Go 1.4 to a different destination, set GOROOT_BOOTSTRAP
// environmental variable accordingly.
//
// Build Go 1.5 and add Go 1.5 bin to your path.
//
//   $ cd $HOME/go/src && ./make.bash
//   $ export PATH=$PATH:$HOME/go/bin
//
// Set a GOPATH if no GOPATH is set, add $GOPATH/bin to your path.
//
//   $ export GOPATH=$HOME
//   $ export PATH=$PATH:$GOPATH/bin
//
// Get the gomobile tool and initialize.
//
//   $ go get golang.org/x/mobile/cmd/gomobile
//   $ gomobile init
//
// It may take a while to initialize gomobile, please wait.
//
// Get the basic example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/basic
//   $ gomobile build golang.org/x/mobile/example/basic # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/basic
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/basic && basic
package main

import (
	"encoding/binary"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/app/debug"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/gl/glutil"
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
	app.Run(app.Callbacks{
		Start: start,
		Stop:  stop,
		Draw:  draw,
		Touch: touch,
	})
}

func start() {
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")
	touchLoc = geom.Point{geom.Width / 2, geom.Height / 2}

	// TODO(crawshaw): the debug package needs to put GL state init here
}

func stop() {
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
}

func touch(t event.Touch) {
	touchLoc = t.Loc
}

func draw() {
	gl.ClearColor(1, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

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

var triangleData = f32.Bytes(binary.LittleEndian,
	0.0, 0.4, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right
)

const (
	coordsPerVertex = 3
	vertexCount     = 3
)

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
