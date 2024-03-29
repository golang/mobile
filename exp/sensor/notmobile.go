// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (linux && !android) || (darwin && !arm && !arm64) || windows

package sensor

import (
	"errors"
	"time"
)

func enable(t Type, delay time.Duration) error {
	return errors.New("sensor: no sensors available")
}

func disable(t Type) error {
	return errors.New("sensor: no sensors available")
}

func destroy() error {
	return nil
}
