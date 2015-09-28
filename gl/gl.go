// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux darwin
// +build !gldebug

package gl

// TODO(crawshaw): build on more host platforms (makes it easier to gobind).
// TODO(crawshaw): expand to cover OpenGL ES 3.
// TODO(crawshaw): should functions on specific types become methods? E.g.
//                 func (t Texture) Bind(target Enum)
//                 this seems natural in Go, but moves us slightly
//                 further away from the underlying OpenGL spec.

// #include "work.h"
import "C"

import (
	"math"
	"unsafe"
)

func (ctx *context) ActiveTexture(texture Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnActiveTexture,
			a0: texture.c(),
		},
	})
}

func (ctx *context) AttachShader(p Program, s Shader) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnAttachShader,
			a0: p.c(),
			a1: s.c(),
		},
	})
}

func (ctx *context) BindAttribLocation(p Program, a Attrib, name string) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBindAttribLocation,
			a0: p.c(),
			a1: a.c(),
			a2: C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name)))),
		},
	})
}

func (ctx *context) BindBuffer(target Enum, b Buffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBindBuffer,
			a0: target.c(),
			a1: b.c(),
		},
	})
}

func (ctx *context) BindFramebuffer(target Enum, fb Framebuffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBindFramebuffer,
			a0: target.c(),
			a1: fb.c(),
		},
	})
}

func (ctx *context) BindRenderbuffer(target Enum, rb Renderbuffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBindRenderbuffer,
			a0: target.c(),
			a1: rb.c(),
		},
	})
}

func (ctx *context) BindTexture(target Enum, t Texture) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBindTexture,
			a0: target.c(),
			a1: t.c(),
		},
	})
}

func (ctx *context) BlendColor(red, green, blue, alpha float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBlendColor,
			a0: C.uintptr_t(math.Float32bits(red)),
			a1: C.uintptr_t(math.Float32bits(green)),
			a2: C.uintptr_t(math.Float32bits(blue)),
			a3: C.uintptr_t(math.Float32bits(alpha)),
		},
	})
}

func (ctx *context) BlendEquation(mode Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBlendEquation,
			a0: mode.c(),
		},
	})
}

func (ctx *context) BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBlendEquationSeparate,
			a0: modeRGB.c(),
			a1: modeAlpha.c(),
		},
	})
}

func (ctx *context) BlendFunc(sfactor, dfactor Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBlendFunc,
			a0: sfactor.c(),
			a1: dfactor.c(),
		},
	})
}

func (ctx *context) BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBlendFuncSeparate,
			a0: sfactorRGB.c(),
			a1: dfactorRGB.c(),
			a2: sfactorAlpha.c(),
			a3: dfactorAlpha.c(),
		},
	})
}

func (ctx *context) BufferData(target Enum, src []byte, usage Enum) {
	parg := unsafe.Pointer(nil)
	if len(src) > 0 {
		parg = unsafe.Pointer(&src[0])
	}
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBufferData,
			a0: target.c(),
			a1: C.uintptr_t(len(src)),
			a2: usage.c(),
		},
		parg:     parg,
		blocking: true,
	})
}

func (ctx *context) BufferInit(target Enum, size int, usage Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBufferData,
			a0: target.c(),
			a1: C.uintptr_t(size),
			a2: 0,
			a3: usage.c(),
		},
	})
}

func (ctx *context) BufferSubData(target Enum, offset int, data []byte) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnBufferSubData,
			a0: target.c(),
			a1: C.uintptr_t(offset),
			a2: C.uintptr_t(len(data)),
		},
		parg:     unsafe.Pointer(&data[0]),
		blocking: true,
	})
}

func (ctx *context) CheckFramebufferStatus(target Enum) Enum {
	return Enum(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCheckFramebufferStatus,
			a0: target.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) Clear(mask Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnClear,
			a0: C.uintptr_t(mask),
		},
	})
}

func (ctx *context) ClearColor(red, green, blue, alpha float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnClearColor,
			a0: C.uintptr_t(math.Float32bits(red)),
			a1: C.uintptr_t(math.Float32bits(green)),
			a2: C.uintptr_t(math.Float32bits(blue)),
			a3: C.uintptr_t(math.Float32bits(alpha)),
		},
	})
}

