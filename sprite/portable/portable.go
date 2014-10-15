// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package portable blah blah blah TODO.
package portable

import (
	"fmt"
	"image"
	"image/draw"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"
)

func Engine(dst *image.RGBA) sprite.Engine {
	return &engine{
		dst:      dst,
		nodes:    []*node{nil},
		sheets:   []*image.RGBA{nil},
		textures: []*image.RGBA{nil},
	}
}

type node struct {
	// TODO: move this into package sprite as Node.EngineFields.RelTransform??
	relTransform f32.Affine
}

var _ sprite.Engine = (*engine)(nil)

type engine struct {
	dst           *image.RGBA
	nodes         []*node
	sheets        []*image.RGBA
	textures      []*image.RGBA
	absTransforms []f32.Affine
}

func (e *engine) Register(n *sprite.Node) {
	if n.EngineFields.Index != 0 {
		panic("portable: sprite.Node already registered")
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
	rgba, ok := a.(*image.RGBA)
	if !ok {
		b := a.Bounds()
		rgba = image.NewRGBA(b)
		draw.Draw(rgba, b, a, b.Min, draw.Src)
	}
	e.sheets = append(e.sheets, rgba)
	return sprite.Sheet(len(e.sheets) - 1), nil
}

func (e *engine) LoadTexture(s sprite.Sheet, bounds image.Rectangle) (sprite.Texture, error) {
	if s < 0 || len(e.sheets) <= int(s) {
		return 0, fmt.Errorf("portable: LoadTexture: sheet %d is out of bounds", s)
	}
	e.textures = append(e.textures, e.sheets[s].SubImage(bounds).(*image.RGBA))
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
		{1 / geom.Scale, 0, 0},
		{0, 1 / geom.Scale, 0},
	})
	e.render(scene, t)
}

func (e *engine) render(n *sprite.Node, t clock.Time) {
	if n.EngineFields.Index == 0 {
		panic("portable: sprite.Node not registered")
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

	if x := e.textures[n.EngineFields.Texture]; x != nil {
		m.Inverse(&m)
		affine(e.dst, x, &m, draw.Over)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.render(c, t)
	}

	// Pop absTransforms.
	e.absTransforms = e.absTransforms[:len(e.absTransforms)-1]
}
