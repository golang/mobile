// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl

//#include <GLES2/gl2.h>
import "C"

// Documented in gl.go.
func f32GL4(f float32) C.GLclampf  { return C.GLclampf(f) }
func f32None(f float32) C.GLclampf { return C.GLclampf(f) }
func f32All(f float32) C.GLfloat   { return C.GLfloat(f) }
