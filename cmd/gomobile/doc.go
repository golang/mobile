// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT. GENERATED BY 'gomobile help documentation doc.go'.

/*
Gomobile is a tool for building and running mobile apps written in Go.

To install:

	$ go get golang.org/x/mobile/cmd/gomobile
	$ gomobile init

At least Go 1.5 is required.
For detailed instructions, see https://golang.org/wiki/Mobile.

Usage:

	gomobile command [arguments]

Commands:

	bind        build a library for Android and iOS
	build       compile android APK and iOS app
	clean       remove object files and cached gomobile files
	init        install android compiler toolchain
	install     compile android APK and install on device
	version     print version

Use 'gomobile help [command]' for more information about that command.


Build a library for Android and iOS

Usage:

	gomobile bind [-target android|ios] [-o output] [build flags] [package]

Bind generates language bindings for the package named by the import
path, and compiles a library for the named target system.

The -target flag takes a target system name, either android (the
default) or ios.

For -target android, the bind command produces an AAR (Android ARchive)
file that archives the precompiled Java API stub classes, the compiled
shared libraries, and all asset files in the /assets subdirectory under
the package directory. The output is named '<package_name>.aar' by
default. This AAR file is commonly used for binary distribution of an
Android library project and most Android IDEs support AAR import. For
example, in Android Studio (1.2+), an AAR file can be imported using
the module import wizard (File > New > New Module > Import .JAR or
.AAR package), and setting it as a new dependency
(File > Project Structure > Dependencies).  This requires 'javac'
(version 1.7+) and Android SDK (API level 15 or newer) to build the
library for Android. The environment variable ANDROID_HOME must be set
to the path to Android SDK. The generated Java class is in the java
package '<package_name>' unless -javapkg flag is specified.

By default, -target=android builds shared libraries for all supported
instruction sets (arm, arm64, 386, amd64). A subset of instruction sets
can be selected by specifying target type with the architecture name. E.g.,
-target=android/arm,android/386.

For -target ios, gomobile must be run on an OS X machine with Xcode
installed. Support is not complete. The -prefix flag can be used to prefix
the names of generated Objective-C types.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -n, -x, -gcflags, -ldflags, -tags, and -work
are shared with the build command. For documentation, see 'go help build'.


Compile android APK and iOS app

Usage:

	gomobile build [-target android|ios] [-o output] [build flags] [package]

Build compiles and encodes the app named by the import path.

The named package must define a main function.

The -target flag takes a target system name, either android (the
default) or ios.

For -target android, if an AndroidManifest.xml is defined in the
package directory, it is added to the APK output. Otherwise, a default
manifest is generated. By default, this builds a fat APK for all supported
instruction sets (arm, 386, amd64, arm64). A subset of instruction sets can
be selected by specifying target type with the architecture name. E.g.
-target=android/arm,android/386.

For -target ios, gomobile must be run on an OS X machine with Xcode
installed. Support is not complete.

If the package directory contains an assets subdirectory, its contents
are copied into the output.

The -o flag specifies the output file name. If not specified, the
output file name depends on the package built.

The -v flag provides verbose output, including the list of packages built.

The build flags -a, -i, -n, -x, -gcflags, -ldflags, -tags, and -work are
shared with the build command. For documentation, see 'go help build'.


Remove object files and cached gomobile files

Usage:

	gomobile clean

Clean removes object files and cached NDK files downloaded by gomobile init


Install android compiler toolchain

Usage:

	gomobile init [-u]

Init installs the Android C++ compiler toolchain and builds copies
of the Go standard library for mobile devices.

When first run, it downloads part of the Android NDK.
The toolchain is installed in $GOPATH/pkg/gomobile.

The -u option forces download and installation of the new toolchain
even when the toolchain exists.


Compile android APK and install on device

Usage:

	gomobile install [-target android] [build flags] [package]

Install compiles and installs the app named by the import path on the
attached mobile device.

Only -target android is supported. The 'adb' tool must be on the PATH.

The build flags -a, -i, -n, -x, -gcflags, -ldflags, -tags, and -work are
shared with the build command.
For documentation, see 'go help build'.


Print version

Usage:

	gomobile version

Version prints versions of the gomobile binary and tools.


Gobind gradle plugin

The gobind gradle plugin integrates gomobile (and gobind) with the Android gradle
build system. The plugin supports two modes, library project and direct integration.

The library project mode is suitable for exporting Go packages to Android apps. To activate it,
create a gradle subproject and configure the gobind plugin in the build.gradle file:

	apply plugin: "org.golang.mobile.bind"

	gobind {
		// The Go package path; must be under one of the GOPATH elements or
		// a relative to the current directory (e.g. ../../hello)
		pkg = "golang.org/x/mobile/example/bind/hello"

		// Optional GOPATH.
		// GOPATH = "/YOUR/GOPATH"

		// Optional path to the go executable.
		// GO = "/path/to/go"

		// Optionally, set the absolute path to the gomobile binary.
		// GOMOBILE = "/path/to/gomobile"
	}

The direct integration mode is suitable for Go packages that import Java API, Android API,
project dependencies as well as generated R.* and Android databinding classes. In the
main Android project build.gradle, apply the gobind plugin after the android plugin:

	apply plugin: 'com.android.application'

	...

	apply plugin: "org.golang.mobile.bind"

	gobind {
		pkg = "golang.org/x/mobile/example/bind/reverse"
	}

*/
package main // import "golang.org/x/mobile/cmd/gomobile"
