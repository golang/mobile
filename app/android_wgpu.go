//go:build android

// WebGPU rendering path for Android. No EGL — exposes the raw
// ANativeWindow pointer so the engine can create a wgpu Surface.

package app

/*
#include <android/native_window.h>
*/
import "C"
import (
	"unsafe"

	"vortex.studio/mobile/event/lifecycle"
	"vortex.studio/mobile/event/paint"
	"vortex.studio/mobile/event/size"
	"vortex.studio/mobile/geom"
)

// nativeWindowEvents carries the ANativeWindow pointer as unsafe.Pointer.
// A non-nil value means the window was created; nil means it was destroyed.
var nativeWindowEvents = make(chan unsafe.Pointer, 1)

// NativeWindowEvents returns a channel that delivers ANativeWindow lifecycle
// events. A non-nil pointer means a new window is available; nil means the
// previous window was destroyed and the surface should be released.
func NativeWindowEvents() <-chan unsafe.Pointer {
	return nativeWindowEvents
}

// GetNativeWindowSize returns the width and height of the given ANativeWindow.
func GetNativeWindowSize(w unsafe.Pointer) (int, int) {
	return nativeWindowWidth(w), nativeWindowHeight(w)
}

func handleNativeWindowCreated(window unsafe.Pointer) {
	// Drain any stale event so we never block.
	select {
	case <-nativeWindowEvents:
	default:
	}
	nativeWindowEvents <- window
}

func handleNativeWindowDestroyed(window unsafe.Pointer) {
	// Signal that the window is gone.
	select {
	case <-nativeWindowEvents:
	default:
	}
	nativeWindowEvents <- nil
}

// mainUI is the WebGPU variant of the main UI loop. It does NOT touch EGL.
func mainUI(vm, jniEnv, ctx uintptr) error {
	workAvailable := theApp.worker.WorkAvailable()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		mainUserFn(theApp)
	}()

	var pixelsPerPt float32
	var orientation size.Orientation

	for {
		select {
		case <-donec:
			return nil
		case cfg := <-windowConfigChange:
			pixelsPerPt = cfg.pixelsPerPt
			orientation = cfg.orientation
		case w := <-windowRedrawNeeded:
			theApp.sendLifecycle(lifecycle.StageFocused)
			widthPx := nativeWindowWidth(w)
			heightPx := nativeWindowHeight(w)
			theApp.eventsIn <- size.Event{
				WidthPx:     widthPx,
				HeightPx:    heightPx,
				WidthPt:     geom.Pt(float32(widthPx) / pixelsPerPt),
				HeightPt:    geom.Pt(float32(heightPx) / pixelsPerPt),
				PixelsPerPt: pixelsPerPt,
				Orientation: orientation,
			}
			theApp.eventsIn <- paint.Event{External: true}
		case <-windowDestroyed:
			theApp.sendLifecycle(lifecycle.StageAlive)
		case <-workAvailable:
			theApp.worker.DoWork()
		case <-theApp.publish:
			// WebGPU presentation is handled by the engine backend, not EGL.
			select {
			case windowRedrawDone <- struct{}{}:
			default:
			}
			theApp.publishResult <- PublishResult{}
		}
	}
}
