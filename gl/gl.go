// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux darwin
// +build !gldebug

package gl

// TODO(crawshaw): build on more host platforms (makes it easier to gobind).
// TODO(crawshaw): expand to cover OpenGL ES 3.
// TODO(crawshaw): should functions on specific types become methods? E.g.
//                 	func (t Texture) Bind(target Enum)
//                 this seems natural in Go, but moves us slightly
//                 further away from the underlying OpenGL spec.

// #include "work.h"
import "C"

import (
	"math"
	"unsafe"
)

// ActiveTexture sets the active texture unit.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glActiveTexture.xhtml
func ActiveTexture(texture Enum) {
	var call call
	call.args.fn = C.glfnActiveTexture
	call.args.a0 = texture.c()
	work <- call
}

// AttachShader attaches a shader to a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glAttachShader.xhtml
func AttachShader(p Program, s Shader) {
	var call call
	call.args.fn = C.glfnAttachShader
	call.args.a0 = p.c()
	call.args.a1 = s.c()
	work <- call
}

// BindAttribLocation binds a vertex attribute index with a named
// variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindAttribLocation.xhtml
func BindAttribLocation(p Program, a Attrib, name string) {
	var call call
	call.args.fn = C.glfnBindAttribLocation
	call.args.a0 = p.c()
	call.args.a1 = a.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
}

// BindBuffer binds a buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindBuffer.xhtml
func BindBuffer(target Enum, b Buffer) {
	var call call
	call.args.fn = C.glfnBindBuffer
	call.args.a0 = target.c()
	call.args.a1 = b.c()
	work <- call
}

// BindFramebuffer binds a framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindFramebuffer.xhtml
func BindFramebuffer(target Enum, fb Framebuffer) {
	var call call
	call.args.fn = C.glfnBindFramebuffer
	call.args.a0 = target.c()
	call.args.a1 = fb.c()
	work <- call
}

// BindRenderbuffer binds a render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindRenderbuffer.xhtml
func BindRenderbuffer(target Enum, rb Renderbuffer) {
	var call call
	call.args.fn = C.glfnBindRenderbuffer
	call.args.a0 = target.c()
	call.args.a1 = rb.c()
	work <- call
}

// BindTexture binds a texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBindTexture.xhtml
func BindTexture(target Enum, t Texture) {
	var call call
	call.args.fn = C.glfnBindTexture
	call.args.a0 = target.c()
	call.args.a1 = t.c()
	work <- call
}

// BlendColor sets the blend color.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendColor.xhtml
func BlendColor(red, green, blue, alpha float32) {
	var call call
	call.args.fn = C.glfnBlendColor
	call.args.a0 = C.uintptr_t(math.Float32bits(red))
	call.args.a1 = C.uintptr_t(math.Float32bits(green))
	call.args.a2 = C.uintptr_t(math.Float32bits(blue))
	call.args.a3 = C.uintptr_t(math.Float32bits(alpha))
	work <- call
}

// BlendEquation sets both RGB and alpha blend equations.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendEquation.xhtml
func BlendEquation(mode Enum) {
	var call call
	call.args.fn = C.glfnBlendEquation
	call.args.a0 = mode.c()
	work <- call
}

// BlendEquationSeparate sets RGB and alpha blend equations separately.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendEquationSeparate.xhtml
func BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	var call call
	call.args.fn = C.glfnBlendEquationSeparate
	call.args.a0 = modeRGB.c()
	call.args.a1 = modeAlpha.c()
	work <- call
}

// BlendFunc sets the pixel blending factors.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendFunc.xhtml
func BlendFunc(sfactor, dfactor Enum) {
	var call call
	call.args.fn = C.glfnBlendFunc
	call.args.a0 = sfactor.c()
	call.args.a1 = dfactor.c()
	work <- call
}

// BlendFunc sets the pixel RGB and alpha blending factors separately.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBlendFuncSeparate.xhtml
func BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	var call call
	call.args.fn = C.glfnBlendFuncSeparate
	call.args.a0 = sfactorRGB.c()
	call.args.a1 = dfactorRGB.c()
	call.args.a2 = sfactorAlpha.c()
	call.args.a3 = dfactorAlpha.c()
	work <- call
}

