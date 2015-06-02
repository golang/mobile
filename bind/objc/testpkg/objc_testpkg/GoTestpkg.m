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
#define _CALL_NewS_ 5
#define _CALL_Sum_ 6

NSData *GoTestpkg_BytesAppend(NSData *a, NSData *b) {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeByteArray(&in, a);
  go_seq_writeByteArray(&in, b);
  go_seq_send(_DESCRIPTOR_, _CALL_BytesAppend_, &in, &out);

  NSData *ret = go_seq_readByteArray(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return ret;
}

void GoTestpkg_Hi() {
  // No input, output.
  GoSeq in = {};
  GoSeq out = {};
  go_seq_send(_DESCRIPTOR_, _CALL_Hi_, &in, &out);
  go_seq_free(&out);
  go_seq_free(&in);
}

void GoTestpkg_Int(int32_t x) {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeInt32(&in, x);
  go_seq_send(_DESCRIPTOR_, _CALL_Int_, &in, &out);
  go_seq_free(&out);
  go_seq_free(&in);
}

int64_t GoTestpkg_Sum(int64_t x, int64_t y) {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeInt64(&in, x);
  go_seq_writeInt64(&in, y);
  go_seq_send(_DESCRIPTOR_, _CALL_Sum_, &in, &out);
  int64_t res = go_seq_readInt64(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return res;
}

NSString *GoTestpkg_Hello(NSString *s) {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeUTF8(&in, s);
  go_seq_send(_DESCRIPTOR_, _CALL_Hello_, &in, &out);
  NSString *res = go_seq_readUTF8(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return res;
}

GoTestpkg_S *GoTestpkg_NewS(double x, double y) {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeFloat64(&in, x);
  go_seq_writeFloat64(&in, y);
  go_seq_send(_DESCRIPTOR_, _CALL_NewS_, &in, &out);
  GoSeqRef *ref = go_seq_readRef(&out);
  GoTestpkg_S *ret = ref.obj;
  if (ret == NULL) { // Go object.
    ret = [[GoTestpkg_S alloc] initWithRef:ref];
  }
  ref = NULL;
  go_seq_free(&out);
  go_seq_free(&in);
  return ret;
}

#define _GO_TESTPKG_S_DESCRIPTOR_ "go.testpkg.S"
#define _GO_TESTPKG_S_FIELD_X_GET_ 0x00f
#define _GO_TESTPKG_S_FIELD_X_SET_ 0x01f
#define _GO_TESTPKG_S_FIELD_Y_GET_ 0x10f
#define _GO_TESTPKG_S_FIELD_Y_SET_ 0x11f
#define _GO_TESTPKG_S_SUM_ 0x00c

@implementation GoTestpkg_S {
}
- (double)X {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeRef(&in, self.ref);
  go_seq_send(_GO_TESTPKG_S_DESCRIPTOR_, _GO_TESTPKG_S_FIELD_X_GET_, &in, &out);
  double res = go_seq_readFloat64(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return res;
}

- (void)setX:(double)x {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeRef(&in, self.ref);
  go_seq_writeFloat64(&in, x);
  go_seq_send(_GO_TESTPKG_S_DESCRIPTOR_, _GO_TESTPKG_S_FIELD_X_SET_, &in, &out);
  go_seq_free(&out);
  go_seq_free(&in);
}

- (double)Y {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeRef(&in, self.ref);
  go_seq_send(_GO_TESTPKG_S_DESCRIPTOR_, _GO_TESTPKG_S_FIELD_Y_GET_, &in, &out);
  double res = go_seq_readFloat64(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return res;
}
- (void)setY:(double)y {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeRef(&in, self.ref);
  go_seq_writeFloat64(&in, y);
  go_seq_send(_GO_TESTPKG_S_DESCRIPTOR_, _GO_TESTPKG_S_FIELD_Y_SET_, &in, &out);
  go_seq_free(&out);
  go_seq_free(&in);
}

- (double)Sum {
  GoSeq in = {};
  GoSeq out = {};
  go_seq_writeRef(&in, self.ref);
  go_seq_send(_GO_TESTPKG_S_DESCRIPTOR_, _GO_TESTPKG_S_SUM_, &in, &out);
  double res = go_seq_readFloat64(&out);
  go_seq_free(&out);
  go_seq_free(&in);
  return res;
}

@end
