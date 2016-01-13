// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -output binres_string.go -type ResType,DataType

// Package binres implements encoding and decoding of android binary resources.
//
// Binary resource structs support unmarshalling the binary output of aapt.
// Implementations of marshalling for each struct must produce the exact input
// sent to unmarshalling. This allows tests to validate each struct representation
// of the binary format as follows:
//
//  * unmarshal the output of aapt
//  * marshal the struct representation
//  * perform byte-to-byte comparison with aapt output per chunk header and body
//
// This process should strive to make structs idiomatic to make parsing xml text
// into structs trivial.
//
// Once the struct representation is validated, tests for parsing xml text
// into structs can become self-referential as the following holds true:
//
//  * the unmarshalled input of aapt output is the only valid target
//  * the unmarshalled input of xml text may be compared to the unmarshalled
//    input of aapt output to identify errors, e.g. text-trims, wrong flags, etc
//
// This provides validation, byte-for-byte, for producing binary xml resources.
//
// It should be made clear that unmarshalling binary resources is currently only
// in scope for proving that the BinaryMarshaler works correctly. Any other use
// is currently out of scope.
//
// A simple view of binary xml document structure:
//
//  XML
//    Pool
//    Map
//    Namespace
//    [...node]
//
// Additional resources:
// https://android.googlesource.com/platform/frameworks/base/+/master/include/androidfw/ResourceTypes.h
// https://justanapplication.wordpress.com/2011/09/13/ (a series of articles, increment date)
package binres

import (
	"encoding"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"unicode"
)

func errWrongType(have ResType, want ...ResType) error {
	return fmt.Errorf("wrong resource type %s, want one of %v", have, want)
}

type ResType uint16

func (t ResType) IsSupported() bool {
	// explicit for clarity
	return t == ResStringPool || t == ResXML ||
		t == ResXMLStartNamespace || t == ResXMLEndNamespace ||
		t == ResXMLStartElement || t == ResXMLEndElement ||
		t == ResXMLCharData ||
		t == ResXMLResourceMap ||
		t == ResTable || t == ResTablePackage ||
		t == ResTableTypeSpec || t == ResTableType
}

// explicitly defined for clarity and resolvability with apt source
const (
	ResNull       ResType = 0x0000
	ResStringPool ResType = 0x0001
	ResTable      ResType = 0x0002
	ResXML        ResType = 0x0003

	ResXMLStartNamespace ResType = 0x0100
	ResXMLEndNamespace   ResType = 0x0101
	ResXMLStartElement   ResType = 0x0102
	ResXMLEndElement     ResType = 0x0103
	ResXMLCharData       ResType = 0x0104

	ResXMLResourceMap ResType = 0x0180

	ResTablePackage  ResType = 0x0200
	ResTableType     ResType = 0x0201
	ResTableTypeSpec ResType = 0x0202
	ResTableLibrary  ResType = 0x0203
)

var (
	btou16 = binary.LittleEndian.Uint16
	btou32 = binary.LittleEndian.Uint32
	putu16 = binary.LittleEndian.PutUint16
	putu32 = binary.LittleEndian.PutUint32
)

// unmarshaler wraps BinaryUnmarshaler to provide byte size of decoded chunks.
type unmarshaler interface {
	encoding.BinaryUnmarshaler

	// size returns the byte size unmarshalled after a call to
	// UnmarshalBinary, or otherwise zero.
	size() int
}

// chunkHeader appears at the front of every data chunk in a resource.
type chunkHeader struct {
	// Type of data that follows this header.
	typ ResType

	// Advance slice index by this value to find its associated data, if any.
	headerByteSize uint16

	// This is the header size plus the size of any data associated with the chunk.
	// Advance slice index by this value to completely skip its contents, including
	// any child chunks. If this value is the same as headerByteSize, there is
	// no data associated with the chunk.
	byteSize uint32
}

// size implements unmarshaler.
func (hdr chunkHeader) size() int { return int(hdr.byteSize) }