func (ctx *context) ClearDepthf(d float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnClearDepthf,
			a0: C.uintptr_t(math.Float32bits(d)),
		},
	})
}

func (ctx *context) ClearStencil(s int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnClearStencil,
			a0: C.uintptr_t(s),
		},
	})
}

func (ctx *context) ColorMask(red, green, blue, alpha bool) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnColorMask,
			a0: glBoolean(red),
			a1: glBoolean(green),
			a2: glBoolean(blue),
			a3: glBoolean(alpha),
		},
	})
}

func (ctx *context) CompileShader(s Shader) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCompileShader,
			a0: s.c(),
		},
	})
}

func (ctx *context) CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data []byte) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCompressedTexImage2D,
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: internalformat.c(),
			a3: C.uintptr_t(width),
			a4: C.uintptr_t(height),
			a5: C.uintptr_t(border),
			a6: C.uintptr_t(len(data)),
		},
		parg:     unsafe.Pointer(&data[0]),
		blocking: true,
	})
}

func (ctx *context) CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data []byte) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCompressedTexSubImage2D,
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: C.uintptr_t(xoffset),
			a3: C.uintptr_t(yoffset),
			a4: C.uintptr_t(width),
			a5: C.uintptr_t(height),
			a6: format.c(),
			a7: C.uintptr_t(len(data)),
		},
		parg:     unsafe.Pointer(&data[0]),
		blocking: true,
	})
}

func (ctx *context) CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCopyTexImage2D,
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: internalformat.c(),
			a3: C.uintptr_t(x),
			a4: C.uintptr_t(y),
			a5: C.uintptr_t(width),
			a6: C.uintptr_t(height),
			a7: C.uintptr_t(border),
		},
	})
}

func (ctx *context) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCopyTexSubImage2D,
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: C.uintptr_t(xoffset),
			a3: C.uintptr_t(yoffset),
			a4: C.uintptr_t(x),
			a5: C.uintptr_t(y),
			a6: C.uintptr_t(width),
			a7: C.uintptr_t(height),
		},
	})
}

func (ctx *context) CreateBuffer() Buffer {
	return Buffer{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGenBuffer,
		},
		blocking: true,
	}))}
}

func (ctx *context) CreateFramebuffer() Framebuffer {
	return Framebuffer{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGenFramebuffer,
		},
		blocking: true,
	}))}
}

func (ctx *context) CreateProgram() Program {
	return Program{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCreateProgram,
		},
		blocking: true,
	}))}
}

func (ctx *context) CreateRenderbuffer() Renderbuffer {
	return Renderbuffer{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGenRenderbuffer,
		},
		blocking: true,
	}))}
}

func (ctx *context) CreateShader(ty Enum) Shader {
	return Shader{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCreateShader,
			a0: C.uintptr_t(ty),
		},
		blocking: true,
	}))}
}

func (ctx *context) CreateTexture() Texture {
	return Texture{Value: uint32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGenTexture,
		},
		blocking: true,
	}))}
}

func (ctx *context) CullFace(mode Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnCullFace,
			a0: mode.c(),
		},
	})
}

func (ctx *context) DeleteBuffer(v Buffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteBuffer,
			a0: C.uintptr_t(v.Value),
		},
	})
}

func (ctx *context) DeleteFramebuffer(v Framebuffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteFramebuffer,
			a0: C.uintptr_t(v.Value),
		},
	})
}

func (ctx *context) DeleteProgram(p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteProgram,
			a0: p.c(),
		},
	})
}

func (ctx *context) DeleteRenderbuffer(v Renderbuffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteRenderbuffer,
			a0: v.c(),
		},
	})
}

func (ctx *context) DeleteShader(s Shader) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteShader,
			a0: s.c(),
		},
	})
}

func (ctx *context) DeleteTexture(v Texture) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDeleteTexture,
			a0: v.c(),
		},
	})
}

func (ctx *context) DepthFunc(fn Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDepthFunc,
			a0: fn.c(),
		},
	})
}

func (ctx *context) DepthMask(flag bool) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDepthMask,
			a0: glBoolean(flag),
		},
	})
}

func (ctx *context) DepthRangef(n, f float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDepthRangef,
			a0: C.uintptr_t(math.Float32bits(n)),
			a1: C.uintptr_t(math.Float32bits(f)),
		},
	})
}