// BufferData creates a new data store for the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferData.xhtml
func BufferData(target Enum, src []byte, usage Enum) {
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

// BufferInit creates a new uninitialized data store for the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferData.xhtml
func BufferInit(target Enum, size int, usage Enum) {
	var call call
	call.args.fn = C.glfnBufferData
	call.args.a0 = target.c()
	call.args.a1 = C.uintptr_t(size)
	call.args.a2 = 0
	call.args.a3 = usage.c()
	work <- call
}

// BufferSubData sets some of data in the bound buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glBufferSubData.xhtml
func BufferSubData(target Enum, offset int, data []byte) {
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

// CheckFramebufferStatus reports the completeness status of the
// active framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCheckFramebufferStatus.xhtml
func CheckFramebufferStatus(target Enum) Enum {
	var call call
	call.args.fn = C.glfnCheckFramebufferStatus
	call.blocking = true
	call.args.a0 = target.c()
	work <- call
	return Enum(<-retvalue)
}

// Clear clears the window.
//
// The behavior of Clear is influenced by the pixel ownership test,
// the scissor test, dithering, and the buffer writemasks.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClear.xhtml
func Clear(mask Enum) {
	var call call
	call.args.fn = C.glfnClear
	call.args.a0 = C.uintptr_t(mask)
	work <- call
}

// ClearColor specifies the RGBA values used to clear color buffers.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearColor.xhtml
func ClearColor(red, green, blue, alpha float32) {
	var call call
	call.args.fn = C.glfnClearColor
	call.args.a0 = C.uintptr_t(math.Float32bits(red))
	call.args.a1 = C.uintptr_t(math.Float32bits(green))
	call.args.a2 = C.uintptr_t(math.Float32bits(blue))
	call.args.a3 = C.uintptr_t(math.Float32bits(alpha))
	work <- call
}

// ClearDepthf sets the depth value used to clear the depth buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearDepthf.xhtml
func ClearDepthf(d float32) {
	var call call
	call.args.fn = C.glfnClearDepthf
	call.args.a0 = C.uintptr_t(math.Float32bits(d))
	work <- call
}

// ClearStencil sets the index used to clear the stencil buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glClearStencil.xhtml
func ClearStencil(s int) {
	var call call
	call.args.fn = C.glfnClearStencil
	call.args.a0 = C.uintptr_t(s)
	work <- call
}

// ColorMask specifies whether color components in the framebuffer
// can be written.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glColorMask.xhtml
func ColorMask(red, green, blue, alpha bool) {
	var call call
	call.args.fn = C.glfnColorMask
	call.args.a0 = glBoolean(red)
	call.args.a1 = glBoolean(green)
	call.args.a2 = glBoolean(blue)
	call.args.a3 = glBoolean(alpha)
	work <- call
}

// CompileShader compiles the source code of s.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompileShader.xhtml
func CompileShader(s Shader) {
	var call call
	call.args.fn = C.glfnCompileShader
	call.args.a0 = s.c()
	work <- call
}

// CompressedTexImage2D writes a compressed 2D texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompressedTexImage2D.xhtml
func CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data []byte) {
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

// CompressedTexSubImage2D writes a subregion of a compressed 2D texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCompressedTexSubImage2D.xhtml
func CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data []byte) {
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

// CopyTexImage2D writes a 2D texture from the current framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCopyTexImage2D.xhtml
func CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
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

// CopyTexSubImage2D writes a 2D texture subregion from the
// current framebuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCopyTexSubImage2D.xhtml
func CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
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

// CreateBuffer creates a buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenBuffers.xhtml
func CreateBuffer() Buffer {
	var call call
	call.args.fn = C.glfnGenBuffer
	call.blocking = true
	work <- call
	return Buffer{Value: uint32(<-retvalue)}
}

// CreateFramebuffer creates a framebuffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenFramebuffers.xhtml
func CreateFramebuffer() Framebuffer {
	var call call
	call.args.fn = C.glfnGenFramebuffer
	call.blocking = true
	work <- call
	return Framebuffer{Value: uint32(<-retvalue)}
}

// CreateProgram creates a new empty program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCreateProgram.xhtml
func CreateProgram() Program {
	var call call
	call.args.fn = C.glfnCreateProgram
	call.blocking = true
	work <- call
	return Program{Value: uint32(<-retvalue)}
}

// CreateRenderbuffer create a renderbuffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenRenderbuffers.xhtml
func CreateRenderbuffer() Renderbuffer {
	var call call
	call.args.fn = C.glfnGenRenderbuffer
	call.blocking = true
	work <- call
	return Renderbuffer{Value: uint32(<-retvalue)}
}

// CreateShader creates a new empty shader object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCreateShader.xhtml
func CreateShader(ty Enum) Shader {
	var call call
	call.args.fn = C.glfnCreateShader
	call.blocking = true
	call.args.a0 = C.uintptr_t(ty)
	work <- call
	return Shader{Value: uint32(<-retvalue)}

}

// CreateTexture creates a texture object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenTextures.xhtml
func CreateTexture() Texture {
	var call call
	call.args.fn = C.glfnGenTexture
	call.blocking = true
	work <- call
	return Texture{Value: uint32(<-retvalue)}
}

// CullFace specifies which polygons are candidates for culling.
//
// Valid modes: FRONT, BACK, FRONT_AND_BACK.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glCullFace.xhtml
func CullFace(mode Enum) {
	var call call
	call.args.fn = C.glfnCullFace
	call.args.a0 = mode.c()
	work <- call
}

// DeleteBuffer deletes the given buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteBuffers.xhtml
func DeleteBuffer(v Buffer) {
	var call call
	call.args.fn = C.glfnDeleteBuffer
	call.args.a0 = C.uintptr_t(v.Value)
	work <- call
}

// DeleteFramebuffer deletes the given framebuffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteFramebuffers.xhtml
func DeleteFramebuffer(v Framebuffer) {
	var call call
	call.args.fn = C.glfnDeleteFramebuffer
	call.args.a0 = C.uintptr_t(v.Value)
	work <- call
}

// DeleteProgram deletes the given program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteProgram.xhtml
func DeleteProgram(p Program) {
	var call call
	call.args.fn = C.glfnDeleteProgram
	call.args.a0 = p.c()
	work <- call
}

// DeleteRenderbuffer deletes the given render buffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteRenderbuffers.xhtml
func DeleteRenderbuffer(v Renderbuffer) {
	var call call
	call.args.fn = C.glfnDeleteRenderbuffer
	call.args.a0 = v.c()
	work <- call
}

// DeleteShader deletes shader s.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteShader.xhtml
func DeleteShader(s Shader) {
	var call call
	call.args.fn = C.glfnDeleteShader
	call.args.a0 = s.c()
	work <- call
}

// DeleteTexture deletes the given texture object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDeleteTextures.xhtml
func DeleteTexture(v Texture) {
	var call call
	call.args.fn = C.glfnDeleteTexture
	call.args.a0 = v.c()
	work <- call
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
	var call call
	call.args.fn = C.glfnDepthFunc
	call.args.a0 = fn.c()
	work <- call
}

// DepthMask sets the depth buffer enabled for writing.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDepthMask.xhtml
func DepthMask(flag bool) {
	var call call
	call.args.fn = C.glfnDepthMask
	call.args.a0 = glBoolean(flag)
	work <- call
}

// DepthRangef sets the mapping from normalized device coordinates to
// window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDepthRangef.xhtml
func DepthRangef(n, f float32) {
	var call call
	call.args.fn = C.glfnDepthRangef
	call.args.a0 = C.uintptr_t(math.Float32bits(n))
	call.args.a1 = C.uintptr_t(math.Float32bits(f))
	work <- call
}

// DetachShader detaches the shader s from the program p.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDetachShader.xhtml
func DetachShader(p Program, s Shader) {
	var call call
	call.args.fn = C.glfnDetachShader
	call.args.a0 = p.c()
	call.args.a1 = s.c()
	work <- call
}

// Disable disables various GL capabilities.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDisable.xhtml
func Disable(cap Enum) {
	var call call
	call.args.fn = C.glfnDisable
	call.args.a0 = cap.c()
	work <- call
}

// DisableVertexAttribArray disables a vertex attribute array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDisableVertexAttribArray.xhtml
func DisableVertexAttribArray(a Attrib) {
	var call call
	call.args.fn = C.glfnDisableVertexAttribArray
	call.args.a0 = a.c()
	work <- call
}

// DrawArrays renders geometric primitives from the bound data.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDrawArrays.xhtml
func DrawArrays(mode Enum, first, count int) {
	var call call
	call.args.fn = C.glfnDrawArrays
	call.args.a0 = mode.c()
	call.args.a1 = C.uintptr_t(first)
	call.args.a2 = C.uintptr_t(count)
	work <- call
}

// DrawElements renders primitives from a bound buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glDrawElements.xhtml
func DrawElements(mode Enum, count int, ty Enum, offset int) {
	var call call
	call.args.fn = C.glfnDrawElements
	call.args.a0 = mode.c()
	call.args.a1 = C.uintptr_t(count)
	call.args.a2 = ty.c()
	call.args.a3 = C.uintptr_t(offset)
	work <- call
}

// TODO(crawshaw): consider DrawElements8 / DrawElements16 / DrawElements32

// Enable enables various GL capabilities.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glEnable.xhtml
func Enable(cap Enum) {
	var call call
	call.args.fn = C.glfnEnable
	call.args.a0 = cap.c()
	work <- call
}

// EnableVertexAttribArray enables a vertex attribute array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glEnableVertexAttribArray.xhtml
func EnableVertexAttribArray(a Attrib) {
	var call call
	call.args.fn = C.glfnEnableVertexAttribArray
	call.args.a0 = a.c()
	work <- call
}

// Finish blocks until the effects of all previously called GL
// commands are complete.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFinish.xhtml
func Finish() {
	var call call
	call.args.fn = C.glfnFinish
	call.blocking = true
	work <- call
	<-retvalue
}

// Flush empties all buffers. It does not block.
//
// An OpenGL implementation may buffer network communication,
// the command stream, or data inside the graphics accelerator.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFlush.xhtml
func Flush() {
	var call call
	call.args.fn = C.glfnFlush
	call.blocking = true
	work <- call
	<-retvalue
}

// FramebufferRenderbuffer attaches rb to the current frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFramebufferRenderbuffer.xhtml
func FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
	var call call
	call.args.fn = C.glfnFramebufferRenderbuffer
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = rbTarget.c()
	call.args.a3 = rb.c()
	work <- call
}

// FramebufferTexture2D attaches the t to the current frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFramebufferTexture2D.xhtml
func FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	var call call
	call.args.fn = C.glfnFramebufferTexture2D
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = texTarget.c()
	call.args.a3 = t.c()
	call.args.a4 = C.uintptr_t(level)
	work <- call
}

