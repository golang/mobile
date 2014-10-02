// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package portable blah blah blah TODO.
package portable

import (
	"fmt"
	"image"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/sprite"
)

func Engine(dst *image.RGBA) sprite.Engine {
	return &engine{
		dst:      dst,
		nodes:    []*node{nil},
		sheets:   []image.Image{nil},
		textures: []*texture{nil},
	}
}

type node struct {
	// TODO: move this into package sprite as Node.EngineFields.RelTransform??
	relTransform f32.Mat3
}

type texture struct {
	img    image.Image
	bounds image.Rectangle
}

var _ sprite.Engine = (*engine)(nil)

type engine struct {
	dst           *image.RGBA
	nodes         []*node
	sheets        []image.Image
	textures      []*texture
	absTransforms []f32.Mat3
}

func (e *engine) Register(n *sprite.Node) {
	if n.EngineFields.Index != 0 {
		panic("portable: sprite.Node already registered")
	}

	o := &node{}
	o.relTransform.Identity()

	e.nodes = append(e.nodes, o)
	n.EngineFields.Index = int32(len(e.nodes))
}

func (e *engine) Unregister(n *sprite.Node) {
	panic("todo")
}

func (e *engine) LoadSheet(a image.Image) (sprite.Sheet, error) {
	e.sheets = append(e.sheets, a)
	return sprite.Sheet(len(e.sheets)), nil
}

func (e *engine) LoadTexture(s sprite.Sheet, bounds image.Rectangle) (sprite.Texture, error) {
	if s < 0 || len(e.sheets) <= int(s) {
		return 0, fmt.Errorf("portable: LoadTexture: sheet %d is out of bounds", s)
	}
	e.textures = append(e.textures, &texture{
		img:    e.sheets[s],
		bounds: bounds,
	})
	return sprite.Texture(len(e.textures)), nil
}

func (e *engine) UnloadSheet(s sprite.Sheet) error {
	panic("todo")
}

func (e *engine) UnloadTexture(x sprite.Texture) error {
	panic("todo")
}

func (e *engine) SetTexture(n *sprite.Node, t sprite.Time, x sprite.Texture) {
	n.EngineFields.Dirty = true // TODO: do we need to propagate dirtiness up/down the tree?
	n.EngineFields.Texture = x
}

func (e *engine) SetTransform(n *sprite.Node, t sprite.Time, m f32.Mat3) {
	n.EngineFields.Dirty = true // TODO: do we need to propagate dirtiness up/down the tree?
	e.nodes[n.EngineFields.Index].relTransform = m
}

func (e *engine) Render(scene *sprite.Node, t sprite.Time) {
	e.absTransforms = e.absTransforms[:0]
	e.render(scene, t)
}

func (e *engine) render(n *sprite.Node, t sprite.Time) {
	if n.Arranger != nil {
		n.Arranger.Arrange(e, n, t)
	}

	// Push absTransforms.
	// TODO: cache absolute transforms and use EngineFields.Dirty?
	rel := &e.nodes[n.EngineFields.Index].relTransform
	if len(e.absTransforms) == 0 {
		e.absTransforms = append(e.absTransforms, *rel)
	} else {
		m := f32.Mat3{}
		m.Mul(rel, &e.absTransforms[len(e.absTransforms)-1]) // TODO: swap args order??
		e.absTransforms = append(e.absTransforms, m)
	}

	if x := e.textures[n.EngineFields.Texture]; x != nil {
		// TODO: draw transformed image onto e.dst.
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.render(c, t)
	}

	// Pop absTransforms.
	e.absTransforms = e.absTransforms[:len(e.absTransforms)-1]
}
