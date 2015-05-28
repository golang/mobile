// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Generated from gl.go using go generate. DO NOT EDIT.
// See doc.go for details.

// +build linux darwin
// +build gldebug

package gl

// #include "work.h"
import "C"

import (
	"fmt"
	"log"
	"math"
	"unsafe"
)

func errDrain() string {
	var errs []Enum
	for {
		e := GetError()
		if e == 0 {
			break
		}
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		return fmt.Sprintf(" error: %v", errs)
	}
	return ""
}

func (v Enum) String() string {
	switch v {
	case 0x0:
		return "0"
	case 0x1:
		return "1"
	case 0x2:
		return "LINE_LOOP"
	case 0x3:
		return "LINE_STRIP"
	case 0x4:
		return "TRIANGLES"
	case 0x5:
		return "TRIANGLE_STRIP"
	case 0x6:
		return "TRIANGLE_FAN"
	case 0x300:
		return "SRC_COLOR"
	case 0x301:
		return "ONE_MINUS_SRC_COLOR"
	case 0x302:
		return "SRC_ALPHA"
	case 0x303:
		return "ONE_MINUS_SRC_ALPHA"
	case 0x304:
		return "DST_ALPHA"
	case 0x305:
		return "ONE_MINUS_DST_ALPHA"
	case 0x306:
		return "DST_COLOR"
	case 0x307:
		return "ONE_MINUS_DST_COLOR"
	case 0x308:
		return "SRC_ALPHA_SATURATE"
	case 0x8006:
		return "FUNC_ADD"
	case 0x8009:
		return "32777"
	case 0x883d:
		return "BLEND_EQUATION_ALPHA"
	case 0x800a:
		return "FUNC_SUBTRACT"
	case 0x800b:
		return "FUNC_REVERSE_SUBTRACT"
	case 0x80c8:
		return "BLEND_DST_RGB"
	case 0x80c9:
		return "BLEND_SRC_RGB"
	case 0x80ca:
		return "BLEND_DST_ALPHA"
	case 0x80cb:
		return "BLEND_SRC_ALPHA"
	case 0x8001:
		return "CONSTANT_COLOR"
	case 0x8002:
		return "ONE_MINUS_CONSTANT_COLOR"
	case 0x8003:
		return "CONSTANT_ALPHA"
	case 0x8004:
		return "ONE_MINUS_CONSTANT_ALPHA"
	case 0x8005:
		return "BLEND_COLOR"
	case 0x8892:
		return "ARRAY_BUFFER"
	case 0x8893:
		return "ELEMENT_ARRAY_BUFFER"
	case 0x8894:
		return "ARRAY_BUFFER_BINDING"
	case 0x8895:
		return "ELEMENT_ARRAY_BUFFER_BINDING"
	case 0x88e0:
		return "STREAM_DRAW"
	case 0x88e4:
		return "STATIC_DRAW"
	case 0x88e8:
		return "DYNAMIC_DRAW"
	case 0x8764:
		return "BUFFER_SIZE"
	case 0x8765:
		return "BUFFER_USAGE"
	case 0x8626:
		return "CURRENT_VERTEX_ATTRIB"
	case 0x404:
		return "FRONT"
	case 0x405:
		return "BACK"
	case 0x408:
		return "FRONT_AND_BACK"
	case 0xde1:
		return "TEXTURE_2D"
	case 0xb44:
		return "CULL_FACE"
	case 0xbe2:
		return "BLEND"
	case 0xbd0:
		return "DITHER"
	case 0xb90:
		return "STENCIL_TEST"
	case 0xb71:
		return "DEPTH_TEST"
	case 0xc11:
		return "SCISSOR_TEST"
	case 0x8037:
		return "POLYGON_OFFSET_FILL"
	case 0x809e:
		return "SAMPLE_ALPHA_TO_COVERAGE"
	case 0x80a0:
		return "SAMPLE_COVERAGE"
	case 0x500:
		return "INVALID_ENUM"
	case 0x501:
		return "INVALID_VALUE"
	case 0x502:
		return "INVALID_OPERATION"
	case 0x505:
		return "OUT_OF_MEMORY"
	case 0x900:
		return "CW"
	case 0x901:
		return "CCW"
	case 0xb21:
		return "LINE_WIDTH"
	case 0x846d:
		return "ALIASED_POINT_SIZE_RANGE"
	case 0x846e:
		return "ALIASED_LINE_WIDTH_RANGE"
	case 0xb45:
		return "CULL_FACE_MODE"
	case 0xb46:
		return "FRONT_FACE"
	case 0xb70:
		return "DEPTH_RANGE"
	case 0xb72:
		return "DEPTH_WRITEMASK"
	case 0xb73:
		return "DEPTH_CLEAR_VALUE"
	case 0xb74:
		return "DEPTH_FUNC"
	case 0xb91:
		return "STENCIL_CLEAR_VALUE"
	case 0xb92:
		return "STENCIL_FUNC"
	case 0xb94:
		return "STENCIL_FAIL"
	case 0xb95:
		return "STENCIL_PASS_DEPTH_FAIL"
	case 0xb96:
		return "STENCIL_PASS_DEPTH_PASS"
	case 0xb97:
		return "STENCIL_REF"
	case 0xb93:
		return "STENCIL_VALUE_MASK"
	case 0xb98:
		return "STENCIL_WRITEMASK"
	case 0x8800:
		return "STENCIL_BACK_FUNC"
	case 0x8801:
		return "STENCIL_BACK_FAIL"
	case 0x8802:
		return "STENCIL_BACK_PASS_DEPTH_FAIL"
	case 0x8803:
		return "STENCIL_BACK_PASS_DEPTH_PASS"
	case 0x8ca3:
		return "STENCIL_BACK_REF"
	case 0x8ca4:
		return "STENCIL_BACK_VALUE_MASK"
	case 0x8ca5:
		return "STENCIL_BACK_WRITEMASK"
	case 0xba2:
		return "VIEWPORT"
	case 0xc10:
		return "SCISSOR_BOX"
	case 0xc22:
		return "COLOR_CLEAR_VALUE"
	case 0xc23:
		return "COLOR_WRITEMASK"
	case 0xcf5:
		return "UNPACK_ALIGNMENT"
	case 0xd05:
		return "PACK_ALIGNMENT"
	case 0xd33:
		return "MAX_TEXTURE_SIZE"
	case 0xd3a:
		return "MAX_VIEWPORT_DIMS"
	case 0xd50:
		return "SUBPIXEL_BITS"
	case 0xd52:
		return "RED_BITS"
	case 0xd53:
		return "GREEN_BITS"
	case 0xd54:
		return "BLUE_BITS"
	case 0xd55:
		return "ALPHA_BITS"
	case 0xd56:
		return "DEPTH_BITS"
	case 0xd57:
		return "STENCIL_BITS"
	case 0x2a00:
		return "POLYGON_OFFSET_UNITS"
	case 0x8038:
		return "POLYGON_OFFSET_FACTOR"
	case 0x8069:
		return "TEXTURE_BINDING_2D"
	case 0x80a8:
		return "SAMPLE_BUFFERS"
	case 0x80a9:
		return "SAMPLES"
	case 0x80aa:
		return "SAMPLE_COVERAGE_VALUE"
	case 0x80ab:
		return "SAMPLE_COVERAGE_INVERT"
	case 0x86a2:
		return "NUM_COMPRESSED_TEXTURE_FORMATS"
	case 0x86a3:
		return "COMPRESSED_TEXTURE_FORMATS"
	case 0x1100:
		return "DONT_CARE"
	case 0x1101:
		return "FASTEST"
	case 0x1102:
		return "NICEST"
	case 0x8192:
		return "GENERATE_MIPMAP_HINT"
	case 0x1400:
		return "BYTE"
	case 0x1401:
		return "UNSIGNED_BYTE"
	case 0x1402:
		return "SHORT"
	case 0x1403:
		return "UNSIGNED_SHORT"
	case 0x1404:
		return "INT"
	case 0x1405:
		return "UNSIGNED_INT"
	case 0x1406:
		return "FLOAT"
	case 0x140c:
		return "FIXED"
	case 0x1902:
		return "DEPTH_COMPONENT"
	case 0x1906:
		return "ALPHA"
	case 0x1907:
		return "RGB"
	case 0x1908:
		return "RGBA"
	case 0x1909:
		return "LUMINANCE"
	case 0x190a:
		return "LUMINANCE_ALPHA"
	case 0x8033:
		return "UNSIGNED_SHORT_4_4_4_4"
	case 0x8034:
		return "UNSIGNED_SHORT_5_5_5_1"
	case 0x8363:
		return "UNSIGNED_SHORT_5_6_5"
	case 0x8869:
		return "MAX_VERTEX_ATTRIBS"
	case 0x8dfb:
		return "MAX_VERTEX_UNIFORM_VECTORS"
	case 0x8dfc:
		return "MAX_VARYING_VECTORS"
	case 0x8b4d:
		return "MAX_COMBINED_TEXTURE_IMAGE_UNITS"
	case 0x8b4c:
		return "MAX_VERTEX_TEXTURE_IMAGE_UNITS"
	case 0x8872:
		return "MAX_TEXTURE_IMAGE_UNITS"
	case 0x8dfd:
		return "MAX_FRAGMENT_UNIFORM_VECTORS"
	case 0x8b4f:
		return "SHADER_TYPE"
	case 0x8b80:
		return "DELETE_STATUS"
	case 0x8b82:
		return "LINK_STATUS"
	case 0x8b83:
		return "VALIDATE_STATUS"
	case 0x8b85:
		return "ATTACHED_SHADERS"
	case 0x8b86:
		return "ACTIVE_UNIFORMS"
	case 0x8b87:
		return "ACTIVE_UNIFORM_MAX_LENGTH"
	case 0x8b89:
		return "ACTIVE_ATTRIBUTES"
	case 0x8b8a:
		return "ACTIVE_ATTRIBUTE_MAX_LENGTH"
	case 0x8b8c:
		return "SHADING_LANGUAGE_VERSION"
	case 0x8b8d:
		return "CURRENT_PROGRAM"
	case 0x200:
		return "NEVER"
	case 0x201:
		return "LESS"
	case 0x202:
		return "EQUAL"
	case 0x203:
		return "LEQUAL"
	case 0x204:
		return "GREATER"
	case 0x205:
		return "NOTEQUAL"
	case 0x206:
		return "GEQUAL"
	case 0x207:
		return "ALWAYS"
	case 0x1e00:
		return "KEEP"
	case 0x1e01:
		return "REPLACE"
	case 0x1e02:
		return "INCR"
	case 0x1e03:
		return "DECR"
	case 0x150a:
		return "INVERT"
	case 0x8507:
		return "INCR_WRAP"
	case 0x8508:
		return "DECR_WRAP"
	case 0x1f00:
		return "VENDOR"
	case 0x1f01:
		return "RENDERER"
	case 0x1f02:
		return "VERSION"
	case 0x1f03:
		return "EXTENSIONS"
	case 0x2600:
		return "NEAREST"
	case 0x2601:
		return "LINEAR"
	case 0x2700:
		return "NEAREST_MIPMAP_NEAREST"
	case 0x2701:
		return "LINEAR_MIPMAP_NEAREST"
	case 0x2702:
		return "NEAREST_MIPMAP_LINEAR"
	case 0x2703:
		return "LINEAR_MIPMAP_LINEAR"
	case 0x2800:
		return "TEXTURE_MAG_FILTER"
	case 0x2801:
		return "TEXTURE_MIN_FILTER"
	case 0x2802:
		return "TEXTURE_WRAP_S"
	case 0x2803:
		return "TEXTURE_WRAP_T"
	case 0x1702:
		return "TEXTURE"
	case 0x8513:
		return "TEXTURE_CUBE_MAP"
	case 0x8514:
		return "TEXTURE_BINDING_CUBE_MAP"
	case 0x8515:
		return "TEXTURE_CUBE_MAP_POSITIVE_X"
	case 0x8516:
		return "TEXTURE_CUBE_MAP_NEGATIVE_X"
	case 0x8517:
		return "TEXTURE_CUBE_MAP_POSITIVE_Y"
	case 0x8518:
		return "TEXTURE_CUBE_MAP_NEGATIVE_Y"
	case 0x8519:
		return "TEXTURE_CUBE_MAP_POSITIVE_Z"
	case 0x851a:
		return "TEXTURE_CUBE_MAP_NEGATIVE_Z"
	case 0x851c:
		return "MAX_CUBE_MAP_TEXTURE_SIZE"
	case 0x84c0:
		return "TEXTURE0"
	case 0x84c1:
		return "TEXTURE1"
	case 0x84c2:
		return "TEXTURE2"
	case 0x84c3:
		return "TEXTURE3"
	case 0x84c4:
		return "TEXTURE4"
	case 0x84c5:
		return "TEXTURE5"
	case 0x84c6:
		return "TEXTURE6"
	case 0x84c7:
		return "TEXTURE7"
	case 0x84c8:
		return "TEXTURE8"
	case 0x84c9:
		return "TEXTURE9"
	case 0x84ca:
		return "TEXTURE10"
	case 0x84cb:
		return "TEXTURE11"
	case 0x84cc:
		return "TEXTURE12"
	case 0x84cd:
		return "TEXTURE13"
	case 0x84ce:
		return "TEXTURE14"
	case 0x84cf:
		return "TEXTURE15"
	case 0x84d0:
		return "TEXTURE16"
	case 0x84d1:
		return "TEXTURE17"
	case 0x84d2:
		return "TEXTURE18"
	case 0x84d3:
		return "TEXTURE19"
	case 0x84d4:
		return "TEXTURE20"
	case 0x84d5:
		return "TEXTURE21"
	case 0x84d6:
		return "TEXTURE22"
	case 0x84d7:
		return "TEXTURE23"
	case 0x84d8:
		return "TEXTURE24"
	case 0x84d9:
		return "TEXTURE25"
	case 0x84da:
		return "TEXTURE26"
	case 0x84db:
		return "TEXTURE27"
	case 0x84dc:
		return "TEXTURE28"
	case 0x84dd:
		return "TEXTURE29"
	case 0x84de:
		return "TEXTURE30"
	case 0x84df:
		return "TEXTURE31"
	case 0x84e0:
		return "ACTIVE_TEXTURE"
	case 0x2901:
		return "REPEAT"
	case 0x812f:
		return "CLAMP_TO_EDGE"
	case 0x8370:
		return "MIRRORED_REPEAT"
	case 0x8622:
		return "VERTEX_ATTRIB_ARRAY_ENABLED"
	case 0x8623:
		return "VERTEX_ATTRIB_ARRAY_SIZE"
	case 0x8624:
		return "VERTEX_ATTRIB_ARRAY_STRIDE"
	case 0x8625:
		return "VERTEX_ATTRIB_ARRAY_TYPE"
	case 0x886a:
		return "VERTEX_ATTRIB_ARRAY_NORMALIZED"
	case 0x8645:
		return "VERTEX_ATTRIB_ARRAY_POINTER"
	case 0x889f:
		return "VERTEX_ATTRIB_ARRAY_BUFFER_BINDING"
	case 0x8b9a:
		return "IMPLEMENTATION_COLOR_READ_TYPE"
	case 0x8b9b:
		return "IMPLEMENTATION_COLOR_READ_FORMAT"
	case 0x8b81:
		return "COMPILE_STATUS"
	case 0x8b84:
		return "INFO_LOG_LENGTH"
	case 0x8b88:
		return "SHADER_SOURCE_LENGTH"
	case 0x8dfa:
		return "SHADER_COMPILER"
	case 0x8df8:
		return "SHADER_BINARY_FORMATS"
	case 0x8df9:
		return "NUM_SHADER_BINARY_FORMATS"
	case 0x8df0:
		return "LOW_FLOAT"
	case 0x8df1:
		return "MEDIUM_FLOAT"
	case 0x8df2:
		return "HIGH_FLOAT"
	case 0x8df3:
		return "LOW_INT"
	case 0x8df4:
		return "MEDIUM_INT"
	case 0x8df5:
		return "HIGH_INT"
	case 0x8d40:
		return "FRAMEBUFFER"
	case 0x8d41:
		return "RENDERBUFFER"
	case 0x8056:
		return "RGBA4"
	case 0x8057:
		return "RGB5_A1"
	case 0x8d62:
		return "RGB565"
	case 0x81a5:
		return "DEPTH_COMPONENT16"
	case 0x8d48:
		return "STENCIL_INDEX8"
	case 0x8d42:
		return "RENDERBUFFER_WIDTH"
	case 0x8d43:
		return "RENDERBUFFER_HEIGHT"
	case 0x8d44:
		return "RENDERBUFFER_INTERNAL_FORMAT"
	case 0x8d50:
		return "RENDERBUFFER_RED_SIZE"
	case 0x8d51:
		return "RENDERBUFFER_GREEN_SIZE"
	case 0x8d52:
		return "RENDERBUFFER_BLUE_SIZE"
	case 0x8d53:
		return "RENDERBUFFER_ALPHA_SIZE"
	case 0x8d54:
		return "RENDERBUFFER_DEPTH_SIZE"
	case 0x8d55:
		return "RENDERBUFFER_STENCIL_SIZE"
	case 0x8cd0:
		return "FRAMEBUFFER_ATTACHMENT_OBJECT_TYPE"
	case 0x8cd1:
		return "FRAMEBUFFER_ATTACHMENT_OBJECT_NAME"
	case 0x8cd2:
		return "FRAMEBUFFER_ATTACHMENT_TEXTURE_LEVEL"
	case 0x8cd3:
		return "FRAMEBUFFER_ATTACHMENT_TEXTURE_CUBE_MAP_FACE"
	case 0x8ce0:
		return "COLOR_ATTACHMENT0"
	case 0x8d00:
		return "DEPTH_ATTACHMENT"
	case 0x8d20:
		return "STENCIL_ATTACHMENT"
	case 0x8cd5:
		return "FRAMEBUFFER_COMPLETE"
	case 0x8cd6:
		return "FRAMEBUFFER_INCOMPLETE_ATTACHMENT"
	case 0x8cd7:
		return "FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT"
	case 0x8cd9:
		return "FRAMEBUFFER_INCOMPLETE_DIMENSIONS"
	case 0x8cdd:
		return "FRAMEBUFFER_UNSUPPORTED"
	case 0x8ca6:
		return "FRAMEBUFFER_BINDING"
	case 0x8ca7:
		return "RENDERBUFFER_BINDING"
	case 0x84e8:
		return "MAX_RENDERBUFFER_SIZE"
	case 0x506:
		return "INVALID_FRAMEBUFFER_OPERATION"
	case 0x100:
		return "DEPTH_BUFFER_BIT"
	case 0x400:
		return "STENCIL_BUFFER_BIT"
	case 0x4000:
		return "COLOR_BUFFER_BIT"
	case 0x8b50:
		return "FLOAT_VEC2"
	case 0x8b51:
		return "FLOAT_VEC3"
	case 0x8b52:
		return "FLOAT_VEC4"
	case 0x8b53:
		return "INT_VEC2"
	case 0x8b54:
		return "INT_VEC3"
	case 0x8b55:
		return "INT_VEC4"
	case 0x8b56:
		return "BOOL"
	case 0x8b57:
		return "BOOL_VEC2"
	case 0x8b58:
		return "BOOL_VEC3"
	case 0x8b59:
		return "BOOL_VEC4"
	case 0x8b5a:
		return "FLOAT_MAT2"
	case 0x8b5b:
		return "FLOAT_MAT3"
	case 0x8b5c:
		return "FLOAT_MAT4"
	case 0x8b5e:
		return "SAMPLER_2D"
	case 0x8b60:
		return "SAMPLER_CUBE"
	case 0x8b30:
		return "FRAGMENT_SHADER"
	case 0x8b31:
		return "VERTEX_SHADER"
	default:
		return fmt.Sprintf("gl.Enum(0x%x)", uint32(v))
	}
}

