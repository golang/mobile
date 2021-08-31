// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var dump = flag.Bool("dump", false, "dump junk.bin binary output")

var (
	origSortPool = sortPool
	origSortAttr = sortAttr
)

func TestBinaryXML(t *testing.T) {
	sortPool, sortAttr = sortToMatchTest, sortAttrToMatchTest
	defer func() { sortPool, sortAttr = origSortPool, origSortAttr }()

	bin, err := binaryXML(bytes.NewBufferString(fmt.Sprintf(input, "")))
	if err != nil {
		t.Fatal(err)
	}
	if *dump {
		if err := ioutil.WriteFile("junk.bin", bin, 0660); err != nil {
			t.Fatal(err)
		}
	}

	if exec.Command("which", "aapt").Run() != nil {
		t.Skip("command aapt not found, skipping")
	}

	apiPath, err := androidAPIPath()
	if err != nil {
		t.Fatal(err)
	}
	androidJar := filepath.Join(apiPath, "android.jar")

	tmpdir, err := ioutil.TempDir("", "gomobile-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	zf, err := zw.Create("AndroidManifest.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zf.Write(bin); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	mblapk := filepath.Join(tmpdir, "mbl.apk")
	if err := ioutil.WriteFile(mblapk, buf.Bytes(), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := exec.Command("aapt", "d", "xmltree", mblapk, "AndroidManifest.xml").Output()
	if err != nil {
		t.Fatal(err)
	}

	manifest := filepath.Join(tmpdir, "AndroidManifest.xml")
	inp := fmt.Sprintf(input, "<uses-sdk android:minSdkVersion=\"15\" />")
	if err := ioutil.WriteFile(manifest, []byte(inp), 0755); err != nil {
		t.Fatal(err)
	}

	sdkapk := filepath.Join(tmpdir, "sdk.apk")
	if _, err := exec.Command("aapt", "p", "-M", manifest, "-I", androidJar, "-F", sdkapk).Output(); err != nil {
		t.Fatal(err)
	}

	sdkout, err := exec.Command("aapt", "d", "xmltree", sdkapk, "AndroidManifest.xml").Output()
	if err != nil {
		t.Fatal(err)
	}
	if *dump {
		b, err := ioutil.ReadFile(sdkapk)
		if err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile("junk.sdk.apk", b, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// manifests contain platformBuildVersionCode and platformBuildVersionName
	// which are not present in gomobile output.
	var fo [][]byte
	for _, line := range bytes.Split(sdkout, []byte{'\n'}) {
		if bytes.Contains(line, []byte("platformBuildVersionCode")) || bytes.Contains(line, []byte("platformBuildVersionName")) {
			continue
		}
		fo = append(fo, line)
	}
	want := bytes.Join(fo, []byte{'\n'})

	if !bytes.Equal(want, got) {
		t.Fatalf("output does not match\nWANT\n%s\nGOT\n%s\n", want, got)
	}
}

// The output of the Android encoder seems to be arbitrary. So for testing,
// we sort the string pool order to match the output we have seen.
func sortToMatchTest(p *binStringPool) {
	var names = []string{
		"versionCode",
		"versionName",
		"minSdkVersion",
		"theme",
		"label",
		"hasCode",
		"debuggable",
		"name",
		"screenOrientation",
		"configChanges",
		"value",
		"android",
		"http://schemas.android.com/apk/res/android",
		"",
		"package",
		"manifest",
		"com.zentus.balloon",
		"1.0",
		"uses-sdk",
		"application",
		"Balloon世界",
		"activity",
		"android.app.NativeActivity",
		"Balloon",
		"meta-data",
		"android.app.lib_name",
		"balloon",
		"intent-filter",
		"\there is some text\n",
		"action",
		"android.intent.action.MAIN",
		"category",
		"android.intent.category.LAUNCHER",
	}

	s := make([]*bstring, 0)
	m := make(map[string]*bstring)

	for _, str := range names {
		bstr := p.m[str]
		if bstr == nil {
			log.Printf("missing %q", str)
			continue
		}
		bstr.ind = uint32(len(s))
		s = append(s, bstr)
		m[str] = bstr
		delete(p.m, str)
	}
	// add unexpected strings
	for str, bstr := range p.m {
		log.Printf("unexpected %q", str)
		bstr.ind = uint32(len(s))
		s = append(s, bstr)
	}
	p.s = s
	p.m = m
}

func sortAttrToMatchTest(e *binStartElement, p *binStringPool) {
	order := []string{
		"versionCode",
		"versionName",
		"versionPackage",

		"theme",
		"label",
		"name",
		"screenOrientation",
		"configChanges",
	}
	ordered := make([]*binAttr, len(order))

outer:
	for i, n := range order {
		for j, a := range e.attr {
			if a != nil && a.name.str == n {
				ordered[i] = a
				e.attr[j] = nil
				continue outer
			}
		}
	}
	var attr []*binAttr
	for _, a := range ordered {
		if a != nil {
			attr = append(attr, a)
		}
	}
	for _, a := range e.attr {
		if a != nil {
			attr = append(attr, a)
		}
	}
	e.attr = attr
}

const input = `<?xml version="1.0" encoding="utf-8"?>
<!--
Copyright 2014 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
-->
<manifest
	xmlns:android="http://schemas.android.com/apk/res/android"
	package="com.zentus.balloon"
	android:versionCode="1"
	android:versionName="1.0">

	%s
	<application android:label="Balloon世界" android:hasCode="false" android:debuggable="true">
	<activity android:name="android.app.NativeActivity"
		android:theme="@android:style/Theme.NoTitleBar.Fullscreen"
		android:label="Balloon"
		android:screenOrientation="portrait"
		android:configChanges="orientation|keyboardHidden">
		<meta-data android:name="android.app.lib_name" android:value="balloon" />
		<intent-filter>
			here is some text
			<action android:name="android.intent.action.MAIN" />
			<category android:name="android.intent.category.LAUNCHER" />
		</intent-filter>
	</activity>
	</application>
</manifest>`