// FrontFace defines which polygons are front-facing.
//
// Valid modes: CW, CCW.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glFrontFace.xhtml
func FrontFace(mode Enum) {
	var call call
	call.args.fn = C.glfnFrontFace
	call.args.a0 = mode.c()
	work <- call
}

// GenerateMipmap generates mipmaps for the current texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGenerateMipmap.xhtml
func GenerateMipmap(target Enum) {
	var call call
	call.args.fn = C.glfnGenerateMipmap
	call.args.a0 = target.c()
	work <- call
}

// GetActiveAttrib returns details about an active attribute variable.
// A value of 0 for index selects the first active attribute variable.
// Permissible values for index range from 0 to the number of active
// attribute variables minus 1.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetActiveAttrib.xhtml
func GetActiveAttrib(p Program, index uint32) (name string, size int, ty Enum) {
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

// GetActiveUniform returns details about an active uniform variable.
// A value of 0 for index selects the first active uniform variable.
// Permissible values for index range from 0 to the number of active
// uniform variables minus 1.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetActiveUniform.xhtml
func GetActiveUniform(p Program, index uint32) (name string, size int, ty Enum) {
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

// GetAttachedShaders returns the shader objects attached to program p.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetAttachedShaders.xhtml
func GetAttachedShaders(p Program) []Shader {
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

// GetAttribLocation returns the location of an attribute variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetAttribLocation.xhtml
func GetAttribLocation(p Program, name string) Attrib {
	var call call
	call.args.fn = C.glfnGetAttribLocation
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
	return Attrib{Value: uint(<-retvalue)}
}

// GetBooleanv returns the boolean values of parameter pname.
//
// Many boolean parameters can be queried more easily using IsEnabled.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetBooleanv(dst []bool, pname Enum) {
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

// GetFloatv returns the float values of parameter pname.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetFloatv(dst []float32, pname Enum) {
	var call call
	call.args.fn = C.glfnGetFloatv
	call.blocking = true
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetIntegerv returns the int values of parameter pname.
//
// Single values may be queried more easily using GetInteger.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetIntegerv(dst []int32, pname Enum) {
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

// GetInteger returns the int value of parameter pname.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGet.xhtml
func GetInteger(pname Enum) int {
	var v [1]int32
	GetIntegerv(v[:], pname)
	return int(v[0])
}

// GetBufferParameteri returns a parameter for the active buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetBufferParameter.xhtml
func GetBufferParameteri(target, value Enum) int {
	var call call
	call.args.fn = C.glfnGetBufferParameteri
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = value.c()
	work <- call
	return int(<-retvalue)
}

// GetError returns the next error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetError.xhtml
func GetError() Enum {
	var call call
	call.args.fn = C.glfnGetError
	call.blocking = true
	work <- call
	return Enum(<-retvalue)
}

// GetFramebufferAttachmentParameteri returns attachment parameters
// for the active framebuffer object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetFramebufferAttachmentParameteriv.xhtml
func GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	var call call
	call.args.fn = C.glfnGetFramebufferAttachmentParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = attachment.c()
	call.args.a2 = pname.c()
	work <- call
	return int(<-retvalue)
}

// GetProgrami returns a parameter value for a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetProgramiv.xhtml
func GetProgrami(p Program, pname Enum) int {
	var call call
	call.args.fn = C.glfnGetProgramiv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

// GetProgramInfoLog returns the information log for a program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetProgramInfoLog.xhtml
func GetProgramInfoLog(p Program) string {
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

// GetRenderbufferParameteri returns a parameter value for a render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetRenderbufferParameteriv.xhtml
func GetRenderbufferParameteri(target, pname Enum) int {
	var call call
	call.args.fn = C.glfnGetRenderbufferParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

// GetShaderi returns a parameter value for a shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderiv.xhtml
func GetShaderi(s Shader, pname Enum) int {
	var call call
	call.args.fn = C.glfnGetShaderiv
	call.blocking = true
	call.args.a0 = s.c()
	call.args.a1 = pname.c()
	work <- call
	return int(<-retvalue)
}

// GetShaderInfoLog returns the information log for a shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderInfoLog.xhtml
func GetShaderInfoLog(s Shader) string {
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

// GetShaderPrecisionFormat returns range and precision limits for
// shader types.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetShaderPrecisionFormat.xhtml
func GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeLow, rangeHigh, precision int) {
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
func GetString(pname Enum) string {
	var call call
	call.args.fn = C.glfnGetString
	call.blocking = true
	call.args.a0 = pname.c()
	work <- call
	return C.GoString((*C.char)((unsafe.Pointer(uintptr(<-retvalue)))))
}

// GetTexParameterfv returns the float values of a texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetTexParameter.xhtml
func GetTexParameterfv(dst []float32, target, pname Enum) {
	var call call
	call.args.fn = C.glfnGetTexParameterfv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetTexParameteriv returns the int values of a texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetTexParameter.xhtml
func GetTexParameteriv(dst []int32, target, pname Enum) {
	var call call
	call.args.fn = C.glfnGetTexParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetUniformfv returns the float values of a uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniform.xhtml
func GetUniformfv(dst []float32, src Uniform, p Program) {
	var call call
	call.args.fn = C.glfnGetUniformfv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = src.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetUniformiv returns the float values of a uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniform.xhtml
func GetUniformiv(dst []int32, src Uniform, p Program) {
	var call call
	call.args.fn = C.glfnGetUniformiv
	call.blocking = true
	call.args.a0 = p.c()
	call.args.a1 = src.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetUniformLocation returns the location of a uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetUniformLocation.xhtml
func GetUniformLocation(p Program, name string) Uniform {
	var call call
	call.blocking = true
	call.args.fn = C.glfnGetUniformLocation
	call.args.a0 = p.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name))))
	work <- call
	return Uniform{Value: int32(<-retvalue)}
}

// GetVertexAttribf reads the float value of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribf(src Attrib, pname Enum) float32 {
	var params [1]float32
	GetVertexAttribfv(params[:], src, pname)
	return params[0]
}

// GetVertexAttribfv reads float values of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
	var call call
	call.args.fn = C.glfnGetVertexAttribfv
	call.blocking = true
	call.args.a0 = src.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// GetVertexAttribi reads the int value of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribi(src Attrib, pname Enum) int32 {
	var params [1]int32
	GetVertexAttribiv(params[:], src, pname)
	return params[0]
}

// GetVertexAttribiv reads int values of a vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glGetVertexAttrib.xhtml
func GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
	var call call
	call.args.fn = C.glfnGetVertexAttribiv
	call.blocking = true
	call.args.a0 = src.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&dst[0])))
	work <- call
	<-retvalue
}

