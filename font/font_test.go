package font

import (
	"testing"

	"code.google.com/p/freetype-go/freetype"
)

func TestLoadFonts(t *testing.T) {
	if _, err := freetype.ParseFont(Default()); err != nil {
		t.Fatalf("default font: %v", err)
	}
	if _, err := freetype.ParseFont(Monospace()); err != nil {
		t.Fatalf("monospace font: %v", err)
	}
}
