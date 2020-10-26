// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package gl

import (
	"syscall/js"
)

func (ctx *context) ActiveTexture(texture Enum) {
	ctx.cachedCall("activeTexture", texture)
}

func (ctx *context) AttachShader(p Program, s Shader) {
	ctx.cachedCall("attachShader", p.c(), s.c())
}

func (ctx *context) BindAttribLocation(p Program, a Attrib, name string) {
	ctx.cachedCall("bindAttribLocation", p.c(), a.c(), name)
}

func (ctx *context) BindBuffer(target Enum, b Buffer) {
	ctx.cachedCall("bindBuffer", target, b.c())
}

func (ctx *context) BindFramebuffer(target Enum, fb Framebuffer) {
	ctx.cachedCall("bindFramebuffer", target, fb.c())
}

func (ctx *context) BindRenderbuffer(target Enum, rb Renderbuffer) {
	ctx.cachedCall("bindRenderbuffer", target, rb.c())
}

func (ctx *context) BindTexture(target Enum, t Texture) {
	ctx.cachedCall("bindTexture", target, t.c())
}

func (ctx *context) BindVertexArray(va VertexArray) {
	ctx.cachedCall("bindVertexArray", va.c())
}

func (ctx *context) BlendColor(red, green, blue, alpha float32) {
	ctx.cachedCall("blendColor", red, green, blue, alpha)
}

func (ctx *context) BlendEquation(mode Enum) {
	ctx.cachedCall("blendEquation", mode)
}

func (ctx *context) BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	ctx.cachedCall("blendEquationSeparate", modeRGB, modeAlpha)
}

func (ctx *context) BlendFunc(sfactor, dfactor Enum) {
	ctx.cachedCall("blendFunc", sfactor, dfactor)
}

func (ctx *context) BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	ctx.cachedCall("blendFuncSeparate", sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha)
}

func (ctx *context) BufferData(target Enum, src []byte, usage Enum) {
	a := js.Null()
	if len(src) > 0 {
		a = jsBytes(src)
	}

	ctx.cachedCall("bufferData", target, a, usage)
}

func (ctx *context) BufferInit(target Enum, size int, usage Enum) {
	ctx.cachedCall("bufferData", target, size, usage)
}

func (ctx *context) BufferSubData(target Enum, offset int, data []byte) {
	ctx.cachedCall("bufferSubData", target, offset, jsBytes(data))
}

func (ctx *context) CheckFramebufferStatus(target Enum) Enum {
	return Enum(toInt(ctx.cachedCall("checkFramebufferStatus", target)))
}

func (ctx *context) Clear(mask Enum) {
	ctx.cachedCall("clear", mask)
}

func (ctx *context) ClearColor(red, green, blue, alpha float32) {
	ctx.cachedCall("clearColor", red, green, blue, alpha)
}

func (ctx *context) ClearDepthf(d float32) {
	ctx.cachedCall("clearDepth", d)
}

func (ctx *context) ClearStencil(s int) {
	ctx.cachedCall("clearStencil", s)
}

func (ctx *context) ColorMask(red, green, blue, alpha bool) {
	ctx.cachedCall("colorMask", red, green, blue, alpha)
}

func (ctx *context) CompileShader(s Shader) {
	ctx.cachedCall("compileShader", s.c())
}

func (ctx *context) CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data []byte) {
	ctx.cachedCall("compressedTexImage2D", target, level, internalformat, width, height, border, jsBytes(data))
}

func (ctx *context) CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data []byte) {
	ctx.cachedCall("compressedTexSubImage2D", target, level, xoffset, yoffset, width, height, format, jsBytes(data))
}

func (ctx *context) CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
	ctx.cachedCall("copyTexImage2D", target, level, internalformat, x, y, width, height, border)
}

func (ctx *context) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	ctx.cachedCall("copyTexSubImage2D", target, level, xoffset, yoffset, x, y, width, height)
}

func (ctx *context) CreateBuffer() Buffer {
	v := ctx.cachedCall("createBuffer")

	return Buffer{Value: jsPtr(v)}
}

