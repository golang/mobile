// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin
// +build arm arm64

package sensor

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreMotion

#import <stdlib.h>

void GoIOS_createManager();

void GoIOS_startAccelerometer(float interval);
void GoIOS_stopAccelerometer();
void GoIOS_readAccelerometer(int64_t* timestamp, float* vector);

void GoIOS_destroyManager();
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var channels struct {
	sync.Mutex
	acceleroDone chan struct{}
}

type manager struct {
}

func (m *manager) initialize() {
	C.GoIOS_createManager()
}

func (m *manager) enable(s Sender, t Type, delay time.Duration) error {
	// TODO(jbd): If delay is smaller than 10 milliseconds, set it to
	// 10 milliseconds. It is highest frequency iOS SDK suppports and
	// we don't want to have time.Tick durations smaller than this value.
	channels.Lock()
	defer channels.Unlock()

	switch t {
	case Accelerometer:
		if channels.acceleroDone != nil {
			return fmt.Errorf("sensor: cannot enable; %v sensor is already enabled", t)
		}
		// TODO(jbd): Check if accelerometer is available.
		interval := float64(delay) / float64(time.Second)
		C.GoIOS_startAccelerometer(C.float(interval))
		channels.acceleroDone = make(chan struct{})
		go m.runAccelerometer(s, delay, channels.acceleroDone)
	case Gyroscope:
	case Magnetometer:
	default:
		return fmt.Errorf("sensor: unknown sensor type: %v", t)
	}
	return nil
}

func (m *manager) disable(t Type) error {
	channels.Lock()
	defer channels.Unlock()

	switch t {
	case Accelerometer:
		if channels.acceleroDone == nil {
			return fmt.Errorf("sensor: cannot disable; %v sensor is not enabled", t)
		}
		close(channels.acceleroDone)
		channels.acceleroDone = nil
		C.GoIOS_stopAccelerometer()
	case Gyroscope:
	case Magnetometer:
	default:
		return fmt.Errorf("sensor: unknown sensor type: %v", t)
	}
	return nil
}

func (m *manager) runAccelerometer(s Sender, d time.Duration, done chan struct{}) {
	var timestamp C.int64_t
	var ev [3]C.float
	var lastTimestamp int64
	for {
		select {
		case <-done:
			return
		default:
			C.GoIOS_readAccelerometer((*C.int64_t)(unsafe.Pointer(&timestamp)), (*C.float)(unsafe.Pointer(&ev[0])))
			t := int64(timestamp)
			if t > lastTimestamp {
				// TODO(jbd): Do we need to convert the values to another unit?
				// How does iOS units compare to the Android units.
				s.Send(Event{
					Sensor:    Accelerometer,
					Timestamp: t,
					Data:      []float64{float64(ev[0]), float64(ev[1]), float64(ev[2])},
				})
				lastTimestamp = t
				time.Sleep(d / 2)
			}
		}
	}
}

// TODO(jbd): Remove close?
func (m *manager) close() error {
	C.GoIOS_destroyManager()
	return nil
}
