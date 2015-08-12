// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

#ifndef GO_SENSORS_ANDROID_H
#define GO_SENSORS_ANDROID_H

typedef struct GoAndroid_SensorManager {
  ASensorEventQueue* queue;
  ALooper* looper;
  int looperId;
} GoAndroid_SensorManager;

void GoAndroid_createManager(int looperId, GoAndroid_SensorManager* dst);
void GoAndroid_destroyManager(GoAndroid_SensorManager* m);
int  GoAndroid_enableSensor(ASensorEventQueue*, int, int32_t);
void GoAndroid_disableSensor(ASensorEventQueue*, int);
int  GoAndroid_readQueue(int looperId, ASensorEventQueue* q, int n, int32_t* types, int64_t* timestamps, float* vectors);

#endif
