// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var tests = []struct {
	name string
	lang string
	pkg  string
	goos string
}{
	{"ObjC-Testpkg", "objc", "golang.org/x/mobile/bind/testpkg", ""},
	{"Java-Testpkg", "java", "golang.org/x/mobile/bind/testpkg", ""},
	{"Go-Testpkg", "go", "golang.org/x/mobile/bind/testpkg", ""},
	{"Java-Javapkg", "java", "golang.org/x/mobile/bind/testpkg/javapkg", "android"},
	{"Go-Javapkg", "go", "golang.org/x/mobile/bind/testpkg/javapkg", "android"},
}

func installGobind() error {
	if out, err := exec.Command("go", "install", "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
		return fmt.Errorf("gobind install failed: %v: %s", err, out)
	}
	return nil
}

func runGobind(lang, pkg, goos string) error {
	cmd := exec.Command("gobind", "-lang", lang, pkg)
	if goos != "" {
		cmd.Env = append(os.Environ(), "GOOS="+goos)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gobind -lang %s %s failed: %v: %s", lang, pkg, err, out)
	}
	return nil
}

func TestGobind(t *testing.T) {
	if err := installGobind(); err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := runGobind(test.lang, test.pkg, test.goos); err != nil {
				t.Error(err)
			}
		})
	}
}

func BenchmarkGobind(b *testing.B) {
	if err := installGobind(); err != nil {
		b.Fatal(err)
	}
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
