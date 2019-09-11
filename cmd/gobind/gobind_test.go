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
	"runtime"
	"testing"
)

var tests = []struct {
	name string
	lang string
	pkg  string
	goos string
	// reverse is true if the test needs to generate reverse bindings using
	// external tools such as javap.
	reverse bool
}{
	{"ObjC-Testpkg", "objc", "golang.org/x/mobile/bind/testdata/testpkg", "", false},
	{"Java-Testpkg", "java", "golang.org/x/mobile/bind/testdata/testpkg", "", false},
	{"Go-Testpkg", "go", "golang.org/x/mobile/bind/testdata/testpkg", "", false},
	{"Java-Javapkg", "java", "golang.org/x/mobile/bind/testdata/testpkg/javapkg", "android", true},
	{"Go-Javapkg", "go", "golang.org/x/mobile/bind/testdata/testpkg/javapkg", "android", true},
	{"Go-Cgopkg", "go,java,objc", "golang.org/x/mobile/bind/testdata/cgopkg", "android", false},
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
	defer os.Remove(bin.Name())
	if runtime.GOOS != "android" {
		if out, err := exec.Command(goBin(), "build", "-o", bin.Name(), "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
			log.Fatalf("gobind build failed: %v: %s", err, out)
		}
		gobindBin = bin.Name()
	}
	return m.Run()
}

func runGobind(t testing.TB, lang, pkg, goos string) error {
	if gobindBin == "" {
		t.Skipf("gobind is not available on %s", runtime.GOOS)
	}
	cmd := exec.Command(gobindBin, "-lang", lang, pkg)
	if goos != "" {
		cmd.Env = append(os.Environ(), "GOOS="+goos)
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gobind -lang %s %s failed: %v: %s", lang, pkg, err, out)
	}
	return nil
}

func TestGobind(t *testing.T) {
	_, javapErr := exec.LookPath("javap")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.reverse && javapErr != nil {
				t.Skip("reverse bind test requires javap which is not available")
			}
			if err := runGobind(t, test.lang, test.pkg, test.goos); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDocs(t *testing.T) {
	if gobindBin == "" {
		t.Skipf("gobind is not available on %s", runtime.GOOS)
	}
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

	gopath, err := exec.Command(goBin(), "env", "GOPATH").Output()
	if err != nil {
		t.Fatal(err)
	}

	const comment = "This is a comment."
	for _, lang := range []string{"java", "objc"} {
		cmd := exec.Command(gobindBin, "-lang", lang, "doctest")
		// TODO(hajimehoshi): Enable this test with Go modules.
		cmd.Env = append(os.Environ(), "GOPATH="+tmpdir+string(filepath.ListSeparator)+string(gopath), "GO111MODULE=off")
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
	_, javapErr := exec.LookPath("javap")
	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			if test.reverse && javapErr != nil {
				b.Skip("reverse bind test requires javap which is not available")
			}
			for i := 0; i < b.N; i++ {
				if err := runGobind(b, test.lang, test.pkg, test.goos); err != nil {
					b.Error(err)
				}
			}
		})
	}
}
