// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!android

package app

/*
Simple on-screen app debugging for X11. Not an officially supported
development target for apps, as screens with mice are very different
than screens with touch panels.

On Ubuntu 14.04 'Trusty', you may have to install these libraries:
sudo apt-get install libegl1-mesa-dev libgles2-mesa-dev libx11-dev
*/

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11

void createWindow(void);
void processEvents(void);
void swapBuffers(void);
*/
import "C"
import (
	"runtime"
	"time"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

func init() {
	registerGLViewportFilter()
}

func main(f func(App)) {
	runtime.LockOSThread()
	C.createWindow()

	// TODO: send lifecycle events when e.g. the X11 window is iconified or moved off-screen.
	sendLifecycle(event.LifecycleStageFocused)

	donec := make(chan struct{})
	go func() {
		f(app{})
		close(donec)
	}()

	// TODO: can we get the actual vsync signal?
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	tc := ticker.C

	for {
		select {
		case <-donec:
			return
		case <-gl.WorkAvailable:
			gl.DoWork()
		case <-endDraw:
			C.swapBuffers()
			tc = ticker.C
		case <-tc:
			tc = nil
			eventsIn <- event.Draw{}
		}
		C.processEvents()
	}
}

//export onResize
func onResize(w, h int) {
	// TODO(nigeltao): don't assume 72 DPI. DisplayWidth and DisplayWidthMM
	// is probably the best place to start looking.
	pixelsPerPt = 1
	eventsIn <- event.Config{
		Width:       geom.Pt(w),
		Height:      geom.Pt(h),
		PixelsPerPt: pixelsPerPt,
	}
}

func sendTouch(c event.Change, x, y float32) {
	eventsIn <- event.Touch{
		ID:     0, // TODO: button??
		Change: c,
		Loc: geom.Point{
			X: geom.Pt(x / pixelsPerPt),
			Y: geom.Pt(y / pixelsPerPt),
		},
	}
}

//export onTouchStart
func onTouchStart(x, y float32) { sendTouch(event.ChangeOn, x, y) }

//export onTouchMove
func onTouchMove(x, y float32) { sendTouch(event.ChangeNone, x, y) }

//export onTouchEnd
func onTouchEnd(x, y float32) { sendTouch(event.ChangeOff, x, y) }

var stopped bool

//export onStop
func onStop() {
	if stopped {
		return
	}
	stopped = true
	sendLifecycle(event.LifecycleStageDead)
	eventsIn <- stopPumping{}
}
