## Ivy Big Number Calculator (Android)

[Ivy (robpike.io/ivy)](https://robpike.io/ivy) is an interpreter for an APL-like language written in Go.
This repository hosts a minimal Android project used for its [Android version](https://play.google.com/store/apps/details?id=org.golang.ivy&hl=en_US&gl=US).

### How to build

Requirements
  - Go 1.17 or newer
  - Android SDK
  - Android NDK
  - `golang.org/x/mobile/cmd/gomobile`

The `gomobile` command uses `ANDROID_HOME` and `ANDROID_NDK_HOME` environment variables.
```
export ANDROID_HOME=/path/to/sdk-directory
export ANDROID_NDK_HOME=/path/to/ndk-directory
```

Note: Both ANDROID_SDK_ROOT and ANDROID_NDK_HOME are deprecated in Android tooling, but `gomobile` still uses it. In many cases, `ANDROID_HOME` corresponds to the new [`ANDROID_SDK_ROOT`](https://developer.android.com/studio/command-line/variables). If you installed NDK using [Android Studio's SDK manager](https://developer.android.com/studio/projects/install-ndk#default-version), use the `$ANDROID_SDK_ROOT/ndk/<ndk-version>/` directory as `ANDROID_NDK_HOME` (where `<ndk-version>` is the NDK version you want to use when compiling `robpike.io/ivy` for Android.

(TODO: update `gomobile` to work with modern Android layout)

From this directory, run:

```sh
  go install golang.org/x/mobile/cmd/gomobile@latest
  go install golang.org/x/mobile/cmd/gobind@latest

  # Make sure `gomobile` and `gobind` is in your `PATH`.
  gomobile bind -o app/ivy.aar robpike.io/ivy/mobile
```

Open this directory from Android Studio, and build.

`robpike.io/ivy` and `golang.org/x/mobile` are required dependencies of this main module. In order to update them:

```
  go get -d golang.org/x/mobile@latest
  go get -d robpike.io/ivy/mobile
  go mod tidy
```
