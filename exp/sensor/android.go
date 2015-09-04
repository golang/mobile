// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package sensor

/*
#cgo LDFLAGS: -landroid

#include <stdlib.h>
#include <android/sensor.h>

void GoAndroid_createManager();
void GoAndroid_destroyManager();
int  GoAndroid_enableSensor(int, int32_t);
void GoAndroid_disableSensor(int);
int  GoAndroid_readQueue(int n, int32_t* types, int64_t* timestamps, float* vectors);
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

var (
	collectingMu sync.Mutex // guards collecting

	// collecting is true if sensor event collecting background
	// job has already started.
	collecting bool
)

// initSignal initializes an underlying looper and event queue.
type initSignal struct{}

// closeSignal destroys the underlying looper and event queue.
type closeSignal struct{}

// readSignal reads up to len(dst) events and mutates n with
// the number of returned events. If error occurs during the read,
// it mutates err.
type readSignal struct {
	dst []Event
	n   *int
	err *error
}

// enableSignal enables the sensors events on the underlying
// event queue for the specified sensor type with the specified
// latency criterion.
type enableSignal struct {
	t     Type
	delay time.Duration
	err   *error
}

// disableSignal disables the events on the underlying event queue
// from the sensor specified.
type disableSignal struct {
	t Type
}

type inOut struct {
	in  interface{}
	out chan struct{}
}

var inout = make(chan inOut)

// init inits the manager and creates a goroutine to proxy the CGO calls.
// All actions related to an ALooper needs to be performed from the same
// OS thread. The goroutine proxy locks itself to an OS thread and handles the
// CGO traffic on the same thread.
func init() {
	go func() {
		runtime.LockOSThread()
		for {
			v := <-inout
			switch s := v.in.(type) {
			case initSignal:
				C.GoAndroid_createManager()
			case enableSignal:
				usecsDelay := s.delay.Nanoseconds() / 1000
				code := int(C.GoAndroid_enableSensor(typeToInt(s.t), C.int32_t(usecsDelay)))
				if code != 0 {
					*s.err = fmt.Errorf("sensor: no default %v sensor on the device", s.t)
				}
			case disableSignal:
				C.GoAndroid_disableSensor(typeToInt(s.t))
			case readSignal:
				n, err := readEvents(s.dst)
				*s.n = n
				*s.err = err
			case closeSignal:
				C.GoAndroid_destroyManager()
				close(v.out)
				return // we don't need this goroutine anymore
			}
			close(v.out)
		}
	}()

	done := make(chan struct{})
	inout <- inOut{
		in:  initSignal{},
		out: done,
	}
	<-done
}

func enable(s Sender, t Type, delay time.Duration) error {
	startCollecting(s)

	var err error
	done := make(chan struct{})
	inout <- inOut{
		in:  enableSignal{t: t, delay: delay, err: &err},
		out: done,
	}
	<-done
	return err
}

func startCollecting(s Sender) {
	collectingMu.Lock()
	defer collectingMu.Unlock()

	if collecting {
		// already collecting.
		return
	}
	collecting = true

	// TODO(jbd): Disable the goroutine if all sensors are disabled?
	// Read will block until there are new events, a goroutine will be
	// parked forever until a sensor is enabled. There must be no
	// performance cost other than allocating blocking an OS thread
	// forever to keep the goroutine running.
	go func() {
		ev := make([]Event, 8)
		var n int
		var err error // TODO(jbd): How to handle the errors? error channel?
		for {
			// TODO(jbd): readSignal is not required anymore. Use the proxying
			// goroutine to continously poll the queue and send the events to s.
			done := make(chan struct{})
			inout <- inOut{
				in:  readSignal{dst: ev, n: &n, err: &err},
				out: done,
			}
			<-done
			for i := 0; i < n; i++ {
				s.Send(ev[i])
			}
		}
	}()
}

func disable(t Type) error {
	done := make(chan struct{})
	inout <- inOut{
		in:  disableSignal{t: t},
		out: done,
	}
	<-done
	return nil
}

func readEvents(e []Event) (n int, err error) {
	num := len(e)
	types := make([]C.int32_t, num)
	timestamps := make([]C.int64_t, num)
	vectors := make([]C.float, 3*num)

	n = int(C.GoAndroid_readQueue(
		C.int(num),
		(*C.int32_t)(unsafe.Pointer(&types[0])),
		(*C.int64_t)(unsafe.Pointer(&timestamps[0])),
		(*C.float)(unsafe.Pointer(&vectors[0]))),
	)
	for i := 0; i < n; i++ {
		e[i] = Event{
			Sensor:    intToType[int(types[i])],
			Timestamp: int64(timestamps[i]),
			Data: []float64{
				float64(vectors[i*3]),
				float64(vectors[i*3+1]),
				float64(vectors[i*3+2]),
			},
		}
	}
	return
}

// TODO(jbd): Remove destroy?
func destroy() error {
	done := make(chan struct{})
	inout <- inOut{
		in:  closeSignal{},
		out: done,
	}
	<-done
	return nil
}

var intToType = map[int]Type{
	C.ASENSOR_TYPE_ACCELEROMETER:  Accelerometer,
	C.ASENSOR_TYPE_GYROSCOPE:      Gyroscope,
	C.ASENSOR_TYPE_MAGNETIC_FIELD: Magnetometer,
}

func typeToInt(t Type) C.int {
	for k, v := range intToType {
		if v == t {
			return C.int(k)
		}
	}
	return C.int(-1)
}
