// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !android

package app

// ShowSoftKeyboard is a no-op on non-Android platforms.
// On Android, this shows the soft keyboard for text input.
func ShowSoftKeyboard() {
	// No-op on desktop/iOS
}

// HideSoftKeyboard is a no-op on non-Android platforms.
// On Android, this hides the soft keyboard.
func HideSoftKeyboard() {
	// No-op on desktop/iOS
}
