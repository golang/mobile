// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

// Run starts the process.
func Run() {
	run()
}

// Draw is called by the render loop to draw the screen.
//
// Drawing is done into a framebuffer, which is then swapped onto the
// screen when Draw returns. It is called 60 times a second.
var Draw func()

/*
TODO(crawshaw): Implement.
var Start func()
var Stop func()
var Resume func()
var Pause func()
*/
