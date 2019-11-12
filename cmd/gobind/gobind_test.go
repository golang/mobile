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
	"runtime"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages/packagestest"
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
	{
		name: "ObjC-Testpkg",
		lang: "objc",
		pkg:  "golang.org/x/mobile/bind/testdata/testpkg",
	},
	{
		name: "Java-Testpkg",
		lang: "java",
		pkg:  "golang.org/x/mobile/bind/testdata/testpkg",
	},
	{
		name: "Go-Testpkg",
		lang: "go",
		pkg:  "golang.org/x/mobile/bind/testdata/testpkg",
	},
	{
		name:    "Java-Javapkg",
		lang:    "java",
		pkg:     "golang.org/x/mobile/bind/testdata/testpkg/javapkg",
		goos:    "android",
		reverse: true,
	},
	{
		name:    "Go-Javapkg",
		lang:    "go",
		pkg:     "golang.org/x/mobile/bind/testdata/testpkg/javapkg",
		goos:    "android",
		reverse: true,
	},
	{
		name: "Go-Cgopkg",
		lang: "go,java,objc",
		pkg:  "golang.org/x/mobile/bind/testdata/cgopkg",
		goos: "android",
	},
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

func runGobind(t testing.TB, lang, pkg, goos string, exported *packagestest.Exported) error {
	if gobindBin == "" {
		t.Skipf("gobind is not available on %s", runtime.GOOS)
	}
	cmd := exec.Command(gobindBin, "-lang", lang, pkg)
	cmd.Dir = exported.Config.Dir
	cmd.Env = exported.Config.Env
	if goos != "" {
		cmd.Env = append(cmd.Env, "GOOS="+goos)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		var cmd string
		for _, env := range exported.Config.Env {
			if strings.HasPrefix(env, "GO111MODULE=") {
				cmd = env + " "
				break
			}
		}
		cmd += fmt.Sprintf("gobind -lang %s %s", lang, pkg)
		return fmt.Errorf("%s failed: %v: %s", cmd, err, out)
	}
	return nil
}

func TestGobind(t *testing.T) { packagestest.TestAll(t, testGobind) }
func testGobind(t *testing.T, exporter packagestest.Exporter) {
	_, javapErr := exec.LookPath("javap")
	exported := packagestest.Export(t, exporter, []packagestest.Module{{
		Name:  "golang.org/x/mobile",
		Files: packagestest.MustCopyFileTree("../.."),
	}})
	defer exported.Cleanup()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if exporter == packagestest.Modules && test.reverse {
				t.Skip("reverse binding does't work with Go modules")
			}
			if test.reverse && javapErr != nil {
				t.Skip("reverse bind test requires javap which is not available")
			}
			if err := runGobind(t, test.lang, test.pkg, test.goos, exported); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDocs(t *testing.T) { packagestest.TestAll(t, testDocs) }
func testDocs(t *testing.T, exporter packagestest.Exporter) {
	if gobindBin == "" {
		t.Skipf("gobind is not available on %s", runtime.GOOS)
	}

	const docsrc = `
package doctest

// This is a comment.
type Struct struct{
}`

	exported := packagestest.Export(t, exporter, []packagestest.Module{
		{
			Name: "example.com/doctest",
			Files: map[string]interface{}{
				"doc.go": docsrc,
			},
		},
		{
			// gobind requires golang.org/x/mobile to generate code for reverse bindings.
			Name:  "golang.org/x/mobile",
			Files: packagestest.MustCopyFileTree("../.."),
		},
	})
	defer exported.Cleanup()

	const comment = "This is a comment."
	for _, lang := range []string{"java", "objc"} {
		cmd := exec.Command(gobindBin, "-lang", lang, "example.com/doctest")
		cmd.Dir = exported.Config.Dir
		cmd.Env = exported.Config.Env
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
	packagestest.BenchmarkAll(b, benchmarkGobind)
}

func benchmarkGobind(b *testing.B, exporter packagestest.Exporter) {
	_, javapErr := exec.LookPath("javap")
	exported := packagestest.Export(b, exporter, []packagestest.Module{{
		Name:  "golang.org/x/mobile",
		Files: packagestest.MustCopyFileTree("../.."),
	}})
	defer exported.Cleanup()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			if exporter == packagestest.Modules && test.reverse {
				b.Skip("reverse binding does't work with Go modules")
			}
			if test.reverse && javapErr != nil {
				b.Skip("reverse bind test requires javap which is not available")
			}
			for i := 0; i < b.N; i++ {
				if err := runGobind(b, test.lang, test.pkg, test.goos, exported); err != nil {
					b.Error(err)
				}
			}
		})
	}
}