// TODO(crawshaw): glGetVertexAttribPointerv

// Hint sets implementation-specific modes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glHint.xhtml
func Hint(target, mode Enum) {
	var call call
	call.args.fn = C.glfnHint
	call.args.a0 = target.c()
	call.args.a1 = mode.c()
	work <- call
}

// IsBuffer reports if b is a valid buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsBuffer.xhtml
func IsBuffer(b Buffer) bool {
	var call call
	call.args.fn = C.glfnIsBuffer
	call.blocking = true
	call.args.a0 = b.c()
	work <- call
	return <-retvalue != 0
}

// IsEnabled reports if cap is an enabled capability.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsEnabled.xhtml
func IsEnabled(cap Enum) bool {
	var call call
	call.args.fn = C.glfnIsEnabled
	call.blocking = true
	call.args.a0 = cap.c()
	work <- call
	return <-retvalue != 0
}

// IsFramebuffer reports if fb is a valid frame buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsFramebuffer.xhtml
func IsFramebuffer(fb Framebuffer) bool {
	var call call
	call.args.fn = C.glfnIsFramebuffer
	call.blocking = true
	call.args.a0 = fb.c()
	work <- call
	return <-retvalue != 0
}

// IsProgram reports if p is a valid program object.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsProgram.xhtml
func IsProgram(p Program) bool {
	var call call
	call.args.fn = C.glfnIsProgram
	call.blocking = true
	call.args.a0 = p.c()
	work <- call
	return <-retvalue != 0
}

