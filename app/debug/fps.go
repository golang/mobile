// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package debug provides GL-based debugging tools for apps.
package debug // import "golang.org/x/mobile/app/debug"

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"
	"sync"
	"time"

	"code.google.com/p/freetype-go/freetype"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/exp/font"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl/glutil"
)

var lastDraw = time.Now()

var monofont = freetype.NewContext()

var fps struct {
	mu sync.Mutex
	c  event.Config
	m  *glutil.Image
}

// TODO(crawshaw): It looks like we need a gl.RegisterInit feature.
// TODO(crawshaw): The gldebug mode needs to complain loudly when GL functions
//                 are called before init, because often they fail silently.
// TODO(nigeltao): no need to re-load the font on every config change (e.g.
//                 phone rotation between portrait and landscape).
func fpsInit() {
	b := font.Monospace()
	f, err := freetype.ParseFont(b)
	if err != nil {
		panic(err)
	}
	monofont.SetFont(f)
	monofont.SetSrc(image.Black)
	monofont.SetHinting(freetype.FullHinting)

	toPx := func(x geom.Pt) int { return int(math.Ceil(float64(geom.Pt(x).Px(fps.c.PixelsPerPt)))) }
	fps.m = glutil.NewImage(toPx(50), toPx(12))
	monofont.SetDst(fps.m.RGBA)
	monofont.SetClip(fps.m.Bounds())
	monofont.SetDPI(72 * float64(fps.c.PixelsPerPt))
	monofont.SetFontSize(12)
}

// DrawFPS draws the per second framerate in the bottom-left of the screen.
func DrawFPS(c event.Config) {
	fps.mu.Lock()
	if fps.c != c || fps.m == nil {
		fps.c = c
		fpsInit()
	}
	fps.mu.Unlock()

	now := time.Now()
	diff := now.Sub(lastDraw)
	str := fmt.Sprintf("%.0f FPS", float32(time.Second)/float32(diff))
	draw.Draw(fps.m.RGBA, fps.m.Rect, image.White, image.Point{}, draw.Src)

	ftpt12 := freetype.Pt(0, int(12*c.PixelsPerPt))
	if _, err := monofont.DrawString(str, ftpt12); err != nil {
		log.Printf("DrawFPS: %v", err)
		return
	}

	fps.m.Upload()
	fps.m.Draw(
		c,
		geom.Point{0, c.Height - 12},
		geom.Point{50, c.Height - 12},
		geom.Point{0, c.Height},
		fps.m.Bounds(),
	)

	lastDraw = now
}