func (hdr *chunkHeader) UnmarshalBinary(bin []byte) error {
	hdr.typ = ResType(btou16(bin))
	if !hdr.typ.IsSupported() {
		return fmt.Errorf("%s not supported", hdr.typ)
	}
	hdr.headerByteSize = btou16(bin[2:])
	hdr.byteSize = btou32(bin[4:])
	if len(bin) < int(hdr.byteSize) {
		return fmt.Errorf("too few bytes to unmarshal chunk body, have %v, need at-least %v", len(bin), hdr.byteSize)
	}
	return nil
}

func (hdr chunkHeader) MarshalBinary() ([]byte, error) {
	if !hdr.typ.IsSupported() {
		return nil, fmt.Errorf("%s not supported", hdr.typ)
	}

	bin := make([]byte, 8)
	putu16(bin, uint16(hdr.typ))
	putu16(bin[2:], hdr.headerByteSize)
	putu32(bin[4:], hdr.byteSize)
	return bin, nil
}

type XML struct {
	chunkHeader

	Pool *Pool
	Map  *Map

	Namespace *Namespace
	Children  []*Element

	// tmp field used when unmarshalling binary
	stack []*Element
}

// TODO this is used strictly for querying in tests and dependent on
// current XML.UnmarshalBinary implementation. Look into moving directly
// into tests.
var debugIndices = make(map[encoding.BinaryMarshaler]int)

const (
	androidSchema = "http://schemas.android.com/apk/res/android"
	toolsSchema   = "http://schemas.android.com/tools"
)

type xattr struct {
	xml.Attr
	bool
}

type xnode struct {
	name  xml.Name
	attrs []xattr
	cdata []string
}

