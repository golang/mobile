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
	"sync"
	"time"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

type windowEventType byte

const (
	start windowEventType = iota
	stop
	resize
)

type windowEvent struct {
	eventType  windowEventType
	arg1, arg2 int
}

var windowEvents struct {
	sync.Mutex
	events  []windowEvent
	touches []event.Touch
}

func run(cbs []Callbacks) {
	runtime.LockOSThread()
	callbacks = cbs

	go gl.Start(func() {
		C.createWindow()
		sendEvent(windowEvent{start, 0, 0})
	})

	for range time.Tick(time.Second / 60) {
		windowEvents.Lock()
		events := windowEvents.events
		touches := windowEvents.touches
		windowEvents.events = nil
		windowEvents.touches = nil
		windowEvents.Unlock()

		for _, ev := range events {
			switch ev.eventType {
			case start:
				close(mainCalled)
				stateStart(callbacks)

			case stop:
				stateStop(callbacks)
				return

			case resize:
				w := ev.arg1
				h := ev.arg2

				// TODO(nigeltao): don't assume 72 DPI. DisplayWidth / DisplayWidthMM
				// is probably the best place to start looking.
				if geom.PixelsPerPt == 0 {
					geom.PixelsPerPt = 1
				}
				configAlt.Width = geom.Pt(w)
				configAlt.Height = geom.Pt(h)
				configSwap(callbacks)
			}
		}

		if !running {
			// Drop touch events before app started.
			continue
		}

		for _, cb := range callbacks {
			if cb.Touch != nil {
				for _, e := range touches {
					cb.Touch(e)
				}
			}
		}

		for _, cb := range callbacks {
			if cb.Draw != nil {
				cb.Draw()
			}
		}

		gl.Do(func() {
			C.swapBuffers()
			C.processEvents()
		})
	}
}

func sendEvent(ev windowEvent) {
	windowEvents.Lock()
	windowEvents.events = append(windowEvents.events, ev)
	windowEvents.Unlock()
}

//export onResize
func onResize(w, h int) {
	gl.Viewport(0, 0, w, h)
	sendEvent(windowEvent{resize, w, h})
}

func sendTouch(ty event.TouchType, x, y float32) {
	windowEvents.Lock()
	windowEvents.touches = append(windowEvents.touches, event.Touch{
		ID:   0,
		Type: ty,
		Loc: geom.Point{
			X: geom.Pt(x / geom.PixelsPerPt),
			Y: geom.Pt(y / geom.PixelsPerPt),
		},
	})
	windowEvents.Unlock()
}

//export onTouchStart
func onTouchStart(x, y float32) { sendTouch(event.TouchStart, x, y) }

//export onTouchMove
func onTouchMove(x, y float32) { sendTouch(event.TouchMove, x, y) }

//export onTouchEnd
func onTouchEnd(x, y float32) { sendTouch(event.TouchEnd, x, y) }

//export onStop
func onStop() {
	sendEvent(windowEvent{stop, 0, 0})
}
