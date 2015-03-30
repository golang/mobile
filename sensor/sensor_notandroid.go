// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !android

package sensor

import "time"

type manager struct {
}

func (m *manager) initialize() {
	panic("not implemented")
}

func (m *manager) enable(t Type, delay time.Duration) error {
	panic("not implemented")
}

func (m *manager) disable(t Type) error {
	panic("not implemented")
}

func (m *manager) read(e []Event) (n int, err error) {
	panic("not implemented")
}

func (m *manager) close() error {
	panic("not implemented")
}
