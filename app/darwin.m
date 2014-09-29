// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

#include "_cgo_export.h"
#include <stdio.h>

#import <Cocoa/Cocoa.h>
#import <Foundation/Foundation.h>
#import <OpenGL/gl.h>
#import <QuartzCore/CVReturn.h>
#import <QuartzCore/CVBase.h>

static CVReturn displayLinkDraw(CVDisplayLinkRef displayLink, const CVTimeStamp* now, const CVTimeStamp* outputTime, CVOptionFlags flagsIn, CVOptionFlags* flagsOut, void* displayLinkContext)
{
	NSOpenGLView* view = displayLinkContext;
	NSOpenGLContext *currentContext = [view openGLContext];
	drawgl((GLintptr)currentContext);
	return kCVReturnSuccess;
}

void lockContext(GLintptr context) {
	NSOpenGLContext* ctx = (NSOpenGLContext*)context;
	[ctx makeCurrentContext];
	CGLLockContext([ctx CGLContextObj]);
}

void unlockContext(GLintptr context) {
	NSOpenGLContext* ctx = (NSOpenGLContext*)context;
	[ctx flushBuffer];
	CGLUnlockContext([ctx CGLContextObj]);

}


@interface MobileGLView : NSOpenGLView
{
	CVDisplayLinkRef displayLink;
}
@end

@implementation MobileGLView
- (void)prepareOpenGL {
	[self setWantsBestResolutionOpenGLSurface:true];
	GLint swapInt = 1;
	[[self openGLContext] setValues:&swapInt forParameter:NSOpenGLCPSwapInterval];

	CVDisplayLinkCreateWithActiveCGDisplays(&displayLink);
	CVDisplayLinkSetOutputCallback(displayLink, &displayLinkDraw, self);

	CGLContextObj cglContext = [[self openGLContext] CGLContextObj];
	CGLPixelFormatObj cglPixelFormat = [[self pixelFormat] CGLPixelFormatObj];
	CVDisplayLinkSetCurrentCGDisplayFromOpenGLContext(displayLink, cglContext, cglPixelFormat);
	CVDisplayLinkStart(displayLink);
}

- (void)reshape {
	NSRect r = [self bounds];
	double scale = [[NSScreen mainScreen] backingScaleFactor];
	setGeom(scale, r.size.width, r.size.height);
}

- (void)mouseDown:(NSEvent *)theEvent {
	NSPoint p = [theEvent locationInWindow];
	eventMouseDown(p.x, p.y);
}

- (void)mouseUp:(NSEvent *)theEvent {
	NSPoint p = [theEvent locationInWindow];
	eventMouseEnd(p.x, p.y);
}

- (void)mouseMoved:(NSEvent *)theEvent {
	NSPoint p = [theEvent locationInWindow];
	eventMouseMove(p.x, p.y);
}
@end

void
runApp(void) {
	[NSAutoreleasePool new];
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

	id menuBar = [[NSMenu new] autorelease];
	id menuItem = [[NSMenuItem new] autorelease];
	[menuBar addItem:menuItem];
	[NSApp setMainMenu:menuBar];

	id menu = [[NSMenu new] autorelease];
	id name = [[NSProcessInfo processInfo] processName];
	id quitMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Quit"
		action:@selector(terminate:) keyEquivalent:@"q"]
		autorelease];
	[menu addItem:quitMenuItem];
	[menuItem setSubmenu:menu];

	NSRect rect = NSMakeRect(0, 0, 200, 200);

	id window = [[[NSWindow alloc] initWithContentRect:rect
			styleMask:NSTitledWindowMask
			backing:NSBackingStoreBuffered
			defer:NO]
		autorelease];
	[window setStyleMask:[window styleMask] | NSResizableWindowMask];
	[window cascadeTopLeftFromPoint:NSMakePoint(20,20)];
	[window makeKeyAndOrderFront:nil];
	[window setTitle:name];

	NSOpenGLPixelFormatAttribute attr[] = {
		NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
		NSOpenGLPFAColorSize,     24,
		NSOpenGLPFAAlphaSize,     8,
		NSOpenGLPFADepthSize,     16,
		NSOpenGLPFAAccelerated,
		NSOpenGLPFADoubleBuffer,
		0
	};
	id pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attr];
	id view = [[MobileGLView alloc] initWithFrame:rect pixelFormat:pixFormat];
	[window setContentView:view];

	[NSApp activateIgnoringOtherApps:YES];
	[NSApp run];
}
