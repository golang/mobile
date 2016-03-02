// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// Package benchmarkpkg contains functions for benchmarking bound functions.
package benchmarkpkg

import (
	"log"
	"time"
)

type Benchmarks interface {
	// It seems to be much faster to call a native function from Java
	// when there is already a native call earlier in the stack.
	//
	// Run runs a named benchmark from a different thread, with
	// no native call prior in the stack.
	Run(name string, n int)
	// RunDirect runs a named benchmark directly, with the native
	// context from the call itself.
	RunDirect(name string, n int)

	// Callbacks for Go benchmarks
	NewI() I
	Noargs()
	Onearg(_ int)
	Oneret() int
	Ref(_ I)
	Manyargs(_, _, _, _, _, _, _, _, _, _ int)
	String(_ string)
	Slice(_ []byte)
}

type (
	I interface {
		F()
	}

	AnI struct {
	}
)

func (_ *AnI) F() {
}

func NewI() I {
	return new(AnI)
}

func runBenchmark(name string, f func(n int)) {
	// Run once for warmup
	f(1)
	n := 1000
	var dt time.Duration
	minDuration := 1 * time.Second
	for dt < minDuration {
		n *= 2
		t0 := time.Now()
		f(n)
		dt = time.Since(t0)
	}
	log.Printf("Benchmark%s	%d	%d ns/op\n", name, n, dt.Nanoseconds()/int64(n))
}

func runGoBenchmark(name string, f func()) {
	runBenchmark("Go"+name, func(n int) {
		for i := 0; i < n; i++ {
			f()
		}
	})
	runBenchmark("Go"+name+"Direct", func(n int) {
		done := make(chan struct{})
		go func() {
			for i := 0; i < n; i++ {
				f()
			}
			close(done)
		}()
		<-done
	})
}

func RunBenchmarks(b Benchmarks) {
	names := []string{
		"Empty",
		"Noargs",
		"Onearg",
		"Oneret",
		"Manyargs",
		"Refjava",
		"Refgo",
		"StringShort",
		"StringLong",
		"StringShortUnicode",
		"StringLongUnicode",
		"SliceShort",
		"SliceLong",
	}
	for _, name := range names {
		runBenchmark("Java"+name, func(n int) {
			b.Run(name, n)
		})
		runBenchmark("Java"+name+"Direct", func(n int) {
			b.RunDirect(name, n)
		})
	}
	runGoBenchmark("Empty", func() {})
	runGoBenchmark("Noarg", func() { b.Noargs() })
	runGoBenchmark("Onearg", func() { b.Onearg(0) })
	runGoBenchmark("Oneret", func() { b.Oneret() })
	javaRef := b.NewI()
	runGoBenchmark("Refjava", func() { b.Ref(javaRef) })
	goRef := NewI()
	runGoBenchmark("Refgo", func() { b.Ref(goRef) })
	runGoBenchmark("Manyargs", func() { b.Manyargs(0, 0, 0, 0, 0, 0, 0, 0, 0, 0) })
	runGoBenchmark("StringShort", func() { b.String(ShortString) })
	runGoBenchmark("StringLong", func() { b.String(LongString) })
	runGoBenchmark("StringShortUnicode", func() { b.String(ShortStringUnicode) })
	runGoBenchmark("StringLongUnicode", func() { b.String(LongStringUnicode) })
	runGoBenchmark("SliceShort", func() { b.Slice(ShortSlice) })
	runGoBenchmark("SliceLong", func() { b.Slice(LongSlice) })
}

func Noargs() {
}

func Onearg(_ int) {
}

func Manyargs(_, _, _, _, _, _, _, _, _, _ int) {
}

func Oneret() int {
	return 0
}

func String(_ string) {
}

func Slice(_ []byte) {
}

func Ref(_ I) {
}

var (
	ShortSlice = make([]byte, 10)
	LongSlice  = make([]byte, 100000)
)

const (
	ShortString        = "Hello, World!"
	LongString         = "Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World!  Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World!  Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World!  Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World! Hello, World!, World!"
	ShortStringUnicode = "Hello, 世界!"
	LongStringUnicode  = "Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界!  Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界!  Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界!  Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界! Hello, 世界!, 世界!"
)
