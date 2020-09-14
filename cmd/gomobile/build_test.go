// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

func TestRFC1034Label(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"a", "a"},
		{"123", "-23"},
		{"a.b.c", "a-b-c"},
		{"a-b", "a-b"},
		{"a:b", "a-b"},
		{"a?b", "a-b"},
		{"αβγ", "---"},
		{"💩", "--"},
		{"My App", "My-App"},
		{"...", ""},
		{".-.", "--"},
	}

	for _, tc := range tests {
		if got := rfc1034Label(tc.in); got != tc.want {
			t.Errorf("rfc1034Label(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAndroidPkgName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"a", "a"},
		{"a123", "a123"},
		{"a.b.c", "a_b_c"},
		{"a-b", "a_b"},
		{"a:b", "a_b"},
		{"a?b", "a_b"},
		{"αβγ", "go___"},
		{"💩", "go_"},
		{"My App", "My_App"},
		{"...", "go___"},
		{".-.", "go___"},
		{"abstract", "abstract_"},
		{"Abstract", "Abstract"},
		{"12345", "go12345"},
	}

	for _, tc := range tests {
		if got := androidPkgName(tc.in); got != tc.want {
			t.Errorf("len %d", len(tc.in))
			t.Errorf("androidPkgName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAndroidBuild(t *testing.T) {
	if runtime.GOOS == "android" {
		t.Skip("not available on Android")
	}
	buf := new(bytes.Buffer)
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
	}()
	xout = buf
	buildN = true
	buildX = true
	buildO = "basic.apk"
	buildTarget = "android/arm"
	gopath = filepath.ToSlash(filepath.SplitList(goEnv("GOPATH"))[0])
	if goos == "windows" {
		os.Setenv("HOMEDRIVE", "C:")
	}
	cmdBuild.flag.Parse([]string{"golang.org/x/mobile/example/basic"})
	oldTags := buildTags
	buildTags = []string{"tag1"}
	defer func() {
		buildTags = oldTags
	}()
	err := runBuild(cmdBuild)
	if err != nil {
		t.Log(buf.String())
		t.Fatal(err)
	}

	diff, err := diffOutput(buf.String(), androidBuildTmpl)
	if err != nil {
		t.Fatalf("computing diff failed: %v", err)
	}
	if diff != "" {
		t.Errorf("unexpected output:\n%s", diff)
	}
}

var androidBuildTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
mkdir -p $WORK/lib/armeabi-v7a
GOOS=android GOARCH=arm CC=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang CXX=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang++ CGO_ENABLED=1 GOARM=7 go build -tags tag1 -x -buildmode=c-shared -o $WORK/lib/armeabi-v7a/libbasic.so golang.org/x/mobile/example/basic
`))

func TestParseBuildTargetFlag(t *testing.T) {
	androidArchs := strings.Join(allArchs("android"), ",")
	iosArchs := strings.Join(allArchs("ios"), ",")

	tests := []struct {
		in        string
		wantErr   bool
		wantOS    string
		wantArchs string
	}{
		{"android", false, "android", androidArchs},
		{"android,android/arm", false, "android", androidArchs},
		{"android/arm", false, "android", "arm"},

		{"ios", false, "darwin", iosArchs},
		{"ios,ios/arm64", false, "darwin", iosArchs},
		{"ios/arm64", false, "darwin", "arm64"},
		{"ios/amd64", false, "darwin", "amd64"},

		{"", true, "", ""},
		{"linux", true, "", ""},
		{"android/x86", true, "", ""},
		{"android/arm5", true, "", ""},
		{"ios/mips", true, "", ""},
		{"android,ios", true, "", ""},
		{"ios,android", true, "", ""},
	}

	for _, tc := range tests {
		gotOS, gotArchs, err := parseBuildTarget(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("-target=%q; want error, got (%q, %q, nil)", tc.in, gotOS, gotArchs)
			}
			continue
		}
		if err != nil || gotOS != tc.wantOS || strings.Join(gotArchs, ",") != tc.wantArchs {
			t.Errorf("-target=%q; want (%v, [%v], nil), got (%q, %q, %v)",
				tc.in, tc.wantOS, tc.wantArchs, gotOS, gotArchs, err)
		}
	}
}

func TestRegexImportGolangXPackage(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantLen int
	}{
		{"ffffffff t golang.org/x/mobile", "golang.org/x/mobile", 2},
		{"ffffffff t github.com/example/repo/vendor/golang.org/x/mobile", "golang.org/x/mobile", 2},
		{"ffffffff t github.com/example/golang.org/x/mobile", "", 0},
		{"ffffffff t github.com/example/repo", "", 0},
		{"ffffffff t github.com/example/repo/vendor", "", 0},
	}

	for _, tc := range tests {
		res := nmRE.FindStringSubmatch(tc.in)
		if len(res) != tc.wantLen {
			t.Errorf("nmRE returned unexpected result for %q: want len(res) = %d, got %d",
				tc.in, tc.wantLen, len(res))
			continue
		}
		if tc.wantLen == 0 {
			continue
		}
		if res[1] != tc.want {
			t.Errorf("nmRE returned unexpected result. want (%v), got (%v)",
				tc.want, res[1])
		}
	}
}

func TestBuildWithGoModules(t *testing.T) {
	if runtime.GOOS == "android" {
		t.Skipf("gomobile are not available on %s", runtime.GOOS)
	}

	dir, err := ioutil.TempDir("", "gomobile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if out, err := exec.Command("go", "build", "-o="+dir, "golang.org/x/mobile/cmd/gomobile").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, string(out))
	}
	path := dir
	if p := os.Getenv("PATH"); p != "" {
		path += string(filepath.ListSeparator) + p
	}

	for _, target := range []string{"android", "ios"} {
		t.Run(target, func(t *testing.T) {
			if target == "ios" {
				t.Skip("gomobile-build doesn't work for iOS. see https://golang.org/issue/32963")
			}

			switch target {
			case "android":
				androidHome := os.Getenv("ANDROID_HOME")
				if androidHome == "" {
					t.Skip("ANDROID_HOME not found, skipping bind")
				}
				if _, err := androidAPIPath(); err != nil {
					t.Skip("No android API platform found in $ANDROID_HOME, skipping bind")
				}
			case "ios":
				if !xcodeAvailable() {
					t.Skip("Xcode is missing")
				}
			}

			var out string
			switch target {
			case "android":
				out = filepath.Join(dir, "basic.apk")
			case "ios":
				out = filepath.Join(dir, "Basic.app")
			}

			tests := []struct {
				Name string
				Path string
				Dir  string
			}{
				{
					Name: "Absolute Path",
					Path: "golang.org/x/mobile/example/basic",
				},
				{
					Name: "Relative Path",
					Path: "./example/basic",
					Dir:  filepath.Join("..", ".."),
				},
			}

			for _, tc := range tests {
				tc := tc
				t.Run(tc.Name, func(t *testing.T) {
					args := []string{"build", "-target=" + target, "-o=" + out}
					if target == "ios" {
						args = append(args, "-bundleid=org.golang.gomobiletest")
					}
					args = append(args, tc.Path)
					cmd := exec.Command(filepath.Join(dir, "gomobile"), args...)
					cmd.Env = append(os.Environ(), "PATH="+path, "GO111MODULE=on")
					cmd.Dir = tc.Dir
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Errorf("gomobile build failed: %v\n%s", err, string(out))
					}
				})
			}
		})
	}
}
