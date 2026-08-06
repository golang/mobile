// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package font

import "os"

func buildDefault() ([]byte, error) {
	return os.ReadFile("/system/fonts/DroidSans.ttf")
}

func buildMonospace() ([]byte, error) {
	return os.ReadFile("/system/fonts/DroidSansMono.ttf")
}
