// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objc

import (
	"ObjC/Foundation"
	"ObjC/Foundation/NSMutableString"
	"ObjC/NetworkExtension/NEPacket"
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

const NSUTF8StringEncoding = 8

func CreateReadNSMutableString() {
	myData := []byte{'A', 'B'}
	// Test byte slices. Use NSMutableString because NSString is
	// represented as Go strings in bindings.
	// Pass slice from Go to native.
	mString := NSMutableString.NewWithData(myData, uint(NSUTF8StringEncoding))
	// Pass slice from native to Go.
	_ = mString.DataUsingEncoding(uint(NSUTF8StringEncoding))
}

// From <sys/socket.h>
const PF_INET = 2

func CallUcharFunction() {
	_ = NEPacket.NewWithData(nil, uint8(PF_INET))
}