func ActiveTexture(texture Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ActiveTexture(%v) %v", texture, errstr)
	}()
	var call call
	call.args.fn = C.glfnActiveTexture
	call.args.a0 = texture.c()
	work <- call
}

func AttachShader(p Program, s Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.AttachShader(%v, %v) %v", p, s, errstr)
	}()
	var call call
	call.args.fn = C.glfnAttachShader
	call.args.a0 = p.c()
	call.args.a1 = s.c()
	work <- call
}

func BindAttribLocation(p Program, a Attrib, name string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BindAttribLocation(%v, %v, %v) %v", p, a, name, errstr)
	}()
	var call call
	call.args.fn = C.glfnBindAttribLocation
	call.args.a0 = p.c()
	call.args.a1 = a.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
}

func BindBuffer(target Enum, b Buffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BindBuffer(%v, %v) %v", target, b, errstr)
	}()
	var call call
	call.args.fn = C.glfnBindBuffer
	call.args.a0 = target.c()
	call.args.a1 = b.c()
	work <- call
}

func BindFramebuffer(target Enum, fb Framebuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BindFramebuffer(%v, %v) %v", target, fb, errstr)
	}()
	var call call
	call.args.fn = C.glfnBindFramebuffer
	call.args.a0 = target.c()
	call.args.a1 = fb.c()
	work <- call
}

