// Copyright 2021 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package dummy

// This is a dummy go file to add required module dependencies to go.mod.

import (
	_ "golang.org/x/mobile/bind"
	_ "robpike.io/ivy"
)
