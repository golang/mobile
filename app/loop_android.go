// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

/*
#include <android/native_activity.h>
#include <android/native_window.h>
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

EGLDisplay display = NULL;
EGLSurface surface = NULL;

char* initEGLDisplay() {
	display = eglGetDisplay(EGL_DEFAULT_DISPLAY);
	if (!eglInitialize(display, 0, 0)) {
		return "EGL initialize failed";
	}
	return NULL;
}

char* createEGLSurface(ANativeWindow* window) {
	char* err;
	EGLint numConfigs, format;
	EGLConfig config;
	EGLContext context;

	if (display == 0) {
		if ((err = initEGLDisplay()) != NULL) {
			return err;
		}
	}

	if (!eglChooseConfig(display, RGB_888, &config, 1, &numConfigs)) {
		return "EGL choose RGB_888 config failed";
	}
	if (numConfigs <= 0) {
		return "EGL no config found";
	}

	eglGetConfigAttrib(display, config, EGL_NATIVE_VISUAL_ID, &format);
	if (ANativeWindow_setBuffersGeometry(window, 0, 0, format) != 0) {
		return "EGL set buffers geometry failed";
	}

	surface = eglCreateWindowSurface(display, config, window, NULL);
	if (surface == EGL_NO_SURFACE) {
		return "EGL create surface failed";
	}

	const EGLint contextAttribs[] = { EGL_CONTEXT_CLIENT_VERSION, 2, EGL_NONE };
	context = eglCreateContext(display, config, EGL_NO_CONTEXT, contextAttribs);

	if (eglMakeCurrent(display, surface, surface, context) == EGL_FALSE) {
		return "eglMakeCurrent failed";
	}
	return NULL;
}

char* destroyEGLSurface() {
	if (!eglDestroySurface(display, surface)) {
		return "EGL destroy surface failed";
	}
	return NULL;
}
*/
import "C"
import (
	"fmt"
	"log"
	"runtime"

	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

func main(f func(App)) {
	// Preserve this OS thread for the GL context created below.
	runtime.LockOSThread()

	donec := make(chan struct{})
	go func() {
		f(app{})
		close(donec)
	}()

	var q *C.AInputQueue

	// Android can send a windowRedrawNeeded event any time, including
	// in the middle of a paint cycle. The redraw event may have changed
	// the size of the screen, so any partial painting is now invalidated.
	// We must also not return to Android (via sending on windowRedrawDone)
	// until a complete paint with the new configuration is complete.
	//
	// When a windowRedrawNeeded request comes in, we increment redrawGen
	// (Gen is short for generation number), and do not make a paint cycle
	// visible on <-endPaint unless Generation agrees. If possible,
	// windowRedrawDone is signalled, allowing onNativeWindowRedrawNeeded
	// to return.
	var redrawGen uint32

	for {
		if q != nil {
			processEvents(q)
		}
		select {
		case <-windowCreated:
		case q = <-inputQueue:
		case <-donec:
			return
		case cfg := <-windowConfigChange:
			// TODO save orientation
			pixelsPerPt = cfg.pixelsPerPt
		case w := <-windowRedrawNeeded:
			newWindow := C.surface == nil
			if newWindow {
				if errStr := C.createEGLSurface(w); errStr != nil {
					log.Printf("app: %s (%s)", C.GoString(errStr), eglGetError())
					return
				}
			}
			sendLifecycle(lifecycle.StageFocused)
			widthPx := int(C.ANativeWindow_getWidth(w))
			heightPx := int(C.ANativeWindow_getHeight(w))
			eventsIn <- config.Event{
				WidthPx:     widthPx,
				HeightPx:    heightPx,
				WidthPt:     geom.Pt(float32(widthPx) / pixelsPerPt),
				HeightPt:    geom.Pt(float32(heightPx) / pixelsPerPt),
				PixelsPerPt: pixelsPerPt,
			}
			redrawGen++
			if newWindow {
				// New window, begin paint loop.
				eventsIn <- paint.Event{redrawGen}
			}
		case <-windowDestroyed:
			if C.surface != nil {
				if errStr := C.destroyEGLSurface(); errStr != nil {
					log.Printf("app: %s (%s)", C.GoString(errStr), eglGetError())
					return
				}
			}
			C.surface = nil
			sendLifecycle(lifecycle.StageAlive)
		case <-gl.WorkAvailable:
			gl.DoWork()
		case p := <-endPaint:
			if p.Generation != redrawGen {
				continue
			}
			if C.surface != nil {
				// eglSwapBuffers blocks until vsync.
				if C.eglSwapBuffers(C.display, C.surface) == C.EGL_FALSE {
					log.Printf("app: failed to swap buffers (%s)", eglGetError())
				}
			}
			select {
			case windowRedrawDone <- struct{}{}:
			default:
			}
			if C.surface != nil {
				redrawGen++
				eventsIn <- paint.Event{redrawGen}
			}
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
		upDownType := touch.TypeMove
		switch C.AMotionEvent_getAction(e) & C.AMOTION_EVENT_ACTION_MASK {
		case C.AMOTION_EVENT_ACTION_DOWN, C.AMOTION_EVENT_ACTION_POINTER_DOWN:
			upDownType = touch.TypeBegin
		case C.AMOTION_EVENT_ACTION_UP, C.AMOTION_EVENT_ACTION_POINTER_UP:
			upDownType = touch.TypeEnd
		}

		for i, n := C.size_t(0), C.AMotionEvent_getPointerCount(e); i < n; i++ {
			t := touch.TypeMove
			if i == upDownIndex {
				t = upDownType
			}
			eventsIn <- touch.Event{
				Sequence: touch.Sequence(C.AMotionEvent_getPointerId(e, i)),
				Type:     t,
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

func eglGetError() string {
	switch errNum := C.eglGetError(); errNum {
	case C.EGL_SUCCESS:
		return "EGL_SUCCESS"
	case C.EGL_NOT_INITIALIZED:
		return "EGL_NOT_INITIALIZED"
	case C.EGL_BAD_ACCESS:
		return "EGL_BAD_ACCESS"
	case C.EGL_BAD_ALLOC:
		return "EGL_BAD_ALLOC"
	case C.EGL_BAD_ATTRIBUTE:
		return "EGL_BAD_ATTRIBUTE"
	case C.EGL_BAD_CONTEXT:
		return "EGL_BAD_CONTEXT"
	case C.EGL_BAD_CONFIG:
		return "EGL_BAD_CONFIG"
	case C.EGL_BAD_CURRENT_SURFACE:
		return "EGL_BAD_CURRENT_SURFACE"
	case C.EGL_BAD_DISPLAY:
		return "EGL_BAD_DISPLAY"
	case C.EGL_BAD_SURFACE:
		return "EGL_BAD_SURFACE"
	case C.EGL_BAD_MATCH:
		return "EGL_BAD_MATCH"
	case C.EGL_BAD_PARAMETER:
		return "EGL_BAD_PARAMETER"
	case C.EGL_BAD_NATIVE_PIXMAP:
		return "EGL_BAD_NATIVE_PIXMAP"
	case C.EGL_BAD_NATIVE_WINDOW:
		return "EGL_BAD_NATIVE_WINDOW"
	case C.EGL_CONTEXT_LOST:
		return "EGL_CONTEXT_LOST"
	default:
		return fmt.Sprintf("Unknown EGL err: %d", errNum)
	}
}
