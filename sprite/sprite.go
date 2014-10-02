// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sprite blah blah blah TODO.
//
// Typical main loop:
//
//	for each frame {
//		quantize time.Now() to a sprite.Time
//		process UI events
//		modify the scene's nodes and animations (Arranger values)
//		e.Render(scene, t)
//	}
package sprite

import (
	"image"

	"code.google.com/p/go.mobile/f32"
)

// TODO: move this to a code.google.com/p/go.mobile/clock package?
type Time int32

type Arranger interface {
	Arrange(e Engine, n *Node, t Time)
}

type Sheet int32

type Texture int32

type Engine interface {
	Register(n *Node)
	Unregister(n *Node)

	LoadSheet(a image.Image) (Sheet, error)
	LoadTexture(s Sheet, bounds image.Rectangle) (Texture, error)
	UnloadSheet(s Sheet) error
	UnloadTexture(x Texture) error

	SetTexture(n *Node, t Time, x Texture)
	SetTransform(n *Node, t Time, m f32.Mat3) // sets transform relative to parent.

	Render(scene *Node, t Time)
}

type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Arranger Arranger

	// EngineFields contains fields that should only be accessed by Engine
	// implementations. It is exported because such implementations can be
	// in other packages.
	EngineFields struct {
		// TODO: separate TexDirty and TransformDirty bits?
		Dirty   bool
		Index   int32
		Texture Texture
	}
}

// TODO: Node parent/sibling/child-related methods.
