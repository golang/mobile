// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java

import (
	"Java/java/io"
	"Java/java/lang"
	"Java/java/util/concurrent"
)

type Runnable struct {
	lang.Runnable
}

func (r *Runnable) Run(this lang.Runnable) {
}

type InputStream struct {
	io.InputStream
}

func (_ *InputStream) Read() (int32, error) {
	return 0, nil
}

func NewInputStream() *InputStream {
	return new(InputStream)
}

type Future struct {
	concurrent.Future
}

func (_ *Future) Get() lang.Object {
	return nil
}

func (_ *Future) Get2(_ int64, _ concurrent.TimeUnit) lang.Object {
	return nil
}

type Object struct {
	lang.Object
}
