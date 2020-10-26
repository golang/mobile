// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package gl

import (
	"syscall/js"
)

type context struct {
	ctx js.Value

	methods map[string]js.Value
}

type context3 struct {
	*context
}

func (c *context) JSValue() js.Value {
	return c.ctx
}

func (c *context) cachedCall(method string, args ...interface{}) js.Value {
	if m, ok := c.methods[method]; ok {
		return m.Invoke(args...)
	}

	if c.methods == nil {
		c.methods = make(map[string]js.Value)
	}

	m := c.ctx.Get(method).Call("bind", c.ctx)
	c.methods[method] = m

	return m.Invoke(args...)
}

func (c *context) DoWork() {
	// WebASM is single-threaded in browsers, so the work is executed synchronously.
}

func (c *context) WorkAvailable() <-chan struct{} {
	return nil
}

func newContext() (canvas, context js.Value, isWebGL2 bool) {
	canvas = js.Global().Get("document").Call("createElement", "canvas")

	if context = canvas.Call("getContext", "webgl2"); context.Truthy() {
		isWebGL2 = true
	} else {
		context = canvas.Call("getContext", "webgl")
	}

	if !context.Truthy() {
		panic("could not obtain WebGL context")
	}

	return
}

func NewContext() (Context, Worker) {
	_, ctx, isWebGL2 := newContext()

	c := &context{ctx: ctx}

	if isWebGL2 {
		return context3{c}, c
	}

	return c, c
}

func Version() string {
	_, ctx, isWebGL2 := newContext()

	if loseContext := ctx.Call("getExtension", "WEBGL_lose_context"); loseContext.Truthy() {
		loseContext.Call("loseContext")
	}

	if isWebGL2 {
		return "GL_ES_3_0"
	}

	return "GL_ES_2_0"
}

var (
	uint8ArrayCtor   = js.Global().Get("Uint8Array")
	float32ArrayCtor = js.Global().Get("Float32Array")
)

func jsBytes(b []byte) js.Value {
	if b == nil {
		return js.Null()
	}

	a := uint8ArrayCtor.New(len(b))

	js.CopyBytesToJS(a, b)

	return a
}

func jsFloats(f []float32) js.Value {
	if f == nil {
		return js.Null()
	}

	a := float32ArrayCtor.New(len(f))

	for i, v := range f {
		a.SetIndex(i, v)
	}

	return a
}

func jsPtr(v js.Value) *js.Value {
	if v.Truthy() {
		return &v
	}

	return nil
}

func toInt(v js.Value) int {
	if v.Type() == js.TypeBoolean {
		if v.Bool() {
			return TRUE
		}

		return FALSE
	}

	return v.Int()
}
