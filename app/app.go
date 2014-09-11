// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import "code.google.com/p/go.mobile/event"

// Run starts the app.
//
// It must be called directly from from the main function and will
// block until the app exits.
func Run(cb Callbacks) {
	run(cb)
}

// Callbacks is the set of functions called by the app.
type Callbacks struct {
	// Draw is called by the render loop to draw the screen.
	//
	// Drawing is done into a framebuffer, which is then swapped onto the
	// screen when Draw returns. It is called 60 times a second.
	Draw func()

	// Touch is called by the app when a touch event occurs.
	Touch func(event.Touch)
}

/*
TODO(crawshaw): Implement.
var Start func()
var Stop func()
var Resume func()
var Pause func()
*/
