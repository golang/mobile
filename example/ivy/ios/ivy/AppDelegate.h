// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#import <WebKit/WebKit.h>
#import "IvyController.h"
#import "Suggestion.h"

@interface AppDelegate : UIResponder <UIApplicationDelegate, UITextFieldDelegate, WKUIDelegate>

@property(strong, nonatomic) UIWindow *window;

@end