func (ctx *context) DetachShader(p Program, s Shader) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDetachShader,
			a0: p.c(),
			a1: s.c(),
		},
	})
}

func (ctx *context) Disable(cap Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDisable,
			a0: cap.c(),
		},
	})
}

func (ctx *context) DisableVertexAttribArray(a Attrib) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDisableVertexAttribArray,
			a0: a.c(),
		},
	})
}

func (ctx *context) DrawArrays(mode Enum, first, count int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDrawArrays,
			a0: mode.c(),
			a1: C.uintptr_t(first),
			a2: C.uintptr_t(count),
		},
	})
}

func (ctx *context) DrawElements(mode Enum, count int, ty Enum, offset int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnDrawElements,
			a0: mode.c(),
			a1: C.uintptr_t(count),
			a2: ty.c(),
			a3: C.uintptr_t(offset),
		},
	})
}

func (ctx *context) Enable(cap Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnEnable,
			a0: cap.c(),
		},
	})
}

func (ctx *context) EnableVertexAttribArray(a Attrib) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnEnableVertexAttribArray,
			a0: a.c(),
		},
	})
}

func (ctx *context) Finish() {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnFinish,
		},
		blocking: true,
	})
}

func (ctx *context) Flush() {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnFlush,
		},
		blocking: true,
	})
}

func (ctx *context) FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnFramebufferRenderbuffer,
			a0: target.c(),
			a1: attachment.c(),
			a2: rbTarget.c(),
			a3: rb.c(),
		},
	})
}

func (ctx *context) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnFramebufferTexture2D,
			a0: target.c(),
			a1: attachment.c(),
			a2: texTarget.c(),
			a3: t.c(),
			a4: C.uintptr_t(level),
		},
	})
}

func (ctx *context) FrontFace(mode Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnFrontFace,
			a0: mode.c(),
		},
	})
}

func (ctx *context) GenerateMipmap(target Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGenerateMipmap,
			a0: target.c(),
		},
	})
}

func (ctx *context) GetActiveAttrib(p Program, index uint32) (name string, size int, ty Enum) {
	bufSize := ctx.GetProgrami(p, ACTIVE_ATTRIBUTE_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)
	var cSize C.GLint
	var cType C.GLenum

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetActiveAttrib,
			a0: p.c(),
			a1: C.uintptr_t(index),
			a2: C.uintptr_t(bufSize),
			a3: 0,
			a4: C.uintptr_t(uintptr(unsafe.Pointer(&cSize))),
			a5: C.uintptr_t(uintptr(unsafe.Pointer(&cType))),
			a6: C.uintptr_t(uintptr(unsafe.Pointer(buf))),
		},
		blocking: true,
	})

	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

func (ctx *context) GetActiveUniform(p Program, index uint32) (name string, size int, ty Enum) {
	bufSize := ctx.GetProgrami(p, ACTIVE_UNIFORM_MAX_LENGTH)
	buf := C.malloc(C.size_t(bufSize))
	defer C.free(buf)
	var cSize C.GLint
	var cType C.GLenum

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetActiveUniform,
			a0: p.c(),
			a1: C.uintptr_t(index),
			a2: C.uintptr_t(bufSize),
			a3: 0,
			a4: C.uintptr_t(uintptr(unsafe.Pointer(&cSize))),
			a5: C.uintptr_t(uintptr(unsafe.Pointer(&cType))),
			a6: C.uintptr_t(uintptr(unsafe.Pointer(buf))),
		},
		blocking: true,
	})

	return C.GoString((*C.char)(buf)), int(cSize), Enum(cType)
}

func (ctx *context) GetAttachedShaders(p Program) []Shader {
	shadersLen := ctx.GetProgrami(p, ATTACHED_SHADERS)
	if shadersLen == 0 {
		return nil
	}
	var n C.GLsizei
	buf := make([]C.GLuint, shadersLen)

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetAttachedShaders,
			a0: p.c(),
			a1: C.uintptr_t(shadersLen),
			a2: C.uintptr_t(uintptr(unsafe.Pointer(&n))),
			a3: C.uintptr_t(uintptr(unsafe.Pointer(&buf[0]))),
		},
		blocking: true,
	})

	buf = buf[:int(n)]
	shaders := make([]Shader, len(buf))
	for i, s := range buf {
		shaders[i] = Shader{Value: uint32(s)}
	}
	return shaders
}

