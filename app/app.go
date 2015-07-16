// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux darwin

package app

import (
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
	_ "golang.org/x/mobile/internal/mobileinit"
)

// Main is called by the main.main function to run the mobile application.
//
// It calls f on the App, in a separate goroutine, as some OS-specific
// libraries require being on 'the main thread'.
func Main(f func(App)) {
	main(f)
}

// App is how a GUI mobile application interacts with the OS.
type App interface {
	// Events returns the events channel. It carries events from the system to
	// the app. The type of such events include:
	//  - config.Event
	//  - lifecycle.Event
	//  - paint.Event
	//  - touch.Event
	// from the golang.org/x/mobile/event/etc packages. Other packages may
	// define other event types that are carried on this channel.
	Events() <-chan interface{}

	// Send sends an event on the events channel. It does not block.
	Send(event interface{})

	// EndPaint flushes any pending OpenGL commands or buffers to the screen.
	EndPaint()
}

var (
	lifecycleStage = lifecycle.StageDead
	pixelsPerPt    = float32(1)

	eventsOut = make(chan interface{})
	eventsIn  = pump(eventsOut)
	endPaint  = make(chan struct{}, 1)
)

func sendLifecycle(to lifecycle.Stage) {
	if lifecycleStage == to {
		return
	}
	eventsIn <- lifecycle.Event{
		From: lifecycleStage,
		To:   to,
	}
	lifecycleStage = to
}

type app struct{}

func (app) Events() <-chan interface{} {
	return eventsOut
}

func (app) Send(event interface{}) {
	eventsIn <- event
}

func (app) EndPaint() {
	// gl.Flush is a lightweight (on modern GL drivers) blocking call
	// that ensures all GL functions pending in the gl package have
	// been passed onto the GL driver before the app package attempts
	// to swap the screen buffer.
	//
	// This enforces that the final receive (for this paint cycle) on
	// gl.WorkAvailable happens before the send on endPaint.
	gl.Flush()

	select {
	case endPaint <- struct{}{}:
	default:
	}
}

var filters []func(interface{}) interface{}

// Filter calls each registered event filter function in sequence.
func Filter(event interface{}) interface{} {
	for _, f := range filters {
		event = f(event)
	}
	return event
}

// RegisterFilter registers a event filter function to be called by Filter. The
// function can return a different event, or return nil to consume the event,
// but the function can also return its argument unchanged, where its purpose
// is to trigger a side effect rather than modify the event.
//
// RegisterFilter should only be called from init functions.
func RegisterFilter(f func(interface{}) interface{}) {
	filters = append(filters, f)
}

type stopPumping struct{}

// pump returns a channel src such that sending on src will eventually send on
// dst, in order, but that src will always be ready to send/receive soon, even
// if dst currently isn't. It is effectively an infinitely buffered channel.
//
// In particular, goroutine A sending on src will not deadlock even if goroutine
// B that's responsible for receiving on dst is currently blocked trying to
// send to A on a separate channel.
//
// Send a stopPumping on the src channel to close the dst channel after all queued
// events are sent on dst. After that, other goroutines can still send to src,
// so that such sends won't block forever, but such events will be ignored.
func pump(dst chan interface{}) (src chan interface{}) {
	src = make(chan interface{})
	go func() {
		// initialSize is the initial size of the circular buffer. It must be a
		// power of 2.
		const initialSize = 16
		i, j, buf, mask := 0, 0, make([]interface{}, initialSize), initialSize-1

		maybeSrc := src
		for {
			maybeDst := dst
			if i == j {
				maybeDst = nil
			}
			if maybeDst == nil && maybeSrc == nil {
				break
			}

			select {
			case maybeDst <- buf[i&mask]:
				buf[i&mask] = nil
				i++

			case e := <-maybeSrc:
				if _, ok := e.(stopPumping); ok {
					maybeSrc = nil
					continue
				}

				// Allocate a bigger buffer if necessary.
				if i+len(buf) == j {
					b := make([]interface{}, 2*len(buf))
					n := copy(b, buf[j&mask:])
					copy(b[n:], buf[:j&mask])
					i, j = 0, len(buf)
					buf, mask = b, len(b)-1
				}

				buf[j&mask] = e
				j++
			}
		}

		close(dst)
		// Block forever.
		for range src {
		}
	}()
	return src
}

// Run starts the mobile application.
//
// It must be called directly from the main function and will block until the
// application exits.
//
// Deprecated: call Main directly instead.
func Run(cb Callbacks) {
	Main(func(a App) {
		var c config.Event
		for e := range a.Events() {
			switch e := Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if cb.Start != nil {
						cb.Start()
					}
				case lifecycle.CrossOff:
					if cb.Stop != nil {
						cb.Stop()
					}
				}
			case config.Event:
				if cb.Config != nil {
					cb.Config(e, c)
				}
				c = e
			case paint.Event:
				if cb.Draw != nil {
					cb.Draw(c)
				}
				a.EndPaint()
			case touch.Event:
				if cb.Touch != nil {
					cb.Touch(e, c)
				}
			}
		}
	})
}

// Callbacks is the set of functions called by the app.
//
// Deprecated: call Main directly instead.
type Callbacks struct {
	// Start is called when the app enters the foreground.
	// The app will start receiving Draw and Touch calls.
	//
	// If the app is responsible for the screen (that is, it is an
	// all-Go app), then Window geometry will be configured and an
	// OpenGL context will be available during Start.
	//
	// If this is a library, Start will be called before the
	// app is told that Go has finished initialization.
	//
	// Start is an equivalent lifecycle state to onStart() on
	// Android and applicationDidBecomeActive on iOS.
	Start func()

	// Stop is called shortly before a program is suspended.
	//
	// When Stop is received, the app is no longer visible and is not
	// receiving events. It should:
	//
	//	- Save any state the user expects saved (for example text).
	//	- Release all resources that are not needed.
	//
	// Execution time in the stop state is limited, and the limit is
	// enforced by the operating system. Stop as quickly as you can.
	//
	// An app that is stopped may be started again. For example, the user
	// opens Recent Apps and switches to your app. A stopped app may also
	// be terminated by the operating system with no further warning.
	//
	// Stop is equivalent to onStop() on Android and
	// applicationDidEnterBackground on iOS.
	Stop func()

	// Draw is called by the render loop to draw the screen.
	//
	// Drawing is done into a framebuffer, which is then swapped onto the
	// screen when Draw returns. It is called 60 times a second.
	Draw func(config.Event)

	// Touch is called by the app when a touch event occurs.
	Touch func(touch.Event, config.Event)

	// Config is called by the app when configuration has changed.
	Config func(new, old config.Event)
}

// TODO: do this for all build targets, not just linux (x11 and Android)? If
// so, should package gl instead of this package call RegisterFilter??
//
// TODO: does Android need this?? It seems to work without it (Nexus 7,
// KitKat). If only x11 needs this, should we move this to x11.go??
func registerGLViewportFilter() {
	RegisterFilter(func(e interface{}) interface{} {
		if e, ok := e.(config.Event); ok {
			w := int(e.PixelsPerPt * float32(e.Width))
			h := int(e.PixelsPerPt * float32(e.Height))
			gl.Viewport(0, 0, w, h)
		}
		return e
	})
}
