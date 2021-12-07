// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobileinit

import (
	"io"
	"log"
	"os"
	"unsafe"
)

/*
#include <stdlib.h>
#include <os/log.h>

os_log_t create_os_log() {
	return os_log_create("org.golang.mobile", "os_log");
}

void os_log_wrap(os_log_t log, const char *str) {
	os_log(log, "%s", str);
}
*/
import "C"

type osWriter struct {
	w C.os_log_t
}

func (o osWriter) Write(p []byte) (n int, err error) {
	cstr := C.CString(string(p))
	C.os_log_wrap(o.w, cstr)
	C.free(unsafe.Pointer(cstr))
	return len(p), nil
}

func init() {
	log.SetOutput(io.MultiWriter(os.Stderr, osWriter{C.create_os_log()}))
}
