// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!android

package app

/*
Simple on-screen app debugging for X11. Not an officially supported
development target for apps, as screens with mice are very different
than screens with touch panels.
*/

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11

#include <X11/Xlib.h>
#include <X11/keysym.h>

extern XIC x_ic;

void createWindow(void);
void processEvents(void);
void swapBuffers(void);
*/
import "C"
import (
	"runtime"
	"time"

	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/geom"
)

func init() {
	theApp.registerGLViewportFilter()
}

func main(f func(App)) {
	runtime.LockOSThread()

	workAvailable := theApp.worker.WorkAvailable()

	C.createWindow()

	// TODO: send lifecycle events when e.g. the X11 window is iconified or moved off-screen.
	theApp.sendLifecycle(lifecycle.StageFocused)

	// TODO: translate X11 expose events to shiny paint events, instead of
	// sending this synthetic paint event as a hack.
	theApp.eventsIn <- paint.Event{}

	donec := make(chan struct{})
	go func() {
		f(theApp)
		close(donec)
	}()

	// TODO: can we get the actual vsync signal?
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	var tc <-chan time.Time

	for {
		select {
		case <-donec:
			return
		case <-workAvailable:
			theApp.worker.DoWork()
		case <-theApp.publish:
			C.swapBuffers()
			tc = ticker.C
		case <-tc:
			tc = nil
			theApp.publishResult <- PublishResult{}
		}
		C.processEvents()
	}
}

//export onResize
func onResize(w, h int) {
	// TODO(nigeltao): don't assume 72 DPI. DisplayWidth and DisplayWidthMM
	// is probably the best place to start looking.
	pixelsPerPt := float32(1)
	theApp.eventsIn <- size.Event{
		WidthPx:     w,
		HeightPx:    h,
		WidthPt:     geom.Pt(w),
		HeightPt:    geom.Pt(h),
		PixelsPerPt: pixelsPerPt,
	}
}

func sendTouch(t touch.Type, x, y float32) {
	theApp.eventsIn <- touch.Event{
		X:        x,
		Y:        y,
		Sequence: 0, // TODO: button??
		Type:     t,
	}
}

//export onTouchBegin
func onTouchBegin(x, y float32) { sendTouch(touch.TypeBegin, x, y) }

//export onTouchMove
func onTouchMove(x, y float32) { sendTouch(touch.TypeMove, x, y) }

//export onTouchEnd
func onTouchEnd(x, y float32) { sendTouch(touch.TypeEnd, x, y) }

