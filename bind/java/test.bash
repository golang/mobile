#!/usr/bin/env bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

function die() {
  echo "FAIL: $1"
  exit 1
}

if [ ! -f test.bash ]; then
  die 'test.bash must be run from $GOPATH/src/golang.org/x/mobile/bind/java'
fi

function cleanup() {
  rm -rf "$ANDROID_APP"
}

if [ -z "$TMPDIR" ]; then
	TMPDIR="/tmp"
fi

if [ -z "$ANDROID_APP" ]; then
	ANDROID_APP=`mktemp -d ${TMPDIR}/android-java.XXXXX` || die 'failed to create a temporary directory'
	echo "Temporary directory for test: $ANDROID_APP"
	trap cleanup EXIT
fi

# Create an android project for test.
# TODO(hyangah): use android update lib-project if the $ANDROID_APP directory
# already exists.
android create lib-project -n BindJavaTest \
  -t "android-19"  -p $ANDROID_APP -k go.testpkg -g -v 0.12.+

gomobile bind -outdir="$ANDROID_APP" ./testpkg

# Copy files to the main src, jni directories.
mv $ANDROID_APP/android/src/main/java/go $ANDROID_APP/src/main/java/
mkdir -p $ANDROID_APP/src/main/jniLibs
cp -r $ANDROID_APP/android/libs/* $ANDROID_APP/src/main/jniLibs

# Add the test file under androidTest directory.
mkdir -p $ANDROID_APP/src/androidTest/java/go
ln -sf $PWD/SeqTest.java $ANDROID_APP/src/androidTest/java/go

# Build the test apk. ($ANDROID_APP/build/outputs/apk).
cd $ANDROID_APP

# If there is no connected device, this will fail after creating the test apk.
# The apk is located in $ANDROID_APP/build/outputs/apk directory.
./gradlew connectedAndroidTest && echo "PASS" && exit 0

# TODO(hyangah): copy the gradle's test output directory in case of test failure?

# TODO(hyangah): gradlew build
# Currently build fails due to a lint error. Disable lint check or
# specify the minSdkVersion in gradle.build to avoid the lint error.