func BindRenderbuffer(target Enum, rb Renderbuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BindRenderbuffer(%v, %v) %v", target, rb, errstr)
	}()
	var call call
	call.args.fn = C.glfnBindRenderbuffer
	call.args.a0 = target.c()
	call.args.a1 = rb.c()
	work <- call
}

func BindTexture(target Enum, t Texture) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BindTexture(%v, %v) %v", target, t, errstr)
	}()
	var call call
	call.args.fn = C.glfnBindTexture
	call.args.a0 = target.c()
	call.args.a1 = t.c()
	work <- call
}

func BlendColor(red, green, blue, alpha float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BlendColor(%v, %v, %v, %v) %v", red, green, blue, alpha, errstr)
	}()
	var call call
	call.args.fn = C.glfnBlendColor
	call.args.a0 = C.uintptr_t(math.Float32bits(red))
	call.args.a1 = C.uintptr_t(math.Float32bits(green))
	call.args.a2 = C.uintptr_t(math.Float32bits(blue))
	call.args.a3 = C.uintptr_t(math.Float32bits(alpha))
	work <- call
}

func BlendEquation(mode Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BlendEquation(%v) %v", mode, errstr)
	}()
	var call call
	call.args.fn = C.glfnBlendEquation
	call.args.a0 = mode.c()
	work <- call
}

func BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BlendEquationSeparate(%v, %v) %v", modeRGB, modeAlpha, errstr)
	}()
	var call call
	call.args.fn = C.glfnBlendEquationSeparate
	call.args.a0 = modeRGB.c()
	call.args.a1 = modeAlpha.c()
	work <- call
}

func BlendFunc(sfactor, dfactor Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BlendFunc(%v, %v) %v", sfactor, dfactor, errstr)
	}()
	var call call
	call.args.fn = C.glfnBlendFunc
	call.args.a0 = sfactor.c()
	call.args.a1 = dfactor.c()
	work <- call
}

func BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BlendFuncSeparate(%v, %v, %v, %v) %v", sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha, errstr)
	}()
	var call call
	call.args.fn = C.glfnBlendFuncSeparate
	call.args.a0 = sfactorRGB.c()
	call.args.a1 = dfactorRGB.c()
	call.args.a2 = sfactorAlpha.c()
	call.args.a3 = dfactorAlpha.c()
	work <- call
}

func BufferData(target Enum, src []byte, usage Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BufferData(%v, len(%d), %v) %v", target, len(src), usage, errstr)
	}()
	var call call
	call.args.fn = C.glfnBufferData
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(len(src))
	call.args.a2 = (C.uintptr_t)(uintptr(unsafe.Pointer(&src[0])))
	call.args.a3 = usage.c()
	work <- call
	<-retvalue
}

func BufferInit(target Enum, size int, usage Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BufferInit(%v, %v, %v) %v", target, size, usage, errstr)
	}()
	var call call
	call.args.fn = C.glfnBufferData
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(size)
	call.args.a2 = 0
	call.args.a3 = usage.c()
	work <- call
}

func BufferSubData(target Enum, offset int, data []byte) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.BufferSubData(%v, %v, len(%d)) %v", target, offset, len(data), errstr)
	}()
	var call call
	call.args.fn = C.glfnBufferSubData
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(offset)
	call.args.a2 = C.uintptr_t(len(data))
	call.args.a3 = (C.uintptr_t)(uintptr(unsafe.Pointer(&data[0])))
	work <- call
	<-retvalue
}

func CheckFramebufferStatus(target Enum) (r0 Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CheckFramebufferStatus(%v) %v%v", target, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnCheckFramebufferStatus
	call.blocking = true
	call.args.a0 = target.c()
	work <- call
	return Enum(<-retvalue)
}

func Clear(mask Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Clear(%v) %v", mask, errstr)
	}()
	var call call
	call.args.fn = C.glfnClear
	call.args.a0 = C.uintptr_t(mask)
	work <- call
}

func ClearColor(red, green, blue, alpha float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ClearColor(%v, %v, %v, %v) %v", red, green, blue, alpha, errstr)
	}()
	var call call
	call.args.fn = C.glfnClearColor
	call.args.a0 = C.uintptr_t(math.Float32bits(red))
	call.args.a1 = C.uintptr_t(math.Float32bits(green))
	call.args.a2 = C.uintptr_t(math.Float32bits(blue))
	call.args.a3 = C.uintptr_t(math.Float32bits(alpha))
	work <- call
}

func ClearDepthf(d float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ClearDepthf(%v) %v", d, errstr)
	}()
	var call call
	call.args.fn = C.glfnClearDepthf
	call.args.a0 = C.uintptr_t(math.Float32bits(d))
	work <- call
}

func ClearStencil(s int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ClearStencil(%v) %v", s, errstr)
	}()
	var call call
	call.args.fn = C.glfnClearStencil
	call.args.a0 = C.uintptr_t(s)
	work <- call
}

func ColorMask(red, green, blue, alpha bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ColorMask(%v, %v, %v, %v) %v", red, green, blue, alpha, errstr)
	}()
	var call call
	call.args.fn = C.glfnColorMask
	call.args.a0 = glBoolean(red)
	call.args.a1 = glBoolean(green)
	call.args.a2 = glBoolean(blue)
	call.args.a3 = glBoolean(alpha)
	work <- call
}

func CompileShader(s Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CompileShader(%v) %v", s, errstr)
	}()
	var call call
	call.args.fn = C.glfnCompileShader
	call.args.a0 = s.c()
	work <- call
}

func CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data []byte) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CompressedTexImage2D(%v, %v, %v, %v, %v, %v, len(%d)) %v", target, level, internalformat, width, height, border, len(data), errstr)
	}()
	var call call
	call.args.fn = C.glfnCompressedTexImage2D
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = internalformat.c()
	call.args.a3 = C.uintptr_t(width)
	call.args.a4 = C.uintptr_t(height)
	call.args.a5 = C.uintptr_t(border)
	call.args.a6 = C.uintptr_t(len(data))
	call.args.a7 = C.uintptr_t(uintptr(unsafe.Pointer(&data[0])))
	work <- call
	<-retvalue
}

func CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data []byte) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CompressedTexSubImage2D(%v, %v, %v, %v, %v, %v, %v, len(%d)) %v", target, level, xoffset, yoffset, width, height, format, len(data), errstr)
	}()
	var call call
	call.args.fn = C.glfnCompressedTexSubImage2D
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = C.uintptr_t(xoffset)
	call.args.a3 = C.uintptr_t(yoffset)
	call.args.a4 = C.uintptr_t(width)
	call.args.a5 = C.uintptr_t(height)
	call.args.a6 = format.c()
	call.args.a7 = C.uintptr_t(len(data))
	call.args.a8 = C.uintptr_t(uintptr(unsafe.Pointer(&data[0])))
	work <- call
	<-retvalue
}

func CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CopyTexImage2D(%v, %v, %v, %v, %v, %v, %v, %v) %v", target, level, internalformat, x, y, width, height, border, errstr)
	}()
	var call call
	call.args.fn = C.glfnCopyTexImage2D
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = internalformat.c()
	call.args.a3 = C.uintptr_t(x)
	call.args.a4 = C.uintptr_t(y)
	call.args.a5 = C.uintptr_t(width)
	call.args.a6 = C.uintptr_t(height)
	call.args.a7 = C.uintptr_t(border)
	work <- call
}

func CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CopyTexSubImage2D(%v, %v, %v, %v, %v, %v, %v, %v) %v", target, level, xoffset, yoffset, x, y, width, height, errstr)
	}()
	var call call
	call.args.fn = C.glfnCopyTexSubImage2D
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = C.uintptr_t(xoffset)
	call.args.a3 = C.uintptr_t(yoffset)
	call.args.a4 = C.uintptr_t(x)
	call.args.a5 = C.uintptr_t(y)
	call.args.a6 = C.uintptr_t(width)
	call.args.a7 = C.uintptr_t(height)
	work <- call
}

func CreateBuffer() (r0 Buffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateBuffer() %v%v", r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGenBuffer
	call.blocking = true
	work <- call
	return Buffer{Value: uint32(<-retvalue)}
}