func (ctx *context) GetAttribLocation(p Program, name string) Attrib {
	return Attrib{Value: uint(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetAttribLocation,
			a0: p.c(),
			a1: C.uintptr_t(uintptr(unsafe.Pointer(C.CString(name)))),
		},
		blocking: true,
	}))}
}

func (ctx *context) GetBooleanv(dst []bool, pname Enum) {
	buf := make([]C.GLboolean, len(dst))

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetBooleanv,
			a0: pname.c(),
			a1: C.uintptr_t(uintptr(unsafe.Pointer(&buf[0]))),
		},
		blocking: true,
	})

	for i, v := range buf {
		dst[i] = v != 0
	}
}

func (ctx *context) GetFloatv(dst []float32, pname Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetFloatv,
			a0: pname.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) GetIntegerv(dst []int32, pname Enum) {
	buf := make([]C.GLint, len(dst))

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetIntegerv,
			a0: pname.c(),
		},
		parg:     unsafe.Pointer(&buf[0]),
		blocking: true,
	})

	for i, v := range buf {
		dst[i] = int32(v)
	}
}

func (ctx *context) GetInteger(pname Enum) int {
	var v [1]int32
	ctx.GetIntegerv(v[:], pname)
	return int(v[0])
}

func (ctx *context) GetBufferParameteri(target, value Enum) int {
	return int(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetBufferParameteri,
			a0: target.c(),
			a1: value.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) GetError() Enum {
	return Enum(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetError,
		},
		blocking: true,
	}))
}

func (ctx *context) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	return int(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetFramebufferAttachmentParameteriv,
			a0: target.c(),
			a1: attachment.c(),
			a2: pname.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) GetProgrami(p Program, pname Enum) int {
	return int(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetProgramiv,
			a0: p.c(),
			a1: pname.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) GetProgramInfoLog(p Program) string {
	infoLen := ctx.GetProgrami(p, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	defer C.free(buf)

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetProgramInfoLog,
			a0: p.c(),
			a1: C.uintptr_t(infoLen),
			a2: 0,
			a3: C.uintptr_t(uintptr(buf)),
		},
		blocking: true,
	})

	return C.GoString((*C.char)(buf))
}

func (ctx *context) GetRenderbufferParameteri(target, pname Enum) int {
	return int(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetRenderbufferParameteriv,
			a0: target.c(),
			a1: pname.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) GetShaderi(s Shader, pname Enum) int {
	return int(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetShaderiv,
			a0: s.c(),
			a1: pname.c(),
		},
		blocking: true,
	}))
}

func (ctx *context) GetShaderInfoLog(s Shader) string {
	infoLen := ctx.GetShaderi(s, INFO_LOG_LENGTH)
	buf := C.malloc(C.size_t(infoLen))
	defer C.free(buf)

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetShaderInfoLog,
			a0: s.c(),
			a1: C.uintptr_t(infoLen),
			a2: 0,
			a3: C.uintptr_t(uintptr(buf)),
		},
		blocking: true,
	})

	return C.GoString((*C.char)(buf))
}

func (ctx *context) GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeLow, rangeHigh, precision int) {
	var cRange [2]C.GLint
	var cPrecision C.GLint

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetShaderPrecisionFormat,
			a0: shadertype.c(),
			a1: precisiontype.c(),
			a2: C.uintptr_t(uintptr(unsafe.Pointer(&cRange[0]))),
			a3: C.uintptr_t(uintptr(unsafe.Pointer(&cPrecision))),
		},
		blocking: true,
	})

	return int(cRange[0]), int(cRange[1]), int(cPrecision)
}

func (ctx *context) GetShaderSource(s Shader) string {
	sourceLen := ctx.GetShaderi(s, SHADER_SOURCE_LENGTH)
	if sourceLen == 0 {
		return ""
	}
	buf := C.malloc(C.size_t(sourceLen))
	defer C.free(buf)

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetShaderSource,
			a0: s.c(),
			a1: C.uintptr_t(sourceLen),
			a2: 0,
			a3: C.uintptr_t(uintptr(buf)),
		},
		blocking: true,
	})

	return C.GoString((*C.char)(buf))
}

