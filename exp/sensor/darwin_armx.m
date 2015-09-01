// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin
// +build arm arm64

#import <CoreMotion/CoreMotion.h>

CMMotionManager* manager = nil;

void GoIOS_createManager() {
  manager = [[CMMotionManager alloc] init];
}

void GoIOS_startAccelerometer() {
  [manager startAccelerometerUpdates];
}

void GoIOS_stopAccelerometer() {
  [manager stopAccelerometerUpdates];
}

void GoIOS_readAccelerometer(int64_t* timestamp, float* v) {
  CMAccelerometerData* data = manager.accelerometerData;
  *timestamp = (int64_t)(data.timestamp * 1000 * 1000);
  v[0] = data.acceleration.x;
  v[1] = data.acceleration.y;
  v[2] = data.acceleration.z;
}

void GoIOS_destroyManager() {
  [manager release];
  manager = nil;
}
