// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	"golang.org/x/mobile/internal/sdkpath"
)

func TestBindAndroid(t *testing.T) {
	platform, err := sdkpath.AndroidAPIPath(minAndroidAPI)
	if err != nil {
		t.Skip("No compatible Android API platform found, skipping bind")
	}
	// platform is a path like "/path/to/Android/sdk/platforms/android-32"
	components := strings.Split(platform, string(filepath.Separator))
	if len(components) < 2 {
		t.Fatalf("API path is too short: %s", platform)
	}
	components = components[len(components)-2:]
	platformRel := filepath.Join("$ANDROID_HOME", components[0], components[1])

	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		buildO = ""
		buildTarget = ""
		bindJavaPkg = ""
	}()
	buildN = true
	buildX = true
	buildO = "asset.aar"
	buildTarget = "android/arm"

	tests := []struct {
		javaPkg string
	}{
		{
			// Empty javaPkg
		},
		{
			javaPkg: "com.example.foo",
		},
	}
	for _, tc := range tests {
		bindJavaPkg = tc.javaPkg

		buf := new(bytes.Buffer)
		xout = buf
		gopath = filepath.SplitList(goEnv("GOPATH"))[0]
		if goos == "windows" {
			os.Setenv("HOMEDRIVE", "C:")
		}
		cmdBind.flag.Parse([]string{"golang.org/x/mobile/asset"})
		err := runBind(cmdBind)
		if err != nil {
			t.Log(buf.String())
			t.Fatal(err)
		}
		got := filepath.ToSlash(buf.String())

		output, err := defaultOutputData("")
		if err != nil {
			t.Fatal(err)
		}
		data := struct {
			outputData
			AndroidPlatform string
			JavaPkg         string
		}{
			outputData:      output,
			AndroidPlatform: platformRel,
			JavaPkg:         tc.javaPkg,
		}

		wantBuf := new(bytes.Buffer)
		if err := bindAndroidTmpl.Execute(wantBuf, data); err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}

		diff, err := diff(got, wantBuf.String())
		if err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}
		if diff != "" {
			t.Errorf("%+v: unexpected output:\n%s", tc, diff)
		}
	}
}

func TestBindApple(t *testing.T) {
	if !xcodeAvailable() {
		t.Skip("Xcode is missing")
	}
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		buildO = ""
		buildTarget = ""
		bindPrefix = ""
	}()
	buildN = true
	buildX = true
	buildO = "Asset.xcframework"
	buildTarget = "ios/arm64"

	tests := []struct {
		prefix string
		out    string
	}{
		{
			// empty prefix
		},
		{
			prefix: "Foo",
		},
		{
			out: "Abcde.xcframework",
		},
	}
	for _, tc := range tests {
		bindPrefix = tc.prefix
		if tc.out != "" {
			buildO = tc.out
		}

		buf := new(bytes.Buffer)
		xout = buf
		gopath = filepath.SplitList(goEnv("GOPATH"))[0]
		if goos == "windows" {
			os.Setenv("HOMEDRIVE", "C:")
		}
		cmdBind.flag.Parse([]string{"golang.org/x/mobile/asset"})
		if err := runBind(cmdBind); err != nil {
			t.Log(buf.String())
			t.Fatal(err)
		}
		got := filepath.ToSlash(buf.String())

		output, err := defaultOutputData("")
		if err != nil {
			t.Fatal(err)
		}

		data := struct {
			outputData
			Output string
			Prefix string
		}{
			outputData: output,
			Output:     buildO[:len(buildO)-len(".xcframework")],
			Prefix:     tc.prefix,
		}

		wantBuf := new(bytes.Buffer)
		if err := bindAppleTmpl.Execute(wantBuf, data); err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}

		diff, err := diff(got, wantBuf.String())
		if err != nil {
			t.Errorf("%+v: computing diff failed: %v", tc, err)
			continue
		}
		if diff != "" {
			t.Errorf("%+v: unexpected output:\n%s", tc, diff)
		}
	}
}

var bindAndroidTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
GOOS=android CGO_ENABLED=1 gobind -lang=go,java -outdir=$WORK{{if .JavaPkg}} -javapkg={{.JavaPkg}}{{end}} golang.org/x/mobile/asset
mkdir -p $WORK/src-android-arm
PWD=$WORK/src-android-arm GOMODCACHE=$GOPATH/pkg/mod GOOS=android GOARCH=arm CC=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang CXX=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang++ CGO_ENABLED=1 GOARM=7 GOPATH=$WORK:$GOPATH go mod tidy
PWD=$WORK/src-android-arm GOMODCACHE=$GOPATH/pkg/mod GOOS=android GOARCH=arm CC=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang CXX=$NDK_PATH/toolchains/llvm/prebuilt/{{.NDKARCH}}/bin/armv7a-linux-androideabi16-clang++ CGO_ENABLED=1 GOARM=7 GOPATH=$WORK:$GOPATH go build -x -buildmode=c-shared -o=$WORK/android/src/main/jniLibs/armeabi-v7a/libgojni.so ./gobind
PWD=$WORK/java javac -d $WORK/javac-output -source 1.8 -target 1.8 -bootclasspath {{.AndroidPlatform}}/android.jar *.java
jar c -C $WORK/javac-output .
`))

var bindAppleTmpl = template.Must(template.New("output").Parse(`GOMOBILE={{.GOPATH}}/pkg/gomobile
WORK=$WORK
rm -r -f "{{.Output}}.xcframework"
GOOS=ios CGO_ENABLED=1 gobind -lang=go,objc -outdir=$WORK/ios -tags=ios{{if .Prefix}} -prefix={{.Prefix}}{{end}} golang.org/x/mobile/asset
mkdir -p $WORK/ios/src-arm64
PWD=$WORK/ios/src-arm64 GOMODCACHE=$GOPATH/pkg/mod GOOS=ios GOARCH=arm64 GOFLAGS=-tags=ios CC=iphoneos-clang CXX=iphoneos-clang++ CGO_CFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_CXXFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_LDFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_ENABLED=1 DARWIN_SDK=iphoneos GOPATH=$WORK/ios:$GOPATH go mod tidy
PWD=$WORK/ios/src-arm64 GOMODCACHE=$GOPATH/pkg/mod GOOS=ios GOARCH=arm64 GOFLAGS=-tags=ios CC=iphoneos-clang CXX=iphoneos-clang++ CGO_CFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_CXXFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_LDFLAGS=-isysroot iphoneos -miphoneos-version-min=13.0 -fembed-bitcode -arch arm64 CGO_ENABLED=1 DARWIN_SDK=iphoneos GOPATH=$WORK/ios:$GOPATH go build -x -buildmode=c-archive -o $WORK/{{.Output}}-ios-arm64.a ./gobind
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers
ln -s A $WORK/ios/iphoneos/{{.Output}}.framework/Versions/Current
ln -s Versions/Current/Headers $WORK/ios/iphoneos/{{.Output}}.framework/Headers
ln -s Versions/Current/{{.Output}} $WORK/ios/iphoneos/{{.Output}}.framework/{{.Output}}
xcrun lipo $WORK/{{.Output}}-ios-arm64.a -create -o $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/{{.Output}}
cp $WORK/ios/src/gobind/{{.Prefix}}Asset.objc.h $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers/{{.Prefix}}Asset.objc.h
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers
cp $WORK/ios/src/gobind/Universe.objc.h $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers/Universe.objc.h
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers
cp $WORK/ios/src/gobind/ref.h $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers/ref.h
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Headers
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Resources
ln -s Versions/Current/Resources $WORK/ios/iphoneos/{{.Output}}.framework/Resources
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Resources
mkdir -p $WORK/ios/iphoneos/{{.Output}}.framework/Versions/A/Modules
ln -s Versions/Current/Modules $WORK/ios/iphoneos/{{.Output}}.framework/Modules
xcodebuild -create-xcframework -framework $WORK/ios/iphoneos/{{.Output}}.framework -output {{.Output}}.xcframework
`))

func TestBindAppleAll(t *testing.T) {
	if !xcodeAvailable() {
		t.Skip("Xcode is missing")
	}
	defer func() {
		xout = os.Stderr
		buildN = false
		buildX = false
		buildO = ""
		buildTarget = ""
		bindPrefix = ""
	}()
	buildN = true
	buildX = true
	buildO = "Asset.xcframework"
	buildTarget = "ios"

	buf := new(bytes.Buffer)
	xout = buf
	gopath = filepath.SplitList(goEnv("GOPATH"))[0]
	if goos == "windows" {
		os.Setenv("HOMEDRIVE", "C:")
	}
	cmdBind.flag.Parse([]string{"golang.org/x/mobile/asset"})
	if err := runBind(cmdBind); err != nil {
		t.Log(buf.String())
		t.Fatal(err)
	}
}

const ambiguousPathsGoMod = `module ambiguouspaths

