// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin
// +build arm arm64

#include "_cgo_export.h"
#include <pthread.h>
#include <stdio.h>
#include <sys/utsname.h>

#import <UIKit/UIKit.h>
#import <GLKit/GLKit.h>

struct utsname sysInfo;

@interface GoAppAppController : GLKViewController
@end

@interface GoAppAppDelegate : UIResponder<UIApplicationDelegate>
@property (strong, nonatomic) UIWindow *window;
@property (strong, nonatomic) GoAppAppController *controller;
@end

@implementation GoAppAppDelegate
- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
	self.window = [[UIWindow alloc] initWithFrame:[[UIScreen mainScreen] bounds]];
	self.controller = [[GoAppAppController alloc] initWithNibName:nil bundle:nil];
	self.window.rootViewController = self.controller;
	[self.window makeKeyAndVisible];
	return YES;
}
@end

@interface GoAppAppController ()
@property (strong, nonatomic) EAGLContext *context;
@end

@implementation GoAppAppController
- (void)viewDidLoad {
	[super viewDidLoad];
	self.context = [[EAGLContext alloc] initWithAPI:kEAGLRenderingAPIOpenGLES2];
	GLKView *view = (GLKView *)self.view;
	view.context = self.context;
	view.drawableDepthFormat = GLKViewDrawableDepthFormat24;
	view.multipleTouchEnabled = true; // TODO expose setting to user.
}

- (void)update {
	GLKView *view = (GLKView *)self.view;
	int w = [view drawableWidth];
	int h = [view drawableHeight];
	setGeom(w, h);

	drawgl((GoUintptr)self.context);
}

#define TOUCH_START 0 // event.TouchStart
#define TOUCH_MOVE  1 // event.TouchMove
#define TOUCH_END   2 // event.TouchEnd

static void sendTouches(int ty, NSSet* touches) {
	CGFloat scale = [UIScreen mainScreen].scale;
	for (UITouch* touch in touches) {
		CGPoint p = [touch locationInView:touch.view];
		sendTouch((GoUintptr)touch, ty, p.x*scale, p.y*scale);
	}
}

- (void)touchesBegan:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_START, touches);
}

- (void)touchesMoved:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_MOVE, touches);
}

- (void)touchesEnded:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_END, touches);
}
@end

void runApp(void) {
	@autoreleasepool {
		UIApplicationMain(0, nil, nil, NSStringFromClass([GoAppAppDelegate class]));
	}
}

void setContext(void* context) {
	EAGLContext* ctx = (EAGLContext*)context;
	if (![EAGLContext setCurrentContext:ctx]) {
		// TODO(crawshaw): determine how terrible this is. Exit?
		NSLog(@"failed to set current context");
	}
}

uint64_t threadID() {
	uint64_t id;
	if (pthread_threadid_np(pthread_self(), &id)) {
		abort();
	}
	return id;
}
