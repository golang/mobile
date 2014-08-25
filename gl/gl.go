// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gl implements Go bindings for OpenGL ES 2.
//
// The bindings are deliberately minimal, staying as close the C API as
// possible. The semantics of each function maps onto functions
// described in the Khronos documentation:
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/
//
// One notable departure from the C API is the introduction of types
// to represent common uses of GLint: Texture, Surface, Buffer, etc.
package gl

// TODO(crawshaw): build on more host platforms (makes it easier to gobind).
// TODO(crawshaw): expand to cover OpenGL ES 3.
// TODO(crawshaw): should functions on specific types become methods? E.g.
//                 	func (t Texture) Bind(target Enum)
//                 this seems natural in Go, but moves us slightly
//                 further away from the underlying OpenGL spec.

//#cgo darwin  LDFLAGS: -framework OpenGL
//#cgo linux   LDFLAGS: -lGLESv2
//#include <stdlib.h>
//#include "gl2.h"
import "C"

import "unsafe"

/*
Partially generated from the Khronos OpenGL API specification in XML
format, which is covered by the license:

	Copyright (c) 2013-2014 The Khronos Group Inc.

	Permission is hereby granted, free of charge, to any person obtaining a
	copy of this software and/or associated documentation files (the
	"Materials"), to deal in the Materials without restriction, including
	without limitation the rights to use, copy, modify, merge, publish,
	distribute, sublicense, and/or sell copies of the Materials, and to
	permit persons to whom the Materials are furnished to do so, subject to
	the following conditions:

	The above copyright notice and this permission notice shall be included
	in all copies or substantial portions of the Materials.

	THE MATERIALS ARE PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
	MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
	CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
	MATERIALS OR THE USE OR OTHER DEALINGS IN THE MATERIALS.

*/

