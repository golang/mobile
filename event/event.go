// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package event defines a model for mobile app events, such as user input
// events.
//
// An event is represented by the empty interface type interface{}. Any value
// can be an event. Packages under this directory define a number of commonly
// used events used by the golang.org/x/mobile/app package:
//	- golang.org/x/mobile/event/config.Event
//	- golang.org/x/mobile/event/lifecycle.Event
//	- golang.org/x/mobile/event/paint.Event
//	- golang.org/x/mobile/event/touch.Event
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
