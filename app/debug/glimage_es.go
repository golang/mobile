// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package debug

const vertexShader = `
uniform mat4 mvp;
attribute vec4 pos;
attribute vec2 inUV;
varying vec2 UV;
void main() {
	gl_Position = mvp * pos;
	UV = inUV;
}
`

const fragmentShader = `
precision mediump float;
varying vec2 UV;
uniform sampler2D textureSample;
void main(){
	gl_FragColor = texture2D(textureSample, UV);
}
`
