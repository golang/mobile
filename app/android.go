// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

/*
Android Apps are built with -buildmode=c-shared. They are loaded by a
running Java process.

Before any entry point is reached, a global constructor initializes the
Go runtime, calling all Go init functions. All cgo calls will block
until this is complete. Next JNI_OnLoad is called. When that is
complete, one of two entry points is called.

All-Go apps built using NativeActivity enter at ANativeActivity_onCreate.

Go libraries (for example, those built with gomobile bind) do not use
the app package initialization.
*/

package app

/*
// The C libraries listed here but not explicitly used in this file are used by
// other *_android.go files. There should be only one #cgo declaration.
#cgo LDFLAGS: -landroid -llog -lEGL -lGLESv2

#include <android/configuration.h>
#include <android/native_activity.h>
#include <time.h>

#include <jni.h>
#include <pthread.h>
#include <stdlib.h>

// current_vm is stored to initialize other cgo packages.
//
// As all the Go packages in a program form a single shared library,
// there can only be one JNI_OnLoad function for iniitialization. In
// OpenJDK there is JNI_GetCreatedJavaVMs, but this is not available
// on android.
JavaVM* current_vm;

// current_ctx is Android's android.context.Context. May be NULL.
jobject current_ctx;

jclass current_ctx_clazz;

jclass app_find_class(JNIEnv* env, const char* name);
*/
import "C"
import (
	"log"
	"os"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/mobile/app/internal/callfn"
)

//export callMain
func callMain(mainPC uintptr) {
	for _, name := range []string{"TMPDIR", "PATH", "LD_LIBRARY_PATH"} {
		n := C.CString(name)
		os.Setenv(name, C.GoString(C.getenv(n)))
		C.free(unsafe.Pointer(n))
	}

	// Set timezone.
	//
	// Note that Android zoneinfo is stored in /system/usr/share/zoneinfo,
	// but it is in some kind of packed TZiff file that we do not support
	// yet. As a stopgap, we build a fixed zone using the tm_zone name.
	var curtime C.time_t
	var curtm C.struct_tm
	C.time(&curtime)
	C.localtime_r(&curtime, &curtm)
	tzOffset := int(curtm.tm_gmtoff)
	tz := C.GoString(curtm.tm_zone)
	time.Local = time.FixedZone(tz, tzOffset)

	go callfn.CallFn(mainPC)
}

//export onCreate
func onCreate(activity *C.ANativeActivity) {
	config := C.AConfiguration_new()
	C.AConfiguration_fromAssetManager(config, activity.assetManager)
	density := C.AConfiguration_getDensity(config)
	C.AConfiguration_delete(config)

	var dpi int
	switch density {
	case C.ACONFIGURATION_DENSITY_DEFAULT:
		dpi = 160
	case C.ACONFIGURATION_DENSITY_LOW,
		C.ACONFIGURATION_DENSITY_MEDIUM,
		213, // C.ACONFIGURATION_DENSITY_TV
		C.ACONFIGURATION_DENSITY_HIGH,
		320, // ACONFIGURATION_DENSITY_XHIGH
		480, // ACONFIGURATION_DENSITY_XXHIGH
		640: // ACONFIGURATION_DENSITY_XXXHIGH
		dpi = int(density)
	case C.ACONFIGURATION_DENSITY_NONE:
		log.Print("android device reports no screen density")
		dpi = 72
	default:
		log.Print("android device reports unknown density: %d", density)
		dpi = int(density) // This is a guess.
	}

	pixelsPerPt = float32(dpi) / 72
}

//export onStart
func onStart(activity *C.ANativeActivity) {
}

//export onResume
func onResume(activity *C.ANativeActivity) {
}

//export onSaveInstanceState
func onSaveInstanceState(activity *C.ANativeActivity, outSize *C.size_t) unsafe.Pointer {
	return nil
}

//export onPause
func onPause(activity *C.ANativeActivity) {
}

//export onStop
func onStop(activity *C.ANativeActivity) {
}

//export onDestroy
func onDestroy(activity *C.ANativeActivity) {
}

//export onWindowFocusChanged
func onWindowFocusChanged(activity *C.ANativeActivity, hasFocus int) {
}

//export onNativeWindowCreated
func onNativeWindowCreated(activity *C.ANativeActivity, w *C.ANativeWindow) {
	windowCreated <- w
}

//export onNativeWindowResized
func onNativeWindowResized(activity *C.ANativeActivity, window *C.ANativeWindow) {
}

//export onNativeWindowRedrawNeeded
func onNativeWindowRedrawNeeded(activity *C.ANativeActivity, window *C.ANativeWindow) {
	windowRedrawNeeded <- window
}

//export onNativeWindowDestroyed
func onNativeWindowDestroyed(activity *C.ANativeActivity, window *C.ANativeWindow) {
	windowDestroyed <- true
}

var queue *C.AInputQueue

//export onInputQueueCreated
func onInputQueueCreated(activity *C.ANativeActivity, q *C.AInputQueue) {
	queue = q
}

//export onInputQueueDestroyed
func onInputQueueDestroyed(activity *C.ANativeActivity, q *C.AInputQueue) {
	queue = nil
}

//export onContentRectChanged
func onContentRectChanged(activity *C.ANativeActivity, rect *C.ARect) {
}

//export onConfigurationChanged
func onConfigurationChanged(activity *C.ANativeActivity) {
}

//export onLowMemory
func onLowMemory(activity *C.ANativeActivity) {
}

// Context holds global OS-specific context.
//
// Its extra methods are deliberately difficult to access because they must be
// used with care. Their use implies the use of cgo, which probably requires
// you understand the initialization process in the app package. Also care must
// be taken to write both Android, iOS, and desktop-testing versions to
// maintain portability.
type Context struct{}

// AndroidContext returns a jobject for the app android.context.Context.
func (Context) AndroidContext() unsafe.Pointer {
	return unsafe.Pointer(C.current_ctx)
}

// JavaVM returns a JNI *JavaVM.
func (Context) JavaVM() unsafe.Pointer {
	return unsafe.Pointer(C.current_vm)
}

var (
	windowDestroyed    = make(chan bool)
	windowCreated      = make(chan *C.ANativeWindow)
	windowRedrawNeeded = make(chan *C.ANativeWindow)
)

func init() {
	registerGLViewportFilter()
}

func main(f func(App)) {
	// Preserve this OS thread for the GL context created in windowDraw.
	runtime.LockOSThread()

	donec := make(chan struct{})
	go func() {
		f(app{})
		close(donec)
	}()

	for w := range windowCreated {
		if windowDraw(w, queue, donec) {
			return
		}
	}
	panic("unreachable")
}
