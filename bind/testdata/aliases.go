// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aliases

type AByte = byte
type AByteSlice = []AByte
type S struct{}
type AStruct = S
type AStructPtr = *AStruct

func TakesAByte(a AByte)                {}
func TakesAByteSlice(a []AByte)         {}
func TakesAByteSliceAlias(a AByteSlice) {}
func TakesAStructPtr(a *AStruct)        {}
func TakesAStructPtrAlias(a AStructPtr) {}
