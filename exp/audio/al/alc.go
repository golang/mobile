// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package al

import (
	"errors"
	"sync"
	"unsafe"
)

var (
	mu sync.Mutex // mu protects device.

	// device is the currently open audio device or nil.
	device unsafe.Pointer
)

// DeviceError returns the last known error from the current device.
func DeviceError() int32 {
	return alcGetError(device)
}

// Context represents an OpenAL context. The listener always
// operates on the current context. In order to make a context
// current, use MakeContextCurrent.
type Context struct {
	ptr unsafe.Pointer
}

// TODO(jbd): Investigate the cases where multiple audio output
// devices might be needed.

// OpenDevice opens the default audio device.
// Calls to OpenDevice are safe for concurrent use.
func OpenDevice() error {
	mu.Lock()
	defer mu.Unlock()

	// already opened
	if device != nil {
		return nil
	}

	ptr := alcOpenDevice("")
	if ptr == nil {
		return errors.New("al: cannot open the default audio device")
	}
	device = ptr
	return nil
}

// CreateContext creates a new context.
func CreateContext() (Context, error) {
	ctx := alcCreateContext(device, nil)
	if ctx == nil {
		return Context{}, errors.New("al: cannot create a new context")
	}
	return Context{ptr: ctx}, nil
}

// Destroy destroys the context and frees the underlying resources.
// The current context cannot be destroyed. Use MakeContextCurrent
// to make a new context current or use nil.
func (c Context) Destroy() {
	alcDestroyContext(c.ptr)
}

// MakeContextCurrent makes the given context current. Calls to
// MakeContextCurrent can accept a zero value Context to allow
// a program not to have an active current context.
func MakeContextCurrent(c Context) bool {
	return alcMakeContextCurrent(c.ptr)
}

// CloseDevice closes the device and frees related resources.
// Calls to CloseDevice are safe for concurrent use.
func CloseDevice() {
	mu.Lock()
	defer mu.Unlock()

	if device != nil {
		alcCloseDevice(device)
		device = nil
	}
}
