// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

#include "_cgo_export.h"
#include <pthread.h>
#include <stdio.h>

#import <Cocoa/Cocoa.h>
#import <Foundation/Foundation.h>
#import <OpenGL/gl3.h>
#import <QuartzCore/CVReturn.h>
#import <QuartzCore/CVBase.h>

static CVReturn displayLinkDraw(CVDisplayLinkRef displayLink, const CVTimeStamp* now, const CVTimeStamp* outputTime, CVOptionFlags flagsIn, CVOptionFlags* flagsOut, void* displayLinkContext)
{
	drawgl();
	return kCVReturnSuccess;
}

void makeCurrentContext(GLintptr context) {
	NSOpenGLContext* ctx = (NSOpenGLContext*)context;
	[ctx makeCurrentContext];
}

uint64 threadID() {
	uint64 id;
	if (pthread_threadid_np(pthread_self(), &id)) {
		abort();
	}
	return id;
}


@interface MobileGLView : NSOpenGLView<NSApplicationDelegate, NSWindowDelegate>
{
	CVDisplayLinkRef displayLink;
}
@end

@implementation MobileGLView
- (void)prepareOpenGL {
	[self setWantsBestResolutionOpenGLSurface:YES];
	GLint swapInt = 1;
	[[self openGLContext] setValues:&swapInt forParameter:NSOpenGLCPSwapInterval];

	CVDisplayLinkCreateWithActiveCGDisplays(&displayLink);
	CVDisplayLinkSetOutputCallback(displayLink, &displayLinkDraw, self);

	CGLContextObj cglContext = [[self openGLContext] CGLContextObj];
	CGLPixelFormatObj cglPixelFormat = [[self pixelFormat] CGLPixelFormatObj];
	CVDisplayLinkSetCurrentCGDisplayFromOpenGLContext(displayLink, cglContext, cglPixelFormat);

	// Using attribute arrays in OpenGL 3.3 requires the use of a VBA.
	// But VBAs don't exist in ES 2. So we bind a default one.
	GLuint vba;
	glGenVertexArrays(1, &vba);
	glBindVertexArray(vba);

	startloop((GLintptr)[self openGLContext]);
}

- (void)reshape {
	[super reshape];

	// Calculate screen PPI.
	//
	// Note that the backingScaleFactor converts from logical
	// pixels to actual pixels, but both of these units vary
	// independently from real world size. E.g.
	//
	// 13" Retina Macbook Pro, 2560x1600, 227ppi, backingScaleFactor=2, scale=3.15
	// 15" Retina Macbook Pro, 2880x1800, 220ppi, backingScaleFactor=2, scale=3.06
	// 27" iMac,               2560x1440, 109ppi, backingScaleFactor=1, scale=1.51
	// 27" Retina iMac,        5120x2880, 218ppi, backingScaleFactor=2, scale=3.03
	NSScreen *screen = [NSScreen mainScreen];
	double screenPixW = [screen frame].size.width * [screen backingScaleFactor];

	CGDirectDisplayID display = (CGDirectDisplayID)[[[screen deviceDescription] valueForKey:@"NSScreenNumber"] intValue];
	CGSize screenSizeMM = CGDisplayScreenSize(display); // in millimeters
	float ppi = 25.4 * screenPixW / screenSizeMM.width;
	float pixelsPerPt = ppi/72.0;

	// The width and height reported to the geom package are the
	// bounds of the OpenGL view. Several steps are necessary.
	// First, [self bounds] gives us the number of logical pixels
	// in the view. Multiplying this by the backingScaleFactor
	// gives us the number of actual pixels.
	NSRect r = [self bounds];
	int w = r.size.width * [screen backingScaleFactor];
	int h = r.size.height * [screen backingScaleFactor];

	setGeom(pixelsPerPt, w, h);
}

- (void)drawRect:(NSRect)theRect {
	// Called during resize. Do an extra draw if we are visible.
	// This gets rid of flicker when resizing.
	if (CVDisplayLinkIsRunning(displayLink)) {
		drawgl();
	}
}

