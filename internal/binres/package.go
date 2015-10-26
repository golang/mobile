// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binres

type TableHeader struct {
	chunkHeader
	packageCount uint32
}

func (hdr *TableHeader) UnmarshalBinary(bin []byte) error {
	if err := (&hdr.chunkHeader).UnmarshalBinary(bin); err != nil {
		return err
	}
	hdr.packageCount = btou32(bin[8:])
	return nil
}

func (hdr TableHeader) MarshalBinary() ([]byte, error) {
	bin := make([]byte, 12)
	b, err := hdr.chunkHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(bin, b)
	putu32(b[8:], hdr.packageCount)
	return bin, nil
}

// TODO next up: package chunk
// https://justanapplication.wordpress.com/2011/09/16/
type Table struct {
	TableHeader
	pool *Pool
	pkg  *Package
}

func (tbl *Table) UnmarshalBinary(bin []byte) error {
	buf := bin
	if err := (&tbl.TableHeader).UnmarshalBinary(buf); err != nil {
		return err
	}
	buf = buf[tbl.headerByteSize:]
	tbl.pool = new(Pool)
	if err := tbl.pool.UnmarshalBinary(buf); err != nil {
		return err
	}
	buf = buf[tbl.pool.size():]
	tbl.pkg = new(Package)
	if err := tbl.pkg.UnmarshalBinary(buf); err != nil {
		return err
	}
	return nil
}

type Package struct {
	chunkHeader

	// If this is a base package, its ID.  Package IDs start
	// at 1 (corresponding to the value of the package bits in a
	// resource identifier).  0 means this is not a base package.
	id uint32

	// name of package, zero terminated
	name [128]uint16

	// Offset to a ResStringPool_header defining the resource
	// type symbol table.  If zero, this package is inheriting from
	// another base package (overriding specific values in it).
	typeStrings uint32

	// Last index into typeStrings that is for public use by others.
	lastPublicType uint32

	// Offset to a ResStringPool_header defining the resource
	// key symbol table.  If zero, this package is inheriting from
	// another base package (overriding specific values in it).
	keyStrings uint32

	// Last index into keyStrings that is for public use by others.
	lastPublicKey uint32

	typeIdOffset uint32
}

func (pkg *Package) UnmarshalBinary(bin []byte) error {
	if err := (&pkg.chunkHeader).UnmarshalBinary(bin); err != nil {
		return err
	}
	pkg.id = btou32(bin[8:])
	for i := range pkg.name {
		pkg.name[i] = btou16(bin[12+i*2:])
	}
	pkg.typeStrings = btou32(bin[140:])
	pkg.lastPublicType = btou32(bin[144:])
	pkg.keyStrings = btou32(bin[148:])
	pkg.lastPublicKey = btou32(bin[152:])
	pkg.typeIdOffset = btou32(bin[156:])
	return nil
}