const (
	POINTS                                       = 0x0000
	LINES                                        = 0x0001
	LINE_LOOP                                    = 0x0002
	LINE_STRIP                                   = 0x0003
	TRIANGLES                                    = 0x0004
	TRIANGLE_STRIP                               = 0x0005
	TRIANGLE_FAN                                 = 0x0006
	SRC_COLOR                                    = 0x0300
	ONE_MINUS_SRC_COLOR                          = 0x0301
	SRC_ALPHA                                    = 0x0302
	ONE_MINUS_SRC_ALPHA                          = 0x0303
	DST_ALPHA                                    = 0x0304
	ONE_MINUS_DST_ALPHA                          = 0x0305
	DST_COLOR                                    = 0x0306
	ONE_MINUS_DST_COLOR                          = 0x0307
	SRC_ALPHA_SATURATE                           = 0x0308
	FUNC_ADD                                     = 0x8006
	BLEND_EQUATION                               = 0x8009
	BLEND_EQUATION_RGB                           = 0x8009
	BLEND_EQUATION_ALPHA                         = 0x883D
	FUNC_SUBTRACT                                = 0x800A
	FUNC_REVERSE_SUBTRACT                        = 0x800B
	BLEND_DST_RGB                                = 0x80C8
	BLEND_SRC_RGB                                = 0x80C9
	BLEND_DST_ALPHA                              = 0x80CA
	BLEND_SRC_ALPHA                              = 0x80CB
	CONSTANT_COLOR                               = 0x8001
	ONE_MINUS_CONSTANT_COLOR                     = 0x8002
	CONSTANT_ALPHA                               = 0x8003
	ONE_MINUS_CONSTANT_ALPHA                     = 0x8004
	BLEND_COLOR                                  = 0x8005
	ARRAY_BUFFER                                 = 0x8892
	ELEMENT_ARRAY_BUFFER                         = 0x8893
	ARRAY_BUFFER_BINDING                         = 0x8894
	ELEMENT_ARRAY_BUFFER_BINDING                 = 0x8895
	STREAM_DRAW                                  = 0x88E0
	STATIC_DRAW                                  = 0x88E4
	DYNAMIC_DRAW                                 = 0x88E8
	BUFFER_SIZE                                  = 0x8764
	BUFFER_USAGE                                 = 0x8765
	CURRENT_VERTEX_ATTRIB                        = 0x8626
	FRONT                                        = 0x0404
	BACK                                         = 0x0405
	FRONT_AND_BACK                               = 0x0408
	TEXTURE_2D                                   = 0x0DE1
	CULL_FACE                                    = 0x0B44
	BLEND                                        = 0x0BE2
	DITHER                                       = 0x0BD0
	STENCIL_TEST                                 = 0x0B90
	DEPTH_TEST                                   = 0x0B71
	SCISSOR_TEST                                 = 0x0C11
	POLYGON_OFFSET_FILL                          = 0x8037
	SAMPLE_ALPHA_TO_COVERAGE                     = 0x809E
	SAMPLE_COVERAGE                              = 0x80A0
	INVALID_ENUM                                 = 0x0500
	INVALID_VALUE                                = 0x0501
	INVALID_OPERATION                            = 0x0502
	OUT_OF_MEMORY                                = 0x0505
	CW                                           = 0x0900
	CCW                                          = 0x0901
	LINE_WIDTH                                   = 0x0B21
	ALIASED_POINT_SIZE_RANGE                     = 0x846D
	ALIASED_LINE_WIDTH_RANGE                     = 0x846E
	CULL_FACE_MODE                               = 0x0B45
	FRONT_FACE                                   = 0x0B46
	DEPTH_RANGE                                  = 0x0B70
	DEPTH_WRITEMASK                              = 0x0B72
	DEPTH_CLEAR_VALUE                            = 0x0B73
	DEPTH_FUNC                                   = 0x0B74
	STENCIL_CLEAR_VALUE                          = 0x0B91
	STENCIL_FUNC                                 = 0x0B92
	STENCIL_FAIL                                 = 0x0B94
	STENCIL_PASS_DEPTH_FAIL                      = 0x0B95
	STENCIL_PASS_DEPTH_PASS                      = 0x0B96
	STENCIL_REF                                  = 0x0B97
	STENCIL_VALUE_MASK                           = 0x0B93
	STENCIL_WRITEMASK                            = 0x0B98
	STENCIL_BACK_FUNC                            = 0x8800
	STENCIL_BACK_FAIL                            = 0x8801
	STENCIL_BACK_PASS_DEPTH_FAIL                 = 0x8802
	STENCIL_BACK_PASS_DEPTH_PASS                 = 0x8803
	STENCIL_BACK_REF                             = 0x8CA3
	STENCIL_BACK_VALUE_MASK                      = 0x8CA4
	STENCIL_BACK_WRITEMASK                       = 0x8CA5
	VIEWPORT                                     = 0x0BA2
	SCISSOR_BOX                                  = 0x0C10
	COLOR_CLEAR_VALUE                            = 0x0C22
	COLOR_WRITEMASK                              = 0x0C23
	UNPACK_ALIGNMENT                             = 0x0CF5
	PACK_ALIGNMENT                               = 0x0D05
	MAX_TEXTURE_SIZE                             = 0x0D33
	MAX_VIEWPORT_DIMS                            = 0x0D3A
	SUBPIXEL_BITS                                = 0x0D50
	RED_BITS                                     = 0x0D52
	GREEN_BITS                                   = 0x0D53
	BLUE_BITS                                    = 0x0D54
	ALPHA_BITS                                   = 0x0D55
	DEPTH_BITS                                   = 0x0D56
	STENCIL_BITS                                 = 0x0D57
	POLYGON_OFFSET_UNITS                         = 0x2A00
	POLYGON_OFFSET_FACTOR                        = 0x8038
	TEXTURE_BINDING_2D                           = 0x8069
	SAMPLE_BUFFERS                               = 0x80A8
	SAMPLES                                      = 0x80A9
	SAMPLE_COVERAGE_VALUE                        = 0x80AA
	SAMPLE_COVERAGE_INVERT                       = 0x80AB
	NUM_COMPRESSED_TEXTURE_FORMATS               = 0x86A2
	COMPRESSED_TEXTURE_FORMATS                   = 0x86A3
	DONT_CARE                                    = 0x1100
	FASTEST                                      = 0x1101
	NICEST                                       = 0x1102
	GENERATE_MIPMAP_HINT                         = 0x8192
	BYTE                                         = 0x1400
	UNSIGNED_BYTE                                = 0x1401
	SHORT                                        = 0x1402
	UNSIGNED_SHORT                               = 0x1403
	INT                                          = 0x1404
	UNSIGNED_INT                                 = 0x1405
	FLOAT                                        = 0x1406
	FIXED                                        = 0x140C
	DEPTH_COMPONENT                              = 0x1902
	ALPHA                                        = 0x1906
	RGB                                          = 0x1907
	RGBA                                         = 0x1908
	LUMINANCE                                    = 0x1909
	LUMINANCE_ALPHA                              = 0x190A
	UNSIGNED_SHORT_4_4_4_4                       = 0x8033
	UNSIGNED_SHORT_5_5_5_1                       = 0x8034
	UNSIGNED_SHORT_5_6_5                         = 0x8363
	MAX_VERTEX_ATTRIBS                           = 0x8869
	MAX_VERTEX_UNIFORM_VECTORS                   = 0x8DFB
	MAX_VARYING_VECTORS                          = 0x8DFC
	MAX_COMBINED_TEXTURE_IMAGE_UNITS             = 0x8B4D
	MAX_VERTEX_TEXTURE_IMAGE_UNITS               = 0x8B4C
	MAX_TEXTURE_IMAGE_UNITS                      = 0x8872
	MAX_FRAGMENT_UNIFORM_VECTORS                 = 0x8DFD
	SHADER_TYPE                                  = 0x8B4F
	DELETE_STATUS                                = 0x8B80
	LINK_STATUS                                  = 0x8B82
	VALIDATE_STATUS                              = 0x8B83
	ATTACHED_SHADERS                             = 0x8B85
	ACTIVE_UNIFORMS                              = 0x8B86
	ACTIVE_UNIFORM_MAX_LENGTH                    = 0x8B87
	ACTIVE_ATTRIBUTES                            = 0x8B89
	ACTIVE_ATTRIBUTE_MAX_LENGTH                  = 0x8B8A
	SHADING_LANGUAGE_VERSION                     = 0x8B8C
	CURRENT_PROGRAM                              = 0x8B8D
	NEVER                                        = 0x0200
	LESS                                         = 0x0201
	EQUAL                                        = 0x0202
	LEQUAL                                       = 0x0203
	GREATER                                      = 0x0204
	NOTEQUAL                                     = 0x0205
	GEQUAL                                       = 0x0206
	ALWAYS                                       = 0x0207
	KEEP                                         = 0x1E00
	REPLACE                                      = 0x1E01
	INCR                                         = 0x1E02
	DECR                                         = 0x1E03
	INVERT                                       = 0x150A
	INCR_WRAP                                    = 0x8507
	DECR_WRAP                                    = 0x8508
	VENDOR                                       = 0x1F00
	RENDERER                                     = 0x1F01
	VERSION                                      = 0x1F02
	EXTENSIONS                                   = 0x1F03
	NEAREST                                      = 0x2600
	LINEAR                                       = 0x2601
	NEAREST_MIPMAP_NEAREST                       = 0x2700
	LINEAR_MIPMAP_NEAREST                        = 0x2701
	NEAREST_MIPMAP_LINEAR                        = 0x2702
	LINEAR_MIPMAP_LINEAR                         = 0x2703
	TEXTURE_MAG_FILTER                           = 0x2800
	TEXTURE_MIN_FILTER                           = 0x2801
	TEXTURE_WRAP_S                               = 0x2802
	TEXTURE_WRAP_T                               = 0x2803
	TEXTURE                                      = 0x1702
	TEXTURE_CUBE_MAP                             = 0x8513
	TEXTURE_BINDING_CUBE_MAP                     = 0x8514
	TEXTURE_CUBE_MAP_POSITIVE_X                  = 0x8515
	TEXTURE_CUBE_MAP_NEGATIVE_X                  = 0x8516
	TEXTURE_CUBE_MAP_POSITIVE_Y                  = 0x8517
	TEXTURE_CUBE_MAP_NEGATIVE_Y                  = 0x8518
	TEXTURE_CUBE_MAP_POSITIVE_Z                  = 0x8519
	TEXTURE_CUBE_MAP_NEGATIVE_Z                  = 0x851A
	MAX_CUBE_MAP_TEXTURE_SIZE                    = 0x851C
	TEXTURE0                                     = 0x84C0
	TEXTURE1                                     = 0x84C1
	TEXTURE2                                     = 0x84C2
	TEXTURE3                                     = 0x84C3
	TEXTURE4                                     = 0x84C4
	TEXTURE5                                     = 0x84C5
	TEXTURE6                                     = 0x84C6
	TEXTURE7                                     = 0x84C7
	TEXTURE8                                     = 0x84C8
	TEXTURE9                                     = 0x84C9
	TEXTURE10                                    = 0x84CA
	TEXTURE11                                    = 0x84CB
	TEXTURE12                                    = 0x84CC
	TEXTURE13                                    = 0x84CD
	TEXTURE14                                    = 0x84CE
	TEXTURE15                                    = 0x84CF
	TEXTURE16                                    = 0x84D0
	TEXTURE17                                    = 0x84D1
	TEXTURE18                                    = 0x84D2
	TEXTURE19                                    = 0x84D3
	TEXTURE20                                    = 0x84D4
	TEXTURE21                                    = 0x84D5
	TEXTURE22                                    = 0x84D6
	TEXTURE23                                    = 0x84D7
	TEXTURE24                                    = 0x84D8
	TEXTURE25                                    = 0x84D9
	TEXTURE26                                    = 0x84DA
	TEXTURE27                                    = 0x84DB
	TEXTURE28                                    = 0x84DC
	TEXTURE29                                    = 0x84DD
	TEXTURE30                                    = 0x84DE
	TEXTURE31                                    = 0x84DF
	ACTIVE_TEXTURE                               = 0x84E0
	REPEAT                                       = 0x2901
	CLAMP_TO_EDGE                                = 0x812F
	MIRRORED_REPEAT                              = 0x8370
	VERTEX_ATTRIB_ARRAY_ENABLED                  = 0x8622
	VERTEX_ATTRIB_ARRAY_SIZE                     = 0x8623
	VERTEX_ATTRIB_ARRAY_STRIDE                   = 0x8624
	VERTEX_ATTRIB_ARRAY_TYPE                     = 0x8625
	VERTEX_ATTRIB_ARRAY_NORMALIZED               = 0x886A
	VERTEX_ATTRIB_ARRAY_POINTER                  = 0x8645
	VERTEX_ATTRIB_ARRAY_BUFFER_BINDING           = 0x889F
	IMPLEMENTATION_COLOR_READ_TYPE               = 0x8B9A
	IMPLEMENTATION_COLOR_READ_FORMAT             = 0x8B9B
	COMPILE_STATUS                               = 0x8B81
	INFO_LOG_LENGTH                              = 0x8B84
	SHADER_SOURCE_LENGTH                         = 0x8B88
	SHADER_COMPILER                              = 0x8DFA
	SHADER_BINARY_FORMATS                        = 0x8DF8
	NUM_SHADER_BINARY_FORMATS                    = 0x8DF9
	LOW_FLOAT                                    = 0x8DF0
	MEDIUM_FLOAT                                 = 0x8DF1
	HIGH_FLOAT                                   = 0x8DF2
	LOW_INT                                      = 0x8DF3
	MEDIUM_INT                                   = 0x8DF4
	HIGH_INT                                     = 0x8DF5
	FRAMEBUFFER                                  = 0x8D40
	RENDERBUFFER                                 = 0x8D41
	RGBA4                                        = 0x8056
	RGB5_A1                                      = 0x8057
	RGB565                                       = 0x8D62
	DEPTH_COMPONENT16                            = 0x81A5
	STENCIL_INDEX8                               = 0x8D48
	RENDERBUFFER_WIDTH                           = 0x8D42
	RENDERBUFFER_HEIGHT                          = 0x8D43
	RENDERBUFFER_INTERNAL_FORMAT                 = 0x8D44
	RENDERBUFFER_RED_SIZE                        = 0x8D50
	RENDERBUFFER_GREEN_SIZE                      = 0x8D51
	RENDERBUFFER_BLUE_SIZE                       = 0x8D52
	RENDERBUFFER_ALPHA_SIZE                      = 0x8D53
	RENDERBUFFER_DEPTH_SIZE                      = 0x8D54
	RENDERBUFFER_STENCIL_SIZE                    = 0x8D55
	FRAMEBUFFER_ATTACHMENT_OBJECT_TYPE           = 0x8CD0
	FRAMEBUFFER_ATTACHMENT_OBJECT_NAME           = 0x8CD1
	FRAMEBUFFER_ATTACHMENT_TEXTURE_LEVEL         = 0x8CD2
	FRAMEBUFFER_ATTACHMENT_TEXTURE_CUBE_MAP_FACE = 0x8CD3
	COLOR_ATTACHMENT0                            = 0x8CE0
	DEPTH_ATTACHMENT                             = 0x8D00
	STENCIL_ATTACHMENT                           = 0x8D20
	FRAMEBUFFER_COMPLETE                         = 0x8CD5
	FRAMEBUFFER_INCOMPLETE_ATTACHMENT            = 0x8CD6
	FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT    = 0x8CD7
	FRAMEBUFFER_INCOMPLETE_DIMENSIONS            = 0x8CD9
	FRAMEBUFFER_UNSUPPORTED                      = 0x8CDD
	FRAMEBUFFER_BINDING                          = 0x8CA6
	RENDERBUFFER_BINDING                         = 0x8CA7
	MAX_RENDERBUFFER_SIZE                        = 0x84E8
	INVALID_FRAMEBUFFER_OPERATION                = 0x0506
)