// IsRenderbuffer reports if rb is a valid render buffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsRenderbuffer.xhtml
func IsRenderbuffer(rb Renderbuffer) bool {
	var call call
	call.args.fn = C.glfnIsRenderbuffer
	call.blocking = true
	call.args.a0 = rb.c()
	work <- call
	return <-retvalue != 0
}

// IsShader reports if s is valid shader.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsShader.xhtml
func IsShader(s Shader) bool {
	var call call
	call.args.fn = C.glfnIsShader
	call.blocking = true
	call.args.a0 = s.c()
	work <- call
	return <-retvalue != 0
}

// IsTexture reports if t is a valid texture.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glIsTexture.xhtml
func IsTexture(t Texture) bool {
	var call call
	call.args.fn = C.glfnIsTexture
	call.blocking = true
	call.args.a0 = t.c()
	work <- call
	return <-retvalue != 0
}

// LineWidth specifies the width of lines.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glLineWidth.xhtml
func LineWidth(width float32) {
	var call call
	call.args.fn = C.glfnLineWidth
	call.args.a0 = C.uintptr_t(math.Float32bits(width))
	work <- call
}

// LinkProgram links the specified program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glLinkProgram.xhtml
func LinkProgram(p Program) {
	var call call
	call.args.fn = C.glfnLinkProgram
	call.args.a0 = p.c()
	work <- call
}