func (ctx *context) CreateFramebuffer() Framebuffer {
	v := ctx.cachedCall("createFramebuffer")

	return Framebuffer{Value: jsPtr(v)}
}

func (ctx *context) CreateProgram() Program {
	v := jsPtr(ctx.cachedCall("createProgram"))

	return Program{Init: v != nil, Value: v}
}

func (ctx *context) CreateRenderbuffer() Renderbuffer {
	v := ctx.cachedCall("createRenderbuffer")

	return Renderbuffer{Value: jsPtr(v)}
}

func (ctx *context) CreateShader(ty Enum) Shader {
	v := ctx.cachedCall("createShader", ty)

	return Shader{Value: jsPtr(v)}
}

func (ctx *context) CreateTexture() Texture {
	v := ctx.cachedCall("createTexture")

	return Texture{Value: jsPtr(v)}
}

func (ctx *context) CreateVertexArray() VertexArray {
	v := ctx.cachedCall("createVertexArray")

	return VertexArray{Value: jsPtr(v)}
}

func (ctx *context) CullFace(mode Enum) {
	ctx.cachedCall("cullFace", mode)
}

func (ctx *context) DeleteBuffer(v Buffer) {
	ctx.cachedCall("deleteBuffer", v.c())
}

func (ctx *context) DeleteFramebuffer(v Framebuffer) {
	ctx.cachedCall("deleteFramebuffer", v.c())
}

func (ctx *context) DeleteProgram(p Program) {
	ctx.cachedCall("deleteProgram", p.c())
}

func (ctx *context) DeleteRenderbuffer(v Renderbuffer) {
	ctx.cachedCall("deleteRenderbuffer", v.c())
}

func (ctx *context) DeleteShader(s Shader) {
	ctx.cachedCall("deleteShader", s.c())
}

func (ctx *context) DeleteTexture(v Texture) {
	ctx.cachedCall("deleteTexture", v)
}

func (ctx *context) DeleteVertexArray(v VertexArray) {
	ctx.cachedCall("deleteVertexArray", v.c())
}

func (ctx *context) DepthFunc(fn Enum) {
	ctx.cachedCall("depthFunc", fn)
}

func (ctx *context) DepthMask(flag bool) {
	ctx.cachedCall("depthMask", flag)
}

func (ctx *context) DepthRangef(n, f float32) {
	ctx.cachedCall("depthRange", n, f)
}

func (ctx *context) DetachShader(p Program, s Shader) {
	ctx.cachedCall("detachShader", p.c(), s.c())
}

func (ctx *context) Disable(cap Enum) {
	ctx.cachedCall("disable", cap)
}

func (ctx *context) DisableVertexAttribArray(a Attrib) {
	ctx.cachedCall("disableVertexAttribArray", a.c())
}

func (ctx *context) DrawArrays(mode Enum, first, count int) {
	ctx.cachedCall("drawArrays", mode, first, count)
}

func (ctx *context) DrawElements(mode Enum, count int, ty Enum, offset int) {
	ctx.cachedCall("drawElements", mode, count, ty, offset)
}

func (ctx *context) Enable(cap Enum) {
	ctx.cachedCall("enable", cap)
}

func (ctx *context) EnableVertexAttribArray(a Attrib) {
	ctx.cachedCall("enableVertexAttribArray", a.c())
}

func (ctx *context) Finish() {
	ctx.cachedCall("finish")
}

func (ctx *context) Flush() {
	ctx.cachedCall("flush")
}

func (ctx *context) FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
	ctx.cachedCall("framebufferRenderbuffer", target, attachment, rbTarget, rb)
}

func (ctx *context) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	ctx.cachedCall("framebufferTexture2D", target, attachment, texTarget, t.c(), level)
}

func (ctx *context) FrontFace(mode Enum) {
	ctx.cachedCall("frontFace", mode)
}

func (ctx *context) GenerateMipmap(target Enum) {
	ctx.cachedCall("generateMipmap", target)
}

func (ctx *context) GetActiveAttrib(p Program, index uint32) (name string, size int, ty Enum) {
	info := ctx.cachedCall("getActiveAttrib", p.c(), index)

	return info.Get("name").String(), info.Get("size").Int(), Enum(info.Get("type").Int())
}

