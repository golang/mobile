// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Output of gobind -lang=objc

#ifndef __GoTestpkg_H__
#define __GoTestpkg_H__

#include "seq.h"

@interface GoTestpkg_S : NSObject {
}
@property(strong, readonly) GoSeqRef *ref;
- (id)initWithRef:(GoSeqRef*)ref;
- (double)X;
- (void)setX:(double)x;
- (double)Y;
- (void)setY:(double)y;
- (double)Sum;
@end

FOUNDATION_EXPORT NSData *GoTestpkg_BytesAppend(NSData *a, NSData *b);
FOUNDATION_EXPORT double GoTestpkg_CallSSum(GoTestpkg_S* s);
FOUNDATION_EXPORT int GoTestpkg_CollectS(int want, int timeoutSec);
FOUNDATION_EXPORT NSString *GoTestpkg_Hello(NSString *s);
FOUNDATION_EXPORT void GoTestpkg_Hi();
FOUNDATION_EXPORT void GoTestpkg_Int(int32_t x);
FOUNDATION_EXPORT GoTestpkg_S *GoTestpkg_NewS(double x, double y);
FOUNDATION_EXPORT int64_t GoTestpkg_Sum(int64_t x, int64_t y);

#endif
