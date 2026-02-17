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
#include <X11/keysym.h>
*/
import "C"
import (
	"golang.org/x/mobile/event/key"
)

//export onKey
func onKey(id uintptr, detail uint16, dir uint8) {
	var modifiers key.Modifiers
	theApp.eventsIn <- key.Event{
		Rune:      convRune(rune(id)),
		Code:      convVirtualKeyCode(detail),
		Modifiers: modifiers,
		Direction: key.Direction(dir),
	}
}

func convRune(r rune) rune {
	if '\uE000' <= r && r <= '\uF8FF' {
		return -1
	}
	return r
}

func convVirtualKeyCode(code uint16) key.Code {
	switch code {
	case C.XK_A:
		return key.CodeA
	case C.XK_B:
		return key.CodeB
	case C.XK_C:
		return key.CodeC
	case C.XK_D:
		return key.CodeD
	case C.XK_E:
		return key.CodeE
	case C.XK_F:
		return key.CodeF
	case C.XK_G:
		return key.CodeG
	case C.XK_H:
		return key.CodeH
	case C.XK_I:
		return key.CodeI
	case C.XK_J:
		return key.CodeJ
	case C.XK_K:
		return key.CodeK
	case C.XK_L:
		return key.CodeL
	case C.XK_M:
		return key.CodeM
	case C.XK_N:
		return key.CodeN
	case C.XK_O:
		return key.CodeO
	case C.XK_P:
		return key.CodeP
	case C.XK_Q:
		return key.CodeQ
	case C.XK_R:
		return key.CodeR
	case C.XK_S:
		return key.CodeS
	case C.XK_T:
		return key.CodeT
	case C.XK_U:
		return key.CodeU
	case C.XK_V:
		return key.CodeV
	case C.XK_W:
		return key.CodeW
	case C.XK_X:
		return key.CodeX
	case C.XK_Y:
		return key.CodeY
	case C.XK_Z:
		return key.CodeZ
	case C.XK_a:
		return key.CodeA
	case C.XK_b:
		return key.CodeB
	case C.XK_c:
		return key.CodeC
	case C.XK_d:
		return key.CodeD
	case C.XK_e:
		return key.CodeE
	case C.XK_f:
		return key.CodeF
	case C.XK_g:
		return key.CodeG
	case C.XK_h:
		return key.CodeH
	case C.XK_i:
		return key.CodeI
	case C.XK_j:
		return key.CodeJ
	case C.XK_k:
		return key.CodeK
	case C.XK_l:
		return key.CodeL
	case C.XK_m:
		return key.CodeM
	case C.XK_n:
		return key.CodeN
	case C.XK_o:
		return key.CodeO
	case C.XK_p:
		return key.CodeP
	case C.XK_q:
		return key.CodeQ
	case C.XK_r:
		return key.CodeR
	case C.XK_s:
		return key.CodeS
	case C.XK_t:
		return key.CodeT
	case C.XK_u:
		return key.CodeU
	case C.XK_v:
		return key.CodeV
	case C.XK_w:
		return key.CodeW
	case C.XK_x:
		return key.CodeX
	case C.XK_y:
		return key.CodeY
	case C.XK_z:
		return key.CodeZ
	case C.XK_1:
		return key.Code1
	case C.XK_2:
		return key.Code2
	case C.XK_3:
		return key.Code3
	case C.XK_4:
		return key.Code4
	case C.XK_5:
		return key.Code5
	case C.XK_6:
		return key.Code6
	case C.XK_7:
		return key.Code7
	case C.XK_8:
		return key.Code8
	case C.XK_9:
		return key.Code9
	case C.XK_0:
		return key.Code0
	case C.XK_Return:
		return key.CodeReturnEnter
	case C.XK_Escape:
		return key.CodeEscape
	case C.XK_BackSpace:
		return key.CodeDeleteBackspace
	case C.XK_Tab:
		return key.CodeTab
	case C.XK_space:
		return key.CodeSpacebar
	case C.XK_minus:
		return key.CodeHyphenMinus
	case C.XK_equal:
		return key.CodeEqualSign
	case C.XK_bracketleft:
		return key.CodeLeftSquareBracket
	case C.XK_bracketright:
		return key.CodeRightSquareBracket
	case C.XK_backslash:
		return key.CodeBackslash
	// 50: Keyboard Non-US "#" and ~
	case C.XK_semicolon:
		return key.CodeSemicolon
	case C.XK_apostrophe:
		return key.CodeApostrophe
	case C.XK_grave:
		return key.CodeGraveAccent
	case C.XK_comma:
		return key.CodeComma
	case C.XK_period:
		return key.CodeFullStop
	case C.XK_slash:
		return key.CodeSlash
	case C.XK_Caps_Lock:
		return key.CodeCapsLock
	case C.XK_F1:
		return key.CodeF1
	case C.XK_F2:
		return key.CodeF2
	case C.XK_F3:
		return key.CodeF3
	case C.XK_F4:
		return key.CodeF4
	case C.XK_F5:
		return key.CodeF5
	case C.XK_F6:
		return key.CodeF6
	case C.XK_F7:
		return key.CodeF7
	case C.XK_F8:
		return key.CodeF8
	case C.XK_F9:
		return key.CodeF9
	case C.XK_F10:
		return key.CodeF10
	case C.XK_F11:
		return key.CodeF11
	case C.XK_F12:
		return key.CodeF12
	// 70: PrintScreen
	// 71: Scroll Lock
	// 72: Pause
	case C.XK_Pause:
		return key.CodePause
	// 73: Insert
	case C.XK_Insert:
		return key.CodeInsert
	case C.XK_Home:
		return key.CodeHome
	case C.XK_Page_Up:
		return key.CodePageUp
	case C.XK_Delete:
		return key.CodeDeleteForward
	case C.XK_End:
		return key.CodeEnd
	case C.XK_Page_Down:
		return key.CodePageDown
	case C.XK_Right:
		return key.CodeRightArrow
	case C.XK_Left:
		return key.CodeLeftArrow
	case C.XK_Down:
		return key.CodeDownArrow
	case C.XK_Up:
		return key.CodeUpArrow
/*
	case C.XK_KeypadClear:
		return key.CodeKeypadNumLock
	case C.XK_KeypadDivide:
		return key.CodeKeypadSlash
	case C.XK_KeypadMultiply:
		return key.CodeKeypadAsterisk
	case C.XK_KeypadMinus:
		return key.CodeKeypadHyphenMinus
	case C.XK_KeypadPlus:
		return key.CodeKeypadPlusSign
	case C.XK_KeypadEnter:
		return key.CodeKeypadEnter
	case C.XK_Keypad1:
		return key.CodeKeypad1
	case C.XK_Keypad2:
		return key.CodeKeypad2
	case C.XK_Keypad3:
		return key.CodeKeypad3
	case C.XK_Keypad4:
		return key.CodeKeypad4
	case C.XK_Keypad5:
		return key.CodeKeypad5
	case C.XK_Keypad6:
		return key.CodeKeypad6
	case C.XK_Keypad7:
		return key.CodeKeypad7
	case C.XK_Keypad8:
		return key.CodeKeypad8
	case C.XK_Keypad9:
		return key.CodeKeypad9
	case C.XK_Keypad0:
		return key.CodeKeypad0
	case C.XK_KeypadDecimal:
		return key.CodeKeypadFullStop
	case C.XK_KeypadEquals:
		return key.CodeKeypadEqualSign
*/
	case C.XK_F13:
		return key.CodeF13
	case C.XK_F14:
		return key.CodeF14
	case C.XK_F15:
		return key.CodeF15
	case C.XK_F16:
		return key.CodeF16
	case C.XK_F17:
		return key.CodeF17
	case C.XK_F18:
		return key.CodeF18
	case C.XK_F19:
		return key.CodeF19
	case C.XK_F20:
		return key.CodeF20
	// 116: Keyboard Execute
	case C.XK_Help:
		return key.CodeHelp
/*
	// 118: Keyboard Menu
	// 119: Keyboard Select
	// 120: Keyboard Stop
	// 121: Keyboard Again
	// 122: Keyboard Undo
	// 123: Keyboard Cut
	// 124: Keyboard Copy
	// 125: Keyboard Paste
	// 126: Keyboard Find
	case C.XK_Mute:
		return key.CodeMute
	case C.XK_VolumeUp:
		return key.CodeVolumeUp
	case C.XK_VolumeDown:
		return key.CodeVolumeDown
*/
	// 130: Keyboard Locking Caps Lock
	// 131: Keyboard Locking Num Lock
	// 132: Keyboard Locking Scroll Lock
	// 133: Keyboard Comma
	// 134: Keyboard Equal Sign
	// ...: Bunch of stuff
	case C.XK_Control_L:
		return key.CodeLeftControl
	//case C.XK_Shift:
	//	return key.CodeLeftShift
	case C.XK_Shift_L:
		return key.CodeLeftShift
	case C.XK_Alt_L:
		return key.CodeLeftAlt
/*
	case C.XK_Command:
		return key.CodeLeftGUI
*/
	case C.XK_Control_R:
		return key.CodeRightControl
	case C.XK_Shift_R:
		return key.CodeRightShift
	case C.XK_Alt_R:
		return key.CodeRightAlt
	case C.XK_Menu:
		return key.CodeRightGUI
	default:
		return key.CodeUnknown
	}
}
