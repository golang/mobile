// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package paint defines an event for the app being ready to paint.
//
// See the golang.org/x/mobile/event package for details on the event model.
package paint // import "golang.org/x/mobile/event/paint"

// TODO: rename App.EndDraw to App.EndPaint.
//
// This is "package paint", not "package draw", to avoid conflicting with the
// standard library's "image/draw" package.

// Event indicates that the app is ready to paint the next frame of the GUI. A
// frame is completed by calling the App's EndDraw method.
type Event struct{}