const (
	DEPTH_BUFFER_BIT   = 0x00000100
	STENCIL_BUFFER_BIT = 0x00000400
	COLOR_BUFFER_BIT   = 0x00004000
)

const (
	FLOAT_VEC2   = 0x8B50
	FLOAT_VEC3   = 0x8B51
	FLOAT_VEC4   = 0x8B52
	INT_VEC2     = 0x8B53
	INT_VEC3     = 0x8B54
	INT_VEC4     = 0x8B55
	BOOL         = 0x8B56
	BOOL_VEC2    = 0x8B57
	BOOL_VEC3    = 0x8B58
	BOOL_VEC4    = 0x8B59
	FLOAT_MAT2   = 0x8B5A
	FLOAT_MAT3   = 0x8B5B
	FLOAT_MAT4   = 0x8B5C
	SAMPLER_2D   = 0x8B5E
	SAMPLER_CUBE = 0x8B60
)

const (
	FRAGMENT_SHADER = 0x8B30
	VERTEX_SHADER   = 0x8B31
)

const (
	FALSE    = 0
	TRUE     = 1
	ZERO     = 0
	ONE      = 1
	NO_ERROR = 0
	NONE     = 0
)

// Attrib is an attribute index.
type Attrib uint

// Enum is equivalent to GLenum, and is normally used with one of the
// constants defined in this package.
type Enum uint32

// Program identifies a compiled shader program.
type Program uint32

// Shader identifies a GLSL shader.
type Shader uint32

// Buffer identifies a GL buffer object.
type Buffer uint32

// Framebuffer identifies a GL framebuffer.
type Framebuffer uint32

// A Renderbuffer is a GL object that holds an image in an internal format.
type Renderbuffer uint32

// A Texture identifies a GL texture unit.
type Texture uint32

// A Uniform identifies a GL uniform attribute value.
type Uniform int32

func (v Attrib) c() C.GLuint { return C.GLuint(v) }

func (v Enum) c() C.GLenum { return C.GLenum(v) }

func (v Program) c() C.GLuint { return C.GLuint(v) }

func (v Shader) c() C.GLuint { return C.GLuint(v) }

func (v Buffer) c() C.GLuint { return C.GLuint(v) }

func (v Framebuffer) c() C.GLuint { return C.GLuint(v) }

func (v Renderbuffer) c() C.GLuint { return C.GLuint(v) }

func (v Texture) c() C.GLuint { return C.GLuint(v) }

func (v Uniform) c() C.GLint { return C.GLint(v) }

// ActiveTexture sets the active texture unit.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glActiveTexture.xhtml
func ActiveTexture(texture Enum) {
	C.glActiveTexture(texture.c())
}