var keyMap = map[C.KeySym]key.Code{
	C.XK_a: key.CodeA,
	C.XK_b: key.CodeB,
	C.XK_c: key.CodeC,
	C.XK_d: key.CodeD,
	C.XK_e: key.CodeE,
	C.XK_f: key.CodeF,
	C.XK_g: key.CodeG,
	C.XK_h: key.CodeH,
	C.XK_i: key.CodeI,
	C.XK_j: key.CodeJ,
	C.XK_k: key.CodeK,
	C.XK_l: key.CodeL,
	C.XK_m: key.CodeM,
	C.XK_n: key.CodeN,
	C.XK_o: key.CodeO,
	C.XK_p: key.CodeP,
	C.XK_q: key.CodeQ,
	C.XK_r: key.CodeR,
	C.XK_s: key.CodeS,
	C.XK_t: key.CodeT,
	C.XK_u: key.CodeU,
	C.XK_v: key.CodeV,
	C.XK_w: key.CodeW,
	C.XK_x: key.CodeX,
	C.XK_y: key.CodeY,
	C.XK_z: key.CodeZ,
	C.XK_A: key.CodeA,
	C.XK_B: key.CodeB,
	C.XK_C: key.CodeC,
	C.XK_D: key.CodeD,
	C.XK_E: key.CodeE,
	C.XK_F: key.CodeF,
	C.XK_G: key.CodeG,
	C.XK_H: key.CodeH,
	C.XK_I: key.CodeI,
	C.XK_J: key.CodeJ,
	C.XK_K: key.CodeK,
	C.XK_L: key.CodeL,
	C.XK_M: key.CodeM,
	C.XK_N: key.CodeN,
	C.XK_O: key.CodeO,
	C.XK_P: key.CodeP,
	C.XK_Q: key.CodeQ,
	C.XK_R: key.CodeR,
	C.XK_S: key.CodeS,
	C.XK_T: key.CodeT,
	C.XK_U: key.CodeU,
	C.XK_V: key.CodeV,
	C.XK_W: key.CodeW,
	C.XK_X: key.CodeX,
	C.XK_Y: key.CodeY,
	C.XK_Z: key.CodeZ,

	C.XK_1: key.Code1,
	C.XK_2: key.Code2,
	C.XK_3: key.Code3,
	C.XK_4: key.Code4,
	C.XK_5: key.Code5,
	C.XK_6: key.Code6,
	C.XK_7: key.Code7,
	C.XK_8: key.Code8,
	C.XK_9: key.Code9,
	C.XK_0: key.Code0,

	C.XK_Return:       key.CodeReturnEnter,
	C.XK_Escape:       key.CodeEscape,
	C.XK_BackSpace:    key.CodeDeleteBackspace,
	C.XK_Tab:          key.CodeTab,
	C.XK_space:        key.CodeSpacebar,
	C.XK_minus:        key.CodeHyphenMinus,
	C.XK_equal:        key.CodeEqualSign,
	C.XK_bracketleft:  key.CodeLeftSquareBracket,
	C.XK_bracketright: key.CodeRightSquareBracket,
	C.XK_backslash:    key.CodeBackslash,
	C.XK_semicolon:    key.CodeSemicolon,
	C.XK_apostrophe:   key.CodeApostrophe,
	C.XK_grave:        key.CodeGraveAccent,
	C.XK_comma:        key.CodeComma,
	C.XK_period:       key.CodeFullStop,
	C.XK_slash:        key.CodeSlash,
	C.XK_Caps_Lock:    key.CodeCapsLock,

	C.XK_F1:  key.CodeF1,
	C.XK_F2:  key.CodeF2,
	C.XK_F3:  key.CodeF3,
	C.XK_F4:  key.CodeF4,
	C.XK_F5:  key.CodeF5,
	C.XK_F6:  key.CodeF6,
	C.XK_F7:  key.CodeF7,
	C.XK_F8:  key.CodeF8,
	C.XK_F9:  key.CodeF9,
	C.XK_F10: key.CodeF10,
	C.XK_F11: key.CodeF11,
	C.XK_F12: key.CodeF12,

	C.XK_Pause:     key.CodePause,
	C.XK_Insert:    key.CodeInsert,
	C.XK_Home:      key.CodeHome,
	C.XK_Page_Up:   key.CodePageUp,
	C.XK_Delete:    key.CodeDeleteForward,
	C.XK_End:       key.CodeEnd,
	C.XK_Page_Down: key.CodePageDown,

	C.XK_Right: key.CodeRightArrow,
	C.XK_Left:  key.CodeLeftArrow,
	C.XK_Down:  key.CodeDownArrow,
	C.XK_Up:    key.CodeUpArrow,

	C.XK_Num_Lock:    key.CodeKeypadNumLock,
	C.XK_KP_Divide:   key.CodeKeypadSlash,
	C.XK_KP_Multiply: key.CodeKeypadAsterisk,
	C.XK_KP_Subtract: key.CodeKeypadHyphenMinus,
	C.XK_KP_Add:      key.CodeKeypadPlusSign,
	C.XK_KP_Enter:    key.CodeKeypadEnter,
	C.XK_KP_1:        key.CodeKeypad1,
	C.XK_KP_2:        key.CodeKeypad2,
	C.XK_KP_3:        key.CodeKeypad3,
	C.XK_KP_4:        key.CodeKeypad4,
	C.XK_KP_5:        key.CodeKeypad5,
	C.XK_KP_6:        key.CodeKeypad6,
	C.XK_KP_7:        key.CodeKeypad7,
	C.XK_KP_8:        key.CodeKeypad8,
	C.XK_KP_9:        key.CodeKeypad9,
	C.XK_KP_0:        key.CodeKeypad0,
	C.XK_KP_Decimal:  key.CodeKeypadFullStop,
	C.XK_KP_Equal:    key.CodeKeypadEqualSign,

	C.XK_F13: key.CodeF13,
	C.XK_F14: key.CodeF14,
	C.XK_F15: key.CodeF15,
	C.XK_F16: key.CodeF16,
	C.XK_F17: key.CodeF17,
	C.XK_F18: key.CodeF18,
	C.XK_F19: key.CodeF19,
	C.XK_F20: key.CodeF20,
	C.XK_F21: key.CodeF21,
	C.XK_F22: key.CodeF22,
	C.XK_F23: key.CodeF23,
	C.XK_F24: key.CodeF24,

	C.XK_Help: key.CodeHelp,

	C.XK_Control_L: key.CodeLeftControl,
	C.XK_Shift_L:   key.CodeLeftShift,
	C.XK_Alt_L:     key.CodeLeftAlt,
	C.XK_Super_L:   key.CodeLeftGUI,
	C.XK_Control_R: key.CodeRightControl,
	C.XK_Shift_R:   key.CodeRightShift,
	C.XK_Alt_R:     key.CodeRightAlt,
	C.XK_Super_R:   key.CodeRightGUI,

	C.XK_Multi_key: key.CodeCompose,
}

//export onKeyEvent
func onKeyEvent(ce *C.XKeyEvent) {
	e := key.Event{}
	// TODO: detect repeats
	if ce._type == C.KeyPress {
		e.Direction = key.DirPress
	} else {
		e.Direction = key.DirRelease
		ce._type = C.KeyPress
	}

	if ce.state&C.ShiftMask != 0 {
		e.Modifiers |= key.ModShift
	}
	if ce.state&C.ControlMask != 0 {
		e.Modifiers |= key.ModControl
	}
	if ce.state&C.Mod1Mask != 0 {
		e.Modifiers |= key.ModAlt
	}

	var keysym C.KeySym
	var status C.Status
	var text [32]C.char
	C.Xutf8LookupString(C.x_ic, ce, &text[0], C.int(len(text)-1), &keysym, &status)

	if status == C.XBufferOverflow {
		// more than 32 bytes for one event - ignore this unlikely case
		return
	}

	if status == C.XLookupChars || status == C.XLookupBoth {
		e.Rune = []rune(C.GoStringN(&text[0], C.int(len(text))))[0]
	}

	if status == C.XLookupKeySym || status == C.XLookupBoth {
		e.Code = keyMap[keysym]
	}

	theApp.eventsIn <- e
}

var stopped bool

//export onStop
func onStop() {
	if stopped {
		return
	}
	stopped = true
	theApp.sendLifecycle(lifecycle.StageDead)
	theApp.eventsIn <- stopPumping{}
}
