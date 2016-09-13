// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objcpkg

import (
	"ObjC/Foundation"
	"ObjC/UIKit"
)

const (
	DescriptionStr = "Descriptrion from Go"
	Hash           = 42
)

type GoNSDate struct {
	Foundation.NSDate
	this Foundation.NSDate
}

func (d *GoNSDate) Hash(this Foundation.NSDate) int {
	return Hash
}

func (d *GoNSDate) Description(this Foundation.NSDate) string {
	// Test this call
	if h := this.Hash(); h != Hash {
		panic("hash mismatch")
	}
	d.this = this
	return DescriptionStr
}

func (d *GoNSDate) This() Foundation.NSDate {
	return d.this
}

func NewGoNSDate() *GoNSDate {
	return new(GoNSDate)
}

type GoNSObject struct {
	C       Foundation.NSObjectC // The class
	P       Foundation.NSObjectP // The protocol
	UseThis bool
}

func (o *GoNSObject) Description(this Foundation.NSObjectC) string {
	if o.UseThis {
		return DescriptionStr
	} else {
		return this.Super().Description()
	}
}

func DupNSDate(date Foundation.NSDate) Foundation.NSDate {
	return date
}

type GoUIResponder struct {
	UIKit.UIResponder
	Called bool
}

func (r *GoUIResponder) PressesBegan(_ Foundation.NSSet, _ UIKit.UIPressesEvent) {
	r.Called = true
}
