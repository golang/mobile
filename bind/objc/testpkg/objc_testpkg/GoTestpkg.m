
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Output of gobind -lang=objc

#include "GoTestpkg.h"
#include <Foundation/Foundation.h>
#include "seq.h"

#define _DESCRIPTOR_ "testpkg"

#define _CALL_BytesAppend_ 1
#define _CALL_Hello_ 2
#define _CALL_Hi_ 3
#define _CALL_Int_ 4
#define _CALL_Sum_ 5

#define INIT_SEQ(X)                                                            \
  GoSeq X;                                                                     \
  X.mem_ptr = NULL;

#define FREE_SEQ(X) go_seq_free(&X);

NSData *GoTestpkg_BytesAppend(NSData *a, NSData *b) {
  INIT_SEQ(in);
  INIT_SEQ(out);
  go_seq_writeByteArray(&in, a);
  go_seq_writeByteArray(&in, b);
  go_seq_send(_DESCRIPTOR_, _CALL_BytesAppend_, &in, &out);

  NSData *ret = go_seq_readByteArray(&out);
  FREE_SEQ(out);
  FREE_SEQ(in);
  return ret;
}

void GoTestpkg_Hi() {
  // No input, output.
  go_seq_send(_DESCRIPTOR_, _CALL_Hi_, NULL, NULL);
}

void GoTestpkg_Int(int32_t x) {
  INIT_SEQ(in);
  go_seq_writeInt32(&in, x);
  go_seq_send(_DESCRIPTOR_, _CALL_Int_, &in, NULL);
  FREE_SEQ(in);
}

int64_t GoTestpkg_Sum(int64_t x, int64_t y) {
  INIT_SEQ(in);
  INIT_SEQ(out);
  go_seq_writeInt64(&in, x);
  go_seq_writeInt64(&in, y);
  go_seq_send(_DESCRIPTOR_, _CALL_Sum_, &in, &out);
  int64_t res = go_seq_readInt64(&out);
  FREE_SEQ(out);
  FREE_SEQ(in);
  return res;
}

NSString *GoTestpkg_Hello(NSString *s) {
  INIT_SEQ(in);
  INIT_SEQ(out);
  go_seq_writeUTF8(&in, s);
  go_seq_send(_DESCRIPTOR_, _CALL_Hello_, &in, &out);
  NSString *res = go_seq_readUTF8(&out);
  FREE_SEQ(in);
  FREE_SEQ(out);
  return res;
}