func (ctx *context) GetActiveUniform(p Program, index uint32) (name string, size int, ty Enum) {
	info := ctx.cachedCall("getActiveUniform", p.c(), index)

	return info.Get("name").String(), info.Get("size").Int(), Enum(info.Get("type").Int())
}

func (ctx *context) GetAttachedShaders(p Program) []Shader {
	a := ctx.cachedCall("getAttachedShaders", p.c())

	shaders := make([]Shader, a.Length())
	for i := range shaders {
		v := a.Index(i)

		shaders[i] = Shader{Value: jsPtr(v)}
	}

	return shaders
}

func (ctx *context) GetAttribLocation(p Program, name string) Attrib {
	v := ctx.cachedCall("getAttribLocation", p.c(), name)

	return Attrib{Value: jsPtr(v)}
}

func (ctx *context) GetBooleanv(dst []bool, pname Enum) {
	v := ctx.cachedCall("getParameter", pname)

	if v.Type() == js.TypeBoolean {
		dst[0] = v.Bool()
	} else {
		for i := range dst {
			dst[i] = v.Index(i).Bool()
		}
	}
}

func (ctx *context) GetFloatv(dst []float32, pname Enum) {
	v := ctx.cachedCall("getParameter", pname)

	if v.Type() == js.TypeNumber {
		dst[0] = float32(v.Float())
	} else {
		for i := range dst {
			dst[i] = float32(v.Index(i).Float())
		}
	}
}

func (ctx *context) GetIntegerv(dst []int32, pname Enum) {
	v := ctx.cachedCall("getParameter", pname)

	if v.Type() == js.TypeObject {
		for i := range dst {
			dst[i] = int32(toInt(v.Index(i)))
		}
	} else {
		dst[0] = int32(toInt(v))
	}
}

func (ctx *context) GetInteger(pname Enum) int {
	var v [1]int32
	ctx.GetIntegerv(v[:], pname)
	return int(v[0])
}

func (ctx *context) GetBufferParameteri(target, value Enum) int {
	return toInt(ctx.cachedCall("getBufferParameter", target, value))
}

func (ctx *context) GetError() Enum {
	return Enum(ctx.cachedCall("getError").Int())
}

func (ctx *context) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	return toInt(ctx.cachedCall("getFramebufferAttachmentParameter", target, attachment, pname))
}

func (ctx *context) GetProgrami(p Program, pname Enum) int {
	return toInt(ctx.cachedCall("getProgramParameter", p.c(), pname))
}

func (ctx *context) GetProgramInfoLog(p Program) string {
	return ctx.cachedCall("getProgramInfoLog", p.c()).String()
}

func (ctx *context) GetRenderbufferParameteri(target, pname Enum) int {
	return toInt(ctx.cachedCall("getRenderbufferParameter", target, pname))
}

func (ctx *context) GetShaderi(s Shader, pname Enum) int {
	return toInt(ctx.cachedCall("getShaderParameter", s.c(), pname))
}

func (ctx *context) GetShaderInfoLog(s Shader) string {
	return ctx.cachedCall("getShaderInfoLog", s.c()).String()
}

func (ctx *context) GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeLow, rangeHigh, precision int) {
	v := ctx.cachedCall("getShaderPrecisionFormat", shadertype, precisiontype)

	return v.Get("rangeMin").Int(), v.Get("rangeMax").Int(), v.Get("precision").Int()
}

func (ctx *context) GetShaderSource(s Shader) string {
	return ctx.cachedCall("getShaderSource", s.c()).String()
}

func (ctx *context) GetString(pname Enum) string {
	return ctx.cachedCall("getParameter", pname).String()
}

func (ctx *context) GetTexParameterfv(dst []float32, target, pname Enum) {
	dst[0] = float32(ctx.cachedCall("getTexParameter", target, pname).Float())
}

func (ctx *context) GetTexParameteriv(dst []int32, target, pname Enum) {
	dst[0] = int32(toInt(ctx.cachedCall("getTexParameter", target, pname)))
}