func (ctx *context) GetString(pname Enum) string {
	ret := ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetString,
			a0: pname.c(),
		},
		blocking: true,
	})
	return C.GoString((*C.char)((unsafe.Pointer(uintptr(ret)))))
}

func (ctx *context) GetTexParameterfv(dst []float32, target, pname Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetTexParameterfv,
			a0: target.c(),
			a1: pname.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) GetTexParameteriv(dst []int32, target, pname Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetTexParameteriv,
			a0: target.c(),
			a1: pname.c(),
		},
		blocking: true,
	})
}

func (ctx *context) GetUniformfv(dst []float32, src Uniform, p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetUniformfv,
			a0: p.c(),
			a1: src.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) GetUniformiv(dst []int32, src Uniform, p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetUniformiv,
			a0: p.c(),
			a1: src.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) GetUniformLocation(p Program, name string) Uniform {
	return Uniform{Value: int32(ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetUniformLocation,
			a0: p.c(),
		},
		parg:     unsafe.Pointer(C.CString(name)),
		blocking: true,
	}))}
}

func (ctx *context) GetVertexAttribf(src Attrib, pname Enum) float32 {
	var params [1]float32
	ctx.GetVertexAttribfv(params[:], src, pname)
	return params[0]
}

func (ctx *context) GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetVertexAttribfv,
			a0: src.c(),
			a1: pname.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) GetVertexAttribi(src Attrib, pname Enum) int32 {
	var params [1]int32
	ctx.GetVertexAttribiv(params[:], src, pname)
	return params[0]
}

func (ctx *context) GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnGetVertexAttribiv,
			a0: src.c(),
			a1: pname.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) Hint(target, mode Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnHint,
			a0: target.c(),
			a1: mode.c(),
		},
	})
}

func (ctx *context) IsBuffer(b Buffer) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsBuffer,
			a0: b.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsEnabled(cap Enum) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsEnabled,
			a0: cap.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsFramebuffer(fb Framebuffer) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsFramebuffer,
			a0: fb.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsProgram(p Program) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsProgram,
			a0: p.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsRenderbuffer(rb Renderbuffer) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsRenderbuffer,
			a0: rb.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsShader(s Shader) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsShader,
			a0: s.c(),
		},
		blocking: true,
	})
}

func (ctx *context) IsTexture(t Texture) bool {
	return 0 != ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnIsTexture,
			a0: t.c(),
		},
		blocking: true,
	})
}

func (ctx *context) LineWidth(width float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnLineWidth,
			a0: C.uintptr_t(math.Float32bits(width)),
		},
	})
}

func (ctx *context) LinkProgram(p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnLinkProgram,
			a0: p.c(),
		},
	})
}

func (ctx *context) PixelStorei(pname Enum, param int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnPixelStorei,
			a0: pname.c(),
			a1: C.uintptr_t(param),
		},
	})
}

func (ctx *context) PolygonOffset(factor, units float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnPolygonOffset,
			a0: C.uintptr_t(math.Float32bits(factor)),
			a1: C.uintptr_t(math.Float32bits(units)),
		},
	})
}

func (ctx *context) ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnReadPixels,
			// TODO(crawshaw): support PIXEL_PACK_BUFFER in GLES3, uses offset.
			a0: C.uintptr_t(x),
			a1: C.uintptr_t(y),
			a2: C.uintptr_t(width),
			a3: C.uintptr_t(height),
			a4: format.c(),
			a5: ty.c(),
		},
		parg:     unsafe.Pointer(&dst[0]),
		blocking: true,
	})
}

func (ctx *context) ReleaseShaderCompiler() {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnReleaseShaderCompiler,
		},
	})
}

func (ctx *context) RenderbufferStorage(target, internalFormat Enum, width, height int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnRenderbufferStorage,
			a0: target.c(),
			a1: internalFormat.c(),
			a2: C.uintptr_t(width),
			a3: C.uintptr_t(height),
		},
	})
}

func (ctx *context) SampleCoverage(value float32, invert bool) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnSampleCoverage,
			a0: C.uintptr_t(math.Float32bits(value)),
			a1: glBoolean(invert),
		},
	})
}

