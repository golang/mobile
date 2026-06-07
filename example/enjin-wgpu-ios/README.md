# enjin-wgpu-ios — Phase 0 iOS rendering spike

Proves the WebGPU (wgpu-native) **Metal** backend renders to a `CAMetalLayer`
handed up from the forked iOS app runtime. If this shows an animated solid color
on the iOS Simulator (or a device), the iOS render path is viable and the rest of
the iOS port (`enjin/docs/ios-implementation-plan.md`, Phases 1–4) is plumbing.

> **macOS + Xcode required.** This cannot be built or run on Linux. There is no
> Linux smoke-check for iOS (unlike Android's `make test-gomobile`).

## What it exercises

- `app/darwin_ios.m` — `UIViewController` backed by a `CAMetalLayer` (converted
  from the upstream GLES `GLKView` runtime).
- `app/darwin_ios_wgpu.go` — `MetalLayerEvents()` / `//export setMetalLayer` bridge
  and the `drainWork` publish handshake.
- `main.go` — `wgpu.CreateInstance(Metal)` → surface from the layer → adapter /
  device → `surface.Configure` → per-frame clear via a render pass → `Present`.

It does **not** import the engine (that would invert the `enjin → mobile` module
dependency); it calls the wgpu binding directly, matching
`enjin/platform/wgpubackend/backend.go`.

## Build & run (Apple-Silicon Mac)

```bash
# 1. Build the fork's gomobile binary (from the engine module).
cd /media/dev-ssd/vortex/enjin-workspace/enjin
make gomobile-build-tool          # -> ../mobile/gomobile
make gomobile-init                # one-time: gomobile init

# 2. Build the spike for the Simulator (no device, signing, or Apple acct needed).
cd ../mobile/example/enjin-wgpu-ios
../../gomobile build -target=iossimulator -bundleid vortex.studio.enjinspike .
#   -> produces enjin-wgpu-ios.app

# 3. Boot a simulator and install/launch.
xcrun simctl boot "iPhone 15" 2>/dev/null || true
open -a Simulator
xcrun simctl install booted enjin-wgpu-ios.app
xcrun simctl launch --console booted vortex.studio.enjinspike
```

`--console` streams the app's `log.Printf` output (adapter name, chosen surface
format/alpha, configure size) so you can confirm the Metal adapter was selected.

### On a physical device

```bash
../../gomobile build -target=ios -bundleid vortex.studio.enjinspike .
```

Requires a signing identity / provisioning profile. The Simulator path above is
enough to validate the render path; do device bring-up after the spike passes.

## What "success" looks like

A window cycling through blue-ish clear colors, plus console lines like:

```
spike: adapter="Apple ..." backend=Metal
spike: surface configured 1170x2532 format=BGRA8Unorm alpha=Opaque
```

## If it fails — likely culprits (these are the open questions Phase 0 answers)

- **Instance is nil / no Metal adapter** — confirm the vendored binding's iOS libs
  and `-framework Metal/QuartzCore` linked (`vendor/.../webgpu/wgpu/wgpu.go:21-29`).
- **GLES probe crash inside `CreateInstance`** — if wgpu-native probes GLES on iOS
  like it does on Android, port the `initTempEGLContext` workaround
  (`enjin/platform/gomobile/platform_android.go:20-65`). Expected NOT needed on iOS.
- **Black screen, no errors** — `CAMetalLayer` `pixelFormat`/`opaque`/`drawableSize`
  mismatch; check `surface.GetCapabilities` formats vs. the layer.
- **Hang on `app.Publish()`** — `drainWork` not running; verify `go drainWork()` in
  `app/darwin_ios.go`'s `main`.

## Cleanup

This spike replaces the upstream OpenGL ES iOS runtime in `app/`. The unrelated
`example/ivy/ios` GL example will no longer build — expected, and consistent with
how the Android WebGPU conversion dropped its GL path.
