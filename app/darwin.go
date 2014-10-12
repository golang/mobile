// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package app

// Simple on-screen app debugging for OS X. Not an officially supported
// development target for apps, as screens with mice are very different
// than screens with touch panels.

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework OpenGL -framework QuartzCore
#import <Cocoa/Cocoa.h>
#import <OpenGL/gl.h>

void glGenVertexArrays(GLsizei n, GLuint* array);
void glBindVertexArray(GLuint array);

void runApp(void);
void lockContext(GLintptr);
void unlockContext(GLintptr);
double backingScaleFactor();
*/
import "C"
import (
	"runtime"
	"sync"

	"code.google.com/p/go.mobile/event"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/gl"
)

func init() {
	runtime.LockOSThread()
}

func run(callbacks Callbacks) {
	cb = callbacks
	C.runApp()
}

//export setGeom
func setGeom(scale, width, height float64) {
	// Macs default to 72 DPI, so scales are equivalent.
	geom.Scale = float32(scale)
	geom.Width = geom.Pt(width)
	geom.Height = geom.Pt(height)
}

func initGL() {
	// Using attribute arrays in OpenGL 3.3 requires the use of a VBA.
	// But VBAs don't exist in ES 2. So we bind one default one.
	var id C.GLuint
	C.glGenVertexArrays(1, &id)
	C.glBindVertexArray(id)
}

var cb Callbacks
var initGLOnce sync.Once
var events = make(chan event.Touch, 1<<6)

func sendTouch(ty event.TouchType, x, y float32) {
	events <- event.Touch{
		Type: ty,
		Loc: geom.Point{
			X: geom.Pt(x),
			Y: geom.Height - geom.Pt(y),
		},
	}
}

//export eventMouseDown
func eventMouseDown(x, y float32) { sendTouch(event.TouchStart, x, y) }

//export eventMouseMove
func eventMouseMove(x, y float32) { sendTouch(event.TouchMove, x, y) }

//export eventMouseEnd
func eventMouseEnd(x, y float32) { sendTouch(event.TouchEnd, x, y) }

//export drawgl
func drawgl(ctx C.GLintptr) {
	runtime.LockOSThread()
	C.lockContext(ctx)

	initGLOnce.Do(initGL)

loop:
	for {
		select {
		case e := <-events:
			if cb.Touch != nil {
				cb.Touch(e)
			}
		default:
			break loop
		}
	}

	// TODO: is the library or the app responsible for clearing the buffers?
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	if cb.Draw != nil {
		cb.Draw()
	}

	C.unlockContext(ctx)
	runtime.UnlockOSThread()
}
