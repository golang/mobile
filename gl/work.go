// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl

/*
#cgo darwin,amd64  LDFLAGS: -framework OpenGL
#cgo darwin,arm    LDFLAGS: -framework OpenGLES
#cgo darwin,arm64  LDFLAGS: -framework OpenGLES
#cgo linux         LDFLAGS: -lGLESv2

#cgo darwin,amd64  CFLAGS: -Dos_osx
#cgo darwin,arm    CFLAGS: -Dos_ios
#cgo darwin,arm64  CFLAGS: -Dos_ios
#cgo linux         CFLAGS: -Dos_linux

#include <stdint.h>
#include "work.h"

struct fnargs cargs[10];
uintptr_t ret;

void process(int count) {
	int i;
	for (i = 0; i < count; i++) {
		processFn(&cargs[i]);
	}
}
*/
import "C"
import "runtime"

// work is a queue of calls to execute.
var work = make(chan call, 10)

// retvalue is sent a return value when blocking calls complete.
// It is safe to use a global unbuffered channel here as calls
// cannot currently be made concurrently.
var retvalue = make(chan C.uintptr_t)

type call struct {
	blocking bool
	fn       func()
	args     C.struct_fnargs
}

// Do calls fn on the OS thread with the GL context.
func Do(fn func()) {
	work <- call{
		fn:       fn,
		blocking: true,
	}
	<-retvalue
}

// Stop stops the current GL processing.
func Stop() {
	var call call
	call.blocking = true
	call.args.fn = C.glfnStop
	work <- call
	<-retvalue
}

// Start executes GL functions on a fixed OS thread, starting with initCtx.
// It blocks until Stop is called. Typical use:
//
//	go gl.Start(func() {
//		// establish a GL context, using for example, EGL.
//	})
//
//	// long running GL calls from any goroutine
//
//	gl.Stop()
func Start(initCtx func()) {
	runtime.LockOSThread()
	initCtx()

	queue := make([]call, 0, len(work))
	for {
		// Wait until at least one piece of work is ready.
		// Accumulate work until a piece is marked as blocking.
		select {
		case w := <-work:
			queue = append(queue, w)
		}
		blocking := queue[len(queue)-1].blocking
	enqueue:
		for len(queue) < cap(queue) && !blocking {
			select {
			case w := <-work:
				queue = append(queue, w)
				blocking = queue[len(queue)-1].blocking
			default:
				break enqueue
			}
		}

		// Process the queued GL functions.
		fn := queue[len(queue)-1].fn
		stop := queue[len(queue)-1].args.fn == C.glfnStop
		if fn != nil || stop {
			queue = queue[:len(queue)-1]
		}
		for i := range queue {
			C.cargs[i] = queue[i].args
		}
		C.process(C.int(len(queue)))
		if fn != nil {
			fn()
		}

		// Cleanup and signal.
		queue = queue[:0]
		if blocking {
			retvalue <- C.ret
		}
		if stop {
			return
		}
	}
}

func glBoolean(b bool) C.uintptr_t {
	if b {
		return TRUE
	}
	return FALSE
}