go 1.18

require golang.org/x/mobile v0.0.0-20230905140555-fbe1c053b6a9

require (
	golang.org/x/exp/shiny v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/image v0.11.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
)
`

const ambiguousPathsGoSum = `github.com/yuin/goldmark v1.4.13/go.mod h1:6yULJ656Px+3vBD8DxQVa3kxgyrAnzto9xy5taEt/CY=
golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2/go.mod h1:djNgcEr1/C05ACkg1iLfiJU5Ep61QUkGW8qpdssI0+w=
golang.org/x/crypto v0.0.0-20210921155107-089bfa567519/go.mod h1:GvvjBRRGRdwPK5ydBHafDWAxML/pGHZbMvKqRZ5+Abc=
golang.org/x/exp/shiny v0.0.0-20230817173708-d852ddb80c63 h1:3AGKexOYqL+ztdWdkB1bDwXgPBuTS/S8A4WzuTvJ8Cg=
golang.org/x/exp/shiny v0.0.0-20230817173708-d852ddb80c63/go.mod h1:UH99kUObWAZkDnWqppdQe5ZhPYESUw8I0zVV1uWBR+0=
golang.org/x/image v0.11.0 h1:ds2RoQvBvYTiJkwpSFDwCcDFNX7DqjL2WsUgTNk0Ooo=
golang.org/x/image v0.11.0/go.mod h1:bglhjqbqVuEb9e9+eNR45Jfu7D+T4Qan+NhQk8Ck2P8=
golang.org/x/mobile v0.0.0-20230905140555-fbe1c053b6a9 h1:LaLfQUz4L1tfuOlrtEouZLZ0qHDwKn87E1NKoiudP/o=
golang.org/x/mobile v0.0.0-20230905140555-fbe1c053b6a9/go.mod h1:2jxcxt/JNJik+N+QcB8q308+SyrE3bu43+sGZDmJ02M=
golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4/go.mod h1:jJ57K6gSWd91VN4djpZkiMVwK6gcyfeH4XE8wZrZaV4=
golang.org/x/mod v0.8.0/go.mod h1:iBbtSCu2XBx23ZKBPSOrRkjjQPZFPuis4dIYUhu/chs=
golang.org/x/net v0.0.0-20190620200207-3b0461eec859/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
golang.org/x/net v0.0.0-20210226172049-e18ecbb05110/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
golang.org/x/net v0.0.0-20220722155237-a158d28d115b/go.mod h1:XRhObCWvk6IyKnWLug+ECip1KBveYUHfp+8e9klMJ9c=
golang.org/x/net v0.6.0/go.mod h1:2Tu9+aMcznHK/AK1HMvgo6xiTLG5rD5rZLDS+rp2Bjs=
golang.org/x/sync v0.0.0-20190423024810-112230192c58/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.1.0/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20201119102817-f84b799fce68/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.5.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.11.0 h1:eG7RXZHdqOJ1i+0lgLgCpSXAp6M3LYlAo6osgSi0xOM=
golang.org/x/sys v0.11.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1/go.mod h1:bj7SfCRtBDWHUb9snDiAeCFNEtKQo2Wmx5Cou7ajbmo=
golang.org/x/term v0.0.0-20210927222741-03fcf44c2211/go.mod h1:jbD1KX2456YbFQfuXm/mYQcufACuNUgVhRMnK/tPxf8=
golang.org/x/term v0.5.0/go.mod h1:jMB1sMXY+tzblOD4FWmEbocvup2/aLOaQEp7JmGp78k=
golang.org/x/text v0.3.0/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
golang.org/x/text v0.3.3/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
golang.org/x/text v0.3.7/go.mod h1:u+2+/6zg+i71rQMx5EYifcz6MCKuco9NR6JIITiCfzQ=
golang.org/x/text v0.7.0/go.mod h1:mrYo+phRRbMaCq/xk9113O4dZlRixOauAjOtrjsXDZ8=
golang.org/x/text v0.12.0/go.mod h1:TvPlkZtksWOMsz7fbANvkp4WM8x/WCo/om8BMLbz+aE=
golang.org/x/tools v0.0.0-20180917221912-90fa682c2a6e/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
golang.org/x/tools v0.1.12/go.mod h1:hNGJHUnrk76NpqgfD5Aqm5Crs+Hm0VOH/i9J2+nxYbc=
golang.org/x/tools v0.6.0/go.mod h1:Xwgl3UAJ/d3gWutnCtw505GrjyAbvKui8lOU390QaIU=
golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
`

const ambiguousPathsGo = `package ambiguouspaths

