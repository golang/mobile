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
#include <pthread.h>

void glGenVertexArrays(GLsizei n, GLuint* array);
void glBindVertexArray(GLuint array);

void runApp(void);
void makeCurrentContext(GLintptr);
double backingScaleFactor();
uint64 threadID();

*/
import "C"
import (
	"log"
	"runtime"
	"sync"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var initThreadID uint64

func init() {
	// Lock the goroutine responsible for initialization to an OS thread.
	// This means the goroutine running main (and calling the run function
	// below) is locked to the OS thread that started the program. This is
	// necessary for the correct delivery of Cocoa events to the process.
	//
	// A discussion on this topic:
	// https://groups.google.com/forum/#!msg/golang-nuts/IiWZ2hUuLDA/SNKYYZBelsYJ
	runtime.LockOSThread()
	initThreadID = uint64(C.threadID())
}

func main(f func(App)) {
	if tid := uint64(C.threadID()); tid != initThreadID {
		log.Fatalf("app.Main called on thread %d, but app.init ran on %d", tid, initThreadID)
	}

	go func() {
		f(app{})
		// TODO(crawshaw): trigger runApp to return
	}()

	C.runApp()
}

var windowHeight geom.Pt

//export setGeom
func setGeom(ppp float32, width, height int) {
	pixelsPerPt = ppp
	windowHeight = geom.Pt(float32(height) / pixelsPerPt)
	eventsIn <- event.Config{
		Width:       geom.Pt(float32(width) / pixelsPerPt),
		Height:      windowHeight,
		PixelsPerPt: pixelsPerPt,
	}
}

var touchEvents struct {
	sync.Mutex
	pending []event.Touch
}

func sendTouch(ty event.TouchType, x, y float32) {
	eventsIn <- event.Touch{
		ID:   0,
		Type: ty,
		Loc: geom.Point{
			X: geom.Pt(x / pixelsPerPt),
			Y: windowHeight - geom.Pt(y/pixelsPerPt),
		},
	}
}

//export eventMouseDown
func eventMouseDown(x, y float32) { sendTouch(event.TouchStart, x, y) }

//export eventMouseDragged
func eventMouseDragged(x, y float32) { sendTouch(event.TouchMove, x, y) }

//export eventMouseEnd
func eventMouseEnd(x, y float32) { sendTouch(event.TouchEnd, x, y) }

var startedgl = false

//export drawgl
func drawgl(ctx C.GLintptr) {
	if !startedgl {
		startedgl = true
		C.makeCurrentContext(ctx)

		// Using attribute arrays in OpenGL 3.3 requires the use of a VBA.
		// But VBAs don't exist in ES 2. So we bind a default one.
		var id C.GLuint
		C.glGenVertexArrays(1, &id)
		C.glBindVertexArray(id)

		sendLifecycle(event.LifecycleStageFocused)
	}

	// TODO: is the library or the app responsible for clearing the buffers?
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	eventsIn <- event.Draw{}

	for {
		select {
		case <-gl.WorkAvailable:
			gl.DoWork()
		case <-endDraw:
			C.CGLFlushDrawable(C.CGLGetCurrentContext())
			return
		}
	}
}