- (void)mouseDown:(NSEvent *)theEvent {
	double scale = [[NSScreen mainScreen] backingScaleFactor];
	NSPoint p = [theEvent locationInWindow];
	eventMouseDown(p.x * scale, p.y * scale);
}

- (void)mouseUp:(NSEvent *)theEvent {
	double scale = [[NSScreen mainScreen] backingScaleFactor];
	NSPoint p = [theEvent locationInWindow];
	eventMouseEnd(p.x * scale, p.y * scale);
}

- (void)mouseDragged:(NSEvent *)theEvent {
	double scale = [[NSScreen mainScreen] backingScaleFactor];
	NSPoint p = [theEvent locationInWindow];
	eventMouseDragged(p.x * scale, p.y * scale);
}

- (void)windowDidBecomeKey:(NSNotification *)notification {
	lifecycleFocused();
}

- (void)windowDidResignKey:(NSNotification *)notification {
	if (![NSApp isHidden]) {
		lifecycleVisible();
	}
}

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
	lifecycleAlive();
	[[NSRunningApplication currentApplication] activateWithOptions:(NSApplicationActivateAllWindows | NSApplicationActivateIgnoringOtherApps)];
	[self.window makeKeyAndOrderFront:self];
	lifecycleVisible();
	CVDisplayLinkStart(displayLink);
}

- (void)applicationWillTerminate:(NSNotification *)aNotification {
	lifecycleDead();
}

- (void)applicationDidHide:(NSNotification *)aNotification {
	CVDisplayLinkStop(displayLink);
	lifecycleAlive();
}

- (void)applicationWillUnhide:(NSNotification *)notification {
	lifecycleVisible();
	CVDisplayLinkStart(displayLink);
}

- (void)windowWillClose:(NSNotification *)notification {
	CVDisplayLinkStop(displayLink);
	lifecycleAlive();
}
@end

void runApp(void) {
	[NSAutoreleasePool new];
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

	id menuBar = [[NSMenu new] autorelease];
	id menuItem = [[NSMenuItem new] autorelease];
	[menuBar addItem:menuItem];
	[NSApp setMainMenu:menuBar];

	id menu = [[NSMenu new] autorelease];
	id name = [[NSProcessInfo processInfo] processName];

	id hideMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Hide"
		action:@selector(hide:) keyEquivalent:@"h"]
		autorelease];
	[menu addItem:hideMenuItem];

	id quitMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Quit"
		action:@selector(terminate:) keyEquivalent:@"q"]
		autorelease];
	[menu addItem:quitMenuItem];
	[menuItem setSubmenu:menu];

	NSRect rect = NSMakeRect(0, 0, 400, 400);

	NSWindow* window = [[[NSWindow alloc] initWithContentRect:rect
			styleMask:NSTitledWindowMask
			backing:NSBackingStoreBuffered
			defer:NO]
		autorelease];
	window.styleMask |= NSResizableWindowMask;
	window.styleMask |= NSMiniaturizableWindowMask ;
	window.styleMask |= NSClosableWindowMask;
	window.title = name;
	[window cascadeTopLeftFromPoint:NSMakePoint(20,20)];

	NSOpenGLPixelFormatAttribute attr[] = {
		NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
		NSOpenGLPFAColorSize,     24,
		NSOpenGLPFAAlphaSize,     8,
		NSOpenGLPFADepthSize,     16,
		NSOpenGLPFAAccelerated,
		NSOpenGLPFADoubleBuffer,
		NSOpenGLPFAAllowOfflineRenderers,
		0
	};
	id pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attr];
	MobileGLView* view = [[MobileGLView alloc] initWithFrame:rect pixelFormat:pixFormat];
	[window setContentView:view];
	[window setDelegate:view];
	[NSApp setDelegate:view];
	[NSApp run];
}

void stopApp(void) {
	[NSApp terminate:nil];
}