// AttachShader attaches a shader to a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glAttachShader.xhtml
func AttachShader(p Program, s Shader) {
	C.glAttachShader(p.c(), s.c())
}

// BindAttribLocation binds a vertex attribute index with a named
// variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindAttribLocation.xhtml
func BindAttribLocation(p Program, a Attrib, name string) {
	str := unsafe.Pointer(C.CString(name))
	defer C.free(str)
	C.glBindAttribLocation(p.c(), a.c(), (*C.GLchar)(str))
}

// BindBuffer binds a buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindBuffer.xhtml
func BindBuffer(target Enum, b Buffer) {
	C.glBindBuffer(target.c(), b.c())
}

// BindFramebuffer binds a framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindFramebuffer.xhtml
func BindFramebuffer(target Enum, fb Framebuffer) {
	C.glBindFramebuffer(target.c(), fb.c())
}

// BindRenderbuffer binds a render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindRenderbuffer.xhtml
func BindRenderbuffer(target Enum, rb Renderbuffer) {
	C.glBindRenderbuffer(target.c(), rb.c())
}

// BindTexture binds a texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindTexture.xhtml
func BindTexture(target Enum, t Texture) {
	C.glBindTexture(target.c(), t.c())
}

// BlendColor sets the blend color.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendColor.xhtml
func BlendColor(red, green, blue, alpha float32) {
	C.glBlendColor(C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}

// BlendEquation sets both RGB and alpha blend equations.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendEquation.xhtml
func BlendEquation(mode Enum) {
	C.glBlendEquation(mode.c())
}

// BlendEquationSeparate sets RGB and alpha blend equations separately.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendEquationSeparate.xhtml
func BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	C.glBlendEquationSeparate(modeRGB.c(), modeAlpha.c())
}

// BlendFunc sets the pixel blending factors.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendFunc.xhtml
func BlendFunc(sfactor, dfactor Enum) {
	C.glBlendFunc(sfactor.c(), dfactor.c())
}

// BlendFunc sets the pixel RGB and alpha blending factors separately.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendFuncSeparate.xhtml
func BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	C.glBlendFuncSeparate(sfactorRGB.c(), dfactorRGB.c(), sfactorAlpha.c(), dfactorAlpha.c())
}

// BufferData creates a new data store for the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferData.xhtml
func BufferData(target Enum, usage Enum, src []byte) {
	C.glBufferData(target.c(), C.GLsizeiptr(len(src)), unsafe.Pointer(&src[0]), usage.c())
}

// BufferInit creates a new unitialized data store for the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferData.xhtml
func BufferInit(target Enum, size int, usage Enum) {
	C.glBufferData(target.c(), C.GLsizeiptr(size), nil, usage.c())
}

// BufferSubData sets some of data in the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferSubData.xhtml
func BufferSubData(target Enum, offset int, data []byte) {
	C.glBufferSubData(target.c(), C.GLintptr(offset), C.GLsizeiptr(len(data)), unsafe.Pointer(&data[0]))
}

// CheckFramebufferStatus reports the completeness status of the
// active framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCheckFramebufferStatus.xhtml
func CheckFramebufferStatus(target Enum) Enum {
	return Enum(C.glCheckFramebufferStatus(target.c()))
}

// Clear clears the window.
//
// The behavior of Clear is influenced by the pixel ownership test,
// the scissor test, dithering, and the buffer writemasks.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClear.xhtml
func Clear(mask Enum) {
	C.glClear(C.GLbitfield(mask))
}

// ClearColor specifies the RGBA values used to clear color buffers.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearColor.xhtml
func ClearColor(red, green, blue, alpha float32) {
	C.glClearColor(C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}

// ClearDepthf sets the depth value used to clear the depth buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearDepthf.xhtml
func ClearDepthf(d float32) {
	C.glClearDepthf(C.GLfloat(d))
}

// ClearStencil sets the index used to clear the stencil buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearStencil.xhtml
func ClearStencil(s int) {
	C.glClearStencil(C.GLint(s))
}

// ColorMask specifies whether color components in the framebuffer
// can be written.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glColorMask.xhtml
func ColorMask(red, green, blue, alpha bool) {
	C.glColorMask(glBoolean(red), glBoolean(green), glBoolean(blue), glBoolean(alpha))
}

// CompileShader compiles the source code of s.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompileShader.xhtml
func CompileShader(s Shader) {
	C.glCompileShader(s.c())
}

// CompressedTexImage2D writes a compressed 2D texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompressedTexImage2D.xhtml
func CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data []byte) {
	C.glCompressedTexImage2D(target.c(), C.GLint(level), internalformat.c(), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLsizei(len(data)), unsafe.Pointer(&data[0]))
}

// CompressedTexSubImage2D writes a subregion of a compressed 2D texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompressedTexSubImage2D.xhtml
func CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data []byte) {
	C.glCompressedTexSubImage2D(target.c(), C.GLint(level), C.GLint(xoffset), C.GLint(yoffset), C.GLsizei(width), C.GLsizei(height), format.c(), C.GLsizei(len(data)), unsafe.Pointer(&data[0]))
}

// CopyTexImage2D writes a 2D texture from the current framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCopyTexImage2D.xhtml
func CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
	C.glCopyTexImage2D(target.c(), C.GLint(level), internalformat.c(), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLint(border))
}

