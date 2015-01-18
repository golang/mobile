// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !android

package font

import (
	"io/ioutil"
	"os"
)

func readFont(fontname string) ([]byte, error) {
	// Common places where fonts are located in linux distributions
	var common_font_dirs []string = []string{
		// Ubuntu
		"/usr/share/fonts/truetype/droid/",
		// Archlinux
		"/usr/share/fonts/TTF/",
	}
	var err error
	// Try in the different directories
	for _, dir := range common_font_dirs {
		path := dir + fontname
		if _, err = os.Stat(path); err == nil {
			return ioutil.ReadFile(path)
		}
	}
	return []byte{}, err
}

func buildDefault() ([]byte, error) {
	return readFont("DroidSans.ttf")
}

func buildMonospace() ([]byte, error) {
	return readFont("DroidSansMono.ttf")
}
