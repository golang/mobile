package util

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/geom"
)

func drawText(img *image.RGBA, x, y int, s string) {
	col := color.RGBA{0, 0, 0, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(s)
}

// Text draws text to screen
type Text struct {
	sz     size.Event
	images *glutil.Images
	m      *glutil.Image
	// TODO: store *gl.Context
}

// NewText creates a Text tied to the current GL context.
func NewText(images *glutil.Images) *Text {
	return &Text{
		images: images,
	}
}

// Draw draws text at the x, y coordinates at the scaleX, scaleY specified by user
// Draw accepts strings with newline characters
func (t *Text) Draw(sz size.Event, x, y int, scaleX, scaleY geom.Pt, s string) {
	if sz.WidthPx == 0 && sz.HeightPx == 0 {
		return
	} else if t.sz != sz {
		t.sz = sz
		if t.m != nil {
			t.m.Release()
		}
		t.m = t.images.NewImage(sz.WidthPx, sz.HeightPx)
	}

	// split string by newline
	lines := strings.Split(s, "\n")

	// draw each string on a seperate line
	for i, v := range lines {
		drawText(t.m.RGBA, int(geom.Pt(x)/scaleX), int(geom.Pt(y)/scaleY)+i*10, v)
	}

	// copy img data to GL device
	t.m.Upload()

	t.m.Draw(
		sz,
		geom.Point{0, 0},                    // topLeft
		geom.Point{sz.WidthPt * scaleX, 0},  // topRight
		geom.Point{0, sz.HeightPt * scaleY}, // bottomLeft
		t.m.RGBA.Bounds(),
	)
}

func (t *Text) Clear() {
	if t.m != nil {
		draw.Draw(t.m.RGBA, t.m.RGBA.Bounds(), image.Transparent, image.Point{}, draw.Src)
	}
}

func (t *Text) Release() {
	if t.m != nil {
		t.m.Release()
		t.m = nil
		t.images = nil
	}
}