// CopyTexSubImage2D writes a 2D texture subregion from the
// current framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCopyTexSubImage2D.xhtml
func CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	C.glCopyTexSubImage2D(target.c(), C.GLint(level), C.GLint(xoffset), C.GLint(yoffset), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

// CreateProgram creates a new empty program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCreateProgram.xhtml
func CreateProgram() Program {
	return Program(C.glCreateProgram())
}

// CreateShader creates a new empty shader object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCreateShader.xhtml
func CreateShader(ty Enum) Shader {
	return Shader(C.glCreateShader(ty.c()))
}

// CullFace specifies which polygons are candidates for culling.
//
// Valid modes: FRONT, BACK, FRONT_AND_BACK.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCullFace.xhtml
func CullFace(mode Enum) {
	C.glCullFace(mode.c())
}

// DeleteBuffers deletes the given buffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteBuffers.xhtml
func DeleteBuffers(v []Buffer) {
	C.glDeleteBuffers(C.GLsizei(len(v)), (*C.GLuint)(unsafe.Pointer(&v[0])))
}

// DeleteFramebuffers deletes the given framebuffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteFramebuffers.xhtml
func DeleteFramebuffers(v []Framebuffer) {
	C.glDeleteFramebuffers(C.GLsizei(len(v)), (*C.GLuint)(unsafe.Pointer(&v[0])))
}

// DeleteProgram deletes the given program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteProgram.xhtml
func DeleteProgram(p Program) {
	C.glDeleteProgram(p.c())
}

// DeleteRenderbuffers deletes the given render buffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteRenderbuffers.xhtml
func DeleteRenderbuffers(v []Renderbuffer) {
	C.glDeleteRenderbuffers(C.GLsizei(len(v)), (*C.GLuint)(unsafe.Pointer(&v[0])))
}

// DeleteShader deletes shader s.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteShader.xhtml
func DeleteShader(s Shader) {
	C.glDeleteShader(s.c())
}

// DeleteTextures deletes the given textures.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteTextures.xhtml
func DeleteTextures(v []Texture) {
	C.glDeleteTextures(C.GLsizei(len(v)), (*C.GLuint)(unsafe.Pointer(&v[0])))
}

// DepthFunc sets the function used for depth buffer comparisons.
//
// Valid fn values:
//	NEVER
//	LESS
//	EQUAL
//	LEQUAL
//	GREATER
//	NOTEQUAL
//	GEQUAL
//	ALWAYS
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDepthFunc.xhtml
func DepthFunc(fn Enum) {
	C.glDepthFunc(fn.c())
}

// DepthMask sets the depth buffer enabled for writing.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDepthMask.xhtml
func DepthMask(flag bool) {
	C.glDepthMask(glBoolean(flag))
}

// DepthRangef sets the mapping from normalized device coordinates to
// window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDepthRangef.xhtml
func DepthRangef(n, f float32) {
	C.glDepthRangef(C.GLfloat(n), C.GLfloat(f))
}

// DetachShader detaches the shader s from the program p.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDetachShader.xhtml
func DetachShader(p Program, s Shader) {
	C.glDetachShader(p.c(), s.c())
}

// Disable disables various GL capabilities.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDisable.xhtml
func Disable(cap Enum) {
	C.glDisable(cap.c())
}

// http://www.khronos.org/opengles/sdk/docs/man3/html/glDisableVertexAttribArray.xhtml
func DisableVertexAttribArray(index Attrib) {
	C.glDisableVertexAttribArray(index.c())
}

// DrawArrays renders geometric primitives from the bound data.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDrawArrays.xhtml
func DrawArrays(mode Enum, first, count int) {
	C.glDrawArrays(mode.c(), C.GLint(first), C.GLsizei(count))
}

// DrawElements renders primitives from a bound buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDrawElements.xhtml
func DrawElements(mode, ty Enum, offset, count int) {
	C.glDrawElements(mode.c(), C.GLsizei(count), ty.c(), unsafe.Pointer(uintptr(offset)))
}

// TODO(crawshaw): consider DrawElements8 / DrawElements16 / DrawElements32

// Enable enables various GL capabilities.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glEnable.xhtml
func Enable(cap Enum) {
	C.glEnable(cap.c())
}

// EnableVertexAttribArray enables a vertex attribute array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glEnableVertexAttribArray.xhtml
func EnableVertexAttribArray(index Attrib) {
	C.glEnableVertexAttribArray(index.c())
}

// Finish blocks until the effects of all previously called GL
// commands are complete.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFinish.xhtml
func Finish() {
	C.glFinish()
}

// Flush empties all buffers. It does not block.
//
// An OpenGL implementation may buffer network communication,
// the command stream, or data inside the graphics accelerator.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFlush.xhtml
func Flush() {
	C.glFlush()
}

// FramebufferRenderbuffer attaches rb to the current frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFramebufferRenderbuffer.xhtml
func FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
	C.glFramebufferRenderbuffer(target.c(), attachment.c(), rbTarget.c(), rb.c())
}

// FramebufferTexture2D attaches the t to the current frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFramebufferTexture2D.xhtml
func FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	C.glFramebufferTexture2D(target.c(), attachment.c(), texTarget.c(), t.c(), C.GLint(level))
}

// FrontFace defines which polygons are front-facing.
//
// Valid modes: CW, CCW.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFrontFace.xhtml
func FrontFace(mode Enum) {
	C.glFrontFace(mode.c())
}

// GenBuffers creates n buffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenBuffers.xhtml
func GenBuffers(n int) []Buffer {
	buf := make([]Buffer, n)
	C.glGenBuffers(C.GLsizei(n), (*C.GLuint)(&buf[0]))
	return buf
}

// GenerateMipmap generates mipmaps for the current texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenerateMipmap.xhtml
func GenerateMipmap(target Enum) {
	C.glGenerateMipmap(target.c())
}

// GenFramebuffers creates n framebuffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenFramebuffers.xhtml
func GenFramebuffers(n int) []Framebuffer {
	buf := make([]Framebuffer, n)
	C.glGenFramebuffers(C.GLsizei(n), (*C.GLuint)(&buf[0]))
	return buf
}

// GenRenderbuffers creates n renderbuffer objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenRenderbuffers.xhtml
func GenRenderbuffers(n int) []Renderbuffer {
	buf := make([]Renderbuffer, n)
	C.glGenRenderbuffers(C.GLsizei(n), (*C.GLuint)(&buf[0]))
	return buf
}

// GenTextures creates n texture objects.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenTextures.xhtml
func GenTextures(n int) []Texture {
	buf := make([]Texture, n)
	C.glGenTextures(C.GLsizei(n), (*C.GLuint)(&buf[0]))
	return buf
}

// GetActiveAttrib returns details about an attribute variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetActiveAttrib.xhtml
func GetActiveAttrib(p Program, a Attrib) (name string, size int, ty Enum) {
	bufSize := GetProgrami(p, ACTIVE_ATTRIBUTE_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)

	var cSize C.GLint
	var cType C.GLenum
	C.glGetActiveAttrib(p.c(), a.c(), C.GLsizei(bufSize), nil, &cSize, &cType, (*C.GLchar)(buf))
	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

// GetActiveUniform returns details about an active uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetActiveUniform.xhtml
func GetActiveUniform(p Program, u Uniform) (name string, size int, ty Enum) {
	bufSize := GetProgrami(p, ACTIVE_UNIFORM_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)

	var cSize C.GLint
	var cType C.GLenum

	C.glGetActiveUniform(p.c(), C.GLuint(u), C.GLsizei(bufSize), nil, &cSize, &cType, (*C.GLchar)(buf))
	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

// GetAttachedShaders returns the shader objects attached to program p.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetAttachedShaders.xhtml
func GetAttachedShaders(p Program) []Shader {
	shadersLen := GetProgrami(p, ATTACHED_SHADERS)
	buf := make([]Shader, shadersLen)
	var n C.GLsizei
	C.glGetAttachedShaders(p.c(), C.GLsizei(shadersLen), &n, (*C.GLuint)(&buf[0]))
	return buf[:int(n)]
}

// GetAttribLocation finds a program attribute variable by name.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetAttribLocation.xhtml
func GetAttribLocation(p Program, name string) Attrib {
	str := unsafe.Pointer(C.CString(name))
	defer C.free(str)
	return Attrib(C.glGetAttribLocation(p.c(), (*C.GLchar)(str)))
}

// GetBooleanv returns the boolean values of parameter pname.
//
// Many boolean parameters can be queried more easily using IsEnabled.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetBooleanv(dst []bool, pname Enum) {
	buf := make([]C.GLboolean, len(dst))
	C.glGetBooleanv(pname.c(), &buf[0])
	for i, v := range buf {
		dst[i] = v != 0
	}
}

// GetFloatv returns the float values of parameter pname.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetFloatv(dst []float32, pname Enum) {
	C.glGetFloatv(pname.c(), (*C.GLfloat)(&dst[0]))
}