func UnmarshalXML(r io.Reader) (*XML, error) {
	tbl, err := OpenTable()
	if err != nil {
		return nil, err
	}

	var nodes []xnode

	dec := xml.NewDecoder(r)
	bx := new(XML)

	for {
		tkn, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		tkn = xml.CopyToken(tkn)

		switch tkn := tkn.(type) {
		case xml.StartElement:
			nodes = append(nodes, xnode{name: tkn.Name})

			for _, attr := range tkn.Attr {
				if attr.Name.Space == toolsSchema || (attr.Name.Space == "xmlns" && attr.Name.Local == "tools") {
					continue // TODO can tbl be queried for schemas to determine validity instead?
				}

				att := xattr{attr, false}
				if attr.Name.Space == "" || attr.Name.Space == "xmlns" {
					att.bool = true
				} else if attr.Name.Space == androidSchema {
					// get type spec and value data type
					ref, err := tbl.RefByName("attr/" + attr.Name.Local)
					if err != nil {
						return nil, err
					}
					nt, err := ref.Resolve(tbl)
					if err != nil {
						return nil, err
					}
					if len(nt.values) == 0 {
						// TODO don't know if this can happen
						panic("TODO don't know how to handle empty values slice")
					}

					if len(nt.values) == 1 {
						val := nt.values[0]
						if val.data.Type != DataIntDec {
							panic("TODO only know how to handle DataIntDec type here")
						}
						t := DataType(val.data.Value)
						switch t {
						case DataString, DataAttribute, DataType(0x3e):
							// TODO identify 0x3e, in bootstrap.xml this is the native lib name
							// TODO why DataAttribute? confirm details of usage
							att.bool = true
						default:
							// TODO resolve other data types
							// fmt.Printf("unhandled data type %0#4x: %s\n", uint32(t), t)
						}
					} else { // attribute value must resolve to one of the values here
						// TODO resolve reference values and assure they match list of values here
					}
				}
				nodes[len(nodes)-1].attrs = append(nodes[len(nodes)-1].attrs, att)
			}
		case xml.CharData:
			if s := poolTrim(string(tkn)); s != "" {
				nodes[len(nodes)-1].cdata = append(nodes[len(nodes)-1].cdata, s)
			}
		case xml.EndElement, xml.Comment, xml.ProcInst:
			// discard
		default:
			panic(fmt.Errorf("unhandled token type: %T %+v", tkn, tkn))
		}
	}

	bvc := xml.Attr{
		Name: xml.Name{
			Space: "",
			Local: "platformBuildVersionCode",
		},
		Value: "15",
	}
	bvn := xml.Attr{
		Name: xml.Name{
			Space: "",
			Local: "platformBuildVersionName",
		},
		Value: "4.0.3",
	}
	nodes[0].attrs = append(nodes[0].attrs, xattr{bvc, true}, xattr{bvn, true})

	// pools appear to be sorted as follows:
	// * attribute names prefixed with android:
	// * "android", [schema-url], [empty-string]
	// * for each node:
	//   * attribute names with no prefix
	//   * node name
	//   * attribute value if data type of name is DataString, DataAttribute, or 0x3e (an unknown)
	bx.Pool = new(Pool)

	for _, node := range nodes {
		for _, attr := range node.attrs {
			if attr.Name.Space == androidSchema {
				bx.Pool.strings = append(bx.Pool.strings, attr.Name.Local)
			}
		}
	}

	// TODO encoding/xml does not enforce namespace prefix and manifest encoding in aapt
	// appears to ignore all other prefixes. Inserting this manually is not strictly correct
	// for the general case, but the effort to do otherwise currently offers nothing.
	bx.Pool.strings = append(bx.Pool.strings, "android", androidSchema)

	// there always appears to be an empty string located after schema, even if one is
	// not present in manifest.
	bx.Pool.strings = append(bx.Pool.strings, "")

	for _, node := range nodes {
		for _, attr := range node.attrs {
			if attr.Name.Space == "" {
				bx.Pool.strings = append(bx.Pool.strings, attr.Name.Local)
			}
		}
		bx.Pool.strings = append(bx.Pool.strings, node.name.Local)
		for _, attr := range node.attrs {
			if attr.bool {
				bx.Pool.strings = append(bx.Pool.strings, attr.Value)
			}
		}
		for _, x := range node.cdata {
			bx.Pool.strings = append(bx.Pool.strings, x)
		}
	}

	// do not eliminate duplicates until the entire slice has been composed.
	// consider <activity android:label="label" .../>
	// all attribute names come first followed by values; in such a case, the value "label"
	// would be a reference to the same "android:label" in the string pool which will occur
	// within the beginning of the pool where other attr names are located.
	bx.Pool.strings = asSet(bx.Pool.strings)

	// TODO consider cases of multiple declarations of the same attr name that should return error
	// before ever reaching this point.
	bx.Map = new(Map)
	for _, s := range bx.Pool.strings {
		ref, err := tbl.RefByName("attr/" + s)
		if err != nil {
			break // break after first non-ref as all strings after are also non-refs.
		}
		bx.Map.rs = append(bx.Map.rs, ref)
	}

	return bx, nil
}

// asSet returns a set from a slice of strings.
func asSet(xs []string) []string {
	m := make(map[string]bool)
	fo := xs[:0]
	for _, x := range xs {
		if !m[x] {
			m[x] = true
			fo = append(fo, x)
		}
	}
	return fo
}

// poolTrim trims all but immediately surrounding space.
// \n\t\tfoobar\n\t\t becomes \tfoobar\n
func poolTrim(s string) string {
	var start, end int
	for i, r := range s {
		if !unicode.IsSpace(r) {
			if i != 0 {
				start = i - 1 // preserve preceding space
			}
			break
		}
	}

	for i := len(s) - 1; i >= 0; i-- {
		r := rune(s[i])
		if !unicode.IsSpace(r) {
			if i != len(s)-1 {
				end = i + 2
			}
			break
		}
	}

	if start == 0 && end == 0 {
		return "" // every char was a space
	}

	return s[start:end]
}

func (bx *XML) UnmarshalBinary(bin []byte) error {
	buf := bin
	if err := (&bx.chunkHeader).UnmarshalBinary(bin); err != nil {
		return err
	}
	buf = buf[8:]

	// TODO this is tracked strictly for querying in tests; look into moving this
	// functionality directly into tests if possible.
	debugIndex := 8

	for len(buf) > 0 {
		t := ResType(btou16(buf))
		k, err := bx.kind(t)
		if err != nil {
			return err
		}
		if err := k.UnmarshalBinary(buf); err != nil {
			return err
		}
		debugIndices[k.(encoding.BinaryMarshaler)] = debugIndex
		debugIndex += int(k.size())
		buf = buf[k.size():]
	}
	return nil
}