// PixelStorei sets pixel storage parameters.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glPixelStorei.xhtml
func PixelStorei(pname Enum, param int32) {
	var call call
	call.args.fn = C.glfnPixelStorei
	call.args.a0 = pname.c()
	call.args.a1 = C.uintptr_t(param)
	work <- call
}

// PolygonOffset sets the scaling factors for depth offsets.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glPolygonOffset.xhtml
func PolygonOffset(factor, units float32) {
	var call call
	call.args.fn = C.glfnPolygonOffset
	call.args.a0 = C.uintptr_t(math.Float32bits(factor))
	call.args.a1 = C.uintptr_t(math.Float32bits(units))
	work <- call
}

// ReadPixels returns pixel data from a buffer.
//
// In GLES 3, the source buffer is controlled with ReadBuffer.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glReadPixels.xhtml
func ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
	var call call
	call.args.fn = C.glfnReadPixels
	call.blocking = true
	// TODO(crawshaw): support PIXEL_PACK_BUFFER in GLES3, uses offset.
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

// ReleaseShaderCompiler frees resources allocated by the shader compiler.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glReleaseShaderCompiler.xhtml
func ReleaseShaderCompiler() {
	var call call
	call.args.fn = C.glfnReleaseShaderCompiler
	work <- call
}

// RenderbufferStorage establishes the data storage, format, and
// dimensions of a renderbuffer object's image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glRenderbufferStorage.xhtml
func RenderbufferStorage(target, internalFormat Enum, width, height int) {
	var call call
	call.args.fn = C.glfnRenderbufferStorage
	call.args.a0 = target.c()
	call.args.a1 = internalFormat.c()
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}

// SampleCoverage sets multisample coverage parameters.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glSampleCoverage.xhtml
func SampleCoverage(value float32, invert bool) {
	var call call
	call.args.fn = C.glfnSampleCoverage
	call.args.a0 = C.uintptr_t(math.Float32bits(value))
	call.args.a1 = glBoolean(invert)
	work <- call
}

// Scissor defines the scissor box rectangle, in window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glScissor.xhtml
func Scissor(x, y, width, height int32) {
	var call call
	call.args.fn = C.glfnScissor
	call.args.a0 = C.uintptr_t(x)
	call.args.a1 = C.uintptr_t(y)
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}

// TODO(crawshaw): ShaderBinary

// ShaderSource sets the source code of s to the given source code.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glShaderSource.xhtml
func ShaderSource(s Shader, src string) {
	var call call
	call.args.fn = C.glfnShaderSource
	call.args.a0 = s.c()
	call.args.a1 = 1

	// We are passing a char**. Make sure both the string and its
	// containing 1-element array are off the stack. Both are freed
	// in work.c.
	cstr := C.CString(src)
	cstrp := (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(cstr))))
	*cstrp = cstr
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(cstrp)))
	work <- call
}

// StencilFunc sets the front and back stencil test reference value.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilFunc.xhtml
func StencilFunc(fn Enum, ref int, mask uint32) {
	var call call
	call.args.fn = C.glfnStencilFunc
	call.args.a0 = fn.c()
	call.args.a1 = C.uintptr_t(ref)
	call.args.a2 = C.uintptr_t(mask)
	work <- call
}

// StencilFunc sets the front or back stencil test reference value.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilFuncSeparate.xhtml
func StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
	var call call
	call.args.fn = C.glfnStencilFuncSeparate
	call.args.a0 = face.c()
	call.args.a1 = fn.c()
	call.args.a2 = C.uintptr_t(ref)
	call.args.a3 = C.uintptr_t(mask)
	work <- call
}

// StencilMask controls the writing of bits in the stencil planes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilMask.xhtml
func StencilMask(mask uint32) {
	var call call
	call.args.fn = C.glfnStencilMask
	call.args.a0 = C.uintptr_t(mask)
	work <- call
}