// GetIntegerv returns the int values of parameter pname.
//
// Single values may be queried more easily using GetInteger.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetIntegerv(pname Enum, data []int32) {
	buf := make([]C.GLint, len(data))
	C.glGetIntegerv(pname.c(), &buf[0])
	for i, v := range buf {
		data[i] = int32(v)
	}
}

// GetInteger returns the int value of parameter pname.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetInteger(pname Enum) int {
	var v C.GLint
	C.glGetIntegerv(pname.c(), &v)
	return int(v)
}

// GetBufferParameteri returns a parameter for the active buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetBufferParameteriv.xhtml
func GetBufferParameteri(target, pname Enum) int {
	var params C.GLint
	C.glGetBufferParameteriv(target.c(), pname.c(), &params)
	return int(params)
}

// GetError returns the next error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetError.xhtml
func GetError() Enum {
	return Enum(C.glGetError())
}

// GetFramebufferAttachmentParameteri returns attachment parameters
// for the active framebuffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetFramebufferAttachmentParameteriv.xhtml
func GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	var params C.GLint
	C.glGetFramebufferAttachmentParameteriv(target.c(), attachment.c(), pname.c(), &params)
	return int(params)
}

// GetProgrami returns a parameter value for a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetProgramiv.xhtml
func GetProgrami(p Program, pname Enum) int {
	var params C.GLint
	C.glGetProgramiv(p.c(), pname.c(), &params)
	return int(params)
}

// GetProgramInfoLog returns the information log for a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetProgramInfoLog.xhtml
func GetProgramInfoLog(p Program) string {
	infoLen := GetProgrami(p, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	C.free(buf)
	C.glGetProgramInfoLog(p.c(), C.GLsizei(infoLen), nil, (*C.GLchar)(buf))
	return C.GoString((*C.char)(buf))
}

// GetShaderi returns a parameter value for a render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetRenderbufferParameteriv.xhtml
func GetRenderbufferParameteri(target, pname Enum) int {
	var params C.GLint
	C.glGetRenderbufferParameteriv(target.c(), pname.c(), &params)
	return int(params)
}

// GetRenderbufferParameteri returns a parameter value for a shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderiv.xhtml
func GetShaderi(s Shader, pname Enum) int {
	var params C.GLint
	C.glGetShaderiv(s.c(), pname.c(), &params)
	return int(params)
}

// GetShaderInfoLog returns the information log for a shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderInfoLog.xhtml
func GetShaderInfoLog(s Shader) string {
	infoLen := GetShaderi(s, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	defer C.free(buf)
	C.glGetShaderInfoLog(s.c(), C.GLsizei(infoLen), nil, (*C.GLchar)(buf))
	return C.GoString((*C.char)(buf))
}

// GetShaderPrecisionFormat returns range and precision limits for
// shader types.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderPrecisionFormat.xhtml
func GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeLow, rangeHigh, precision int) {
	const glintSize = 4
	var cRange [2]C.GLint
	var cPrecision C.GLint

	C.glGetShaderPrecisionFormat(shadertype.c(), precisiontype.c(), &cRange[0], &cPrecision)
	return int(cRange[0]), int(cRange[1]), int(cPrecision)
}

// GetShaderSource returns source code of shader s.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderSource.xhtml
func GetShaderSource(s Shader) string {
	sourceLen := GetShaderi(s, SHADER_SOURCE_LENGTH)
	if sourceLen == 0 {
		return ""
	}
	buf := C.malloc(C.size_t(sourceLen))
	defer C.free(buf)
	C.glGetShaderSource(s.c(), C.GLsizei(sourceLen), nil, (*C.GLchar)(buf))
	return C.GoString((*C.char)(buf))
}

// GetString reports current GL state.
//
// Valid name values:
//	EXTENSIONS
//	RENDERER
//	SHADING_LANGUAGE_VERSION
//	VENDOR
//	VERSION
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetString.xhtml
func GetString(name Enum) string {
	// Bounce through unsafe.Pointer, because on some platforms
	// GetString returns an *unsigned char which doesn't convert.
	return C.GoString((*C.char)((unsafe.Pointer)(C.glGetString(name.c()))))
}

// GetTexParameterfv returns the float values of a texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetTexParameter.xhtml
func GetTexParameterfv(dst []float32, target, pname Enum) {
	C.glGetTexParameterfv(target.c(), pname.c(), (*C.GLfloat)(&dst[0]))
}

// GetTexParameteriv returns the int values of a texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetTexParameter.xhtml
func GetTexParameteriv(dst []int32, target, pname Enum) {
	C.glGetTexParameteriv(target.c(), pname.c(), (*C.GLint)(&dst[0]))
}

// GetUniformfv returns the float values of a uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniform.xhtml
func GetUniformfv(dst []float32, src Uniform, p Program) {
	C.glGetUniformfv(p.c(), src.c(), (*C.GLfloat)(&dst[0]))
}

// GetUniformiv returns the float values of a uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniform.xhtml
func GetUniformiv(dst []int32, src Uniform, p Program) {
	C.glGetUniformiv(p.c(), src.c(), (*C.GLint)(&dst[0]))
}

// GetUniformLocation returns the location of uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniformLocation.xhtml
func GetUniformLocation(p Program, name string) Uniform {
	str := C.CString(name)
	defer C.free((unsafe.Pointer)(str))
	return Uniform(C.glGetUniformLocation(p.c(), (*C.GLchar)(str)))
}

// GetVertexAttribf reads the float value of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribf(src Attrib, pname Enum) float32 {
	var params C.GLfloat
	C.glGetVertexAttribfv(src.c(), pname.c(), &params)
	return float32(params)
}

// GetVertexAttribfv reads float values of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
	C.glGetVertexAttribfv(src.c(), pname.c(), (*C.GLfloat)(&dst[0]))
}

// GetVertexAttribi reads the int value of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribi(src Attrib, pname Enum) int32 {
	var params C.GLint
	C.glGetVertexAttribiv(src.c(), pname.c(), &params)
	return int32(params)
}

// GetVertexAttribiv reads int values of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
	C.glGetVertexAttribiv(src.c(), pname.c(), (*C.GLint)(&dst[0]))
}

// TODO(crawshaw): glGetVertexAttribPointerv

// Hint sets implementation-specific modes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glHint.xhtml
func Hint(target, mode Enum) {
	C.glHint(target.c(), mode.c())
}