func (ctx *context) GetUniformfv(dst []float32, src Uniform, p Program) {
	v := ctx.cachedCall("getUniform", p.c(), src.c())

	if v.Type() == js.TypeNumber {
		dst[0] = float32(v.Float())
	} else {
		for i := range dst {
			dst[i] = float32(v.Index(i).Float())
		}
	}
}

func (ctx *context) GetUniformiv(dst []int32, src Uniform, p Program) {
	v := ctx.cachedCall("getUniform", p.c(), src.c())

	if v.Type() == js.TypeObject {
		for i := range dst {
			dst[i] = int32(toInt(v.Index(i)))
		}
	} else {
		dst[0] = int32(toInt(v))
	}
}

func (ctx *context) GetUniformLocation(p Program, name string) Uniform {
	v := ctx.cachedCall("getUniformLocation", p.c(), name)

	return Uniform{Value: jsPtr(v)}
}

func (ctx *context) GetVertexAttribf(src Attrib, pname Enum) float32 {
	var params [1]float32
	ctx.GetVertexAttribfv(params[:], src, pname)
	return params[0]
}

func (ctx *context) GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
	v := ctx.cachedCall("getVertexAttrib", src.c(), pname)

	if v.Type() == js.TypeNumber {
		dst[0] = float32(v.Float())
	} else {
		for i := range dst {
			dst[i] = float32(v.Index(i).Float())
		}
	}
}

func (ctx *context) GetVertexAttribi(src Attrib, pname Enum) int32 {
	var params [1]int32
	ctx.GetVertexAttribiv(params[:], src, pname)
	return params[0]
}

func (ctx *context) GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
	v := ctx.cachedCall("getVertexAttrib", src.c(), pname)

	if v.Type() == js.TypeObject {
		for i := range dst {
			dst[i] = int32(toInt(v.Index(i)))
		}
	} else {
		dst[0] = int32(toInt(v))
	}
}

func (ctx *context) Hint(target, mode Enum) {
	ctx.cachedCall("hint", target, mode)
}

func (ctx *context) IsBuffer(b Buffer) bool {
	return ctx.cachedCall("isBuffer", b.c()).Bool()
}

func (ctx *context) IsEnabled(cap Enum) bool {
	return ctx.cachedCall("isEnabled", cap).Bool()
}

func (ctx *context) IsFramebuffer(fb Framebuffer) bool {
	return ctx.cachedCall("isFramebuffer", fb.c()).Bool()
}

func (ctx *context) IsProgram(p Program) bool {
	return ctx.cachedCall("isProgram", p.c()).Bool()
}

func (ctx *context) IsRenderbuffer(rb Renderbuffer) bool {
	return ctx.cachedCall("isRenderbuffer", rb.c()).Bool()
}

func (ctx *context) IsShader(s Shader) bool {
	return ctx.cachedCall("isShader", s.c()).Bool()
}

func (ctx *context) IsTexture(t Texture) bool {
	return ctx.cachedCall("isTexture", t.c()).Bool()
}

func (ctx *context) LineWidth(width float32) {
	ctx.cachedCall("lineWidth", width)
}

func (ctx *context) LinkProgram(p Program) {
	ctx.cachedCall("linkProgram", p.c())
}

func (ctx *context) PixelStorei(pname Enum, param int32) {
	ctx.cachedCall("pixelStorei", pname, param)
}

func (ctx *context) PolygonOffset(factor, units float32) {
	ctx.cachedCall("polygonOffset", factor, units)
}

func (ctx *context) ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
	pixels := uint8ArrayCtor.New(len(dst))

	ctx.cachedCall("readPixels", x, y, width, height, format, ty, pixels)

	js.CopyBytesToGo(dst, pixels)
}

func (ctx *context) ReleaseShaderCompiler() {
	// not supported in WebGL
}

func (ctx *context) RenderbufferStorage(target, internalFormat Enum, width, height int) {
	ctx.cachedCall("renderbufferStorage", target, internalFormat, width, height)
}

func (ctx *context) SampleCoverage(value float32, invert bool) {
	ctx.cachedCall("sampleCoverage", value, invert)
}

func (ctx *context) Scissor(x, y, width, height int32) {
	ctx.cachedCall("scissor", x, y, width, height)
}

