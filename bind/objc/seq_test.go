// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objc

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Use the Xcode XCTestCase framework to run the regular tests and the special SeqBench.m benchmarks.
//
// Regular tests run in the xcodetest project as normal unit test (logic test in Xcode lingo).
// Unit tests execute faster but cannot run on a real device. The benchmarks in SeqBench.m run as
// UI unit tests.
//
// The Xcode files embedded in this file were constructed in Xcode 9 by:
//
// - Creating a new project through Xcode. Both unit tests and UI tests were checked off.
// - Xcode schemes are per-user by default. The shared scheme is created by selecting
//   Project => Schemes => Manage Schemes from the Xcode menu and selecting "Shared".
// - Remove files not needed for xcodebuild (determined empirically). In particular, the empty
//   tests Xcode creates can be removed and the unused user scheme.
//
// All tests here require the Xcode command line tools.

var destination = flag.String("device", "platform=iOS Simulator,name=iPhone 6s Plus", "Specify the -destination flag to xcodebuild")

var gomobileBin string

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	binDir, err := ioutil.TempDir("", "bind-objc-test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(binDir)

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	if runtime.GOOS != "android" {
		gocmd := filepath.Join(runtime.GOROOT(), "bin", "go")
		gomobileBin = filepath.Join(binDir, "gomobile"+exe)
		gobindBin := filepath.Join(binDir, "gobind"+exe)
		if out, err := exec.Command(gocmd, "build", "-o", gomobileBin, "golang.org/x/mobile/cmd/gomobile").CombinedOutput(); err != nil {
			log.Fatalf("gomobile build failed: %v: %s", err, out)
		}
		if out, err := exec.Command(gocmd, "build", "-o", gobindBin, "golang.org/x/mobile/cmd/gobind").CombinedOutput(); err != nil {
			log.Fatalf("gobind build failed: %v: %s", err, out)
		}
		path := binDir
		if oldPath := os.Getenv("PATH"); oldPath != "" {
			path += string(filepath.ListSeparator) + oldPath
		}
		os.Setenv("PATH", path)
	}

	return m.Run()
}

// TestObjcSeqTest runs ObjC test SeqTest.m.
func TestObjcSeqTest(t *testing.T) {
	runTest(t, []string{
		"golang.org/x/mobile/bind/testdata/testpkg",
		"golang.org/x/mobile/bind/testdata/testpkg/secondpkg",
		"golang.org/x/mobile/bind/testdata/testpkg/simplepkg",
	}, "", "SeqTest.m", "Testpkg.framework", false, false)
}

// TestObjcSeqBench runs ObjC test SeqBench.m.
func TestObjcSeqBench(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark in short mode.")
	}
	runTest(t, []string{"golang.org/x/mobile/bind/testdata/benchmark"}, "", "SeqBench.m", "Benchmark.framework", true, true)
}

// TestObjcSeqWrappers runs ObjC test SeqWrappers.m.
func TestObjcSeqWrappers(t *testing.T) {
	runTest(t, []string{"golang.org/x/mobile/bind/testdata/testpkg/objcpkg"}, "", "SeqWrappers.m", "Objcpkg.framework", false, false)
}

// TestObjcCustomPkg runs the ObjC test SeqCustom.m.
func TestObjcCustomPkg(t *testing.T) {
	runTest(t, []string{"golang.org/x/mobile/bind/testdata/testpkg"}, "Custom", "SeqCustom.m", "Testpkg.framework", false, false)
}

