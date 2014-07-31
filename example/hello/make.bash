#!/usr/bin/env bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

if [ ! -f make.bash ]; then
	echo 'make.bash must be run from $GOPATH/src/code.google.com/p/go.mobile/example/hello'
	exit 1
fi
if [ -z "$ANDROID_APP" ]; then
	echo 'ERROR: Environment variable ANDROID_APP is unset.'
	exit 1
fi
ANDROID_APP=$(cd $ANDROID_APP; pwd)

mkdir -p $ANDROID_APP/src/main/jniLibs/armeabi \
	$ANDROID_APP/src/main/java/go/hi
(cd ../.. && ln -sf $PWD/app/*.java $ANDROID_APP/src/main/java/go)
(cd ../.. && ln -sf $PWD/bind/java/*.java $ANDROID_APP/src/main/java/go)
ln -sf $PWD/hi/*.java $ANDROID_APP/src/main/java/go/hi
CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 \
	go build -ldflags="-shared" .
mv -f hello $ANDROID_APP/src/main/jniLibs/armeabi/libgojni.so