// StencilMaskSeparate controls the writing of bits in the stencil planes.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilMaskSeparate.xhtml
func StencilMaskSeparate(face Enum, mask uint32) {
	var call call
	call.args.fn = C.glfnStencilMaskSeparate
	call.args.a0 = face.c()
	call.args.a1 = C.uintptr_t(mask)
	work <- call
}

// StencilOp sets front and back stencil test actions.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilOp.xhtml
func StencilOp(fail, zfail, zpass Enum) {
	var call call
	call.args.fn = C.glfnStencilOp
	call.args.a0 = fail.c()
	call.args.a1 = zfail.c()
	call.args.a2 = zpass.c()
	work <- call
}

// StencilOpSeparate sets front or back stencil tests.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glStencilOpSeparate.xhtml
func StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
	var call call
	call.args.fn = C.glfnStencilOpSeparate
	call.args.a0 = face.c()
	call.args.a1 = sfail.c()
	call.args.a2 = dpfail.c()
	call.args.a3 = dppass.c()
	work <- call
}

// TexImage2D writes a 2D texture image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexImage2D.xhtml
func TexImage2D(target Enum, level int, width, height int, format Enum, ty Enum, data []byte) {
	// It is common to pass TexImage2D a nil data, indicating that a
	// bound GL buffer is being used as the source. In that case, it
	// is not necessary to block.
	var call call
	call.args.fn = C.glfnTexImage2D
	// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
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

// TexSubImage2D writes a subregion of a 2D texture image.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexSubImage2D.xhtml
func TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	var call call
	call.args.fn = C.glfnTexSubImage2D
	call.blocking = true
	// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
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

// TexParameterf sets a float texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameterf(target, pname Enum, param float32) {
	var call call
	call.args.fn = C.glfnTexParameterf
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(math.Float32bits(param))
	work <- call
}

// TexParameterfv sets a float texture parameter array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameterfv(target, pname Enum, params []float32) {
	var call call
	call.args.fn = C.glfnTexParameterfv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&params[0])))
	work <- call
	<-retvalue
}

// TexParameteri sets an integer texture parameter.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameteri(target, pname Enum, param int) {
	var call call
	call.args.fn = C.glfnTexParameteri
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(param)
	work <- call
}

// TexParameteriv sets an integer texture parameter array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glTexParameter.xhtml
func TexParameteriv(target, pname Enum, params []int32) {
	var call call
	call.args.fn = C.glfnTexParameteriv
	call.blocking = true
	call.args.a0 = target.c()
	call.args.a1 = pname.c()
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&params[0])))
	work <- call
	<-retvalue
}

// Uniform1f writes a float uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1f(dst Uniform, v float32) {
	var call call
	call.args.fn = C.glfnUniform1f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v))
	work <- call
}

// Uniform1fv writes a [len(src)]float uniform array.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniform1fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src))
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform1i writes an int uniform variable.
//
// Uniform1i and Uniform1iv are the only two functions that may be used
// to load uniform variables defined as sampler types. Loading samplers
// with any other function will result in a INVALID_OPERATION error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1i(dst Uniform, v int) {
	var call call
	call.args.fn = C.glfnUniform1i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v)
	work <- call
}

// Uniform1iv writes a int uniform array of len(src) elements.
//
// Uniform1i and Uniform1iv are the only two functions that may be used
// to load uniform variables defined as sampler types. Loading samplers
// with any other function will result in a INVALID_OPERATION error.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform1iv(dst Uniform, src []int32) {
	var call call
	call.args.fn = C.glfnUniform1iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src))
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform2f writes a vec2 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2f(dst Uniform, v0, v1 float32) {
	var call call
	call.args.fn = C.glfnUniform2f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	work <- call
}

// Uniform2fv writes a vec2 uniform array of len(src)/2 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniform2fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 2)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform2i writes an ivec2 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2i(dst Uniform, v0, v1 int) {
	var call call
	call.args.fn = C.glfnUniform2i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	work <- call
}

// Uniform2iv writes an ivec2 uniform array of len(src)/2 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform2iv(dst Uniform, src []int32) {
	var call call
	call.args.fn = C.glfnUniform2iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 2)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform3f writes a vec3 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3f(dst Uniform, v0, v1, v2 float32) {
	var call call
	call.args.fn = C.glfnUniform3f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	call.args.a3 = C.uintptr_t(math.Float32bits(v2))
	work <- call
}

// Uniform3fv writes a vec3 uniform array of len(src)/3 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniform3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 3)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform3i writes an ivec3 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3i(dst Uniform, v0, v1, v2 int32) {
	var call call
	call.args.fn = C.glfnUniform3i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	call.args.a3 = C.uintptr_t(v2)
	work <- call
}

