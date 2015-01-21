// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

// Package alc provides OpenAL's ALC (Audio Library Context) bindings for Go.
package alc

/*
#cgo darwin   CFLAGS:  -DGOOS_darwin
#cgo darwin   LDFLAGS: -framework OpenAL

#ifdef GOOS_darwin
#include <stdlib.h>
#include <OpenAL/alc.h>
#endif
*/
import "C"
import "unsafe"

// Error returns one of these values.
const (
	InvalidDevice  = 0xA001
	InvalidContext = 0xA002
	InvalidEnum    = 0xA003
	InvalidValue   = 0xA004
	OutOfMemory    = 0xA005
)

// Device represents an audio device.
type Device struct {
	d *C.ALCdevice
}

// Error returns the last known error from the current device.
func (d *Device) Error() int32 {
	return int32(C.alcGetError(d.d))
}

// Context represents a context created in the OpenAL layer. A valid current
// context is required to run OpenAL functions.
// The returned context will be available process-wide if it's made the
// current by calling MakeContextCurrent.
type Context struct {
	c *C.ALCcontext
}

// Open opens a new device in the OpenAL layer.
func Open(name string) *Device {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return &Device{d: C.alcOpenDevice((*C.ALCchar)(unsafe.Pointer(n)))}
}

// Close closes the device.
func (d *Device) Close() bool {
	return C.alcCloseDevice(d.d) == 1
}

// CreateContext creates a new context.
func (d *Device) CreateContext(attrs []int32) *Context {
	// TODO(jbd): Handle attributes.
	c := C.alcCreateContext(d.d, nil)
	return &Context{c: c}
}

// MakeContextCurrent makes a context current process-wide.
func MakeContextCurrent(c *Context) bool {
	return C.alcMakeContextCurrent(c.c) == 1
}