// IsBuffer reports if b is a valid buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsBuffer.xhtml
func IsBuffer(b Buffer) bool {
	return C.glIsBuffer(b.c()) != 0
}

// IsEnabled reports if cap is an enabled capability.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsEnabled.xhtml
func IsEnabled(cap Enum) bool {
	return C.glIsEnabled(cap.c()) != 0
}

// IsFramebuffer reports if fb is a valid frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsFramebuffer.xhtml
func IsFramebuffer(fb Framebuffer) bool {
	return C.glIsFramebuffer(fb.c()) != 0
}

// IsProgram reports if p is a valid program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsProgram.xhtml
func IsProgram(p Program) bool {
	return C.glIsProgram(p.c()) != 0
}

// IsRenderbuffer reports if rb is a valid render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsRenderbuffer.xhtml
func IsRenderbuffer(rb Renderbuffer) bool {
	return C.glIsRenderbuffer(rb.c()) != 0
}

// IsShader reports if s is valid shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsShader.xhtml
func IsShader(s Shader) bool {
	return C.glIsShader(s.c()) != 0
}

// IsTexture reports if t is a valid texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsTexture.xhtml
func IsTexture(t Texture) bool {
	return C.glIsTexture(t.c()) != 0
}

// LineWidth specifies the width of lines.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glLineWidth.xhtml
func LineWidth(width float32) {
	C.glLineWidth(C.GLfloat(width))
}

// LinkProgram links the specified program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glLinkProgram.xhtml
func LinkProgram(p Program) {
	C.glLinkProgram(p.c())
}

// PixelStorei sets pixel storage parameters.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glPixelStorei.xhtml
func PixelStorei(pname Enum, param int32) {
	C.glPixelStorei(pname.c(), C.GLint(param))
}

// PolygonOffset sets the scaling factors for depth offsets.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glPolygonOffset.xhtml
func PolygonOffset(factor, units float32) {
	C.glPolygonOffset(C.GLfloat(factor), C.GLfloat(units))
}

// ReadPixels returns pixel data from a buffer.
//
// In GLES 3, the source buffer is controlled with ReadBuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glReadPixels.xhtml
func ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
	// TODO(crawshaw): support PIXEL_PACK_BUFFER in GLES3, uses offset.
	C.glReadPixels(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), format.c(), ty.c(), unsafe.Pointer(&dst[0]))
}

// ReleaseShaderCompiler frees resources allocated by the shader compiler.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glReleaseShaderCompiler.xhtml
func ReleaseShaderCompiler() {
	C.glReleaseShaderCompiler()
}

// RenderbufferStorage establishes the data storage, format, and
// dimensions of a renderbuffer object's image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glRenderbufferStorage.xhtml
func RenderbufferStorage(target, internalFormat Enum, width, height int) {
	C.glRenderbufferStorage(target.c(), internalFormat.c(), C.GLsizei(width), C.GLsizei(height))
}

// SampleCoverage sets multisample coverage parameters.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glSampleCoverage.xhtml
func SampleCoverage(value float32, invert bool) {
	C.glSampleCoverage(C.GLfloat(value), glBoolean(invert))
}

func glBoolean(b bool) C.GLboolean {
	if b {
		return 0
	}
	return 1
}

// Scissor defines the scissor box rectangle, in window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glScissor.xhtml
func Scissor(x, y, width, height int32) {
	C.glScissor(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

// TODO(crawshaw): ShaderBinary

// ShaderSource sets the source code of s to the given source code.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glShaderSource.xhtml
func ShaderSource(s Shader, src string) {
	str := (*C.GLchar)(C.CString(src))
	defer C.free(unsafe.Pointer(str))
	C.glShaderSource(s.c(), 1, &str, nil)
}

//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilFunc.xhtml
func StencilFunc(fn Enum, ref int, mask uint32) {
	C.glStencilFunc(fn.c(), C.GLint(ref), C.GLuint(mask))
}

//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilFuncSeparate.xhtml
func StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
	C.glStencilFuncSeparate(face.c(), fn.c(), C.GLint(ref), C.GLuint(mask))
}

// StencilMask controls the writing of bits in the stencil planes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilMask.xhtml
func StencilMask(mask uint32) {
	C.glStencilMask(C.GLuint(mask))
}

// StencilMaskSeparate controls the writing of bits in the stencil planes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilMaskSeparate.xhtml
func StencilMaskSeparate(face Enum, mask uint32) {
	C.glStencilMaskSeparate(face.c(), C.GLuint(mask))
}

// StencilOp sets front and back stencil test actions.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilOp.xhtml
func StencilOp(fail, zfail, zpass Enum) {
	C.glStencilOp(fail.c(), zfail.c(), zpass.c())
}

// StencilOpSeparate sets front or back stencil tests.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilOpSeparate.xhtml
func StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
	C.glStencilOpSeparate(face.c(), sfail.c(), dpfail.c(), dppass.c())
}

// TexImage2D writes a 2D texture image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexImage2D.xhtml
func TexImage2D(target Enum, level int, width, height int, format Enum, ty Enum, data []byte) {
	// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
	p := unsafe.Pointer(nil)
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glTexImage2D(target.c(), C.GLint(level), C.GLint(format), C.GLsizei(width), C.GLsizei(height), 0, format.c(), ty.c(), p)
}

// TexSubImage2D writes a subregion of a 2D texture image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexSubImage2D.xhtml
func TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
	C.glTexSubImage2D(target.c(), C.GLint(level), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), format.c(), ty.c(), unsafe.Pointer(&data[0]))
}

// TexParameterf sets a float texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameterf(target, pname Enum, param float32) {
	C.glTexParameterf(target.c(), pname.c(), C.GLfloat(param))
}

// TexParameterfv sets a float texture parameter array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameterfv(target, pname Enum, params []float32) {
	C.glTexParameterfv(target.c(), pname.c(), (*C.GLfloat)(&params[0]))
}

// TexParameteri sets an integer texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameteri(target, pname Enum, param int) {
	C.glTexParameteri(target.c(), pname.c(), C.GLint(param))
}

// TexParameteriv sets an integer texture parameter array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameteriv(target, pname Enum, params []int32) {
	C.glTexParameteriv(target.c(), pname.c(), (*C.GLint)(&params[0]))
}

// Uniform1f writes a float uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1f(dst Uniform, v float32) {
	C.glUniform1f(dst.c(), C.GLfloat(v))
}

// Uniform1fv writes a [len(src)]float uniform array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1fv(dst Uniform, src []float32) {
	C.glUniform1fv(dst.c(), C.GLsizei(len(src)), (*C.GLfloat)(&src[0]))
}

// Uniform1i writes an int uniform variable.
//
// Uniform1i and Uniform1iv are the only two functions that may be used
// to load uniform variables defined as sampler types. Loading samplers
// with any other function will result in a INVALID_OPERATION error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1i(dst Uniform, v int) {
	C.glUniform1i(dst.c(), C.GLint(v))
}

