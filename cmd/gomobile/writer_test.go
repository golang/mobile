// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestWriter(t *testing.T) {
	block, _ := pem.Decode([]byte(debugCert))
	if block == nil {
		t.Fatal("no cert")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	f, err := ioutil.TempFile("", "testapk-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())
	apkPath := f.Name() + ".apk"

	f, err = os.Create(apkPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(apkPath)

	apkw := NewWriter(f, privKey)

	w, err := apkw.Create("AndroidManifest.xml")
	if err != nil {
		t.Fatalf("could not create AndroidManifest.xml: %v", err)
	}
	if _, err := w.Write([]byte(androidManifest)); err != nil {
		t.Errorf("could not write AndroidManifest.xml: %v", err)
	}

	w, err = apkw.Create("assets/hello_world.txt")
	if err != nil {
		t.Fatalf("could not create assets/hello_world.txt: %v", err)
	}
	if _, err := w.Write([]byte("Hello, 世界")); err != nil {
		t.Errorf("could not write assets/hello_world.txt: %v", err)
	}

	if err := apkw.Close(); err != nil {
		t.Fatal(err)
	}

	if exec.Command("which", "aapt").Run() != nil {
		t.Skip("command aapt not found, skipping")
	}

	out, err := exec.Command("aapt", "list", "-a", apkPath).CombinedOutput()
	aaptGot := string(out)
	if err != nil {
		t.Logf("aapt:\n%s", aaptGot)
		t.Fatalf("aapt failed: %v", err)
	}

	if aaptGot != aaptWant {
		t.Errorf("unexpected output from aapt")
		d, err := diff(aaptWant, aaptGot)
		if err != nil {
			t.Errorf("diff failed: %v", err)
		} else {
			t.Logf("%s", d)
		}
	}
}

const aaptWant = `AndroidManifest.xml
assets/hello_world.txt
META-INF/MANIFEST.MF
META-INF/CERT.SF
META-INF/CERT.RSA

Resource table:
Package Groups (0)

Android manifest:
N: android=http://schemas.android.com/apk/res/android
  E: manifest (line=2)
    A: package="org.golang.fakeapp" (Raw: "org.golang.fakeapp")
    A: android:versionCode(0x0101021b)=(type 0x10)0x1
    A: android:versionName(0x0101021c)="1.0" (Raw: "1.0")
    E: uses-sdk (line=8)
      A: android:minSdkVersion(0x0101020c)=(type 0x10)0x9
    E: application (line=9)
      A: android:label(0x01010001)="FakeApp" (Raw: "FakeApp")
      A: android:hasCode(0x0101000c)=(type 0x12)0x0
      A: android:debuggable(0x0101000f)=(type 0x12)0xffffffff
      E: activity (line=10)
        A: android:name(0x01010003)="android.app.NativeActivity" (Raw: "android.app.NativeActivity")
        A: android:label(0x01010001)="FakeApp" (Raw: "FakeApp")
        A: android:configChanges(0x0101001f)=(type 0x11)0xa0
        E: intent-filter (line=14)
          E: action (line=15)
            A: android:name(0x01010003)="android.intent.action.MAIN" (Raw: "android.intent.action.MAIN")
          E: category (line=16)
            A: android:name(0x01010003)="android.intent.category.LAUNCHER" (Raw: "android.intent.category.LAUNCHER")
`

const androidManifest = `
<manifest
	xmlns:android="http://schemas.android.com/apk/res/android"
	package="org.golang.fakeapp"
	android:versionCode="1"
	android:versionName="1.0">

	<uses-sdk android:minSdkVersion="9" />
	<application android:label="FakeApp" android:hasCode="false" android:debuggable="true">
		<activity android:name="android.app.NativeActivity"
			android:label="FakeApp"
			android:configChanges="orientation|keyboardHidden">

			<intent-filter>
				<action android:name="android.intent.action.MAIN" />
				<category android:name="android.intent.category.LAUNCHER" />
			</intent-filter>
		</activity>
	</application>
</manifest>
`

func diff(s1, s2 string) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "apk-writer-diff")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "apk-writer-diff")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	io.WriteString(f1, s1)
	io.WriteString(f2, s2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}
