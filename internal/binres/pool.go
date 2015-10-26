// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binres

import (
	"fmt"
	"unicode/utf16"
)

const (
	SortedFlag uint32 = 1 << 0
	UTF8Flag          = 1 << 8
)

// PoolRef is the i'th string in a pool.
type PoolRef uint32

// Pool has the following structure marshalled:
//
// 	binChunkHeader
//  StringCount  uint32 // Number of strings in this pool
//  StyleCount   uint32 // Number of style spans in pool
//  Flags        uint32 // SortedFlag, UTF8Flag
//  StringsStart uint32 // Index of string data from header
//  StylesStart  uint32 // Index of style data from header
//
//  StringIndices []uint32 // starting at zero
//
//  // UTF16 entries are concatenations of the following:
//  // [2]byte uint16 string length, exclusive
//  // [2]byte [optional] low word if high bit of length was set
//  // [n]byte data
//  // [2]byte 0x0000 terminator
//  Strings []uint16
type Pool struct {
	chunkHeader

	strings []string
	styles  []*Span
	flags   uint32 // SortedFlag, UTF8Flag
}

func (pl *Pool) IsSorted() bool { return pl.flags&SortedFlag == SortedFlag }
func (pl *Pool) IsUTF8() bool   { return pl.flags&UTF8Flag == UTF8Flag }

func (pl *Pool) UnmarshalBinary(bin []byte) error {
	if err := (&pl.chunkHeader).UnmarshalBinary(bin); err != nil {
		return err
	}
	if pl.typ != ResStringPool {
		return fmt.Errorf("have type %s, want %s", pl.typ, ResStringPool)
	}

	nstrings := btou32(bin[8:])
	pl.strings = make([]string, nstrings)

	nstyles := btou32(bin[12:])
	pl.styles = make([]*Span, nstyles)

	pl.flags = btou32(bin[16:])

	offstrings := btou32(bin[20:])
	offstyle := btou32(bin[24:])

	hdrlen := 28 // header size is always 28
	if pl.IsUTF8() {
		for i := range pl.strings {
			ii := btou32(bin[hdrlen+i*4:]) // index of string index
			r := int(offstrings + ii)      // read index of string

			// for char and byte sizes below, if leading bit set,
			// treat first as 7-bit high word and the next byte as low word.

			// get char size of string
			cn := uint8(bin[r])
			rcn := int(cn)
			r++
			if cn&(1<<7) != 0 {
				cn0 := int(cn ^ (1 << 7)) // high word
				cn1 := int(bin[r])        // low word
				rcn = cn0*256 + cn1
				r++
			}

			// get byte size of string
			// TODO(d) i've seen at least one case in android.jar resource table that has only
			// highbit set, effectively making 7-bit highword zero. The reason for this is currently
			// unknown but would make tests that unmarshal-marshal to match bytes impossible to pass.
			// The values noted were high-word: 0 (after highbit unset), low-word: 141
			// I don't recall character count but was around ?78?
			// The case here may very well be that the size is treated like int8 triggering the use of
			// two bytes to store size even though the implementation uses uint8.
			n := uint8(bin[r])
			r++
			rn := int(n)
			if n&(1<<7) != 0 {
				n0 := int(n ^ (1 << 7)) // high word
				n1 := int(bin[r])       // low word
				rn = n0*(1<<8) + n1
				r++
			}

			//
			data := bin[r : r+rn]
			if x := uint8(bin[r+rn]); x != 0 {
				return fmt.Errorf("expected zero terminator, got %v for byte size %v char len %v", x, rn, rcn)
			}
			pl.strings[i] = string(data)
		}
	} else {
		for i := range pl.strings {
			ii := btou32(bin[hdrlen+i*4:]) // index of string index
			r := int(offstrings + ii)      // read index of string
			n := btou16(bin[r:])           // string length
			rn := int(n)
			r += 2

			if n&(1<<15) != 0 { // TODO this is untested
				n0 := int(n ^ (1 << 15)) // high word
				n1 := int(btou16(bin[r:]))
				rn = n0*(1<<16) + n1
				r += 2
			}

			data := make([]uint16, int(rn))
			for i := range data {
				data[i] = btou16(bin[r+(2*i):])
			}

			r += int(n * 2)
			if x := btou16(bin[r:]); x != 0 {
				return fmt.Errorf("expected zero terminator, got 0x%04X\n%s", x)
			}

			pl.strings[i] = string(utf16.Decode(data))
		}
	}

	// TODO
	_ = offstyle
	// styii := hdrlen + int(nstrings*4)
	// for i := range pl.styles {
	// ii := btou32(bin[styii+i*4:])
	// r := int(offstyle + ii)
	// spn := new(binSpan)
	// spn.UnmarshalBinary(bin[r:])
	// pl.styles[i] = spn
	// }

	return nil
}

