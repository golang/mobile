// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

// Go runtime entry point for apps running on android.
// Sets up everything the runtime needs and exposes
// the entry point to JNI.

package app

/*
#cgo LDFLAGS: -llog -landroid
#include <android/log.h>
#include <android/configuration.h>
#include <android/native_activity.h>
#include <pthread.h>
#include <jni.h>

pthread_cond_t go_started_cond;
pthread_mutex_t go_started_mu;
int go_started;

// current_vm is stored to initialize other cgo packages.
//
// As all the Go packages in a program form a single shared library,
// there can only be one JNI_OnLoad function for iniitialization. In
// OpenJDK there is JNI_GetCreatedJavaVMs, but this is not available
// on android.
JavaVM* current_vm;
*/
import "C"
import (
	"log"
	"runtime"
	"unsafe"

	"code.google.com/p/go.mobile/geom"
)

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

	geom.Scale = float32(dpi) / 72
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
func onInputQueueDestroyed(activity *C.ANativeActivity, queue *C.AInputQueue) {
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

// JavaInit is an initialization function registered by the package
// code.google.com/p/go.mobile/bind/java. It gives the Java language
// bindings access to the JNI *JavaVM object.
var JavaInit func(javaVM uintptr)

var (
	windowDestroyed = make(chan bool)
	windowCreated   = make(chan *C.ANativeWindow)
)

func run(cb Callbacks) {
	// We want to keep the event loop on a consistent OS thread.
	runtime.LockOSThread()

	ctag := C.CString("Go")
	cstr := C.CString("app.Run")
	C.__android_log_write(C.ANDROID_LOG_INFO, ctag, cstr)
	C.free(unsafe.Pointer(ctag))
	C.free(unsafe.Pointer(cstr))

	if JavaInit != nil {
		JavaInit(uintptr(unsafe.Pointer(C.current_vm)))
	}

	// Inform Java that the program is initialized.
	C.pthread_mutex_lock(&C.go_started_mu)
	C.go_started = 1
	C.pthread_cond_signal(&C.go_started_cond)
	C.pthread_mutex_unlock(&C.go_started_mu)

	for {
		select {
		case w := <-windowCreated:
			windowDrawLoop(cb, w, queue)
		}
	}
}
