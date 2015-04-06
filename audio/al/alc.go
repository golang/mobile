// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package al

import "unsafe"

// Device represents an audio device.
type Device struct {
	ptr unsafe.Pointer
}

// Error returns the last known error from the current device.
func (d *Device) Error() int32 {
	return alcGetError(d.ptr)
}

// Context represents a context created in the OpenAL layer. A valid current
// context is required to run OpenAL functions.
// The returned context will be available process-wide if it's made the
// current by calling MakeContextCurrent.
type Context struct {
	ptr unsafe.Pointer
}

// Open opens a new device in the OpenAL layer.
func Open(name string) *Device {
	ptr := alcOpenDevice(name)
	if ptr == nil {
		return nil
	}
	return &Device{ptr: ptr}
}

// Close closes the device.
func (d *Device) Close() bool {
	return alcCloseDevice(d.ptr)
}

// CreateContext creates a new context.
func (d *Device) CreateContext(attrs []int32) *Context {
	ptr := alcCreateContext(d.ptr, attrs)
	if ptr == nil {
		return nil
	}
	return &Context{ptr: ptr}
}

// MakeContextCurrent makes a context current. The context available
// process-wide, you don't need to lock the current OS thread to
// access the current context.
func (c *Context) MakeContextCurrent() bool {
	return alcMakeContextCurrent(c.ptr)
}

// Destroy destroys the current context and frees the related resources.
func (c *Context) Destroy() {
	alcDestroyContext(c.ptr)
}