func CreateFramebuffer() (r0 Framebuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateFramebuffer() %v%v", r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGenFramebuffer
	call.blocking = true
	work <- call
	return Framebuffer{Value: uint32(<-retvalue)}
}

func CreateProgram() (r0 Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateProgram() %v%v", r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnCreateProgram
	call.blocking = true
	work <- call
	return Program{Value: uint32(<-retvalue)}
}

func CreateRenderbuffer() (r0 Renderbuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateRenderbuffer() %v%v", r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGenRenderbuffer
	call.blocking = true
	work <- call
	return Renderbuffer{Value: uint32(<-retvalue)}
}

func CreateShader(ty Enum) (r0 Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateShader(%v) %v%v", ty, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnCreateShader
	call.blocking = true
	call.args.a0 = C.uintptr_t(ty)
	work <- call
	return Shader{Value: uint32(<-retvalue)}
}

func CreateTexture() (r0 Texture) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CreateTexture() %v%v", r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGenTexture
	call.blocking = true
	work <- call
	return Texture{Value: uint32(<-retvalue)}
}

func CullFace(mode Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.CullFace(%v) %v", mode, errstr)
	}()
	var call call
	call.args.fn = C.glfnCullFace
	call.args.a0 = mode.c()
	work <- call
}

func DeleteBuffer(v Buffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteBuffer(%v) %v", v, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteBuffer
	call.args.a0 = C.uintptr_t(v.Value)
	work <- call
}

func DeleteFramebuffer(v Framebuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteFramebuffer(%v) %v", v, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteFramebuffer
	call.args.a0 = C.uintptr_t(v.Value)
	work <- call
}

func DeleteProgram(p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteProgram(%v) %v", p, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteProgram
	call.args.a0 = p.c()
	work <- call
}

func DeleteRenderbuffer(v Renderbuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteRenderbuffer(%v) %v", v, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteRenderbuffer
	call.args.a0 = v.c()
	work <- call
}

func DeleteShader(s Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteShader(%v) %v", s, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteShader
	call.args.a0 = s.c()
	work <- call
}

func DeleteTexture(v Texture) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DeleteTexture(%v) %v", v, errstr)
	}()
	var call call
	call.args.fn = C.glfnDeleteTexture
	call.args.a0 = v.c()
	work <- call
}

func DepthFunc(fn Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DepthFunc(%v) %v", fn, errstr)
	}()
	var call call
	call.args.fn = C.glfnDepthFunc
	call.args.a0 = fn.c()
	work <- call
}

func DepthMask(flag bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DepthMask(%v) %v", flag, errstr)
	}()
	var call call
	call.args.fn = C.glfnDepthMask
	call.args.a0 = glBoolean(flag)
	work <- call
}

func DepthRangef(n, f float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DepthRangef(%v, %v) %v", n, f, errstr)
	}()
	var call call
	call.args.fn = C.glfnDepthRangef
	call.args.a0 = C.uintptr_t(math.Float32bits(n))
	call.args.a1 = C.uintptr_t(math.Float32bits(f))
	work <- call
}

func DetachShader(p Program, s Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DetachShader(%v, %v) %v", p, s, errstr)
	}()
	var call call
	call.args.fn = C.glfnDetachShader
	call.args.a0 = p.c()
	call.args.a1 = s.c()
	work <- call
}

func Disable(cap Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Disable(%v) %v", cap, errstr)
	}()
	var call call
	call.args.fn = C.glfnDisable
	call.args.a0 = cap.c()
	work <- call
}

func DisableVertexAttribArray(a Attrib) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DisableVertexAttribArray(%v) %v", a, errstr)
	}()
	var call call
	call.args.fn = C.glfnDisableVertexAttribArray
	call.args.a0 = a.c()
	work <- call
}

func DrawArrays(mode Enum, first, count int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DrawArrays(%v, %v, %v) %v", mode, first, count, errstr)
	}()
	var call call
	call.args.fn = C.glfnDrawArrays
	call.args.a0 = mode.c()
	call.args.a1 = C.uintptr_t(first)
	call.args.a2 = C.uintptr_t(count)
	work <- call
}

func DrawElements(mode Enum, count int, ty Enum, offset int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.DrawElements(%v, %v, %v, %v) %v", mode, count, ty, offset, errstr)
	}()
	var call call
	call.args.fn = C.glfnDrawElements
	call.args.a0 = mode.c()
	call.args.a1 = C.uintptr_t(count)
	call.args.a2 = ty.c()
	call.args.a3 = C.uintptr_t(offset)
	work <- call
}

func Enable(cap Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Enable(%v) %v", cap, errstr)
	}()
	var call call
	call.args.fn = C.glfnEnable
	call.args.a0 = cap.c()
	work <- call
}

func EnableVertexAttribArray(a Attrib) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.EnableVertexAttribArray(%v) %v", a, errstr)
	}()
	var call call
	call.args.fn = C.glfnEnableVertexAttribArray
	call.args.a0 = a.c()
	work <- call
}

func Finish() {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Finish() %v", errstr)
	}()
	var call call
	call.args.fn = C.glfnFinish
	call.blocking = true
	work <- call
	<-retvalue
}

func Flush() {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Flush() %v", errstr)
	}()
	var call call
	call.args.fn = C.glfnFlush
	call.blocking = true
	work <- call
	<-retvalue
}

func FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.FramebufferRenderbuffer(%v, %v, %v, %v) %v", target, attachment, rbTarget, rb, errstr)
	}()
	var call call
	call.args.fn = C.glfnFramebufferRenderbuffer
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = rbTarget.c()
	call.args.a3 = rb.c()
	work <- call
}

func FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.FramebufferTexture2D(%v, %v, %v, %v, %v) %v", target, attachment, texTarget, t, level, errstr)
	}()
	var call call
	call.args.fn = C.glfnFramebufferTexture2D
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = texTarget.c()
	call.args.a3 = t.c()
	call.args.a4 = C.uintptr_t(level)
	work <- call
}

func FrontFace(mode Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.FrontFace(%v) %v", mode, errstr)
	}()
	var call call
	call.args.fn = C.glfnFrontFace
	call.args.a0 = mode.c()
	work <- call
}

func GenerateMipmap(target Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GenerateMipmap(%v) %v", target, errstr)
	}()
	var call call
	call.args.fn = C.glfnGenerateMipmap
	call.args.a0 = target.c()
	work <- call
}

func GetActiveAttrib(p Program, index uint32) (name string, size int, ty Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetActiveAttrib(%v, %v) (%v, %v, %v) %v", p, index, name, size, ty, errstr)
	}()
	bufSize := GetProgrami(p, ACTIVE_ATTRIBUTE_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)
	var cSize C.GLint
	var cType C.GLenum
	var call call
	call.args.fn = C.glfnGetActiveAttrib
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(index)
	call.args.a2 = C.uintptr_t(bufSize)
	call.args.a3 = 0
	call.args.a4 = C.uintptr_t(uintptr(unsafe.Pointer(&cSize)))
	call.args.a5 = C.uintptr_t(uintptr(unsafe.Pointer(&cType)))
	call.args.a6 = C.uintptr_t(uintptr(unsafe.Pointer(buf)))
	work <- call
	<-retvalue
	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

