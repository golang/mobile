// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

/*
#include <android/log.h>
#include <android/native_activity.h>
#include <android/input.h>
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

void querySurfaceWidthAndHeight() {
	eglQuerySurface(display, surface, EGL_WIDTH, &windowWidth);
	eglQuerySurface(display, surface, EGL_HEIGHT, &windowHeight);
}

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

	querySurfaceWidthAndHeight();
}

#undef LOG_ERROR
*/
import "C"
import (
	"log"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var firstWindowDraw = true

func windowDraw(w *C.ANativeWindow, queue *C.AInputQueue, donec chan struct{}) (done bool) {
	C.createEGLWindow(w)

	// TODO: is this needed if we also have the "case <-windowRedrawNeeded:" below??
	sendLifecycle(event.LifecycleStageFocused)
	eventsIn <- event.Config{
		Width:       geom.Pt(float32(C.windowWidth) / pixelsPerPt),
		Height:      geom.Pt(float32(C.windowHeight) / pixelsPerPt),
		PixelsPerPt: pixelsPerPt,
	}
	if firstWindowDraw {
		firstWindowDraw = false
		// TODO: be more principled about when to send a draw event.
		eventsIn <- event.Draw{}
	}

	for {
		processEvents(queue)
		select {
		case <-donec:
			return true
		case <-windowRedrawNeeded:
			// Re-query the width and height.
			C.querySurfaceWidthAndHeight()
			sendLifecycle(event.LifecycleStageFocused)
			eventsIn <- event.Config{
				Width:       geom.Pt(float32(C.windowWidth) / pixelsPerPt),
				Height:      geom.Pt(float32(C.windowHeight) / pixelsPerPt),
				PixelsPerPt: pixelsPerPt,
			}
		case <-windowDestroyed:
			sendLifecycle(event.LifecycleStageAlive)
			return false
		case <-gl.WorkAvailable:
			gl.DoWork()
		case <-endDraw:
			// eglSwapBuffers blocks until vsync.
			C.eglSwapBuffers(C.display, C.surface)
			eventsIn <- event.Draw{}
		}
	}
}

func processEvents(queue *C.AInputQueue) {
	var event *C.AInputEvent
	for C.AInputQueue_getEvent(queue, &event) >= 0 {
		if C.AInputQueue_preDispatchEvent(queue, event) != 0 {
			continue
		}
		processEvent(event)
		C.AInputQueue_finishEvent(queue, event, 0)
	}
}

func processEvent(e *C.AInputEvent) {
	switch C.AInputEvent_getType(e) {
	case C.AINPUT_EVENT_TYPE_KEY:
		log.Printf("TODO input event: key")
	case C.AINPUT_EVENT_TYPE_MOTION:
		// At most one of the events in this batch is an up or down event; get its index and change.
		upDownIndex := C.size_t(C.AMotionEvent_getAction(e)&C.AMOTION_EVENT_ACTION_POINTER_INDEX_MASK) >> C.AMOTION_EVENT_ACTION_POINTER_INDEX_SHIFT
		upDownChange := event.ChangeNone
		switch C.AMotionEvent_getAction(e) & C.AMOTION_EVENT_ACTION_MASK {
		case C.AMOTION_EVENT_ACTION_DOWN, C.AMOTION_EVENT_ACTION_POINTER_DOWN:
			upDownChange = event.ChangeOn
		case C.AMOTION_EVENT_ACTION_UP, C.AMOTION_EVENT_ACTION_POINTER_UP:
			upDownChange = event.ChangeOff
		}

		for i, n := C.size_t(0), C.AMotionEvent_getPointerCount(e); i < n; i++ {
			change := event.ChangeNone
			if i == upDownIndex {
				change = upDownChange
			}
			eventsIn <- event.Touch{
				ID:     event.TouchSequenceID(C.AMotionEvent_getPointerId(e, i)),
				Change: change,
				Loc: geom.Point{
					X: geom.Pt(float32(C.AMotionEvent_getX(e, i)) / pixelsPerPt),
					Y: geom.Pt(float32(C.AMotionEvent_getY(e, i)) / pixelsPerPt),
				},
			}
		}
	default:
		log.Printf("unknown input event, type=%d", C.AInputEvent_getType(e))
	}
}
