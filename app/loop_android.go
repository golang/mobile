// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

/*
#cgo android LDFLAGS: -llog -landroid -lEGL -lGLESv2
#include <android/log.h>
#include <android/native_activity.h>
#include <EGL/egl.h>
#include <GLES/gl.h>

// TODO(crawshaw): Test configuration on more devices.
const EGLint RGB_888[] = {
	EGL_RENDERABLE_TYPE, EGL_OPENGL_ES2_BIT,
	EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
	EGL_BLUE_SIZE, 8,
	EGL_GREEN_SIZE, 8,
	EGL_RED_SIZE, 8,
	EGL_DEPTH_SIZE, 16,
	EGL_CONFIG_CAVEAT, EGL_NONE,
	EGL_NONE
};

EGLint windowWidth;
EGLint windowHeight;
EGLDisplay display;
EGLSurface surface;

#define LOG_ERROR(...) __android_log_print(ANDROID_LOG_ERROR, "Go", __VA_ARGS__)

void createEGLWindow(ANativeWindow* window) {
	EGLint numConfigs, format;
	EGLConfig config;
	EGLContext context;

	display = eglGetDisplay(EGL_DEFAULT_DISPLAY);
	if (!eglInitialize(display, 0, 0)) {
		LOG_ERROR("EGL initialize failed");
		return;
	}

	if (!eglChooseConfig(display, RGB_888, &config, 1, &numConfigs)) {
		LOG_ERROR("EGL choose RGB_888 config failed");
		return;
	}
	if (numConfigs <= 0) {
		LOG_ERROR("EGL no config found");
		return;
	}

	eglGetConfigAttrib(display, config, EGL_NATIVE_VISUAL_ID, &format);
	if (ANativeWindow_setBuffersGeometry(window, 0, 0, format) != 0) {
		LOG_ERROR("EGL set buffers geometry failed");
		return;
	}

	surface = eglCreateWindowSurface(display, config, window, NULL);
	if (surface == EGL_NO_SURFACE) {
		LOG_ERROR("EGL create surface failed");
		return;
	}

	const EGLint contextAttribs[] = { EGL_CONTEXT_CLIENT_VERSION, 2, EGL_NONE };
	context = eglCreateContext(display, config, EGL_NO_CONTEXT, contextAttribs);

	if (eglMakeCurrent(display, surface, surface, context) == EGL_FALSE) {
		LOG_ERROR("eglMakeCurrent failed");
		return;
	}

	eglQuerySurface(display, surface, EGL_WIDTH, &windowWidth);
	eglQuerySurface(display, surface, EGL_HEIGHT, &windowHeight);

	glDisable(GL_DEPTH_TEST);
	glViewport(0, 0, windowWidth, windowHeight);
	glClearColor(0, 0, 0, 1);
	glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT);
	eglSwapBuffers(display, surface);

	GLenum err;
	if ((err = glGetError()) != GL_NO_ERROR) {
		LOG_ERROR("GL error in createEGLWindow: 0x%x", err);
	}
}

#undef LOG_ERROR
*/
import "C"

// windowDrawLoop calls Draw at 60 FPS processes input queue events.
func windowDrawLoop(cb Callbacks, window *C.ANativeWindow) {
	C.createEGLWindow(window)

	for {
		select {
		case <-windowDestroyed:
			return
		// TODO(crawshaw): Event processing goes here.
		default:
			cb.Draw()
			C.eglSwapBuffers(C.display, C.surface)
		}
	}
}
