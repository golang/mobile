## Ivy Big Number Calculator (Android)

[Ivy (robpike.io/ivy)](https://robpike.io/ivy) is an interpreter for an APL-like language written in Go.
This repository hosts a minimal Android project used for its [Android version](https://play.google.com/store/apps/details?id=org.golang.ivy&hl=en_US&gl=US).

### How to build

Requirements
  - Go 1.17 or newer
  - Android SDK
  - Android NDK
  - `golang.org/x/mobile/cmd/gomobile`

The `gomobile` command respects the `ANDROID_HOME` and `ANDROID_NDK_HOME` environment variables.  If `gomobile` can't find your SDK and NDK, you can set these environment variables to specify their locations:
```
export ANDROID_HOME=/path/to/sdk-directory
export ANDROID_NDK_HOME=/path/to/ndk-directory
```

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
