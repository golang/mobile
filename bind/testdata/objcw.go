// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objc

import (
	"ObjC/Foundation"
	"ObjC/UIKit"
)

type GoNSDate struct {
	Foundation.NSDate
	this Foundation.NSDate
}

func (d *GoNSDate) Hash(this Foundation.NSDate) int {
	return 0
}

type GoNSObject struct {
	C Foundation.NSObjectC // The class
	P Foundation.NSObjectP // The protocol
}

func (o *GoNSObject) Description(this Foundation.NSObjectC) string {
	return ""
}

func DupNSDate(date Foundation.NSDate) Foundation.NSDate {
	return date
}

type GoUIResponder struct {
	UIKit.UIResponder
}

func (r *GoUIResponder) PressesBegan(_ Foundation.NSSet, _ UIKit.UIPressesEvent) {
}
