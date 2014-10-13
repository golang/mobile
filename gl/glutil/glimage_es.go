// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

const vertexShader = `#version 100
uniform mat3 mvp;
uniform mat3 uvp;
attribute vec3 pos;
attribute vec2 inUV;
varying vec2 UV;
void main() {
	vec3 p = pos;
	p.z = 1.0;
	gl_Position = vec4(mvp * p, 1);
	UV = (uvp * vec3(inUV, 1)).xy;
}
`

const fragmentShader = `#version 100
precision mediump float;
varying vec2 UV;
uniform sampler2D textureSample;
void main(){
	gl_FragColor = texture2D(textureSample, UV);
}
`
