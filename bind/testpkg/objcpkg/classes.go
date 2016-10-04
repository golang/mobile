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
}

func (d *GoNSDate) Hash(self Foundation.NSDate) int {
	return Hash
}

func (d *GoNSDate) Description(self Foundation.NSDate) string {
	// Test self call
	if h := self.Hash(); h != Hash {
		panic("hash mismatch")
	}
	return DescriptionStr
}

func (d *GoNSDate) GetSelf(self Foundation.NSDate) Foundation.NSDate {
	return self
}

func NewGoNSDate() *GoNSDate {
	return new(GoNSDate)
}

type GoNSObject struct {
	C       Foundation.NSObjectC // The class
	P       Foundation.NSObjectP // The protocol
	UseSelf bool
}

func (o *GoNSObject) Description(self Foundation.NSObjectC) string {
	if o.UseSelf {
		return DescriptionStr
	} else {
		return self.Super().Description()
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
