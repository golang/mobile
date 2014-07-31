// Package testpkg contains bound functions for testing the cgo-JNI interface.
package testpkg

import (
	"fmt"
	"runtime"
	"time"
)

type I interface {
	F()
}

func Call(i I) {
	i.F()
}

var keep []I

func Keep(i I) {
	keep = append(keep, i)
}

var numSCollected int

type S struct {
	// *S already has a finalizer, so we need another object
	// to count successful collections.
	innerObj *int
}

func (*S) F() {
	fmt.Println("called F on *S")
}

func finalizeInner(*int) {
	numSCollected++
}

func New() *S {
	s := &S{innerObj: new(int)}
	runtime.SetFinalizer(s.innerObj, finalizeInner)
	return s
}

func GC() {
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	runtime.GC()
}

func Add(x, y int) int {
	return x + y
}

func NumSCollected() int {
	return numSCollected
}