func (ctx *context) ShaderSource(s Shader, src string) {
	ctx.cachedCall("shaderSource", s.c(), src)
}

func (ctx *context) StencilFunc(fn Enum, ref int, mask uint32) {
	ctx.cachedCall("stencilFunc", fn, ref, mask)
}

func (ctx *context) StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
	ctx.cachedCall("stencilFuncSeparate", face, fn, ref, mask)
}

func (ctx *context) StencilMask(mask uint32) {
	ctx.cachedCall("stencilMask", mask)
}

func (ctx *context) StencilMaskSeparate(face Enum, mask uint32) {
	ctx.cachedCall("stencilMaskSeparate", face, mask)
}

func (ctx *context) StencilOp(fail, zfail, zpass Enum) {
	ctx.cachedCall("stencilOp", fail, zfail, zpass)
}

func (ctx *context) StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
	ctx.cachedCall("stencilOpSeparate", face, sfail, dpfail, dppass)
}

func (ctx *context) TexImage2D(target Enum, level int, internalFormat int, width, height int, format Enum, ty Enum, data []byte) {
	ctx.cachedCall("texImage2D", target, level, internalFormat, width, height, 0, format, ty, jsBytes(data))
}

func (ctx *context) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	ctx.cachedCall("texSubImage2D", target, level, x, y, width, height, format, ty, jsBytes(data))
}

func (ctx *context) TexParameterf(target, pname Enum, param float32) {
	ctx.cachedCall("texParameterf", target, pname, param)
}

func (ctx *context) TexParameterfv(target, pname Enum, params []float32) {
	args := []interface{}{target, pname}
	for _, f := range params {
		args = append(args, f)
	}

	ctx.cachedCall("texParameterf", args...)
}

func (ctx *context) TexParameteri(target, pname Enum, param int) {
	ctx.cachedCall("texParameteri", target, pname, param)
}

func (ctx *context) TexParameteriv(target, pname Enum, params []int32) {
	args := []interface{}{target, pname}
	for _, i := range params {
		args = append(args, i)
	}

	ctx.cachedCall("texParameteri", args...)
}

func (ctx *context) Uniform1f(dst Uniform, v float32) {
	ctx.cachedCall("uniform1f", dst.c(), v)
}

func (ctx *context) Uniform1fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniform1f", dst.c(), src[0])
}

func (ctx *context) Uniform1i(dst Uniform, v int) {
	ctx.cachedCall("uniform1i", dst.c(), v)
}

func (ctx *context) Uniform1iv(dst Uniform, src []int32) {
	ctx.cachedCall("uniform1i", dst.c(), src[0])
}

func (ctx *context) Uniform2f(dst Uniform, v0, v1 float32) {
	ctx.cachedCall("uniform2f", dst.c(), v0, v1)
}

func (ctx *context) Uniform2fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniform2f", dst.c(), src[0], src[1])
}

func (ctx *context) Uniform2i(dst Uniform, v0, v1 int) {
	ctx.cachedCall("uniform2i", dst.c(), v0, v1)
}

func (ctx *context) Uniform2iv(dst Uniform, src []int32) {
	ctx.cachedCall("uniform2i", dst.c(), src[0], src[1])
}

func (ctx *context) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	ctx.cachedCall("uniform3f", dst.c(), v0, v1, v2)
}

func (ctx *context) Uniform3fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniform3f", dst.c(), src[0], src[1], src[2])
}

func (ctx *context) Uniform3i(dst Uniform, v0, v1, v2 int32) {
	ctx.cachedCall("uniform3i", dst.c(), v0, v1, v2)
}

func (ctx *context) Uniform3iv(dst Uniform, src []int32) {
	ctx.cachedCall("uniform3i", dst.c(), src[0], src[1], src[2])
}

func (ctx *context) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	ctx.cachedCall("uniform4f", dst.c(), v0, v1, v2, v3)
}

func (ctx *context) Uniform4fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniform4f", dst.c(), src[0], src[1], src[2], src[3])
}

func (ctx *context) Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
	ctx.cachedCall("uniform4i", dst.c(), v0, v1, v2, v3)
}