func runTest(t *testing.T, pkgNames []string, prefix, testfile, framework string, uitest, dumpOutput bool) {
	if gomobileBin == "" {
		t.Skipf("no gomobile on %s", runtime.GOOS)
	}
	if _, err := run("which xcodebuild"); err != nil {
		t.Skip("command xcodebuild not found, skipping")
	}

	tmpdir, err := ioutil.TempDir("", "bind-objc-seq-test-")
	if err != nil {
		t.Fatalf("failed to prepare temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	t.Logf("tmpdir = %s", tmpdir)

	if err := createProject(tmpdir, testfile, framework); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	if err := cp(filepath.Join(tmpdir, testfile), testfile); err != nil {
		t.Fatalf("failed to copy %s: %v", testfile, err)
	}

	cmd := exec.Command(gomobileBin, "bind", "-target", "ios", "-tags", "aaa bbb")
	if prefix != "" {
		cmd.Args = append(cmd.Args, "-prefix", prefix)
	}
	cmd.Args = append(cmd.Args, pkgNames...)
	cmd.Dir = filepath.Join(tmpdir, "xcodetest")
	// Reverse binding doesn't work with Go module since imports starting with Java or ObjC are not valid FQDNs.
	// Disable Go module explicitly until this problem is solved. See golang/go#27234.
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", buf)
		t.Fatalf("failed to run gomobile bind: %v", err)
	}

	testPattern := "xcodetestTests"
	if uitest {
		testPattern = "xcodetestUITests"
	}
	cmd = exec.Command("xcodebuild", "test", "-scheme", "xcodetest", "-destination", *destination, "-only-testing:"+testPattern)
	cmd.Dir = tmpdir
	buf, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", buf)
		t.Errorf("failed to run xcodebuild: %v", err)
	}
	if dumpOutput {
		t.Logf("%s", buf)
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

// createProject generates the files required for xcodebuild test to run a
// Objective-C testfile with a gomobile bind framework.
func createProject(dir, testfile, framework string) error {
	for _, d := range []string{"xcodetest", "xcodetest.xcodeproj/xcshareddata/xcschemes", "xcodetestTests", "xcodetestUITests"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0700); err != nil {
			return err
		}
	}
	files := []struct {
		path    string
		content string
	}{
		{"xcodetest/Info.plist", infoPlist},
		{"xcodetest.xcodeproj/project.pbxproj", fmt.Sprintf(pbxproj, testfile, framework)},
		{"xcodetest.xcodeproj/xcshareddata/xcschemes/xcodetest.xcscheme", xcodescheme},
		{"xcodetestTests/Info.plist", testInfoPlist},
		// For UI tests. Only UI tests run on a real idevice.
		{"xcodetestUITests/Info.plist", testInfoPlist},
		{"xcodetest/AppDelegate.h", appdelegateh},
		{"xcodetest/main.m", mainm},
		{"xcodetest/AppDelegate.m", appdelegatem},
	}
	for _, f := range files {
		if err := ioutil.WriteFile(filepath.Join(dir, f.path), []byte(f.content), 0700); err != nil {
			return err
		}
	}
	return nil
}

const infoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>$(EXECUTABLE_NAME)</string>
	<key>CFBundleIdentifier</key>
	<string>$(PRODUCT_BUNDLE_IDENTIFIER)</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>$(PRODUCT_NAME)</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>CFBundleSignature</key>
	<string>????</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>LSRequiresIPhoneOS</key>
	<true/>
	<key>UIRequiredDeviceCapabilities</key>
	<array>
		<string>armv7</string>
	</array>
	<key>UISupportedInterfaceOrientations</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
	<key>UISupportedInterfaceOrientations~ipad</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationPortraitUpsideDown</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
</dict>
</plist>`

const testInfoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>$(EXECUTABLE_NAME)</string>
	<key>CFBundleIdentifier</key>
	<string>$(PRODUCT_BUNDLE_IDENTIFIER)</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>$(PRODUCT_NAME)</string>
	<key>CFBundlePackageType</key>
	<string>BNDL</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>CFBundleSignature</key>
	<string>????</string>
	<key>CFBundleVersion</key>
	<string>1</string>
</dict>
</plist>`

const pbxproj = `// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXBuildFile section */
		642D058D2094883B00FE587C /* AppDelegate.m in Sources */ = {isa = PBXBuildFile; fileRef = 642D058C2094883B00FE587C /* AppDelegate.m */; };
		642D05952094883C00FE587C /* Assets.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = 642D05942094883C00FE587C /* Assets.xcassets */; };
		642D059B2094883C00FE587C /* main.m in Sources */ = {isa = PBXBuildFile; fileRef = 642D059A2094883C00FE587C /* main.m */; };
		642D05A52094883C00FE587C /* xcodetestTests.m in Sources */ = {isa = PBXBuildFile; fileRef = 642D05A42094883C00FE587C /* xcodetestTests.m */; };
		642D05B02094883C00FE587C /* xcodetestUITests.m in Sources */ = {isa = PBXBuildFile; fileRef = 642D05AF2094883C00FE587C /* xcodetestUITests.m */; };
		642D05BE209488E400FE587C /* Testpkg.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = 642D05BD209488E400FE587C /* Testpkg.framework */; };
		642D05BF209488E400FE587C /* Testpkg.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = 642D05BD209488E400FE587C /* Testpkg.framework */; };
		642D05C0209488E400FE587C /* Testpkg.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = 642D05BD209488E400FE587C /* Testpkg.framework */; };
/* End PBXBuildFile section */

/* Begin PBXContainerItemProxy section */
		642D05A12094883C00FE587C /* PBXContainerItemProxy */ = {
			isa = PBXContainerItemProxy;
			containerPortal = 642D05802094883B00FE587C /* Project object */;
			proxyType = 1;
			remoteGlobalIDString = 642D05872094883B00FE587C;
			remoteInfo = xcodetest;
		};
		642D05AC2094883C00FE587C /* PBXContainerItemProxy */ = {
			isa = PBXContainerItemProxy;
			containerPortal = 642D05802094883B00FE587C /* Project object */;
			proxyType = 1;
			remoteGlobalIDString = 642D05872094883B00FE587C;
			remoteInfo = xcodetest;
		};
/* End PBXContainerItemProxy section */

/* Begin PBXFileReference section */
		642D05882094883B00FE587C /* xcodetest.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = xcodetest.app; sourceTree = BUILT_PRODUCTS_DIR; };
		642D058B2094883B00FE587C /* AppDelegate.h */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.h; path = AppDelegate.h; sourceTree = "<group>"; };
		642D058C2094883B00FE587C /* AppDelegate.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = AppDelegate.m; sourceTree = "<group>"; };
		642D05942094883C00FE587C /* Assets.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Assets.xcassets; sourceTree = "<group>"; };
		642D05992094883C00FE587C /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		642D059A2094883C00FE587C /* main.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = main.m; sourceTree = "<group>"; };
		642D05A02094883C00FE587C /* xcodetestTests.xctest */ = {isa = PBXFileReference; explicitFileType = wrapper.cfbundle; includeInIndex = 0; path = xcodetestTests.xctest; sourceTree = BUILT_PRODUCTS_DIR; };
		642D05A42094883C00FE587C /* xcodetestTests.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = ../%[1]s; sourceTree = "<group>"; };
		642D05A62094883C00FE587C /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		642D05AB2094883C00FE587C /* xcodetestUITests.xctest */ = {isa = PBXFileReference; explicitFileType = wrapper.cfbundle; includeInIndex = 0; path = xcodetestUITests.xctest; sourceTree = BUILT_PRODUCTS_DIR; };
		642D05AF2094883C00FE587C /* xcodetestUITests.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = ../%[1]s; sourceTree = "<group>"; };
		642D05B12094883C00FE587C /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		642D05BD209488E400FE587C /* Testpkg.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = Testpkg.framework; path = %[2]s; sourceTree = "<group>"; };
/* End PBXFileReference section */

/* Begin PBXFrameworksBuildPhase section */
		642D05852094883B00FE587C /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05BE209488E400FE587C /* Testpkg.framework in Frameworks */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D059D2094883C00FE587C /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05BF209488E400FE587C /* Testpkg.framework in Frameworks */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D05A82094883C00FE587C /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05C0209488E400FE587C /* Testpkg.framework in Frameworks */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXFrameworksBuildPhase section */

/* Begin PBXGroup section */
		642D057F2094883B00FE587C = {
			isa = PBXGroup;
			children = (
				642D058A2094883B00FE587C /* xcodetest */,
				642D05A32094883C00FE587C /* xcodetestTests */,
				642D05AE2094883C00FE587C /* xcodetestUITests */,
				642D05892094883B00FE587C /* Products */,
				642D05BD209488E400FE587C /* Testpkg.framework */,
			);
			sourceTree = "<group>";
		};
		642D05892094883B00FE587C /* Products */ = {
			isa = PBXGroup;
			children = (
				642D05882094883B00FE587C /* xcodetest.app */,
				642D05A02094883C00FE587C /* xcodetestTests.xctest */,
				642D05AB2094883C00FE587C /* xcodetestUITests.xctest */,
			);
			name = Products;
			sourceTree = "<group>";
		};
		642D058A2094883B00FE587C /* xcodetest */ = {
			isa = PBXGroup;
			children = (
				642D058B2094883B00FE587C /* AppDelegate.h */,
				642D058C2094883B00FE587C /* AppDelegate.m */,
				642D05942094883C00FE587C /* Assets.xcassets */,
				642D05992094883C00FE587C /* Info.plist */,
				642D059A2094883C00FE587C /* main.m */,
			);
			path = xcodetest;
			sourceTree = "<group>";
		};
		642D05A32094883C00FE587C /* xcodetestTests */ = {
			isa = PBXGroup;
			children = (
				642D05A42094883C00FE587C /* xcodetestTests.m */,
				642D05A62094883C00FE587C /* Info.plist */,
			);
			path = xcodetestTests;
			sourceTree = "<group>";
		};
		642D05AE2094883C00FE587C /* xcodetestUITests */ = {
			isa = PBXGroup;
			children = (
				642D05AF2094883C00FE587C /* xcodetestUITests.m */,
				642D05B12094883C00FE587C /* Info.plist */,
			);
			path = xcodetestUITests;
			sourceTree = "<group>";
		};
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
		642D05872094883B00FE587C /* xcodetest */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 642D05B42094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetest" */;
			buildPhases = (
				642D05842094883B00FE587C /* Sources */,
				642D05852094883B00FE587C /* Frameworks */,
				642D05862094883B00FE587C /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = xcodetest;
			productName = xcodetest;
			productReference = 642D05882094883B00FE587C /* xcodetest.app */;
			productType = "com.apple.product-type.application";
		};
		642D059F2094883C00FE587C /* xcodetestTests */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 642D05B72094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetestTests" */;
			buildPhases = (
				642D059C2094883C00FE587C /* Sources */,
				642D059D2094883C00FE587C /* Frameworks */,
				642D059E2094883C00FE587C /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
				642D05A22094883C00FE587C /* PBXTargetDependency */,
			);
			name = xcodetestTests;
			productName = xcodetestTests;
			productReference = 642D05A02094883C00FE587C /* xcodetestTests.xctest */;
			productType = "com.apple.product-type.bundle.unit-test";
		};
		642D05AA2094883C00FE587C /* xcodetestUITests */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 642D05BA2094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetestUITests" */;
			buildPhases = (
				642D05A72094883C00FE587C /* Sources */,
				642D05A82094883C00FE587C /* Frameworks */,
				642D05A92094883C00FE587C /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
				642D05AD2094883C00FE587C /* PBXTargetDependency */,
			);
			name = xcodetestUITests;
			productName = xcodetestUITests;
			productReference = 642D05AB2094883C00FE587C /* xcodetestUITests.xctest */;
			productType = "com.apple.product-type.bundle.ui-testing";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		642D05802094883B00FE587C /* Project object */ = {
			isa = PBXProject;
			attributes = {
				LastUpgradeCheck = 0930;
				ORGANIZATIONNAME = golang;
				TargetAttributes = {
					642D05872094883B00FE587C = {
						CreatedOnToolsVersion = 9.3;
					};
					642D059F2094883C00FE587C = {
						CreatedOnToolsVersion = 9.3;
						TestTargetID = 642D05872094883B00FE587C;
					};
					642D05AA2094883C00FE587C = {
						CreatedOnToolsVersion = 9.3;
						TestTargetID = 642D05872094883B00FE587C;
					};
				};
			};
			buildConfigurationList = 642D05832094883B00FE587C /* Build configuration list for PBXProject "xcodetest" */;
			compatibilityVersion = "Xcode 9.3";
			developmentRegion = en;
			hasScannedForEncodings = 0;
			knownRegions = (
				en,
				Base,
			);
			mainGroup = 642D057F2094883B00FE587C;
			productRefGroup = 642D05892094883B00FE587C /* Products */;
			projectDirPath = "";
			projectRoot = "";
			targets = (
				642D05872094883B00FE587C /* xcodetest */,
				642D059F2094883C00FE587C /* xcodetestTests */,
				642D05AA2094883C00FE587C /* xcodetestUITests */,
			);
		};
/* End PBXProject section */

/* Begin PBXResourcesBuildPhase section */
		642D05862094883B00FE587C /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05952094883C00FE587C /* Assets.xcassets in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D059E2094883C00FE587C /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D05A92094883C00FE587C /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXResourcesBuildPhase section */

/* Begin PBXSourcesBuildPhase section */
		642D05842094883B00FE587C /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D059B2094883C00FE587C /* main.m in Sources */,
				642D058D2094883B00FE587C /* AppDelegate.m in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D059C2094883C00FE587C /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05A52094883C00FE587C /* xcodetestTests.m in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		642D05A72094883C00FE587C /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				642D05B02094883C00FE587C /* xcodetestUITests.m in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXSourcesBuildPhase section */

/* Begin PBXTargetDependency section */
		642D05A22094883C00FE587C /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 642D05872094883B00FE587C /* xcodetest */;
			targetProxy = 642D05A12094883C00FE587C /* PBXContainerItemProxy */;
		};
		642D05AD2094883C00FE587C /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 642D05872094883B00FE587C /* xcodetest */;
			targetProxy = 642D05AC2094883C00FE587C /* PBXContainerItemProxy */;
		};
/* End PBXTargetDependency section */

/* Begin XCBuildConfiguration section */
		642D05B22094883C00FE587C /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ANALYZER_NONNULL = YES;
				CLANG_ANALYZER_NUMBER_OBJECT_CONVERSION = YES_AGGRESSIVE;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++14";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_ENABLE_OBJC_WEAK = YES;
				CLANG_WARN_BLOCK_CAPTURE_AUTORELEASING = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_COMMA = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DEPRECATED_OBJC_IMPLEMENTATIONS = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_DOCUMENTATION_COMMENTS = YES;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INFINITE_RECURSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_NON_LITERAL_NULL_CONVERSION = YES;
				CLANG_WARN_OBJC_IMPLICIT_RETAIN_SELF = YES;
				CLANG_WARN_OBJC_LITERAL_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_RANGE_LOOP_ANALYSIS = YES;
				CLANG_WARN_STRICT_PROTOTYPES = YES;
				CLANG_WARN_SUSPICIOUS_MOVE = YES;
				CLANG_WARN_UNGUARDED_AVAILABILITY = YES_AGGRESSIVE;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				CODE_SIGN_IDENTITY = "iPhone Developer";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = dwarf;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				ENABLE_TESTABILITY = YES;
				GCC_C_LANGUAGE_STANDARD = gnu11;
				GCC_DYNAMIC_NO_PIC = NO;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_OPTIMIZATION_LEVEL = 0;
				GCC_PREPROCESSOR_DEFINITIONS = (
					"DEBUG=1",
					"$(inherited)",
				);
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				IPHONEOS_DEPLOYMENT_TARGET = 11.3;
				MTL_ENABLE_DEBUG_INFO = YES;
				ONLY_ACTIVE_ARCH = YES;
				SDKROOT = iphoneos;
			};
			name = Debug;
		};
		642D05B32094883C00FE587C /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ANALYZER_NONNULL = YES;
				CLANG_ANALYZER_NUMBER_OBJECT_CONVERSION = YES_AGGRESSIVE;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++14";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_ENABLE_OBJC_WEAK = YES;
				CLANG_WARN_BLOCK_CAPTURE_AUTORELEASING = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_COMMA = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DEPRECATED_OBJC_IMPLEMENTATIONS = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_DOCUMENTATION_COMMENTS = YES;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INFINITE_RECURSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_NON_LITERAL_NULL_CONVERSION = YES;
				CLANG_WARN_OBJC_IMPLICIT_RETAIN_SELF = YES;
				CLANG_WARN_OBJC_LITERAL_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_RANGE_LOOP_ANALYSIS = YES;
				CLANG_WARN_STRICT_PROTOTYPES = YES;
				CLANG_WARN_SUSPICIOUS_MOVE = YES;
				CLANG_WARN_UNGUARDED_AVAILABILITY = YES_AGGRESSIVE;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				CODE_SIGN_IDENTITY = "iPhone Developer";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
				ENABLE_NS_ASSERTIONS = NO;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				GCC_C_LANGUAGE_STANDARD = gnu11;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				IPHONEOS_DEPLOYMENT_TARGET = 11.3;
				MTL_ENABLE_DEBUG_INFO = NO;
				SDKROOT = iphoneos;
				VALIDATE_PRODUCT = YES;
			};
			name = Release;
		};
		642D05B52094883C00FE587C /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetest/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetest;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Debug;
		};
		642D05B62094883C00FE587C /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetest/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetest;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Release;
		};
		642D05B82094883C00FE587C /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				BUNDLE_LOADER = "$(TEST_HOST)";
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetestTests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetestTests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_HOST = "$(BUILT_PRODUCTS_DIR)/xcodetest.app/xcodetest";
			};
			name = Debug;
		};
		642D05B92094883C00FE587C /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				BUNDLE_LOADER = "$(TEST_HOST)";
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetestTests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetestTests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_HOST = "$(BUILT_PRODUCTS_DIR)/xcodetest.app/xcodetest";
			};
			name = Release;
		};
		642D05BB2094883C00FE587C /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetestUITests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetestUITests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_TARGET_NAME = xcodetest;
			};
			name = Debug;
		};
		642D05BC2094883C00FE587C /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				FRAMEWORK_SEARCH_PATHS = (
					"$(inherited)",
					"$(PROJECT_DIR)/xcodetest",
				);
				INFOPLIST_FILE = xcodetestUITests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = org.golang.xcodetestUITests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_TARGET_NAME = xcodetest;
			};
			name = Release;
		};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
		642D05832094883B00FE587C /* Build configuration list for PBXProject "xcodetest" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				642D05B22094883C00FE587C /* Debug */,
				642D05B32094883C00FE587C /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		642D05B42094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetest" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				642D05B52094883C00FE587C /* Debug */,
				642D05B62094883C00FE587C /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		642D05B72094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetestTests" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				642D05B82094883C00FE587C /* Debug */,
				642D05B92094883C00FE587C /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		642D05BA2094883C00FE587C /* Build configuration list for PBXNativeTarget "xcodetestUITests" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				642D05BB2094883C00FE587C /* Debug */,
				642D05BC2094883C00FE587C /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
/* End XCConfigurationList section */
	};
	rootObject = 642D05802094883B00FE587C /* Project object */;
}`

const xcodescheme = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme
   LastUpgradeVersion = "0930"
   version = "1.3">
   <BuildAction
      parallelizeBuildables = "YES"
      buildImplicitDependencies = "YES">
      <BuildActionEntries>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "YES"
            buildForArchiving = "YES"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "642D05872094883B00FE587C"
               BuildableName = "xcodetest.app"
               BlueprintName = "xcodetest"
               ReferencedContainer = "container:xcodetest.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
      </BuildActionEntries>
   </BuildAction>
   <TestAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      shouldUseLaunchSchemeArgsEnv = "YES">
      <Testables>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "642D059F2094883C00FE587C"
               BuildableName = "xcodetestTests.xctest"
               BlueprintName = "xcodetestTests"
               ReferencedContainer = "container:xcodetest.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "642D05AA2094883C00FE587C"
               BuildableName = "xcodetestUITests.xctest"
               BlueprintName = "xcodetestUITests"
               ReferencedContainer = "container:xcodetest.xcodeproj">
            </BuildableReference>
         </TestableReference>
      </Testables>
      <MacroExpansion>
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "642D05872094883B00FE587C"
            BuildableName = "xcodetest.app"
            BlueprintName = "xcodetest"
            ReferencedContainer = "container:xcodetest.xcodeproj">
         </BuildableReference>
      </MacroExpansion>
      <AdditionalOptions>
      </AdditionalOptions>
   </TestAction>
   <LaunchAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      launchStyle = "0"
      useCustomWorkingDirectory = "NO"
      ignoresPersistentStateOnLaunch = "NO"
      debugDocumentVersioning = "YES"
      debugServiceExtension = "internal"
      allowLocationSimulation = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "642D05872094883B00FE587C"
            BuildableName = "xcodetest.app"
            BlueprintName = "xcodetest"
            ReferencedContainer = "container:xcodetest.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
      <AdditionalOptions>
      </AdditionalOptions>
   </LaunchAction>
   <ProfileAction
      buildConfiguration = "Release"
      shouldUseLaunchSchemeArgsEnv = "YES"
      savedToolIdentifier = ""
      useCustomWorkingDirectory = "NO"
      debugDocumentVersioning = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "642D05872094883B00FE587C"
            BuildableName = "xcodetest.app"
            BlueprintName = "xcodetest"
            ReferencedContainer = "container:xcodetest.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </ProfileAction>
   <AnalyzeAction
      buildConfiguration = "Debug">
   </AnalyzeAction>
   <ArchiveAction
      buildConfiguration = "Release"
      revealArchiveInOrganizer = "YES">
   </ArchiveAction>
</Scheme>`

const appdelegateh = `#import <UIKit/UIKit.h>

@interface AppDelegate : UIResponder <UIApplicationDelegate>

@property (strong, nonatomic) UIWindow *window;

@end`

const appdelegatem = `#import "AppDelegate.h"

@interface AppDelegate ()

@end

@implementation AppDelegate


- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
    return YES;
}

- (void)applicationWillResignActive:(UIApplication *)application {
}

- (void)applicationDidEnterBackground:(UIApplication *)application {
}

- (void)applicationWillEnterForeground:(UIApplication *)application {
}

- (void)applicationDidBecomeActive:(UIApplication *)application {
}

- (void)applicationWillTerminate:(UIApplication *)application {
}

@end`

const mainm = `#import <UIKit/UIKit.h>
#import "AppDelegate.h"

int main(int argc, char * argv[]) {
    @autoreleasepool {
        return UIApplicationMain(argc, argv, nil, NSStringFromClass([AppDelegate class]));
    }
}`
