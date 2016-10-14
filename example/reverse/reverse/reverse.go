// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package reverse implements an Android in Go
package reverse

import (
	"Java/android/databinding/DataBindingUtil"
	"Java/android/os"
	"Java/android/support/v7/app"
	rlayout "Java/go/reverse/R/layout"
	"Java/go/reverse/databinding/ActivityMainBinding"
)

type MainActivity struct {
	app.AppCompatActivity
}

func (a *MainActivity) OnCreate1(this app.AppCompatActivity, b os.Bundle) {
	this.Super().OnCreate1(b)
	db := DataBindingUtil.SetContentView2(this, rlayout.Activity_main)
	mainBind := ActivityMainBinding.Cast(db)
	mainBind.SetAct(this)
}

func (a *MainActivity) GetLabel() string {
	return "Hello from Go!"
}
