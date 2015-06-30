// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package event defines mobile app events, such as user input events.
//
// An event is represented by the empty interface type interface{}. Any value
// can be an event. This package defines a number of commonly used events used
// by the golang.org/x/mobile/app package:
//	- Config
//	- Draw
//	- Lifecycle
//	- Touch
// Other packages may define their own events, and post them onto an app's
// event channel.
//
// Other packages can also register event filters, e.g. to manage resources in
// response to lifecycle events. Such packages should call:
//	event.RegisterFilter(etc)
// in an init function inside that package.
//
// The program code that consumes an app's events is expected to call
// event.Filter on every event they receive, and then switch on its type:
//	for e := range a.Events() {
//		switch e := event.Filter(e).(type) {
//		etc
//		}
//	}
package event // import "golang.org/x/mobile/event"

// The best source on android input events is the NDK: include/android/input.h
//
// iOS event handling guide:
// https://developer.apple.com/library/ios/documentation/EventHandling/Conceptual/EventHandlingiPhoneOS

// TODO: keyboard events.

import (
	"fmt"

	"golang.org/x/mobile/geom"
)

var filters []func(interface{}) interface{}

// Filter calls each registered filter function in sequence.
func Filter(event interface{}) interface{} {
	for _, f := range filters {
		event = f(event)
	}
	return event
}

// RegisterFilter registers a filter function to be called by Filter. The
// function can return a different event, or return nil to consume the event,
// but the function can also return its argument unchanged, where its purpose
// is to trigger a side effect rather than modify the event.
//
// RegisterFilter should only be called from init functions.
func RegisterFilter(f func(interface{}) interface{}) {
	filters = append(filters, f)
}

// Change is a change in state, such as key or mouse button being up (off) or
// down (on). Some events with a Change-typed field can have no change in
// state, such as a key repeat or a mouse or touch drag.
type Change uint32

func (c Change) String() string {
	switch c {
	case ChangeOn:
		return "on"
	case ChangeOff:
		return "off"
	}
	return "none"
}

const (
	ChangeNone Change = 0
	ChangeOn   Change = 1
	ChangeOff  Change = 2
)

// Config holds the dimensions and physical resolution of the app's window.
type Config struct {
	// Width and Height are the window's dimensions.
	Width, Height geom.Pt

	// PixelsPerPt is the window's physical resolution. It is the number of
	// pixels in a single geom.Pt, from the golang.org/x/mobile/geom package.
	//
	// There are a wide variety of pixel densities in existing phones and
	// tablets, so apps should be written to expect various non-integer
	// PixelsPerPt values. In general, work in geom.Pt.
	PixelsPerPt float32
}

// Draw indicates that the app is ready to draw the next frame of the GUI. A
// frame is completed by calling the App's EndDraw method.
type Draw struct{}

// Lifecycle is a lifecycle change from an old stage to a new stage.
type Lifecycle struct {
	From, To LifecycleStage
}

// Crosses returns whether the transition from From to To crosses the stage s:
// 	- It returns ChangeOn if it does, and the Lifecycle change is positive.
// 	- It returns ChangeOff if it does, and the Lifecycle change is negative.
//	- Otherwise, it returns ChangeNone.
// See the documentation for LifecycleStage for more discussion of positive and
// negative changes, and crosses.
func (l Lifecycle) Crosses(s LifecycleStage) Change {
	switch {
	case l.From < s && l.To >= s:
		return ChangeOn
	case l.From >= s && l.To < s:
		return ChangeOff
	}
	return ChangeNone
}

// LifecycleStage is a stage in the app's lifecycle. The values are ordered, so
// that a lifecycle change from stage From to stage To implicitly crosses every
// stage in the range (min, max], exclusive on the low end and inclusive on the
// high end, where min is the minimum of From and To, and max is the maximum.
//
// The documentation for individual stages talk about positive and negative
// crosses. A positive Lifecycle change is one where its From stage is less
// than its To stage. Similarly, a negative Lifecycle change is one where From
// is greater than To. Thus, a positive Lifecycle change crosses every stage in
// the range (From, To] in increasing order, and a negative Lifecycle change
// crosses every stage in the range (To, From] in decreasing order.
type LifecycleStage uint32

// TODO: how does iOS map to these stages? What do cross-platform mobile
// abstractions do?

const (
	// LifecycleStageDead is the zero stage. No Lifecycle change crosses this
	// stage, but:
	//	- A positive change from this stage is the very first lifecycle change.
	//	- A negative change to this stage is the very last lifecycle change.
	LifecycleStageDead LifecycleStage = iota

	// LifecycleStageAlive means that the app is alive.
	//	- A positive cross means that the app has been created.
	//	- A negative cross means that the app is being destroyed.
	// Each cross, either from or to LifecycleStageDead, will occur only once.
	// On Android, these correspond to onCreate and onDestroy.
	LifecycleStageAlive

	// LifecycleStageVisible means that the app window is visible.
	//	- A positive cross means that the app window has become visible.
	//	- A negative cross means that the app window has become invisible.
	// On Android, these correspond to onStart and onStop.
	// On Desktop, an app window can become invisible if e.g. it is minimized,
	// unmapped, or not on a visible workspace.
	LifecycleStageVisible

	// LifecycleStageFocused means that the app window has the focus.
	//	- A positive cross means that the app window has gained the focus.
	//	- A negative cross means that the app window has lost the focus.
	// On Android, these correspond to onResume and onFreeze.
	LifecycleStageFocused
)

func (l LifecycleStage) String() string {
	switch l {
	case LifecycleStageDead:
		return "LifecycleStageDead"
	case LifecycleStageAlive:
		return "LifecycleStageAlive"
	case LifecycleStageVisible:
		return "LifecycleStageVisible"
	case LifecycleStageFocused:
		return "LifecycleStageFocused"
	default:
		return fmt.Sprintf("LifecycleStageInvalid:%d", l)
	}
}

// Touch is a user touch event.
//
// The same ID is shared by all events in a sequence. A sequence starts with a
// single TouchStart, is followed by zero or more TouchMoves, and ends with a
// single TouchEnd. An ID distinguishes concurrent sequences but is
// subsequently reused.
//
// On Android, Touch is an AInputEvent with AINPUT_EVENT_TYPE_MOTION.
// On iOS, Touch is the UIEvent delivered to a UIView.
type Touch struct {
	ID   TouchSequenceID
	Type TouchType
	Loc  geom.Point
}

func (t Touch) String() string {
	var ty string
	switch t.Type {
	case TouchStart:
		ty = "start"
	case TouchMove:
		ty = "move "
	case TouchEnd:
		ty = "end  "
	}
	return fmt.Sprintf("Touch{ %s, %s }", ty, t.Loc)
}

// TouchSequenceID identifies a sequence of Touch events.
type TouchSequenceID int64

// TODO: change TouchType to Change.

// TouchType describes the type of a touch event.
type TouchType byte

const (
	// TouchStart is a user first touching the device.
	//
	// On Android, this is a AMOTION_EVENT_ACTION_DOWN.
	// On iOS, this is a call to touchesBegan.
	TouchStart TouchType = iota

	// TouchMove is a user dragging across the device.
	//
	// A TouchMove is delivered between a TouchStart and TouchEnd.
	//
	// On Android, this is a AMOTION_EVENT_ACTION_MOVE.
	// On iOS, this is a call to touchesMoved.
	TouchMove

	// TouchEnd is a user no longer touching the device.
	//
	// On Android, this is a AMOTION_EVENT_ACTION_UP.
	// On iOS, this is a call to touchesEnded.
	TouchEnd
)