// Uniform3iv writes an ivec3 uniform array of len(src)/3 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform3iv(dst Uniform, src []int32) {
	var call call
	call.args.fn = C.glfnUniform3iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 3)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform4f writes a vec4 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	var call call
	call.args.fn = C.glfnUniform4f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(v0))
	call.args.a2 = C.uintptr_t(math.Float32bits(v1))
	call.args.a3 = C.uintptr_t(math.Float32bits(v2))
	call.args.a4 = C.uintptr_t(math.Float32bits(v3))
	work <- call
}

// Uniform4fv writes a vec4 uniform array of len(src)/4 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniform4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// Uniform4i writes an ivec4 uniform variable.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
	var call call
	call.args.fn = C.glfnUniform4i
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(v0)
	call.args.a2 = C.uintptr_t(v1)
	call.args.a3 = C.uintptr_t(v2)
	call.args.a4 = C.uintptr_t(v3)
	work <- call
}

// Uniform4i writes an ivec4 uniform array of len(src)/4 elements.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func Uniform4iv(dst Uniform, src []int32) {
	var call call
	call.args.fn = C.glfnUniform4iv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// UniformMatrix2fv writes 2x2 matrices. Each matrix uses four
// float32 values, so the number of matrices written is len(src)/4.
//
// Each matrix must be supplied in column major order.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix2fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniformMatrix2fv
	call.blocking = true
	// OpenGL ES 2 does not support transpose.
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 4)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// UniformMatrix3fv writes 3x3 matrices. Each matrix uses nine
// float32 values, so the number of matrices written is len(src)/9.
//
// Each matrix must be supplied in column major order.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix3fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniformMatrix3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 9)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// UniformMatrix4fv writes 4x4 matrices. Each matrix uses 16
// float32 values, so the number of matrices written is len(src)/16.
//
// Each matrix must be supplied in column major order.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUniform.xhtml
func UniformMatrix4fv(dst Uniform, src []float32) {
	var call call
	call.args.fn = C.glfnUniformMatrix4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(len(src) / 16)
	call.args.a2 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// UseProgram sets the active program.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glUseProgram.xhtml
func UseProgram(p Program) {
	var call call
	call.args.fn = C.glfnUseProgram
	call.args.a0 = p.c()
	work <- call
}

// ValidateProgram checks to see whether the executables contained in
// program can execute given the current OpenGL state.
//
// Typically only used for debugging.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glValidateProgram.xhtml
func ValidateProgram(p Program) {
	var call call
	call.args.fn = C.glfnValidateProgram
	call.args.a0 = p.c()
	work <- call
}

// VertexAttrib1f writes a float vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib1f(dst Attrib, x float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib1f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	work <- call
}

// VertexAttrib1fv writes a float vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib1fv(dst Attrib, src []float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib1fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// VertexAttrib2f writes a vec2 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib2f(dst Attrib, x, y float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib2f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	work <- call
}

// VertexAttrib2fv writes a vec2 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib2fv(dst Attrib, src []float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib2fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// VertexAttrib3f writes a vec3 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib3f(dst Attrib, x, y, z float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib3f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	call.args.a3 = C.uintptr_t(math.Float32bits(z))
	work <- call
}

// VertexAttrib3fv writes a vec3 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib3fv(dst Attrib, src []float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib3fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// VertexAttrib4f writes a vec4 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib4f(dst Attrib, x, y, z, w float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib4f
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(math.Float32bits(x))
	call.args.a2 = C.uintptr_t(math.Float32bits(y))
	call.args.a3 = C.uintptr_t(math.Float32bits(z))
	call.args.a4 = C.uintptr_t(math.Float32bits(w))
	work <- call
}

// VertexAttrib4fv writes a vec4 vertex attribute.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glVertexAttrib.xhtml
func VertexAttrib4fv(dst Attrib, src []float32) {
	var call call
	call.args.fn = C.glfnVertexAttrib4fv
	call.blocking = true
	call.args.a0 = dst.c()
	call.args.a1 = C.uintptr_t(uintptr(unsafe.Pointer(&src[0])))
	work <- call
	<-retvalue
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
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

// Viewport sets the viewport, an affine transformation that
// normalizes device coordinates to window coordinates.
//
// http://www.khronos.org/opengles/sdk/docs/man3/html/glViewport.xhtml
func Viewport(x, y, width, height int) {
	var call call
	call.args.fn = C.glfnViewport
	call.args.a0 = C.uintptr_t(x)
	call.args.a1 = C.uintptr_t(y)
	call.args.a2 = C.uintptr_t(width)
	call.args.a3 = C.uintptr_t(height)
	work <- call
}
