//go:build darwin && ios

// WebGPU/Metal rendering bridge for iOS. No OpenGL ES — exposes the
// CAMetalLayer pointer so the engine (or the spike example) can create a wgpu
// Surface directly. This is the iOS counterpart of android_wgpu.go.

package app

import "C"

import "unsafe"

// metalLayerEvents carries the CAMetalLayer pointer as unsafe.Pointer. A
// non-nil value means a layer is available (created or re-laid-out); the
// consumer creates or reconfigures its wgpu surface from it.
var metalLayerEvents = make(chan unsafe.Pointer, 1)

// MetalLayerEvents returns a channel that delivers CAMetalLayer lifecycle
// events. A non-nil pointer means a layer is available; nil means it went away.
func MetalLayerEvents() <-chan unsafe.Pointer {
	return metalLayerEvents
}

// setMetalLayer is called from Objective-C (darwin_ios.m) on the main thread
// once the view's CAMetalLayer exists and whenever it is re-laid-out. It is
// idempotent — the channel is drained first so we never block UIKit.
//
//export setMetalLayer
func setMetalLayer(layer unsafe.Pointer) {
	select {
	case <-metalLayerEvents:
	default:
	}
	metalLayerEvents <- layer
}

// drainWork consumes the gomobile worker/publish channels. With WebGPU there is
// no GL context to make current and no buffers to swap — presentation is done by
// the wgpu surface in the render loop. We still complete the publish handshake so
// callers of app.Publish() don't block, mirroring android_wgpu.go's mainUI.
func drainWork() {
	workAvailable := theApp.worker.WorkAvailable()
	for {
		select {
		case <-workAvailable:
			theApp.worker.DoWork()
		case <-theApp.publish:
			theApp.publishResult <- PublishResult{}
		}
	}
}
