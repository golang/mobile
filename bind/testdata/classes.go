// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java

import (
	gopkg "Java/java"
	"Java/java/io"
	"Java/java/lang"
	"Java/java/lang/System"
	"Java/java/util/Spliterators"
	"Java/java/util/concurrent"
)

type Runnable struct {
	lang.Runnable
}

func (r *Runnable) Run(this gopkg.Runnable) {
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

func (_ *Future) Get() (lang.Object, error) {
	return nil, nil
}

// Use a trailing underscore to override multiple overloaded methods.
func (_ *Future) Get_(_ int64, _ concurrent.TimeUnit) (lang.Object, error) {
	return nil, nil
}

type Object struct {
	lang.Object
}

func innerClassTypes() {
	// java.util.Spliterators.iterator use inner class types
	// for the return value as well as parameters.
	Spliterators.Iterator(nil)
}

func returnType() {
	// Implicit types (java.io.Console) should be wrapped.
	cons := System.Console()
	cons.Flush()
}