func (pl *Pool) MarshalBinary() ([]byte, error) {
	if pl.IsUTF8() {
		return nil, fmt.Errorf("encode utf8 not supported")
	}

	var (
		hdrlen = 28
		// indices of string indices
		iis    = make([]uint32, len(pl.strings))
		iislen = len(iis) * 4
		// utf16 encoded strings concatenated together
		strs []uint16
	)
	for i, x := range pl.strings {
		if len(x)>>16 > 0 {
			panic(fmt.Errorf("string lengths over 1<<15 not yet supported, got len %d", len(x)))
		}
		p := utf16.Encode([]rune(x))
		if len(p) == 0 {
			strs = append(strs, 0x0000, 0x0000)
		} else {
			strs = append(strs, uint16(len(p))) // string length (implicitly includes zero terminator to follow)
			strs = append(strs, p...)
			strs = append(strs, 0) // zero terminated
		}
		// indices start at zero
		if i+1 != len(iis) {
			iis[i+1] = uint32(len(strs) * 2) // utf16 byte index
		}
	}

	// check strings is 4-byte aligned, pad with zeros if not.
	for x := (len(strs) * 2) % 4; x != 0; x -= 2 {
		strs = append(strs, 0x0000)
	}

	strslen := len(strs) * 2

	hdr := chunkHeader{
		typ:            ResStringPool,
		headerByteSize: 28,
		byteSize:       uint32(28 + iislen + strslen),
	}

	bin := make([]byte, hdr.byteSize)

	hdrbin, err := hdr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(bin, hdrbin)

	putu32(bin[8:], uint32(len(pl.strings)))
	putu32(bin[12:], uint32(len(pl.styles)))
	putu32(bin[16:], pl.flags)
	putu32(bin[20:], uint32(hdrlen+iislen))
	putu32(bin[24:], 0) // index of styles start, is 0 when styles length is 0

	buf := bin[28:]
	for _, x := range iis {
		putu32(buf, x)
		buf = buf[4:]
	}
	for _, x := range strs {
		putu16(buf, x)
		buf = buf[2:]
	}

	if len(buf) != 0 {
		panic(fmt.Errorf("failed to fill allocated buffer, %v bytes left over", len(buf)))
	}

	return bin, nil
}

type Span struct {
	name PoolRef

	firstChar, lastChar uint32
}

func (spn *Span) UnmarshalBinary(bin []byte) error {
	const end = 0xFFFFFFFF
	spn.name = PoolRef(btou32(bin))
	if spn.name == end {
		return nil
	}
	spn.firstChar = btou32(bin[4:])
	spn.lastChar = btou32(bin[8:])
	return nil
}

// Map contains a uint32 slice mapping strings in the string
// pool back to resource identifiers. The i'th element of the slice
// is also the same i'th element of the string pool.
type Map struct {
	chunkHeader
	rs []uint32
}

func (m *Map) UnmarshalBinary(bin []byte) error {
	(&m.chunkHeader).UnmarshalBinary(bin)
	buf := bin[m.headerByteSize:m.byteSize]
	m.rs = make([]uint32, len(buf)/4)
	for i := range m.rs {
		m.rs[i] = btou32(buf[i*4:])
	}
	return nil
}

func (m *Map) MarshalBinary() ([]byte, error) {
	bin := make([]byte, 8+len(m.rs)*4)
	b, err := m.chunkHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(bin, b)
	for i, r := range m.rs {
		putu32(bin[8+i*4:], r)
	}
	return bin, nil
}
