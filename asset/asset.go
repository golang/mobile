// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// Package asset provides access to application-bundled assets.
//
// On Android, assets are accessed via android.content.res.AssetManager.
// These files are stored in the assets/ directory of the app. Any raw asset
// can be accessed by its direct relative name. For example assets/img.png
// can be opened with Open("img.png").
//
// On iOS an asset is a resource stored in the application bundle.
// Resources can be loaded using the same relative paths.
//
// For consistency when debugging on a desktop, assets are read from a
// directory named assets under the current working directory.
package asset

import "io"

// Open opens a named asset.
//
// Errors are of type *os.PathError.
//
// This must not be called from init when used in android apps.
func Open(name string) (File, error) {
	return openAsset(name)
}

// File is an open asset.
type File interface {
	io.ReadSeeker
	io.Closer
}
