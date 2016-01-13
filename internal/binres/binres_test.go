// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binres

import (
	"bytes"
	"encoding"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func printrecurse(t *testing.T, pl *Pool, el *Element, ws string) {
	for _, attr := range el.attrs {
		ns := ""
		if attr.NS != math.MaxUint32 {
			ns = pl.strings[int(attr.NS)]
			nss := strings.Split(ns, "/")
			ns = nss[len(nss)-1]
		}

		val := ""
		if attr.RawValue != math.MaxUint32 {
			val = pl.strings[int(attr.RawValue)]
		} else {
			switch attr.TypedValue.Type {
			case DataIntDec:
				val = fmt.Sprintf("%v", attr.TypedValue.Value)
			case DataIntBool:
				val = fmt.Sprintf("%v", attr.TypedValue.Value == 1)
			default:
				val = fmt.Sprintf("0x%08X", attr.TypedValue.Value)
			}
		}
		dt := attr.TypedValue.Type

		t.Logf("%s|attr:ns(%v) name(%s) val(%s) valtyp(%s)\n", ws, ns, pl.strings[int(attr.Name)], val, dt)
	}
	t.Log()
	for _, e := range el.Children {
		printrecurse(t, pl, e, ws+"        ")
	}
}

func TestBootstrap(t *testing.T) {
	bin, err := ioutil.ReadFile("testdata/bootstrap.bin")
	if err != nil {
		log.Fatal(err)
	}

	checkMarshal := func(res encoding.BinaryMarshaler, bsize int) {
		b, err := res.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		idx := debugIndices[res]
		a := bin[idx : idx+bsize]
		if !bytes.Equal(a, b) {
			x, y := len(a), len(b)
			if x != y {
				t.Errorf("%v: %T: byte length does not match, have %v, want %v", idx, res, y, x)
			}
			if x > y {
				x, y = y, x
			}
			mismatch := false
			for i := 0; i < x; i++ {
				if mismatch = a[i] != b[i]; mismatch {
					t.Errorf("%v: %T: first byte mismatch at %v of %v", idx, res, i, bsize)
					break
				}
			}
			if mismatch {
				// print out a reasonable amount of data to help identify issues
				truncate := x > 1300
				if truncate {
					x = 1300
				}
				t.Log("      HAVE               WANT")
				for i := 0; i < x; i += 4 {
					he, we := 4, 4
					if i+he >= x {
						he = x - i
					}
					if i+we >= y {
						we = y - i
					}
					t.Logf("%3v | % X        % X\n", i, b[i:i+he], a[i:i+we])
				}
				if truncate {
					t.Log("... output truncated.")
				}
			}
		}
	}

	bxml := new(XML)
	if err := bxml.UnmarshalBinary(bin); err != nil {
		t.Fatal(err)
	}

	for i, x := range bxml.Pool.strings {
		t.Logf("Pool(%v): %q\n", i, x)
	}

	for _, e := range bxml.Children {
		printrecurse(t, bxml.Pool, e, "")
	}

	checkMarshal(&bxml.chunkHeader, int(bxml.headerByteSize))
	checkMarshal(bxml.Pool, bxml.Pool.size())
	checkMarshal(bxml.Map, bxml.Map.size())
	checkMarshal(bxml.Namespace, bxml.Namespace.size())

	for el := range bxml.iterElements() {
		checkMarshal(el, el.size())
		checkMarshal(el.end, el.end.size())
	}

	checkMarshal(bxml.Namespace.end, bxml.Namespace.end.size())
	checkMarshal(bxml, bxml.size())
}

func TestEncode(t *testing.T) {
	f, err := os.Open("testdata/bootstrap.xml")
	if err != nil {
		t.Fatal(err)
	}

	bx, err := UnmarshalXML(f)
	if err != nil {
		t.Fatal(err)
	}

	//
	bin, err := ioutil.ReadFile("testdata/bootstrap.bin")
	if err != nil {
		log.Fatal(err)
	}
	bxml := new(XML)
	if err := bxml.UnmarshalBinary(bin); err != nil {
		t.Fatal(err)
	}

	if err := compareStrings(t, bxml.Pool.strings, bx.Pool.strings); err != nil {
		t.Fatal(err)
	}

	if err := compareUint32s(t, rtou(bxml.Map.rs), rtou(bx.Map.rs)); err != nil {
		t.Fatal(err)
	}
}

func rtou(a []TableRef) (b []uint32) {
	for _, x := range a {
		b = append(b, uint32(x))
	}
	return
}

func compareUint32s(t *testing.T, a, b []uint32) error {
	var err error
	if len(a) != len(b) {
		err = fmt.Errorf("lengths do not match")
	}

	n := len(a)
	if n < len(b) {
		n = len(b)
	}

	t.Log("a.Map.rs    b.Map.rs")
	for i := 0; i < n; i++ {
		var c, d string
		if i < len(a) {
			c = fmt.Sprintf("%0#8x ", a[i])
		} else {
			c = "__________"
		}
		if i < len(b) {
			d = fmt.Sprintf("%0#8x ", b[i])
		} else {
			d = "__________"
		}
		if err == nil && c != d {
			err = fmt.Errorf("has missing/incorrect values")
		}
		t.Log(c, d)
	}

	return err
}

func compareStrings(t *testing.T, a, b []string) error {
	var err error
	if len(a) != len(b) {
		err = fmt.Errorf("lengths do not match")
	}

	for i, x := range a {
		v := "__"
		for j, y := range b {
			if x == y {
				v = fmt.Sprintf("%2v", j)
				break
			}
		}
		if err == nil && v == "__" {
			if !strings.HasPrefix(x, "4.0.") {
				// as of the time of this writing, the current version of build tools being targetted
				// reports 4.0.4-1406430. Previously, this was 4.0.3. This number is likely still due
				// to change so only report error if 4.x incremented.
				//
				// TODO this check has the potential to hide real errors but can be fixed once more
				// of the xml document is unmarshalled and XML can be queried to assure this is related
				// to platformBuildVersionName.
				err = fmt.Errorf("has missing/incorrect values")
			}
		}
		t.Logf("Pool(%2v, %s) %q", i, v, x)
	}

	contains := func(xs []string, a string) bool {
		for _, x := range xs {
			if x == a {
				return true
			}
		}
		return false
	}

	if err != nil {
		t.Log()
		t.Logf("## only in var a")
		for i, x := range a {
			if !contains(b, x) {
				t.Logf("Pool(%2v) %q", i, x)
			}
		}

		t.Log()
		t.Logf("## only in var b")
		for i, x := range b {
			if !contains(a, x) {
				t.Logf("Pool(%2v) %q", i, x)
			}
		}
	}

	return err
}

func TestOpenTable(t *testing.T) {
	sdkdir := os.Getenv("ANDROID_HOME")
	if sdkdir == "" {
		t.Skip("ANDROID_HOME env var not set")
	}
	tbl, err := OpenTable()
	if err != nil {
		t.Fatal(err)
	}
	if len(tbl.pkgs) == 0 {
		t.Fatal("failed to decode any resource packages")
	}

	pkg := tbl.pkgs[0]

	t.Log("package name:", pkg.name)

	for i, x := range pkg.typePool.strings {
		t.Logf("typePool[i=%v]: %s\n", i, x)
	}

	for i, spec := range pkg.specs {
		t.Logf("spec[i=%v]: %v %q\n", i, spec.id, pkg.typePool.strings[spec.id-1])
		for j, typ := range spec.types {
			t.Logf("\ttype[i=%v]: %v\n", j, typ.id)
			for k, nt := range typ.entries {
				if nt == nil { // NoEntry
					continue
				}
				t.Logf("\t\tentry[i=%v]: %v %q\n", k, nt.key, pkg.keyPool.strings[nt.key])
				if k > 5 {
					t.Logf("\t\t... truncating output")
					break
				}
			}
		}
	}
}

func TestTableRefByName(t *testing.T) {
	sdkdir := os.Getenv("ANDROID_HOME")
	if sdkdir == "" {
		t.Skip("ANDROID_HOME env var not set")
	}
	tbl, err := OpenTable()
	if err != nil {
		t.Fatal(err)
	}
	if len(tbl.pkgs) == 0 {
		t.Fatal("failed to decode any resource packages")
	}

	ref, err := tbl.RefByName("@android:style/Theme.NoTitleBar.Fullscreen")
	if err != nil {
		t.Fatal(err)
	}

	if want := uint32(0x01030007); uint32(ref) != want {
		t.Fatalf("RefByName does not match expected result, have %0#8x, want %0#8x", ref, want)
	}
}

func testPackResources(t *testing.T) {
	packResources()
	f, err := os.Open(filepath.Join("data", "packed.arsc.gz"))
	if err != nil {
		t.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("packed.arsc.gz %vKB", fi.Size()/1024)
}

func TestTableMarshal(t *testing.T) {
	tbl, err := OpenSDKTable()
	if err != nil {
		t.Fatal(err)
	}

	bin, err := tbl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	xtbl := new(Table)
	if err := xtbl.UnmarshalBinary(bin); err != nil {
		t.Fatal(err)
	}

	if len(tbl.pool.strings) != len(xtbl.pool.strings) {
		t.Fatal("tbl.pool lengths don't match")
	}
	if len(tbl.pkgs) != len(xtbl.pkgs) {
		t.Fatal("tbl.pkgs lengths don't match")
	}

	pkg, xpkg := tbl.pkgs[0], xtbl.pkgs[0]
	if err := compareStrings(t, pkg.typePool.strings, xpkg.typePool.strings); err != nil {
		t.Fatal(err)
	}
	if err := compareStrings(t, pkg.keyPool.strings, xpkg.keyPool.strings); err != nil {
		t.Fatal(err)
	}

	if len(pkg.specs) != len(xpkg.specs) {
		t.Fatal("pkg.specs lengths don't match")
	}

	for i, spec := range pkg.specs {
		xspec := xpkg.specs[i]
		if spec.id != xspec.id {
			t.Fatal("spec.id doesn't match")
		}
		if spec.entryCount != xspec.entryCount {
			t.Fatal("spec.entryCount doesn't match")
		}
		if len(spec.entries) != len(xspec.entries) {
			t.Fatal("spec.entries lengths don't match")
		}
		for j, mask := range spec.entries {
			xmask := xspec.entries[j]
			if mask != xmask {
				t.Fatal("entry mask doesn't match")
			}
		}
		if len(spec.types) != len(xspec.types) {
			t.Fatal("spec.types length don't match")
		}
		for j, typ := range spec.types {
			xtyp := xspec.types[j]
			if typ.id != xtyp.id {
				t.Fatal("typ.id doesn't match")
			}
			if typ.entryCount != xtyp.entryCount {
				t.Fatal("typ.entryCount doesn't match")
			}
			if typ.entriesStart != xtyp.entriesStart {
				t.Fatal("typ.entriesStart doesn't match")
			}
			if len(typ.indices) != len(xtyp.indices) {
				t.Fatal("typ.indices length don't match")
			}
			for k, index := range typ.indices {
				xindex := xtyp.indices[k]
				if index != xindex {
					t.Errorf("type index doesn't match at %v, have %v, want %v", k, xindex, index)
				}
			}
			if len(typ.entries) != len(xtyp.entries) {
				t.Fatal("typ.entries lengths don't match")
			}
			for k, nt := range typ.entries {
				xnt := xtyp.entries[k]
				if nt == nil {
					if xnt != nil {
						t.Fatal("nt is nil but xnt is not")
					}
					continue
				}
				if nt.size != xnt.size {
					t.Fatal("entry.size doesn't match")
				}
				if nt.flags != xnt.flags {
					t.Fatal("entry.flags don't match")
				}
				if nt.key != xnt.key {
					t.Fatal("entry.key doesn't match")
				}

				if nt.parent != xnt.parent {
					t.Fatal("entry.parent doesn't match")
				}
				if nt.count != xnt.count {
					t.Fatal("entry.count doesn't match")
				}
				for l, val := range nt.values {
					xval := xnt.values[l]
					if val.name != xval.name {
						t.Fatal("value.name doesn't match")
					}
				}
			}
		}
	}
}

func BenchmarkTableRefByName(b *testing.B) {
	sdkdir := os.Getenv("ANDROID_HOME")
	if sdkdir == "" {
		b.Fatal("ANDROID_HOME env var not set")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tbl, err := OpenTable()
		if err != nil {
			b.Fatal(err)
		}
		_, err = tbl.RefByName("@android:style/Theme.NoTitleBar.Fullscreen")
		if err != nil {
			b.Fatal(err)
		}
	}
}
