// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages/packagestest"
)

func TestMain(m *testing.M) {
	// To avoid recompiling the gobind command (and to support compiler options
	// like -race and -coverage), allow the test binary itself to re-exec itself
	// as the gobind command by setting an environment variable.
	if os.Getenv("GOBIND_TEST_IS_GOBIND") != "" {
		main()
		os.Exit(0)
	}
	os.Setenv("GOBIND_TEST_IS_GOBIND", "1")

	os.Exit(m.Run())
}

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

func mustHaveBindTestdata(t testing.TB) {
	switch runtime.GOOS {
	case "android", "ios":
		t.Skipf("skipping: test cannot access ../../bind/testdata on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func gobindBin(t testing.TB) string {
	switch runtime.GOOS {
	case "js", "ios":
		t.Skipf("skipping: cannot exec subprocess on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func runGobind(t testing.TB, lang, pkg, goos string, exported *packagestest.Exported) error {
	cmd := exec.Command(gobindBin(t), "-lang", lang, pkg)
	cmd.Dir = exported.Config.Dir
	cmd.Env = exported.Config.Env
	if goos != "" {
		// Add CGO_ENABLED=1 explicitly since Cgo is disabled when GOOS is different from host OS.
		cmd.Env = append(cmd.Env, "GOOS="+goos, "CGO_ENABLED=1")
	}
	stderr := new(strings.Builder)
	cmd.Stderr = stderr
	stdout := new(strings.Builder)
	cmd.Stdout = stdout
	err := cmd.Run()
	if testing.Verbose() && stdout.Len() > 0 {
		t.Logf("stdout (%v):\n%s", cmd, stderr)
	}
	if stderr.Len() > 0 {
		t.Logf("stderr (%v):\n%s", cmd, stderr)
	}
	if err != nil {
		return fmt.Errorf("%v: %w", cmd, err)
	}
	return nil
}

func TestGobind(t *testing.T) { packagestest.TestAll(t, testGobind) }
func testGobind(t *testing.T, exporter packagestest.Exporter) {
	mustHaveBindTestdata(t)

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
	mustHaveBindTestdata(t)

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
		cmd := exec.Command(gobindBin(t), "-lang", lang, "example.com/doctest")
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
