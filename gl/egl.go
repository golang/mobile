// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package gl

// #include <EGL/egl.h>
// #cgo LDFLAGS: -lEGL
import "C"

func init() {
	display := C.eglGetDisplay(C.EGL_DEFAULT_DISPLAY)

	C.eglInitialize(display, nil, nil)

	attributes := []C.EGLint{
		C.EGL_RED_SIZE, 1,
		C.EGL_GREEN_SIZE, 1,
		C.EGL_BLUE_SIZE, 1,
		C.EGL_NONE,
	}

	var (
		config C.EGLConfig
		count  C.EGLint
	)
	C.eglChooseConfig(display, &attributes[0], &config, 1, &count)

	C.eglBindAPI(C.EGL_OPENGL_ES_API)

	attributes = []C.EGLint{
		C.EGL_CONTEXT_CLIENT_VERSION, 3,
		C.EGL_NONE,
	}

	context := C.eglCreateContext(display, config, C.EGLContext(C.EGL_NO_CONTEXT), &attributes[0])
	if context == nil {
		version = "GL_ES_2_0"

		return
	}

	C.eglDestroyContext(display, context)

	version = "GL_ES_3_0"
}
