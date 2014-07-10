// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package app hooks into the Android runtime via cgo/JNI.
//
// Typical use of this package is in a shared library loaded by
// a mobile app:
//
//	package main
//
//	import "google.golang.org/mobile/app"
//
//	func main() {
//		app.Run()
//	}
package app

// Run starts the process.
func Run() {
	run()
}