// Uniform1iv writes a int uniform array of len(src) elements.
//
// Uniform1i and Uniform1iv are the only two functions that may be used
// to load uniform variables defined as sampler types. Loading samplers
// with any other function will result in a INVALID_OPERATION error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1iv(dst Uniform, src []int32) {
	C.glUniform1iv(dst.c(), C.GLsizei(len(src)), (*C.GLint)(&src[0]))
}

// Uniform2f writes a vec2 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2f(dst Uniform, v0, v1 float32) {
	C.glUniform2f(dst.c(), C.GLfloat(v0), C.GLfloat(v1))
}

// Uniform2fv writes a vec2 uniform array of len(src)/2 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2fv(dst Uniform, src []float32) {
	C.glUniform2fv(dst.c(), C.GLsizei(len(src)/2), (*C.GLfloat)(&src[0]))
}

// Uniform2i writes an ivec2 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2i(location, v0, v1 int) {
	C.glUniform2i(C.GLint(location), C.GLint(v0), C.GLint(v1))
}

// Uniform2iv writes an ivec2 uniform array of len(src)/2 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2iv(dst Uniform, src []int32) {
	C.glUniform2iv(dst.c(), C.GLsizei(len(src)/2), (*C.GLint)(&src[0]))
}

// Uniform3f writes a vec3 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3f(dst Uniform, v0, v1, v2 float32) {
	C.glUniform3f(dst.c(), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2))
}

// Uniform3fv writes a vec3 uniform array of len(src)/3 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3fv(dst Uniform, src []float32) {
	C.glUniform3fv(dst.c(), C.GLsizei(len(src)/3), (*C.GLfloat)(&src[0]))
}

// Uniform3i writes an ivec3 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3i(location, v0, v1, v2 int32) {
	C.glUniform3i(C.GLint(location), C.GLint(v0), C.GLint(v1), C.GLint(v2))
}

// Uniform3iv writes an ivec3 uniform array of len(src)/3 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3iv(dst Uniform, src []int32) {
	C.glUniform3iv(dst.c(), C.GLsizei(len(src)/3), (*C.GLint)(&src[0]))
}

// Uniform4f writes a vec4 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	C.glUniform4f(dst.c(), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2), C.GLfloat(v3))
}

// Uniform4fv writes a vec4 uniform array of len(src)/4 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4fv(dst Uniform, src []float32) {
	C.glUniform4fv(dst.c(), C.GLsizei(len(src)/4), (*C.GLfloat)(&src[0]))
}

// Uniform4i writes an ivec4 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
	C.glUniform4i(dst.c(), C.GLint(v0), C.GLint(v1), C.GLint(v2), C.GLint(v3))
}

// Uniform4i writes an ivec4 uniform array of len(src)/4 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4iv(dst Uniform, src []int32) {
	C.glUniform4iv(dst.c(), C.GLsizei(len(src)/4), (*C.GLint)(&src[0]))
}

// UniformMatrix2fv writes 2x2 matricies. Each matrix uses four
// float32 values, so the number of matricies written is len(src)/4.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix2fv(dst Uniform, src []float32) {
	/// OpenGL ES 2 does not support transpose.
	C.glUniformMatrix2fv(dst.c(), C.GLsizei(len(src)/4), 0, (*C.GLfloat)(&src[0]))
}

// UniformMatrix3fv writes 3x3 matricies. Each matrix uses nine
// float32 values, so the number of matricies written is len(src)/9.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix3fv(dst Uniform, src []float32) {
	C.glUniformMatrix3fv(dst.c(), C.GLsizei(len(src)/9), 0, (*C.GLfloat)(&src[0]))
}

// UniformMatrix4fv writes 4x4 matricies. Each matrix uses 16
// float32 values, so the number of matricies written is len(src)/16.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix4fv(dst Uniform, src []float32) {
	C.glUniformMatrix4fv(dst.c(), C.GLsizei(len(src)/16), 0, (*C.GLfloat)(&src[0]))
}

// UseProgram sets the active program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUseProgram.xhtml
func UseProgram(p Program) {
	C.glUseProgram(p.c())
}

// ValidateProgram checks to see whether the executables contained in
// program can execute given the current OpenGL state.
//
// Typically only used for debugging.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glValidateProgram.xhtml
func ValidateProgram(p Program) {
	C.glValidateProgram(p.c())
}

// VertexAttrib1f writes a float vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib1f(dst Attrib, x float32) {
	C.glVertexAttrib1f(dst.c(), C.GLfloat(x))
}

// VertexAttrib1fv writes a float vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib1fv(dst Attrib, src []float32) {
	C.glVertexAttrib1fv(dst.c(), (*C.GLfloat)(&src[0]))
}

// VertexAttrib2f writes a vec2 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib2f(dst Attrib, x, y float32) {
	C.glVertexAttrib2f(dst.c(), C.GLfloat(x), C.GLfloat(y))
}

// VertexAttrib2fv writes a vec2 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib2fv(dst Attrib, src []float32) {
	C.glVertexAttrib2fv(dst.c(), (*C.GLfloat)(&src[0]))
}

// VertexAttrib3f writes a vec3 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib3f(dst Attrib, x, y, z float32) {
	C.glVertexAttrib3f(dst.c(), C.GLfloat(x), C.GLfloat(y), C.GLfloat(z))
}

// VertexAttrib3fv writes a vec3 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib3fv(dst Attrib, src []float32) {
	C.glVertexAttrib3fv(dst.c(), (*C.GLfloat)(&src[0]))
}

// VertexAttrib4f writes a vec4 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib4f(dst Attrib, x, y, z, w float32) {
	C.glVertexAttrib4f(dst.c(), C.GLfloat(x), C.GLfloat(y), C.GLfloat(z), C.GLfloat(w))
}

// VertexAttrib4fv writes a vec4 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib4fv(dst Attrib, src []float32) {
	C.glVertexAttrib4fv(dst.c(), (*C.GLfloat)(&src[0]))
}

// VertexAttribPointer uses a bound bufer to define vertex attribute data.
//
// Direct use of VertexAttribPointer to load data into OpenGL is not
// supported via the Go bindings. Instead, use BindBuffer with an
// ARRAY_BUFFER and then fill it using BufferData.
//
// The size argument specifies the number of components per attribute,
// between 1-4. The stride argument specifies the byte offset between
// consecutive vertex attributes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttribPointer.xhtml
func VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	n := glBoolean(normalized)
	s := C.GLsizei(stride)
	C.glVertexAttribPointer(dst.c(), C.GLint(size), ty.c(), n, s, unsafe.Pointer(uintptr(offset)))
}

// Viewport sets the viewport, an affine transformation that
// normalizes device coordinates to window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glViewport.xhtml
func Viewport(x, y, width, height int) {
	C.glViewport(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}
