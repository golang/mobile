#!/usr/bin/env bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# TODO(crawshaw):
#	This script does not yet run the tests, it just sets them
#	up so they can be run by Android Studio.

set -e

if [ ! -f test.bash ]; then
	echo 'test.bash must be run from $GOPATH/src/code.google.com/p/go.mobile/bind/java'
        exit 1
fi

if [ -z "$ANDROID_APP" ]; then
	echo 'ERROR: Environment variable ANDROID_APP is unset.'
	exit 1
fi

mkdir -p $ANDROID_APP/src/main/jniLibs/armeabi \
	$ANDROID_APP/src/main/java/go/testpkg
ln -sf $PWD/*.java $ANDROID_APP/src/main/java/go
ln -sf $PWD/testpkg/Testpkg.java $ANDROID_APP/src/main/java/go/testpkg
CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 \
	go build -ldflags="-shared" javatest.go
mv -f javatest $ANDROID_APP/src/main/jniLibs/armeabi/libgojni.so