func GetActiveUniform(p Program, index uint32) (name string, size int, ty Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetActiveUniform(%v, %v) (%v, %v, %v) %v", p, index, name, size, ty, errstr)
	}()
	bufSize := GetProgrami(p, ACTIVE_UNIFORM_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)
	var cSize C.GLint
	var cType C.GLenum
	var call call
	call.args.fn = C.glfnGetActiveUniform
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(index)
	call.args.a2 = C.uintptr_t(bufSize)
	call.args.a3 = 0
	call.args.a4 = C.uintptr_t(uintptr(unsafe.Pointer(&cSize)))
	call.args.a5 = C.uintptr_t(uintptr(unsafe.Pointer(&cType)))
	call.args.a6 = C.uintptr_t(uintptr(unsafe.Pointer(buf)))
	work <- call
	<-retvalue
	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

func GetAttachedShaders(p Program) (r0 []Shader) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetAttachedShaders(%v) %v%v", p, r0, errstr)
	}()
	shadersLen := GetProgrami(p, ATTACHED_SHADERS)
	var n C.GLsizei
	buf := make([]C.GLuint, shadersLen)
	var call call
	call.blocking = true
	call.args.fn = C.glfnGetAttachedShaders
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(shadersLen)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&n)))
	call.args.a3 = C.uintptr_t(uintptr(unsafe.Pointer(&buf[0])))
	work <- call
	<-retvalue
	buf = buf[:int(n)]
	shaders := make([]Shader, len(buf))
	for i, s := range buf {
		shaders[i] = Shader{Value: uint32(s)}
	}
	return shaders
}

func GetAttribLocation(p Program, name string) (r0 Attrib) {
	defer func() {
		errstr := errDrain()
		r0.name = name
		log.Printf("gl.GetAttribLocation(%v, %v) %v%v", p, name, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetAttribLocation
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
	return Attrib{Value: uint(<-retvalue)}
}

func GetBooleanv(dst []bool, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetBooleanv(%v, %v) %v", dst, pname, errstr)
	}()
	buf := make([]C.GLboolean, len(dst))
	var call call
	call.args.fn = C.glfnGetBooleanv
	call.blocking = true
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&buf[0])))
	work <- call
	<-retvalue
	for i, v := range buf {
		dst[i] = v != 0
	}
}

func GetFloatv(dst []float32, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetFloatv(len(%d), %v) %v", len(dst), pname, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetFloatv
	call.blocking = true
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetIntegerv(dst []int32, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetIntegerv(%v, %v) %v", dst, pname, errstr)
	}()
	buf := make([]C.GLint, len(dst))
	var call call
	call.args.fn = C.glfnGetIntegerv
	call.blocking = true
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&buf[0])))
	work <- call
	<-retvalue
	for i, v := range buf {
		dst[i] = int32(v)
	}
}

func GetInteger(pname Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetInteger(%v) %v%v", pname, r0, errstr)
	}()
	var v [1]int32
	GetIntegerv(v[:], pname)
	return int(v[0])
}

func GetBufferParameteri(target, value Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetBufferParameteri(%v, %v) %v%v", target, value, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetBufferParameteri
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = value.c()
	work <- call
	return int(<-retvalue)
}

func GetError() (r0 Enum) {
	var call call
	call.args.fn = C.glfnGetError
	call.blocking = true
	work <- call
	return Enum(<-retvalue)
}

func GetFramebufferAttachmentParameteri(target, attachment, pname Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetFramebufferAttachmentParameteri(%v, %v, %v) %v%v", target, attachment, pname, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetFramebufferAttachmentParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = pname.c()
	work <- call
	return int(<-retvalue)
}

func GetProgrami(p Program, pname Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetProgrami(%v, %v) %v%v", p, pname, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetProgramiv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

func GetProgramInfoLog(p Program) (r0 string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetProgramInfoLog(%v) %v%v", p, r0, errstr)
	}()
	infoLen := GetProgrami(p, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	defer C.free(buf)
	var call call
	call.args.fn = C.glfnGetProgramInfoLog
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(infoLen)
	call.args.a2 = 0
	call.args.a3 = C.uintptr_t(uintptr(buf))
	work <- call
	<-retvalue
	return C.GoString((*C.char)(buf))
}

func GetRenderbufferParameteri(target, pname Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetRenderbufferParameteri(%v, %v) %v%v", target, pname, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetRenderbufferParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

func GetShaderi(s Shader, pname Enum) (r0 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetShaderi(%v, %v) %v%v", s, pname, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetShaderiv
	call.blocking = true
	call.args.a0 = s.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

func GetShaderInfoLog(s Shader) (r0 string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetShaderInfoLog(%v) %v%v", s, r0, errstr)
	}()
	infoLen := GetShaderi(s, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	defer C.free(buf)
	var call call
	call.args.fn = C.glfnGetShaderInfoLog
	call.blocking = true
	call.args.a0 = s.c()
	call.args.a1 = C.uintptr_t(infoLen)
	call.args.a2 = 0
	call.args.a3 = C.uintptr_t(uintptr(buf))
	work <- call
	<-retvalue
	return C.GoString((*C.char)(buf))
}

func GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeLow, rangeHigh, precision int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetShaderPrecisionFormat(%v, %v) (%v, %v, %v) %v", shadertype, precisiontype, rangeLow, rangeHigh, precision, errstr)
	}()
	var cRange [2]C.GLint
	var cPrecision C.GLint
	var call call
	call.args.fn = C.glfnGetShaderPrecisionFormat
	call.blocking = true
	call.args.a0 = shadertype.c()
	call.args.a1 = precisiontype.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&cRange[0])))
	call.args.a3 = C.uintptr_t(uintptr(unsafe.Pointer(&cPrecision)))
	work <- call
	<-retvalue
	return int(cRange[0]), int(cRange[1]), int(cPrecision)
}

func GetShaderSource(s Shader) (r0 string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetShaderSource(%v) %v%v", s, r0, errstr)
	}()
	sourceLen := GetShaderi(s, SHADER_SOURCE_LENGTH)
	if sourceLen == 0 {
		return ""
	}
	buf := C.malloc(C.size_t(sourceLen))
	defer C.free(buf)
	var call call
	call.args.fn = C.glfnGetShaderSource
	call.blocking = true
	call.args.a0 = s.c()
	call.args.a1 = C.uintptr_t(sourceLen)
	call.args.a2 = 0
	call.args.a3 = C.uintptr_t(uintptr(buf))
	work <- call
	<-retvalue
	return C.GoString((*C.char)(buf))
}

func GetString(pname Enum) (r0 string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetString(%v) %v%v", pname, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetString
	call.blocking = true
	call.args.a0 = pname.c()
	work <- call
	return C.GoString((*C.char)((unsafe.Pointer(uintptr(<-retvalue)))))
}

func GetTexParameterfv(dst []float32, target, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetTexParameterfv(len(%d), %v, %v) %v", len(dst), target, pname, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetTexParameterfv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetTexParameteriv(dst []int32, target, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetTexParameteriv(%v, %v, %v) %v", dst, target, pname, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetTexParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetUniformfv(dst []float32, src Uniform, p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetUniformfv(len(%d), %v, %v) %v", len(dst), src, p, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetUniformfv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = src.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetUniformiv(dst []int32, src Uniform, p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetUniformiv(%v, %v, %v) %v", dst, src, p, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetUniformiv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = src.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetUniformLocation(p Program, name string) (r0 Uniform) {
	defer func() {
		errstr := errDrain()
		r0.name = name
		log.Printf("gl.GetUniformLocation(%v, %v) %v%v", p, name, r0, errstr)
	}()
	var call call
	call.blocking = true
	call.args.fn = C.glfnGetUniformLocation
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
	return Uniform{Value: int32(<-retvalue)}
}

func GetVertexAttribf(src Attrib, pname Enum) (r0 float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetVertexAttribf(%v, %v) %v%v", src, pname, r0, errstr)
	}()
	var params [1]float32
	GetVertexAttribfv(params[:], src, pname)
	return params[0]
}

func GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetVertexAttribfv(len(%d), %v, %v) %v", len(dst), src, pname, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetVertexAttribfv
	call.blocking = true
	call.args.a0 = src.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func GetVertexAttribi(src Attrib, pname Enum) (r0 int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetVertexAttribi(%v, %v) %v%v", src, pname, r0, errstr)
	}()
	var params [1]int32
	GetVertexAttribiv(params[:], src, pname)
	return params[0]
}

func GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.GetVertexAttribiv(%v, %v, %v) %v", dst, src, pname, errstr)
	}()
	var call call
	call.args.fn = C.glfnGetVertexAttribiv
	call.blocking = true
	call.args.a0 = src.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func Hint(target, mode Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Hint(%v, %v) %v", target, mode, errstr)
	}()
	var call call
	call.args.fn = C.glfnHint
	call.args.a0 = target.c()
	call.args.a1 = mode.c()
	work <- call
}

func IsBuffer(b Buffer) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsBuffer(%v) %v%v", b, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsBuffer
	call.blocking = true
	call.args.a0 = b.c()
	work <- call
	return <-retvalue != 0
}

func IsEnabled(cap Enum) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsEnabled(%v) %v%v", cap, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsEnabled
	call.blocking = true
	call.args.a0 = cap.c()
	work <- call
	return <-retvalue != 0
}

func IsFramebuffer(fb Framebuffer) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsFramebuffer(%v) %v%v", fb, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsFramebuffer
	call.blocking = true
	call.args.a0 = fb.c()
	work <- call
	return <-retvalue != 0
}

func IsProgram(p Program) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsProgram(%v) %v%v", p, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsProgram
	call.blocking = true
	call.args.a0 = p.c()
	work <- call
	return <-retvalue != 0
}

func IsRenderbuffer(rb Renderbuffer) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsRenderbuffer(%v) %v%v", rb, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsRenderbuffer
	call.blocking = true
	call.args.a0 = rb.c()
	work <- call
	return <-retvalue != 0
}

func IsShader(s Shader) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsShader(%v) %v%v", s, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsShader
	call.blocking = true
	call.args.a0 = s.c()
	work <- call
	return <-retvalue != 0
}

func IsTexture(t Texture) (r0 bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.IsTexture(%v) %v%v", t, r0, errstr)
	}()
	var call call
	call.args.fn = C.glfnIsTexture
	call.blocking = true
	call.args.a0 = t.c()
	work <- call
	return <-retvalue != 0
}

func LineWidth(width float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.LineWidth(%v) %v", width, errstr)
	}()
	var call call
	call.args.fn = C.glfnLineWidth
	call.args.a0 = C.uintptr_t(math.Float32bits(width))
	work <- call
}

func LinkProgram(p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.LinkProgram(%v) %v", p, errstr)
	}()
	var call call
	call.args.fn = C.glfnLinkProgram
	call.args.a0 = p.c()
	work <- call
}

func PixelStorei(pname Enum, param int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.PixelStorei(%v, %v) %v", pname, param, errstr)
	}()
	var call call
	call.args.fn = C.glfnPixelStorei
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(param)
	work <- call
}

func PolygonOffset(factor, units float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.PolygonOffset(%v, %v) %v", factor, units, errstr)
	}()
	var call call
	call.args.fn = C.glfnPolygonOffset
	call.args.a0 = C.uintptr_t(math.Float32bits(factor))
	call.args.a1 = C.uintptr_t(math.Float32bits(units))
	work <- call
}

func ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ReadPixels(len(%d), %v, %v, %v, %v, %v, %v) %v", len(dst), x, y, width, height, format, ty, errstr)
	}()
	var call call
	call.args.fn = C.glfnReadPixels
	call.blocking = true
	call.args.a0 = C.uintptr_t(x)
	call.args.a1 = C.uintptr_t(y)
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	call.args.a4 = format.c()
	call.args.a5 = ty.c()
	call.args.a6 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

func ReleaseShaderCompiler() {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ReleaseShaderCompiler() %v", errstr)
	}()
	var call call
	call.args.fn = C.glfnReleaseShaderCompiler
	work <- call
}

func RenderbufferStorage(target, internalFormat Enum, width, height int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.RenderbufferStorage(%v, %v, %v, %v) %v", target, internalFormat, width, height, errstr)
	}()
	var call call
	call.args.fn = C.glfnRenderbufferStorage
	call.args.a0 = target.c()
	call.args.a1 = internalFormat.c()
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}

func SampleCoverage(value float32, invert bool) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.SampleCoverage(%v, %v) %v", value, invert, errstr)
	}()
	var call call
	call.args.fn = C.glfnSampleCoverage
	call.args.a0 = C.uintptr_t(math.Float32bits(value))
	call.args.a1 = glBoolean(invert)
	work <- call
}

func Scissor(x, y, width, height int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Scissor(%v, %v, %v, %v) %v", x, y, width, height, errstr)
	}()
	var call call
	call.args.fn = C.glfnScissor
	call.args.a0 = C.uintptr_t(x)
	call.args.a1 = C.uintptr_t(y)
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}

func ShaderSource(s Shader, src string) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ShaderSource(%v, %v) %v", s, src, errstr)
	}()
	var call call
	call.args.fn = C.glfnShaderSource
	call.args.a0 = s.c()
	call.args.a1 = 1
	cstr := C.CString(src)
	cstrp := (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(cstr))))
	*cstrp = cstr
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(cstrp)))
	work <- call
}

func StencilFunc(fn Enum, ref int, mask uint32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilFunc(%v, %v, %v) %v", fn, ref, mask, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilFunc
	call.args.a0 = fn.c()
	call.args.a1 = C.uintptr_t(ref)
	call.args.a2 = C.uintptr_t(mask)
	work <- call
}

func StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilFuncSeparate(%v, %v, %v, %v) %v", face, fn, ref, mask, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilFuncSeparate
	call.args.a0 = face.c()
	call.args.a1 = fn.c()
	call.args.a2 = C.uintptr_t(ref)
	call.args.a3 = C.uintptr_t(mask)
	work <- call
}

func StencilMask(mask uint32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilMask(%v) %v", mask, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilMask
	call.args.a0 = C.uintptr_t(mask)
	work <- call
}

func StencilMaskSeparate(face Enum, mask uint32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilMaskSeparate(%v, %v) %v", face, mask, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilMaskSeparate
	call.args.a0 = face.c()
	call.args.a1 = C.uintptr_t(mask)
	work <- call
}

func StencilOp(fail, zfail, zpass Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilOp(%v, %v, %v) %v", fail, zfail, zpass, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilOp
	call.args.a0 = fail.c()
	call.args.a1 = zfail.c()
	call.args.a2 = zpass.c()
	work <- call
}

func StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.StencilOpSeparate(%v, %v, %v, %v) %v", face, sfail, dpfail, dppass, errstr)
	}()
	var call call
	call.args.fn = C.glfnStencilOpSeparate
	call.args.a0 = face.c()
	call.args.a1 = sfail.c()
	call.args.a2 = dpfail.c()
	call.args.a3 = dppass.c()
	work <- call
}

