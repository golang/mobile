// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package javapkg

import (
	"Java/java/beans"
	"Java/java/io"
	"Java/java/io/IOException"
	"Java/java/lang"
	"Java/java/lang/Character"
	"Java/java/lang/Integer"
	"Java/java/lang/Object"
	"Java/java/lang/Runnable"
	"Java/java/util"
	"Java/java/util/concurrent"
)

const (
	ToStringPrefix     = "Go toString: "
	IOExceptionMessage = "GoInputStream IOException"
)

type GoRunnable struct {
	lang.Object
	lang.Runnable
	this lang.Runnable

	Field string
}

func (r *GoRunnable) ToString(this lang.Runnable) string {
	return ToStringPrefix
}

func (r *GoRunnable) Run(this lang.Runnable) {
}

func (r *GoRunnable) GetThis(this lang.Runnable) lang.Runnable {
	return this
}

type GoInputStream struct {
	io.InputStream
}

func (_ *GoInputStream) Read() (int32, error) {
	return 0, IOException.New_Ljava_lang_String_2(IOExceptionMessage)
}

func NewGoInputStream() *GoInputStream {
	return new(GoInputStream)
}

type GoFuture struct {
	concurrent.Future
}

func (_ *GoFuture) Cancel(_ bool) bool {
	return false
}

func (_ *GoFuture) Get() lang.Object {
	return nil
}

func (_ *GoFuture) Get2(_ int64, _ concurrent.TimeUnit) lang.Object {
	return nil
}

func (_ *GoFuture) IsCancelled() bool {
	return false
}

func (_ *GoFuture) IsDone() bool {
	return false
}

type GoObject struct {
	lang.Object
	this lang.Object
}

func (o *GoObject) ToString(this lang.Object) string {
	o.this = this
	return ToStringPrefix + this.Super().ToString()
}

func (r *GoObject) CheckClass() bool {
	// Verify that GetClass returns interface{} because java.lang.Class
	// is unreferenced.
	var f func() interface{} = r.this.GetClass
	// But it should return a value
	return f() != nil
}

func (_ *GoObject) HashCode() int32 {
	return 42
}

func RunRunnable(r lang.Runnable) {
	r.Run()
}

func RunnableRoundtrip(r lang.Runnable) lang.Runnable {
	return r
}

// Test constructing and returning Go instances of GoObject and GoRunnable
// outside a constructor
func ConstructGoRunnable() *GoRunnable {
	return new(GoRunnable)
}

func ConstructGoObject() *GoObject {
	return new(GoObject)
}

// java.beans.PropertyChangeEvent is a class a with no default constructors.
type GoPCE struct {
	beans.PropertyChangeEvent
}

func NewGoPCE(_ lang.Object, _ string, _ lang.Object, _ lang.Object) *GoPCE {
	return new(GoPCE)
}

// java.util.ArrayList is a class with multiple constructors
type GoArrayList struct {
	util.ArrayList
}

func NewGoArrayList() *GoArrayList {
	return new(GoArrayList)
}

func NewGoArrayListWithCap(_ int32) *GoArrayList {
	return new(GoArrayList)
}

func CallSubset(s Character.Subset) {
	s.ToString()
}

type GoSubset struct {
	Character.Subset
}

func NewGoSubset(_ string) *GoSubset {
	return new(GoSubset)
}

func NewJavaObject() lang.Object {
	return Object.New()
}

func NewJavaInteger() lang.Integer {
	return Integer.New_I(42)
}

type NoargConstructor struct {
	util.BitSet // An otherwise unused class with a no-arg constructor
}

type GoRand struct {
	util.Random
}

func (_ *GoRand) Next(this util.Random, i int32) int32 {
	return this.Super().Next(i)
}

type I interface{}

func CastInterface(intf I) lang.Runnable {
	var r lang.Runnable = Runnable.Cast(intf)
	r.Run()
	return r
}

func CastRunnable(o lang.Object) lang.Runnable {
	defer func() {
		recover() // swallow the panic
	}()
	var r lang.Runnable = Runnable.Cast(o)
	r.Run()
	return r
}