func (ctx *context) Scissor(x, y, width, height int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnScissor,
			a0: C.uintptr_t(x),
			a1: C.uintptr_t(y),
			a2: C.uintptr_t(width),
			a3: C.uintptr_t(height),
		},
	})
}

func (ctx *context) ShaderSource(s Shader, src string) {
	// We are passing a char**. Make sure both the string and its
	// containing 1-element array are off the stack. Both are freed
	// in work.c.
	cstr := C.CString(src)
	cstrp := (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(cstr))))
	*cstrp = cstr

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnShaderSource,
			a0: s.c(),
			a1: 1,
			a2: C.uintptr_t(uintptr(unsafe.Pointer(cstrp))),
		},
	})
}

func (ctx *context) StencilFunc(fn Enum, ref int, mask uint32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilFunc,
			a0: fn.c(),
			a1: C.uintptr_t(ref),
			a2: C.uintptr_t(mask),
		},
	})
}

func (ctx *context) StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilFuncSeparate,
			a0: face.c(),
			a1: fn.c(),
			a2: C.uintptr_t(ref),
			a3: C.uintptr_t(mask),
		},
	})
}

func (ctx *context) StencilMask(mask uint32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilMask,
			a0: C.uintptr_t(mask),
		},
	})
}

func (ctx *context) StencilMaskSeparate(face Enum, mask uint32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilMaskSeparate,
			a0: face.c(),
			a1: C.uintptr_t(mask),
		},
	})
}

func (ctx *context) StencilOp(fail, zfail, zpass Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilOp,
			a0: fail.c(),
			a1: zfail.c(),
			a2: zpass.c(),
		},
	})
}

func (ctx *context) StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnStencilOpSeparate,
			a0: face.c(),
			a1: sfail.c(),
			a2: dpfail.c(),
			a3: dppass.c(),
		},
	})
}

func (ctx *context) TexImage2D(target Enum, level int, width, height int, format Enum, ty Enum, data []byte) {
	// It is common to pass TexImage2D a nil data, indicating that a
	// bound GL buffer is being used as the source. In that case, it
	// is not necessary to block.
	parg := unsafe.Pointer(nil)
	if len(data) > 0 {
		parg = unsafe.Pointer(&data[0])
	}

	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexImage2D,
			// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: C.uintptr_t(format),
			a3: C.uintptr_t(width),
			a4: C.uintptr_t(height),
			a5: format.c(),
			a6: ty.c(),
		},
		parg:     parg,
		blocking: parg != nil,
	})
}

func (ctx *context) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexSubImage2D,
			// TODO(crawshaw): GLES3 offset for PIXEL_UNPACK_BUFFER and PIXEL_PACK_BUFFER.
			a0: target.c(),
			a1: C.uintptr_t(level),
			a2: C.uintptr_t(x),
			a3: C.uintptr_t(y),
			a4: C.uintptr_t(width),
			a5: C.uintptr_t(height),
			a6: format.c(),
			a7: ty.c(),
		},
		parg:     unsafe.Pointer(&data[0]),
		blocking: true,
	})
}

func (ctx *context) TexParameterf(target, pname Enum, param float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexParameterf,
			a0: target.c(),
			a1: pname.c(),
			a2: C.uintptr_t(math.Float32bits(param)),
		},
	})
}

func (ctx *context) TexParameterfv(target, pname Enum, params []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexParameterfv,
			a0: target.c(),
			a1: pname.c(),
		},
		parg:     unsafe.Pointer(&params[0]),
		blocking: true,
	})
}

func (ctx *context) TexParameteri(target, pname Enum, param int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexParameteri,
			a0: target.c(),
			a1: pname.c(),
			a2: C.uintptr_t(param),
		},
	})
}

func (ctx *context) TexParameteriv(target, pname Enum, params []int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnTexParameteriv,
			a0: target.c(),
			a1: pname.c(),
		},
		parg:     unsafe.Pointer(&params[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform1f(dst Uniform, v float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform1f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(v)),
		},
	})
}

func (ctx *context) Uniform1fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform1fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src)),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform1i(dst Uniform, v int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform1i,
			a0: dst.c(),
			a1: C.uintptr_t(v),
		},
	})
}

func (ctx *context) Uniform1iv(dst Uniform, src []int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform1iv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src)),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform2f(dst Uniform, v0, v1 float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform2f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(v0)),
			a2: C.uintptr_t(math.Float32bits(v1)),
		},
	})
}

