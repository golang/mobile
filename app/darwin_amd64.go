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
#include <pthread.h>

void runApp(void);
void stopApp(void);
void makeCurrentContext(GLintptr);
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
	// This means the goroutine running main (and calling runApp below)
	// is locked to the OS thread that started the program. This is
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
		C.stopApp()
		// TODO(crawshaw): trigger runApp to return
	}()

	C.runApp()
}

// loop is the primary drawing loop.
//
// After Cocoa has captured the initial OS thread for processing Cocoa
// events in runApp, it starts loop on another goroutine. It is locked
// to an OS thread for its OpenGL context.
//
// Two Cocoa threads deliver draw signals to loop. The primary source of
// draw events is the CVDisplayLink timer, which is tied to the display
// vsync. Secondary draw events come from [NSView drawRect:] when the
// window is resized.
func loop(ctx C.GLintptr) {
	runtime.LockOSThread()
	C.makeCurrentContext(ctx)

	for range draw {
		eventsIn <- event.Draw{}
	loop1:
		for {
			select {
			case <-gl.WorkAvailable:
				gl.DoWork()
			case <-endDraw:
				C.CGLFlushDrawable(C.CGLGetCurrentContext())
				break loop1
			}
		}
		drawDone <- struct{}{}
	}
}

var (
	draw     = make(chan struct{})
	drawDone = make(chan struct{})
)

//export drawgl
func drawgl() {
	draw <- struct{}{}
	<-drawDone
}

//export startloop
func startloop(ctx C.GLintptr) {
	go loop(ctx)
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

func sendTouch(c event.Change, x, y float32) {
	eventsIn <- event.Touch{
		ID:     0,
		Change: c,
		Loc: geom.Point{
			X: geom.Pt(x / pixelsPerPt),
			Y: windowHeight - geom.Pt(y/pixelsPerPt),
		},
	}
}

//export eventMouseDown
func eventMouseDown(x, y float32) { sendTouch(event.ChangeOn, x, y) }

//export eventMouseDragged
func eventMouseDragged(x, y float32) { sendTouch(event.ChangeNone, x, y) }

//export eventMouseEnd
func eventMouseEnd(x, y float32) { sendTouch(event.ChangeOff, x, y) }

//export lifecycleDead
func lifecycleDead() { sendLifecycle(event.LifecycleStageDead) }

//export lifecycleAlive
func lifecycleAlive() { sendLifecycle(event.LifecycleStageAlive) }

//export lifecycleVisible
func lifecycleVisible() { sendLifecycle(event.LifecycleStageVisible) }

//export lifecycleFocused
func lifecycleFocused() { sendLifecycle(event.LifecycleStageFocused) }