import (
	_ "golang.org/x/mobile/app"
)

func Dummy() {}
`

func TestBindWithGoModules(t *testing.T) {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skipf("gomobile and gobind are not available on %s", runtime.GOOS)
	}

	dir := t.TempDir()

	if out, err := exec.Command("go", "build", "-o="+dir, "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, string(out))
	}
	if out, err := exec.Command("go", "build", "-o="+dir, "golang.org/x/mobile/cmd/gomobile").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, string(out))
	}
	path := dir
	if p := os.Getenv("PATH"); p != "" {
		path += string(filepath.ListSeparator) + p
	}

	// Create a source package dynamically to avoid go.mod files in this repository. See golang/go#34352 for more details.
	if err := os.Mkdir(filepath.Join(dir, "ambiguouspaths"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ambiguouspaths", "go.mod"), []byte(ambiguousPathsGoMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ambiguouspaths", "go.sum"), []byte(ambiguousPathsGoSum), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ambiguouspaths", "ambiguouspaths.go"), []byte(ambiguousPathsGo), 0644); err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{"android", "ios"} {
		target := target
		t.Run(target, func(t *testing.T) {
			switch target {
			case "android":
				if _, err := sdkpath.AndroidAPIPath(minAndroidAPI); err != nil {
					t.Skip("No compatible Android API platform found, skipping bind")
				}
			case "ios":
				if !xcodeAvailable() {
					t.Skip("Xcode is missing")
				}
			}

			var out string
			switch target {
			case "android":
				out = filepath.Join(dir, "cgopkg.aar")
			case "ios":
				out = filepath.Join(dir, "Cgopkg.xcframework")
			}

			tests := []struct {
				Name string
				Path string
				Dir  string
			}{
				{
					Name: "Absolute Path",
					Path: "golang.org/x/mobile/bind/testdata/cgopkg",
				},
				{
					Name: "Relative Path",
					Path: "./bind/testdata/cgopkg",
					Dir:  filepath.Join("..", ".."),
				},
				{
					Name: "Ambiguous Paths",
					Path: ".",
					Dir:  filepath.Join(dir, "ambiguouspaths"),
				},
			}

			for _, tc := range tests {
				tc := tc
				t.Run(tc.Name, func(t *testing.T) {
					cmd := exec.Command(filepath.Join(dir, "gomobile"), "bind", "-target="+target, "-o="+out, tc.Path)
					cmd.Env = append(os.Environ(), "PATH="+path, "GO111MODULE=on")
					cmd.Dir = tc.Dir
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Errorf("gomobile bind failed: %v\n%s", err, string(out))
					}
				})
			}
		})
	}
}
