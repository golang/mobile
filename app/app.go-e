// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"io"

	"golang.org/x/mobile/event"
)

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

// Open opens a named asset.
//
// On Android, assets are accessed via android.content.res.AssetManager.
// These files are stored in the assets/ directory of the app. Any raw asset
// can be accessed by its direct relative name. For example assets/img.png
// can be opened with Open("img.png").
//
// On iOS an asset is a resource stored in the application bundle.
// Resources can be loaded using the same relative paths.
//
// For consistency when debugging on a desktop, assets are read from a
// directoy named assets under the current working directory.
func Open(name string) (ReadSeekCloser, error) {
	return openAsset(name)
}

// ReadSeekCloser is an io.ReadSeeker and io.Closer.
type ReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
}
