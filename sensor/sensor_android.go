// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sensor

import "time"

type manager struct {
}

func enable(m *manager, t Type, delay time.Duration) error {
	panic("not implemented")
}

func disable(m *manager, t Type) error {
	panic("not implemented")
}

func read(m *manager, e []Event) (n int, err error) {
	panic("not implemented")
}

func close(m *manager) error {
	panic("not implemented")
}
