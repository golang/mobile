// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package java // import "golang.org/x/mobile/bind/java"

//#cgo LDFLAGS: -llog
//#include <android/log.h>
//#include <jni.h>
//#include <stdint.h>
//#include <string.h>
//#include "seq_android.h"
import "C"
import (
	"fmt"
	"unsafe"

	"golang.org/x/mobile/bind/seq"
	"golang.org/x/mobile/internal/mobileinit"
)

const maxSliceLen = 1<<31 - 1

const debug = false

// Send is called by Java to send a request to run a Go function.
//export Send
func Send(descriptor string, code int, req *C.uint8_t, reqlen C.size_t, res **C.uint8_t, reslen *C.size_t) {
	fn := seq.Registry[descriptor][code]
	if fn == nil {
		panic(fmt.Sprintf("invalid descriptor(%s) and code(0x%x)", descriptor, code))
	}

	var in, out *seq.Buffer
	if req != nil && reqlen > 0 {
		in = &seq.Buffer{
			Data: (*[maxSliceLen]byte)(unsafe.Pointer(req))[:reqlen],
		}
	}
	if res != nil {
		out = new(seq.Buffer)
	}

	fn(out, in)

	if res != nil {
		// BUG(hyangah): the function returning a go byte slice (so fn writes a pointer into 'out') is unsafe.
		// After fn is complete here, Go runtime is free to collect or move the pointed byte slice
		// contents. (Explicitly calling runtime.GC here will surface the problem?)
		// Without pinning support from Go side, it will be hard to fix it without extra copying.
		seqToBuf(res, reslen, out)
	}
}

// DestroyRef is called by Java to inform Go it is done with a reference.
//export DestroyRef
func DestroyRef(refnum C.int32_t) {
	seq.Delete(int32(refnum))
}

func seqToBuf(bufptr **C.uint8_t, lenptr *C.size_t, buf *seq.Buffer) {
	if debug {
		fmt.Printf("seqToBuf tag 1, len(buf.Data)=%d, *lenptr=%d\n", len(buf.Data), *lenptr)
	}
	if len(buf.Data) == 0 {
		*lenptr = 0
		return
	}
	if len(buf.Data) > int(*lenptr) {
		// TODO(crawshaw): realloc
		C.free(unsafe.Pointer(*bufptr))
		m := C.malloc(C.size_t(len(buf.Data)))
		if uintptr(m) == 0 {
			panic(fmt.Sprintf("malloc failed, size=%d", len(buf.Data)))
		}
		*bufptr = (*C.uint8_t)(m)
		*lenptr = C.size_t(len(buf.Data))
	}
	C.memcpy(unsafe.Pointer(*bufptr), unsafe.Pointer(&buf.Data[0]), C.size_t(len(buf.Data)))
}

// transact calls a method on a Java object instance.
// It blocks until the call is complete.
func transact(ref *seq.Ref, _ string, code int, inBuf *seq.Buffer) *seq.Buffer {
	var (
		out    *C.uint8_t = nil
		outLen C.size_t   = 0
		in     *C.uint8_t = nil
		inLen  C.size_t   = 0
	)

	if len(inBuf.Data) > 0 {
		in = (*C.uint8_t)(unsafe.Pointer(&inBuf.Data[0]))
		inLen = C.size_t(len(inBuf.Data))
	}

	C.recv(C.int32_t(ref.Num), C.int(code), in, inLen, &out, &outLen)
	if outLen > 0 {
		outBuf := &seq.Buffer{
			Data: make([]byte, outLen),
		}
		copy(outBuf.Data, (*[maxSliceLen]byte)(unsafe.Pointer(out))[:outLen])
		return outBuf
	}
	return nil
}

func encodeString(out *seq.Buffer, v string) {
	out.WriteUTF16(v)
}

func decodeString(in *seq.Buffer) string {
	return in.ReadUTF16()
}

func init() {
	seq.FinalizeRef = func(ref *seq.Ref) {
		if ref.Num < 0 {
			panic(fmt.Sprintf("not a Java ref: %d", ref.Num))
		}
		transact(ref, "", -1, new(seq.Buffer))
	}

	seq.Transact = transact
	seq.EncString = encodeString
	seq.DecString = decodeString
}

//export setContext
func setContext(vm *C.JavaVM, ctx C.jobject) {
	mobileinit.SetCurrentContext(unsafe.Pointer(vm), unsafe.Pointer(ctx))
}
