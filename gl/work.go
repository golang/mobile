// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux cgo

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

// TODO: return uintptr_t instead of taking ret.
void process(struct fnargs* cargs, int count, uintptr_t* ret) {
	int i;
	for (i = 0; i < count; i++) {
		*ret = processFn(&cargs[i]);
	}
}
*/
import "C"

const workbufLen = 10

type context struct {
	cptr uintptr

	workAvailable chan struct{}

	// work is a queue of calls to execute.
	work chan call

	// retvalue is sent a return value when blocking calls complete.
	// It is safe to use a global unbuffered channel here as calls
	// cannot currently be made concurrently.
	//
	// TODO: the comment above about concurrent calls isn't actually true: package
	// app calls package gl, but it has to do so in a separate goroutine, which
	// means that its gl calls (which may be blocking) can race with other gl calls
	// in the main program. We should make it safe to issue blocking gl calls
	// concurrently, or get the gl calls out of package app, or both.
	retvalue chan C.uintptr_t

	cargs [workbufLen]C.struct_fnargs
	ret   C.uintptr_t
}

func (ctx *context) WorkAvailable() <-chan struct{} { return ctx.workAvailable }

// NewContext creates a cgo OpenGL context.
//
// See the Worker interface for more details on how it is used.
func NewContext() (Context, Worker) {
	glctx := &context{
		workAvailable: make(chan struct{}, 1),
		work:          make(chan call, workbufLen),
		retvalue:      make(chan C.uintptr_t),
	}
	return glctx, glctx
}

type call struct {
	args     C.struct_fnargs
	blocking bool
}

func (ctx *context) enqueue(c call) C.uintptr_t {
	ctx.work <- c

	select {
	case ctx.workAvailable <- struct{}{}:
	default:
	}

	if c.blocking {
		return <-ctx.retvalue
	}
	return 0
}

func (ctx *context) DoWork() {
	queue := make([]call, 0, len(ctx.work))
	for {
		// Wait until at least one piece of work is ready.
		// Accumulate work until a piece is marked as blocking.
		select {
		case w := <-ctx.work:
			queue = append(queue, w)
		default:
			return
		}
		blocking := queue[len(queue)-1].blocking
	enqueue:
		for len(queue) < cap(queue) && !blocking {
			select {
			case w := <-ctx.work:
				queue = append(queue, w)
				blocking = queue[len(queue)-1].blocking
			default:
				break enqueue
			}
		}

		// Process the queued GL functions.
		for i, q := range queue {
			ctx.cargs[i] = q.args
		}
		C.process(&ctx.cargs[0], C.int(len(queue)), &ctx.ret)

		// Cleanup and signal.
		queue = queue[:0]
		if blocking {
			ctx.retvalue <- ctx.ret
		}
	}
}

func glBoolean(b bool) C.uintptr_t {
	if b {
		return TRUE
	}
	return FALSE
}
