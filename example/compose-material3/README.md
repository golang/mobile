Go bind example using Kotlin + Compose + Material 3

Android project generated using Android Studio: `File > New > New Project > Empty Compose Activity (Material 3)`

### Requirements

1. `ANDROID_HOME` is set (e.g: `export ANDROID_HOME="$HOME/Android/Sdk"`)
2. `gomobile` and `gobind` is available - (e.g: `go install golang.org/x/mobile/cmd/gomobile@latest`)
3. Android NDK 19 to 23 is installed


### Installing the NDK

1. Download [latest Android NDK LTS - r23c](https://dl.google.com/android/repository/android-ndk-r23c-linux.zip)
2. Extract it under `~/Android/Sdk/ndk/`
3. Create the `ndk-bundle` folder as symbolic link:
    ```
    ln -sf ~/Android/Sdk/ndk/android-ndk-r23c ~/Android/Sdk/ndk-bundle
    ```

_For context: https://github.com/golang/go/issues/35030#issuecomment-641319703_


### Building project

1. From this directory run:
    ```
    gomobile bind -o android/app/libs/hello.aar -androidapi 32 ./hello
    ```

2. Import this project in Android Studio and build;

During tests `-androidapi` also worked with: `19`, `21`, `22`, `23`, `24`, `26`, `27`, `28`, `29`, `30`, `31` and `32`

> Note that Compose Previews that uses external modules is not working in Android Studio Chipmunk | 2021.2.1 Patch 1 - [see thread](https://issuetracker.google.com/issues/206862224?pli=1)

> You need to run the `gomobile bind` command again every time you make a change to Go code.