func (ctx *context) Uniform4iv(dst Uniform, src []int32) {
	ctx.cachedCall("uniform4i", dst.c(), src[0], src[1], src[2], src[3])
}

func (ctx *context) UniformMatrix2fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix2fv", dst.c(), false, jsFloats(src))
}

func (ctx *context) UniformMatrix3fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix3fv", dst.c(), false, jsFloats(src))
}

func (ctx *context) UniformMatrix4fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix4fv", dst.c(), false, jsFloats(src))
}

func (ctx *context) UseProgram(p Program) {
	ctx.cachedCall("useProgram", p.c())
}

func (ctx *context) ValidateProgram(p Program) {
	ctx.cachedCall("validateProgram", p.c())
}

func (ctx *context) VertexAttrib1f(dst Attrib, x float32) {
	ctx.cachedCall("vertexAttrib1f", dst.c(), x)
}

func (ctx *context) VertexAttrib1fv(dst Attrib, src []float32) {
	ctx.cachedCall("vertexAttrib1f", dst.c(), src[0])
}

func (ctx *context) VertexAttrib2f(dst Attrib, x, y float32) {
	ctx.cachedCall("vertexAttrib2f", dst.c(), x, y)
}

func (ctx *context) VertexAttrib2fv(dst Attrib, src []float32) {
	ctx.cachedCall("vertexAttrib2f", dst.c(), src[0], src[1])
}

func (ctx *context) VertexAttrib3f(dst Attrib, x, y, z float32) {
	ctx.cachedCall("vertexAttrib3f", dst.c(), x, y, z)
}

func (ctx *context) VertexAttrib3fv(dst Attrib, src []float32) {
	ctx.cachedCall("vertexAttrib3f", dst.c(), src[0], src[1], src[2])
}

func (ctx *context) VertexAttrib4f(dst Attrib, x, y, z, w float32) {
	ctx.cachedCall("vertexAttrib4f", dst.c(), x, y, z, w)
}

func (ctx *context) VertexAttrib4fv(dst Attrib, src []float32) {
	ctx.cachedCall("vertexAttrib4f", dst.c(), src[0], src[1], src[2], src[3])
}

func (ctx *context) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	ctx.cachedCall("vertexAttribPointer", dst.c(), size, ty, normalized, stride, offset)
}

func (ctx *context) Viewport(x, y, width, height int) {
	ctx.cachedCall("viewport", x, y, width, height)
}

func (ctx context3) UniformMatrix2x3fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix2x3fv", dst, false, jsFloats(src))
}

func (ctx context3) UniformMatrix3x2fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix3x2fv", dst, false, jsFloats(src))
}

func (ctx context3) UniformMatrix2x4fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix2x4fv", dst, false, jsFloats(src))
}

func (ctx context3) UniformMatrix4x2fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix4x2fv", dst, false, jsFloats(src))
}

func (ctx context3) UniformMatrix3x4fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix3x4fv", dst, false, jsFloats(src))
}

func (ctx context3) UniformMatrix4x3fv(dst Uniform, src []float32) {
	ctx.cachedCall("uniformMatrix4x3fv", dst, false, jsFloats(src))
}

func (ctx context3) BlitFramebuffer(srcX0, srcY0, srcX1, srcY1, dstX0, dstY0, dstX1, dstY1 int, mask uint, filter Enum) {
	ctx.cachedCall("blitFramebuffer", srcX0, srcY0, srcX1, srcY1, dstX0, dstY0, dstX1, dstY1, mask, filter)
}

func (ctx context3) Uniform1ui(dst Uniform, v uint32) {
	ctx.cachedCall("uniform1ui", dst.c(), v)
}

func (ctx context3) Uniform2ui(dst Uniform, v0, v1 uint32) {
	ctx.cachedCall("uniform2ui", dst.c(), v0, v1)
}

func (ctx context3) Uniform3ui(dst Uniform, v0, v1, v2 uint) {
	ctx.cachedCall("uniform3ui", dst.c(), v0, v1, v2)
}

func (ctx context3) Uniform4ui(dst Uniform, v0, v1, v2, v3 uint32) {
	ctx.cachedCall("uniform4ui", dst.c(), v0, v1, v2, v3)
}