func TexImage2D(target Enum, level int, width, height int, format Enum, ty Enum, data []byte) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexImage2D(%v, %v, %v, %v, %v, %v, len(%d)) %v", target, level, width, height, format, ty, len(data), errstr)
	}()
	// It is common to pass TexImage2D a nil data, indicating that a
	// bound GL buffer is being used as the source. In that case, it
	// is not necessary to block.
	var call call
	call.args.fn = C.glfnTexImage2D
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = C.uintptr_t(format)
	call.args.a3 = C.uintptr_t(width)
	call.args.a4 = C.uintptr_t(height)
	call.args.a5 = format.c()
	call.args.a6 = ty.c()
	if len(data) > 0 {
		call.blocking = true
		call.args.a7 = C.uintptr_t(uintptr(unsafe.Pointer(&data[0])))
	} else {
		call.args.a7 = 0
	}
	work <- call
	if call.blocking {
		<-retvalue
	}
}

func TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexSubImage2D(%v, %v, %v, %v, %v, %v, %v, %v, len(%d)) %v", target, level, x, y, width, height, format, ty, len(data), errstr)
	}()
	var call call
	call.args.fn = C.glfnTexSubImage2D
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(level)
	call.args.a2 = C.uintptr_t(x)
	call.args.a3 = C.uintptr_t(y)
	call.args.a4 = C.uintptr_t(width)
	call.args.a5 = C.uintptr_t(height)
	call.args.a6 = format.c()
	call.args.a7 = ty.c()
	call.args.a8 = C.uintptr_t(uintptr(unsafe.Pointer(&data[0])))
	work <- call
	<-retvalue
}

func TexParameterf(target, pname Enum, param float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexParameterf(%v, %v, %v) %v", target, pname, param, errstr)
	}()
	var call call
	call.args.fn = C.glfnTexParameterf
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(math.Float32bits(param))
	work <- call
}

func TexParameterfv(target, pname Enum, params []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexParameterfv(%v, %v, len(%d)) %v", target, pname, len(params), errstr)
	}()
	var call call
	call.args.fn = C.glfnTexParameterfv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&params[0])))
	work <- call
	<-retvalue
}

func TexParameteri(target, pname Enum, param int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexParameteri(%v, %v, %v) %v", target, pname, param, errstr)
	}()
	var call call
	call.args.fn = C.glfnTexParameteri
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(param)
	work <- call
}

func TexParameteriv(target, pname Enum, params []int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.TexParameteriv(%v, %v, %v) %v", target, pname, params, errstr)
	}()
	var call call
	call.args.fn = C.glfnTexParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&params[0])))
	work <- call
	<-retvalue
}

func Uniform1f(dst Uniform, v float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform1f(%v, %v) %v", dst, v, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform1f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v))
	work <- call
}

func Uniform1fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform1fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform1fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src))
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform1i(dst Uniform, v int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform1i(%v, %v) %v", dst, v, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform1i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v)
	work <- call
}

func Uniform1iv(dst Uniform, src []int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform1iv(%v, %v) %v", dst, src, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform1iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src))
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform2f(dst Uniform, v0, v1 float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform2f(%v, %v, %v) %v", dst, v0, v1, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform2f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	work <- call
}

func Uniform2fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform2fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform2fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 2)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform2i(dst Uniform, v0, v1 int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform2i(%v, %v, %v) %v", dst, v0, v1, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform2i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	work <- call
}

func Uniform2iv(dst Uniform, src []int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform2iv(%v, %v) %v", dst, src, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform2iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 2)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform3f(dst Uniform, v0, v1, v2 float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform3f(%v, %v, %v, %v) %v", dst, v0, v1, v2, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform3f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	call.args.a3 = C.uintptr_t(math.Float32bits(v2))
	work <- call
}

func Uniform3fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform3fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 3)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform3i(dst Uniform, v0, v1, v2 int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform3i(%v, %v, %v, %v) %v", dst, v0, v1, v2, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform3i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	call.args.a3 = C.uintptr_t(v2)
	work <- call
}

func Uniform3iv(dst Uniform, src []int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform3iv(%v, %v) %v", dst, src, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform3iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 3)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform4f(%v, %v, %v, %v, %v) %v", dst, v0, v1, v2, v3, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform4f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	call.args.a3 = C.uintptr_t(math.Float32bits(v2))
	call.args.a4 = C.uintptr_t(math.Float32bits(v3))
	work <- call
}

func Uniform4fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform4fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform4i(%v, %v, %v, %v, %v) %v", dst, v0, v1, v2, v3, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform4i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	call.args.a3 = C.uintptr_t(v2)
	call.args.a4 = C.uintptr_t(v3)
	work <- call
}

func Uniform4iv(dst Uniform, src []int32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Uniform4iv(%v, %v) %v", dst, src, errstr)
	}()
	var call call
	call.args.fn = C.glfnUniform4iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func UniformMatrix2fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.UniformMatrix2fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniformMatrix2fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func UniformMatrix3fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.UniformMatrix3fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniformMatrix3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 9)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func UniformMatrix4fv(dst Uniform, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.UniformMatrix4fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnUniformMatrix4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 16)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func UseProgram(p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.UseProgram(%v) %v", p, errstr)
	}()
	var call call
	call.args.fn = C.glfnUseProgram
	call.args.a0 = p.c()
	work <- call
}

func ValidateProgram(p Program) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.ValidateProgram(%v) %v", p, errstr)
	}()
	var call call
	call.args.fn = C.glfnValidateProgram
	call.args.a0 = p.c()
	work <- call
}

func VertexAttrib1f(dst Attrib, x float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib1f(%v, %v) %v", dst, x, errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib1f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	work <- call
}

func VertexAttrib1fv(dst Attrib, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib1fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib1fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func VertexAttrib2f(dst Attrib, x, y float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib2f(%v, %v, %v) %v", dst, x, y, errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib2f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	work <- call
}

func VertexAttrib2fv(dst Attrib, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib2fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib2fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func VertexAttrib3f(dst Attrib, x, y, z float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib3f(%v, %v, %v, %v) %v", dst, x, y, z, errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib3f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	call.args.a3 = C.uintptr_t(math.Float32bits(z))
	work <- call
}

func VertexAttrib3fv(dst Attrib, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib3fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func VertexAttrib4f(dst Attrib, x, y, z, w float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib4f(%v, %v, %v, %v, %v) %v", dst, x, y, z, w, errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib4f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	call.args.a3 = C.uintptr_t(math.Float32bits(z))
	call.args.a4 = C.uintptr_t(math.Float32bits(w))
	work <- call
}

func VertexAttrib4fv(dst Attrib, src []float32) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttrib4fv(%v, len(%d)) %v", dst, len(src), errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttrib4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

func VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.VertexAttribPointer(%v, %v, %v, %v, %v, %v) %v", dst, size, ty, normalized, stride, offset, errstr)
	}()
	var call call
	call.args.fn = C.glfnVertexAttribPointer
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(size)
	call.args.a2 = ty.c()
	call.args.a3 = glBoolean(normalized)
	call.args.a4 = C.uintptr_t(stride)
	call.args.a5 = C.uintptr_t(offset)
	work <- call
}

func Viewport(x, y, width, height int) {
	defer func() {
		errstr := errDrain()
		log.Printf("gl.Viewport(%v, %v, %v, %v) %v", x, y, width, height, errstr)
	}()
	var call call
	call.args.fn = C.glfnViewport
	call.args.a0 = C.uintptr_t(x)
	call.args.a1 = C.uintptr_t(y)
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}