func (bx *XML) kind(t ResType) (unmarshaler, error) {
	switch t {
	case ResStringPool:
		if bx.Pool != nil {
			return nil, fmt.Errorf("pool already exists")
		}
		bx.Pool = new(Pool)
		return bx.Pool, nil
	case ResXMLResourceMap:
		if bx.Map != nil {
			return nil, fmt.Errorf("resource map already exists")
		}
		bx.Map = new(Map)
		return bx.Map, nil
	case ResXMLStartNamespace:
		if bx.Namespace != nil {
			return nil, fmt.Errorf("namespace start already exists")
		}
		bx.Namespace = new(Namespace)
		return bx.Namespace, nil
	case ResXMLEndNamespace:
		if bx.Namespace.end != nil {
			return nil, fmt.Errorf("namespace end already exists")
		}
		bx.Namespace.end = new(Namespace)
		return bx.Namespace.end, nil
	case ResXMLStartElement:
		el := new(Element)
		if len(bx.stack) == 0 {
			bx.Children = append(bx.Children, el)
		} else {
			n := len(bx.stack)
			var p *Element
			p, bx.stack = bx.stack[n-1], bx.stack[:n-1]
			p.Children = append(p.Children, el)
			bx.stack = append(bx.stack, p)
		}
		bx.stack = append(bx.stack, el)
		return el, nil
	case ResXMLEndElement:
		n := len(bx.stack)
		var el *Element
		el, bx.stack = bx.stack[n-1], bx.stack[:n-1]
		if el.end != nil {
			return nil, fmt.Errorf("element end already exists")
		}
		el.end = new(ElementEnd)
		return el.end, nil
	case ResXMLCharData: // TODO assure correctness
		cdt := new(CharData)
		el := bx.stack[len(bx.stack)-1]
		if el.head == nil {
			el.head = cdt
		} else if el.tail == nil {
			el.tail = cdt
		} else {
			return nil, fmt.Errorf("element head and tail already contain chardata")
		}
		return cdt, nil
	default:
		return nil, fmt.Errorf("unexpected type %s", t)
	}
}

func (bx *XML) MarshalBinary() ([]byte, error) {
	var (
		bin, b []byte
		err    error
	)
	b, err = bx.chunkHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bin = append(bin, b...)

	b, err = bx.Pool.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bin = append(bin, b...)

	b, err = bx.Map.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bin = append(bin, b...)

	b, err = bx.Namespace.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bin = append(bin, b...)

	for _, child := range bx.Children {
		if err := marshalRecurse(child, &bin); err != nil {
			return nil, err
		}
	}

	b, err = bx.Namespace.end.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bin = append(bin, b...)

	return bin, nil
}

func marshalRecurse(el *Element, bin *[]byte) error {
	b, err := el.MarshalBinary()
	if err != nil {
		return err
	}
	*bin = append(*bin, b...)

	if el.head != nil {
		b, err := el.head.MarshalBinary()
		if err != nil {
			return err
		}
		*bin = append(*bin, b...)
	}

	for _, child := range el.Children {
		if err := marshalRecurse(child, bin); err != nil {
			return err
		}
	}

	b, err = el.end.MarshalBinary()
	if err != nil {
		return err
	}
	*bin = append(*bin, b...)

	return nil
}

func (bx *XML) iterElements() <-chan *Element {
	ch := make(chan *Element, 1)
	go func() {
		for _, el := range bx.Children {
			iterElementsRecurse(el, ch)
		}
		close(ch)
	}()
	return ch
}

func iterElementsRecurse(el *Element, ch chan *Element) {
	ch <- el
	for _, e := range el.Children {
		iterElementsRecurse(e, ch)
	}
}
