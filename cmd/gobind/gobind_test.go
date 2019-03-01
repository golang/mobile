// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var tests = []struct {
	name string
	lang string
	pkg  string
	goos string
}{
	{"ObjC-Testpkg", "objc", "golang.org/x/mobile/bind/testdata/testpkg", ""},
	{"Java-Testpkg", "java", "golang.org/x/mobile/bind/testdata/testpkg", ""},
	{"Go-Testpkg", "go", "golang.org/x/mobile/bind/testdata/testpkg", ""},
	{"Java-Javapkg", "java", "golang.org/x/mobile/bind/testdata/testpkg/javapkg", "android"},
	{"Go-Javapkg", "go", "golang.org/x/mobile/bind/testdata/testpkg/javapkg", "android"},
	{"Go-Javapkg", "go,java,objc", "golang.org/x/mobile/bind/testdata/cgopkg", "android"},
}

var gobindBin string

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	bin, err := ioutil.TempFile("", "*.exe")
	if err != nil {
		log.Fatal(err)
	}
	bin.Close()
	gobindBin = bin.Name()
	defer os.Remove(gobindBin)
	if out, err := exec.Command("go", "build", "-o", gobindBin, "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
		log.Fatalf("gobind build failed: %v: %s", err, out)
	}
	return m.Run()
}

func runGobind(lang, pkg, goos string) error {
	cmd := exec.Command(gobindBin, "-lang", lang, pkg)
	if goos != "" {
		cmd.Env = append(os.Environ(), "GOOS="+goos)
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gobind -lang %s %s failed: %v: %s", lang, pkg, err, out)
	}
	return nil
}

func TestGobind(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := runGobind(test.lang, test.pkg, test.goos); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDocs(t *testing.T) {
	// Create a fake package for doc.go
	tmpdir, err := ioutil.TempDir("", "gobind-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	docPkg := filepath.Join(tmpdir, "src", "doctest")
	if err := os.MkdirAll(docPkg, 0700); err != nil {
		t.Fatal(err)
	}
	const docsrc = `
package doctest

// This is a comment.
type Struct struct{
}`
	if err := ioutil.WriteFile(filepath.Join(docPkg, "doc.go"), []byte(docsrc), 0700); err != nil {
		t.Fatal(err)
	}

	const comment = "This is a comment."
	for _, lang := range []string{"java", "objc"} {
		cmd := exec.Command(gobindBin, "-lang", lang, "doctest")
		cmd.Env = append(os.Environ(), "GOROOT="+tmpdir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("gobind -lang %s failed: %v: %s", lang, err, out)
			continue
		}
		if bytes.Index(out, []byte(comment)) == -1 {
			t.Errorf("gobind output for language %s did not contain the comment %q", lang, comment)
		}
	}
}

func BenchmarkGobind(b *testing.B) {
	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := runGobind(test.lang, test.pkg, test.goos); err != nil {
					b.Error(err)
				}
			}
		})
	}
}
