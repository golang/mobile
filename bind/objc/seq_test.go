// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objc

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var destination = flag.String("device", "platform=iOS Simulator,name=iPhone 6s Plus", "Specify the -destination flag to xcodebuild")

// TestObjcSeqTest runs ObjC test SeqTest.m.
// This requires the xcode command lines tools.
func TestObjcSeqTest(t *testing.T) {
	runTest(t, "golang.org/x/mobile/bind/objc/testpkg", "xcodetest", "SeqTest.m")
}

func runTest(t *testing.T, pkgName, project, testfile string) {
	if _, err := run("which xcodebuild"); err != nil {
		t.Skip("command xcodebuild not found, skipping")
	}
	if _, err := run("which gomobile"); err != nil {
		_, err := run("go install golang.org/x/mobile/cmd/gomobile")
		if err != nil {
			t.Skip("gomobile not available, skipping")
		}
	}

	// TODO(hyangah): gomobile init if necessary.

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed pwd: %v", err)
	}
	tmpdir, err := ioutil.TempDir("", "bind-objc-seq-test-")
	if err != nil {
		t.Fatalf("failed to prepare temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	t.Logf("tmpdir = %s", tmpdir)

	if buf, err := exec.Command("cp", "-a", project, tmpdir).CombinedOutput(); err != nil {
		t.Logf("%s", buf)
		t.Fatalf("failed to copy %s to tmp dir: %v", project, err)
	}

	if err := cp(filepath.Join(tmpdir, testfile), testfile); err != nil {
		t.Fatalf("failed to copy %s: %v", testfile, err)
	}

	if err := os.Chdir(filepath.Join(tmpdir, project)); err != nil {
		t.Fatalf("failed chdir: %v", err)
	}
	defer os.Chdir(cwd)

	buf, err := run("gomobile bind -target=ios " + pkgName)
	if err != nil {
		t.Logf("%s", buf)
		t.Fatalf("failed to run gomobile bind: %v", err)
	}

	cmd := exec.Command("xcodebuild", "test", "-scheme", "xcodetest", "-destination", *destination)
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Logf("%s", buf)
		t.Errorf("failed to run xcodebuild: %v", err)
	}
}

func run(cmd string) ([]byte, error) {
	c := strings.Split(cmd, " ")
	return exec.Command(c[0], c[1:]...).CombinedOutput()
}

func cp(dst, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %v", err)
	}
	defer r.Close()
	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to open destination: %v", err)
	}
	_, err = io.Copy(w, r)
	cerr := w.Close()
	if err != nil {
		return err
	}
	return cerr
}
