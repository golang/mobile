// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin
// +build arm arm64

package app

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework UIKit -framework GLKit -framework OpenGLES -framework QuartzCore
#include <sys/utsname.h>
#include <stdint.h>
#include <pthread.h>

extern struct utsname sysInfo;

void runApp(void);
void setContext(void* context);
uint64_t threadID();
*/
import "C"
import (
	"log"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var initThreadID uint64

func init() {
	// Lock the goroutine responsible for initialization to an OS thread.
	// This means the goroutine running main (and calling the run function
	// below) is locked to the OS thread that started the program. This is
	// necessary for the correct delivery of UIKit events to the process.
	//
	// A discussion on this topic:
	// https://groups.google.com/forum/#!msg/golang-nuts/IiWZ2hUuLDA/SNKYYZBelsYJ
	runtime.LockOSThread()
	initThreadID = uint64(C.threadID())
}

func run(cbs []Callbacks) {
	if tid := uint64(C.threadID()); tid != initThreadID {
		log.Fatalf("app.Run called on thread %d, but app.init ran on %d", tid, initThreadID)
	}
	close(mainCalled)
	callbacks = cbs
	C.runApp()
}

// TODO(crawshaw): determine minimum iOS version and remove irrelevant devices.
var machinePPI = map[string]int{
	"i386":      163, // simulator
	"x86_64":    163, // simulator
	"iPod1,1":   163, // iPod Touch gen1
	"iPod2,1":   163, // iPod Touch gen2
	"iPod3,1":   163, // iPod Touch gen3
	"iPod4,1":   326, // iPod Touch gen4
	"iPod5,1":   326, // iPod Touch gen5
	"iPhone1,1": 163, // iPhone
	"iPhone1,2": 163, // iPhone 3G
	"iPhone2,1": 163, // iPhone 3GS
	"iPad1,1":   132, // iPad gen1
	"iPad2,1":   132, // iPad gen2
	"iPad2,2":   132, // iPad gen2 GSM
	"iPad2,3":   132, // iPad gen2 CDMA
	"iPad2,4":   132, // iPad gen2
	"iPad2,5":   163, // iPad Mini gen1
	"iPad2,6":   163, // iPad Mini gen1 AT&T
	"iPad2,7":   163, // iPad Mini gen1 VZ
	"iPad3,1":   264, // iPad gen3
	"iPad3,2":   264, // iPad gen3 VZ
	"iPad3,3":   264, // iPad gen3 AT&T
	"iPad3,4":   264, // iPad gen4
	"iPad3,5":   264, // iPad gen4 AT&T
	"iPad3,6":   264, // iPad gen4 VZ
	"iPad4,1":   264, // iPad Air wifi
	"iPad4,2":   264, // iPad Air LTE
	"iPad4,3":   264, // iPad Air LTE China
	"iPad4,4":   326, // iPad Mini gen2 wifi
	"iPad4,5":   326, // iPad Mini gen2 LTE
	"iPad4,6":   326, // iPad Mini 3
	"iPad4,7":   326, // iPad Mini 3
	"iPhone3,1": 326, // iPhone 4
	"iPhone4,1": 326, // iPhone 4S
	"iPhone5,1": 326, // iPhone 5
	"iPhone5,2": 326, // iPhone 5
	"iPhone5,3": 326, // iPhone 5c
	"iPhone5,4": 326, // iPhone 5c
	"iPhone6,1": 326, // iPhone 5s
	"iPhone6,2": 326, // iPhone 5s
	"iPhone7,1": 401, // iPhone 6 Plus
	"iPhone7,2": 326, // iPhone 6
}

func ppi() int {
	C.uname(&C.sysInfo)
	name := C.GoString(&C.sysInfo.machine[0])
	v, ok := machinePPI[name]
	if !ok {
		log.Printf("unknown machine: %s", name)
		v = 163 // emergency fallback
	}
	return v
}

//export setGeom
func setGeom(width, height int) {
	if geom.PixelsPerPt == 0 {
		geom.PixelsPerPt = float32(ppi()) / 72
	}
	configAlt.Width = geom.Pt(float32(width) / geom.PixelsPerPt)
	configAlt.Height = geom.Pt(float32(height) / geom.PixelsPerPt)
	configSwap(callbacks)
}

var startedgl = false

// touchIDs is the current active touches. The position in the array
// is the ID, the value is the UITouch* pointer value.
//
// It is widely reported that the iPhone can handle up to 5 simultaneous
// touch events, while the iPad can handle 11.
var touchIDs [11]uintptr

var touchEvents struct {
	sync.Mutex
	pending []event.Touch
}

//export sendTouch
func sendTouch(touch uintptr, touchType int, x, y float32) {
	id := -1
	for i, val := range touchIDs {
		if val == touch {
			id = i
			break
		}
	}
	if id == -1 {
		for i, val := range touchIDs {
			if val == 0 {
				touchIDs[i] = touch
				id = i
				break
			}
		}
		if id == -1 {
			panic("out of touchIDs")
		}
	}

	ty := event.TouchType(touchType)
	if ty == event.TouchEnd {
		touchIDs[id] = 0
	}

	touchEvents.Lock()
	touchEvents.pending = append(touchEvents.pending, event.Touch{
		ID:   event.TouchSequenceID(id),
		Type: ty,
		Loc: geom.Point{
			X: geom.Pt(x / geom.PixelsPerPt),
			Y: geom.Pt(y / geom.PixelsPerPt),
		},
	})
	touchEvents.Unlock()
}

//export drawgl
func drawgl(ctx uintptr) {
	if !startedgl {
		startedgl = true
		go gl.Start(func() {
			C.setContext(unsafe.Pointer(ctx))
		})
		stateStart(cb)
	}

	touchEvents.Lock()
	pending := touchEvents.pending
	touchEvents.pending = nil
	touchEvents.Unlock()
	for _, cb := range callbacks {
		if cb.Touch != nil {
			for _, e := range pending {
				cb.Touch(e)
			}
		}
	}

	// TODO not here?
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, cb := range callbacks {
		if cb.Draw != nil {
			cb.Draw()
		}
	}
}