func (ctx *context) Uniform2fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform2fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 2),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform2i(dst Uniform, v0, v1 int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform2i,
			a0: dst.c(),
			a1: C.uintptr_t(v0),
			a2: C.uintptr_t(v1),
		},
	})
}

func (ctx *context) Uniform2iv(dst Uniform, src []int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform2iv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 2),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform3f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(v0)),
			a2: C.uintptr_t(math.Float32bits(v1)),
			a3: C.uintptr_t(math.Float32bits(v2)),
		},
	})
}

func (ctx *context) Uniform3fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform3fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 3),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform3i(dst Uniform, v0, v1, v2 int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform3i,
			a0: dst.c(),
			a1: C.uintptr_t(v0),
			a2: C.uintptr_t(v1),
			a3: C.uintptr_t(v2),
		},
	})
}

func (ctx *context) Uniform3iv(dst Uniform, src []int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform3iv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 3),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform4f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(v0)),
			a2: C.uintptr_t(math.Float32bits(v1)),
			a3: C.uintptr_t(math.Float32bits(v2)),
			a4: C.uintptr_t(math.Float32bits(v3)),
		},
	})
}

func (ctx *context) Uniform4fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform4fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 4),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform4i,
			a0: dst.c(),
			a1: C.uintptr_t(v0),
			a2: C.uintptr_t(v1),
			a3: C.uintptr_t(v2),
			a4: C.uintptr_t(v3),
		},
	})
}

func (ctx *context) Uniform4iv(dst Uniform, src []int32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniform4iv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 4),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) UniformMatrix2fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniformMatrix2fv,
			// OpenGL ES 2 does not support transpose.
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 4),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) UniformMatrix3fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniformMatrix3fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 9),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) UniformMatrix4fv(dst Uniform, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUniformMatrix4fv,
			a0: dst.c(),
			a1: C.uintptr_t(len(src) / 16),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) UseProgram(p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnUseProgram,
			a0: p.c(),
		},
	})
}

func (ctx *context) ValidateProgram(p Program) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnValidateProgram,
			a0: p.c(),
		},
	})
}

func (ctx *context) VertexAttrib1f(dst Attrib, x float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib1f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(x)),
		},
	})
}

func (ctx *context) VertexAttrib1fv(dst Attrib, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib1fv,
			a0: dst.c(),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) VertexAttrib2f(dst Attrib, x, y float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib2f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(x)),
			a2: C.uintptr_t(math.Float32bits(y)),
		},
	})
}

func (ctx *context) VertexAttrib2fv(dst Attrib, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib2fv,
			a0: dst.c(),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) VertexAttrib3f(dst Attrib, x, y, z float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib3f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(x)),
			a2: C.uintptr_t(math.Float32bits(y)),
			a3: C.uintptr_t(math.Float32bits(z)),
		},
	})
}

func (ctx *context) VertexAttrib3fv(dst Attrib, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib3fv,
			a0: dst.c(),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) VertexAttrib4f(dst Attrib, x, y, z, w float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib4f,
			a0: dst.c(),
			a1: C.uintptr_t(math.Float32bits(x)),
			a2: C.uintptr_t(math.Float32bits(y)),
			a3: C.uintptr_t(math.Float32bits(z)),
			a4: C.uintptr_t(math.Float32bits(w)),
		},
	})
}

func (ctx *context) VertexAttrib4fv(dst Attrib, src []float32) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttrib4fv,
			a0: dst.c(),
		},
		parg:     unsafe.Pointer(&src[0]),
		blocking: true,
	})
}

func (ctx *context) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnVertexAttribPointer,
			a0: dst.c(),
			a1: C.uintptr_t(size),
			a2: ty.c(),
			a3: glBoolean(normalized),
			a4: C.uintptr_t(stride),
			a5: C.uintptr_t(offset),
		},
	})
}

func (ctx *context) Viewport(x, y, width, height int) {
	ctx.enqueue(call{
		args: C.struct_fnargs{
			fn: C.glfnViewport,
			a0: C.uintptr_t(x),
			a1: C.uintptr_t(y),
			a2: C.uintptr_t(width),
			a3: C.uintptr_t(height),
		},
	})
}
