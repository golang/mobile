// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Converted from the upstream OpenGL ES (GLKView) runtime to a Metal /
// CAMetalLayer runtime so the engine can drive a wgpu (WebGPU) surface.
// This mirrors the Android WebGPU bridge (android_wgpu.go). The view's
// backing layer is a CAMetalLayer; its pointer is handed to Go via
// setMetalLayer, and Go creates the wgpu surface from it.

//go:build darwin && ios
// +build darwin
// +build ios

#include "_cgo_export.h"
#include <pthread.h>
#include <stdio.h>
#include <sys/utsname.h>

#import <UIKit/UIKit.h>
#import <QuartzCore/CAMetalLayer.h>

struct utsname sysInfo;

// GoMetalView is a UIView whose backing layer is a CAMetalLayer. Overriding
// +layerClass is the canonical way to make a UIView render with Metal.
@interface GoMetalView : UIView
@end

@implementation GoMetalView
+ (Class)layerClass {
	return [CAMetalLayer class];
}
@end

@interface GoAppAppController : UIViewController
@property (strong, nonatomic) GoMetalView *metalView;
- (void)updateDrawableSizeAndNotify;
@end

@interface GoAppAppDelegate : UIResponder<UIApplicationDelegate>
@property (strong, nonatomic) UIWindow *window;
@property (strong, nonatomic) GoAppAppController *controller;
@end

@implementation GoAppAppDelegate
- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
	lifecycleAlive();
	self.window = [[UIWindow alloc] initWithFrame:[[UIScreen mainScreen] bounds]];
	self.controller = [[GoAppAppController alloc] initWithNibName:nil bundle:nil];
	self.window.rootViewController = self.controller;
	[self.window makeKeyAndVisible];
	return YES;
}

- (void)applicationDidBecomeActive:(UIApplication *)application {
	lifecycleFocused();
}

- (void)applicationWillResignActive:(UIApplication *)application {
	lifecycleVisible();
}

- (void)applicationDidEnterBackground:(UIApplication *)application {
	lifecycleAlive();
}

- (void)applicationWillTerminate:(UIApplication *)application {
	lifecycleDead();
}
@end

@implementation GoAppAppController
- (void)loadView {
	self.metalView = [[GoMetalView alloc] initWithFrame:[[UIScreen mainScreen] bounds]];
	self.metalView.multipleTouchEnabled = YES; // TODO expose setting to user.
	self.metalView.userInteractionEnabled = YES;
	self.view = self.metalView;
}

- (CAMetalLayer *)metalLayer {
	return (CAMetalLayer *)self.metalView.layer;
}

- (void)viewDidLoad {
	[super viewDidLoad];

	int scale = 1;
	if ([[UIScreen mainScreen] respondsToSelector:@selector(displayLinkWithTarget:selector:)]) {
		scale = (int)[UIScreen mainScreen].scale; // either 1.0, 2.0, or 3.0.
	}
	setScreen(scale);

	CAMetalLayer *layer = [self metalLayer];
	// Opaque, non-blended composition (matches the Android RGBX-opaque choice).
	layer.opaque = YES;
	layer.contentsScale = [UIScreen mainScreen].scale;
	// pixelFormat defaults to MTLPixelFormatBGRA8Unorm; wgpu negotiates the
	// surface format against the layer's capabilities.

	[self updateDrawableSizeAndNotify];
}

- (void)viewDidLayoutSubviews {
	[super viewDidLayoutSubviews];
	[self updateDrawableSizeAndNotify];
}

// updateDrawableSizeAndNotify sizes the CAMetalLayer's drawable to the current
// view bounds in pixels, hands the layer pointer to Go, and pushes a size event.
- (void)updateDrawableSizeAndNotify {
	CAMetalLayer *layer = [self metalLayer];
	CGFloat scale = [UIScreen mainScreen].scale;
	CGSize bounds = self.view.bounds.size;
	layer.drawableSize = CGSizeMake(bounds.width * scale, bounds.height * scale);

	// Hand the CAMetalLayer pointer to Go. Idempotent — Go drains stale events.
	setMetalLayer((void *)layer);

	UIInterfaceOrientation orientation = [[UIApplication sharedApplication] statusBarOrientation];
	updateConfig((int)bounds.width, (int)bounds.height, orientation);
}

- (void)viewWillTransitionToSize:(CGSize)size withTransitionCoordinator:(id<UIViewControllerTransitionCoordinator>)coordinator {
	[coordinator animateAlongsideTransition:^(id<UIViewControllerTransitionCoordinatorContext> context) {
		// TODO(crawshaw): come up with a plan to handle animations.
	} completion:^(id<UIViewControllerTransitionCoordinatorContext> context) {
		[self updateDrawableSizeAndNotify];
	}];
}

#define TOUCH_TYPE_BEGIN 0 // touch.TypeBegin
#define TOUCH_TYPE_MOVE  1 // touch.TypeMove
#define TOUCH_TYPE_END   2 // touch.TypeEnd

static void sendTouches(int change, NSSet* touches) {
	CGFloat scale = [UIScreen mainScreen].scale;
	for (UITouch* touch in touches) {
		CGPoint p = [touch locationInView:touch.view];
		sendTouch((GoUintptr)touch, (GoUintptr)change, p.x*scale, p.y*scale);
	}
}

- (void)touchesBegan:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_BEGIN, touches);
}

- (void)touchesMoved:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_MOVE, touches);
}

- (void)touchesEnded:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_END, touches);
}

- (void)touchesCanceled:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_END, touches);
}
@end

void runApp(void) {
	char* argv[] = {};
	@autoreleasepool {
		UIApplicationMain(0, argv, nil, NSStringFromClass([GoAppAppDelegate class]));
	}
}

uint64_t threadID() {
	uint64_t id;
	if (pthread_threadid_np(pthread_self(), &id)) {
		abort();
	}
	return id;
}
