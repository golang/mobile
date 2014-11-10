// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package glsprite blah blah blah TODO.
package glsprite

import (
	"image"
	"image/draw"

	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl/glutil"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"
)

type node struct {
	// TODO: move this into package sprite as Node.EngineFields.RelTransform??
	relTransform f32.Affine
}

type sheet struct {
	glImage *glutil.Image
	b       image.Rectangle
}

type texture struct {
	sheet sprite.Sheet
	b     image.Rectangle
}

func Engine() sprite.Engine {
	return &engine{
		nodes:    []*node{nil},
		sheets:   []sheet{{}},
		textures: []texture{{}},
	}
}

type engine struct {
	glImages map[sprite.Texture]*glutil.Image
	nodes    []*node
	sheets   []sheet
	textures []texture

	absTransforms []f32.Affine
}

func (e *engine) Register(n *sprite.Node) {
	if n.EngineFields.Index != 0 {
		panic("glsprite: sprite.Node already registered")
	}
	o := &node{}
	o.relTransform.Identity()

	e.nodes = append(e.nodes, o)
	n.EngineFields.Index = int32(len(e.nodes) - 1)
}

func (e *engine) Unregister(n *sprite.Node) {
	panic("todo")
}

func (e *engine) LoadSheet(a image.Image) (sprite.Sheet, error) {
	b := a.Bounds()
	glImage := glutil.NewImage(b.Dx(), b.Dy())
	draw.Draw(glImage.RGBA, glImage.Bounds(), a, b.Min, draw.Src)
	glImage.Upload()
	// TODO: set "glImage.Pix = nil"?? We don't need the CPU-side image any more.
	e.sheets = append(e.sheets, sheet{
		glImage: glImage,
		b:       b,
	})
	return sprite.Sheet(len(e.sheets) - 1), nil
}

func (e *engine) LoadTexture(s sprite.Sheet, bounds image.Rectangle) (sprite.Texture, error) {
	e.textures = append(e.textures, texture{
		sheet: s,
		b:     bounds,
	})
	return sprite.Texture(len(e.textures) - 1), nil
}

func (e *engine) UnloadSheet(s sprite.Sheet) error {
	panic("todo")
}

func (e *engine) UnloadTexture(x sprite.Texture) error {
	panic("todo")
}

func (e *engine) SetTexture(n *sprite.Node, t clock.Time, x sprite.Texture) {
	n.EngineFields.Dirty = true // TODO: do we need to propagate dirtiness up/down the tree?
	n.EngineFields.Texture = x
}

func (e *engine) SetTransform(n *sprite.Node, t clock.Time, m f32.Affine) {
	n.EngineFields.Dirty = true // TODO: do we need to propagate dirtiness up/down the tree?
	e.nodes[n.EngineFields.Index].relTransform = m
}

func (e *engine) Render(scene *sprite.Node, t clock.Time) {
	e.absTransforms = append(e.absTransforms[:0], f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})
	e.render(scene, t)
}

func (e *engine) render(n *sprite.Node, t clock.Time) {
	if n.EngineFields.Index == 0 {
		panic("glsprite: sprite.Node not registered")
	}
	if n.Arranger != nil {
		n.Arranger.Arrange(e, n, t)
	}

	// Push absTransforms.
	// TODO: cache absolute transforms and use EngineFields.Dirty?
	rel := &e.nodes[n.EngineFields.Index].relTransform
	m := f32.Affine{}
	m.Mul(&e.absTransforms[len(e.absTransforms)-1], rel)
	e.absTransforms = append(e.absTransforms, m)

	if x := e.textures[n.EngineFields.Texture]; x.sheet != 0 {
		if 0 < x.sheet && int(x.sheet) < len(e.sheets) {
			e.sheets[x.sheet].glImage.Draw(
				geom.Point{
					geom.Pt(m[0][2]),
					geom.Pt(m[1][2]),
				},
				geom.Point{
					geom.Pt(m[0][2] + m[0][0]),
					geom.Pt(m[1][2] + m[1][0]),
				},
				geom.Point{
					geom.Pt(m[0][2] + m[0][1]),
					geom.Pt(m[1][2] + m[1][1]),
				},
				x.b,
			)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.render(c, t)
	}

	// Pop absTransforms.
	e.absTransforms = e.absTransforms[:len(e.absTransforms)-1]
}
