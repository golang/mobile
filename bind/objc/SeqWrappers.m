// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

@import ObjectiveC.message;
@import Foundation;
@import XCTest;
@import Objcpkg;

@interface TestNSObject : NSObject

- (NSString *)description;
- (NSString *)super_description;

@end

@implementation TestNSObject

- (NSString *)description {
	return @"hej";
}

- (NSString *)super_description {
	return [super description];
}

@end

@interface wrappers : XCTestCase

@end

@implementation wrappers

- (void)setUp {
    [super setUp];
    // Put setup code here. This method is called before the invocation of each test method in the class.
}

- (void)tearDown {
    // Put teardown code here. This method is called after the invocation of each test method in the class.
    [super tearDown];
}

- (void)testFunction {
	GoObjcpkgFunc();
}

- (void)testMethod {
	GoObjcpkgMethod();
}

- (void)testNew {
	GoObjcpkgNew();
}

- (void)testError {
	GoObjcpkgError();
}
@end
