//go:build darwin && ios

// Command enjin-wgpu-ios is a Phase 0 spike: it proves that the WebGPU
// (wgpu-native) Metal backend renders to a CAMetalLayer handed up from the
// forked iOS app runtime (app/darwin_ios.m + app/darwin_ios_wgpu.go).
//
// It deliberately does NOT import the engine — that would invert the
// enjin -> mobile module dependency. Instead it calls the wgpu binding
// directly, matching the call pattern in
// enjin/platform/wgpubackend/backend.go, so a success here transfers cleanly
// to the real platform_ios.go.
//
// It clears the screen to a slowly cycling color and re-arms a paint event each
// frame. If you see an animated solid color on the simulator/device, the Metal
// surface path works.
//
// Build & run: see README.md in this directory.
package main

import (
	"log"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"

	"vortex.studio/mobile/app"
	"vortex.studio/mobile/event/lifecycle"
	"vortex.studio/mobile/event/paint"
	"vortex.studio/mobile/event/size"
)

func main() {
	instance := wgpu.CreateInstance(&wgpu.InstanceDescriptor{
		Backends: wgpu.InstanceBackendMetal,
	})
	if instance == nil {
		log.Fatal("spike: failed to create wgpu instance (Metal unavailable?)")
	}

	layerEvents := app.MetalLayerEvents()

	var (
		surface    *wgpu.Surface
		adapter    *wgpu.Adapter
		device     *wgpu.Device
		queue      *wgpu.Queue
		format     wgpu.TextureFormat
		alphaMode  wgpu.CompositeAlphaMode
		width      uint32
		height     uint32
		configured bool
		paused     bool
		frame      int
	)

	configure := func() {
		if surface == nil || adapter == nil || device == nil || width == 0 || height == 0 {
			return
		}
		surface.Configure(adapter, device, &wgpu.SurfaceConfiguration{
			Usage:       wgpu.TextureUsageRenderAttachment,
			Format:      format,
			Width:       width,
			Height:      height,
			PresentMode: wgpu.PresentModeFifo,
			AlphaMode:   alphaMode,
		})
		configured = true
		log.Printf("spike: surface configured %dx%d format=%v alpha=%v", width, height, format, alphaMode)
	}

	// initSurface is called the first time a CAMetalLayer pointer arrives.
	initSurface := func(layer unsafe.Pointer) {
		surface = instance.CreateSurface(&wgpu.SurfaceDescriptor{
			MetalLayer: &wgpu.SurfaceDescriptorFromMetalLayer{
				Layer: layer,
			},
		})
		if surface == nil {
			log.Fatal("spike: failed to create surface from CAMetalLayer")
		}

		var err error
		adapter, err = instance.RequestAdapter(&wgpu.RequestAdapterOptions{
			CompatibleSurface: surface,
			PowerPreference:   wgpu.PowerPreferenceHighPerformance,
		})
		if err != nil {
			log.Fatalf("spike: RequestAdapter: %v", err)
		}
		device, err = adapter.RequestDevice(&wgpu.DeviceDescriptor{Label: "enjin-spike"})
		if err != nil {
			log.Fatalf("spike: RequestDevice: %v", err)
		}
		queue = device.GetQueue()

		caps := surface.GetCapabilities(adapter)
		format = caps.Formats[0]
		for _, f := range caps.Formats {
			if f == wgpu.TextureFormatBGRA8Unorm || f == wgpu.TextureFormatRGBA8Unorm {
				format = f
				break
			}
		}
		alphaMode = caps.AlphaModes[0]
		for _, a := range caps.AlphaModes {
			if a == wgpu.CompositeAlphaModeOpaque {
				alphaMode = a
				break
			}
		}
		log.Printf("spike: adapter=%q backend=%v", adapter.GetInfo().Name, adapter.GetInfo().BackendType)
		configure()
	}

	render := func() {
		tex, err := surface.GetCurrentTexture()
		if err != nil || tex == nil {
			log.Printf("spike: GetCurrentTexture: %v", err)
			return
		}
		view, err := tex.CreateView(nil)
		if err != nil {
			log.Printf("spike: CreateView: %v", err)
			return
		}
		defer view.Release()

		// Cycle the clear color so we can see frames advancing.
		frame++
		g := float64(frame%120) / 120.0

		enc, err := device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{})
		if err != nil {
			log.Printf("spike: CreateCommandEncoder: %v", err)
			return
		}
		pass := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
			ColorAttachments: []wgpu.RenderPassColorAttachment{{
				View:       view,
				LoadOp:     wgpu.LoadOpClear,
				StoreOp:    wgpu.StoreOpStore,
				ClearValue: wgpu.Color{R: 0.6, G: g, B: 0.9, A: 1.0},
			}},
		})
		pass.End()
		pass.Release()

		cmd, err := enc.Finish(nil)
		if err != nil {
			log.Printf("spike: encoder.Finish: %v", err)
			return
		}
		queue.Submit(cmd)
		surface.Present()
		device.Poll(false, nil)
	}

	app.Main(func(a app.App) {
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					paused = false
					a.Send(paint.Event{}) // kick the paint loop
				case lifecycle.CrossOff:
					paused = true
				}

			case size.Event:
				width = uint32(e.WidthPx)
				height = uint32(e.HeightPx)
				if configured {
					configure() // reconfigure on rotation/resize
				}

			case paint.Event:
				if paused {
					continue
				}
				// Pick up the CAMetalLayer the first time it's delivered.
				select {
				case lp := <-layerEvents:
					if lp != nil && surface == nil {
						initSurface(lp)
					}
				default:
				}

				if configured {
					render()
				}
				a.Publish()
				a.Send(paint.Event{}) // request the next frame
			}
		}
	})
}
