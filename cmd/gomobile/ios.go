// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// TODO(jbd): iOS builds are achievable only if the host machine is darwin.
// TODO(jbd): Build darwin/arm cross compiler during gomobile init.

func goIOSBuild(src string) error {
	dir := "$XCODEPROJ"
	if !buildN {
		tmp, err := ioutil.TempDir("", "xcodeproject")
		if err != nil {
			return err
		}
		dir = tmp
		defer os.RemoveAll(dir)
	}
	if buildX {
		printcmd("mkdir %s", dir)
	}

	layout := map[string][]byte{
		dir + "/main.xcodeproj/project.pbxproj":                        []byte(projPbxproj),
		dir + "/main/Info.plist":                                       []byte(infoPlist),
		dir + "/main/Images.xcassets/AppIcon.appiconset/Contents.json": []byte(contentsJSON),
	}

	for dst, v := range layout {
		if buildX {
			printcmd("echo \"%s\" > %s", v, dst)
		}
		if !buildN {
			if err := os.MkdirAll(filepath.Dir(dst), 0775|os.ModeDir); err != nil {
				return err
			}
			if err := ioutil.WriteFile(dst, v, 0644); err != nil {
				return err
			}
		}
	}

	armPath := filepath.Join(dir, "arm")
	if err := goBuild(src, armPath, []string{
		`GOOS=darwin`,
		`GOARCH=arm`,
		`GOARM=7`,
	}); err != nil {
		return err
	}

	arm64Path := filepath.Join(dir, "arm64")
	if err := goBuild(src, arm64Path, []string{
		`GOOS=darwin`,
		`GOARCH=arm64`,
	}); err != nil {
		return err
	}

	// Apple requires builds to target both darwin/arm and darwin/arm64.
	// We are using lipo tool to build multiarchitecture binaries.
	// TODO(jbd): Investigate the new announcements about iO9's fat binary
	// size limitations are breaking this feature.
	if buildX {
		printcmd("xcrun lipo -create %s %s -o %s", armPath, arm64Path, filepath.Join(dir, "main/main"))
	}
	if !buildN {
		cmd := exec.Command(
			"xcrun", "lipo",
			"-create", armPath, arm64Path,
			"-o", filepath.Join(dir, "main/main"),
		)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// TODO(jbd): Set the launcher icon.
	if err := iosCopyAssets(dir); err != nil {
		return err
	}

	// Build and move the release build to the output directory.
	if buildX {
		printcmd("xcrun xcodebuild -configuration Release -project %s", dir+"/main.xcodeproj")
	}
	if !buildN {
		cmd := exec.Command(
			"xcrun", "xcodebuild",
			"-configuration", "Release",
			"-project", dir+"/main.xcodeproj",
		)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// TODO(jbd): Fallback to copying if renaming fails.
	out := path.Base(pkg.ImportPath) + ".app"
	if buildO != "" {
		out = filepath.Join(buildO, out)
	}
	if buildX {
		printcmd("mv %s %s", dir+"/build/Release-iphoneos/main.app", out)
	}
	if !buildN {
		if err := os.Rename(dir+"/build/Release-iphoneos/main.app", out); err != nil {
			return err
		}
	}
	return nil
}

func iosCopyAssets(xcodeProjDir string) error {
	assetsDir := filepath.Join(pkg.Dir, "assets")
	assetsDirExists := true
	fi, err := os.Stat(assetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			assetsDirExists = false
		} else {
			return err
		}
	} else {
		assetsDirExists = fi.IsDir()
	}
	if !assetsDirExists {
		return nil
	}

	if buildX {
		printcmd("cp -R %s %s", filepath.Join(pkg.Dir, "assets"), xcodeProjDir+"/main/assets")
	}
	if buildN {
		return nil
	}

	return filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dst := xcodeProjDir + "/main/assets/" + path[len(assetsDir)+1:]
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		w, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	return nil
}

func goBuild(src, o string, env []string) error {
	goroot := goEnv("GOROOT")
	gopath := goEnv("GOPATH")
	cmd := exec.Command(
		`go`,
		`build`,
		`-tags=`+strconv.Quote(strings.Join(ctx.BuildTags, ",")),
	)
	if buildV {
		cmd.Args = append(cmd.Args, "-v")
	}
	if buildI {
		cmd.Args = append(cmd.Args, "-i")
	}
	if buildX {
		cmd.Args = append(cmd.Args, "-x")
	}
	cmd.Args = append(cmd.Args, "-o="+o)
	cmd.Args = append(cmd.Args, src)
	cmd.Env = append(env, []string{
		`CGO_ENABLED=1`,
		`GOROOT=` + goroot,
		`GOPATH=` + gopath,
	}...)
	cmd.Stderr = os.Stderr
	if buildX {
		printcmd("%s", strings.Join(cmd.Env, " ")+" "+strings.Join(cmd.Args, " "))
	}
	if !buildN {
		cmd.Env = environ(cmd.Env)
		return cmd.Run()
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
  <string>main</string>
  <key>CFBundleIdentifier</key>
  <string>org.golang.todo.$(PRODUCT_NAME:rfc1034identifier)</string>
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
  <key>UILaunchStoryboardName</key>
  <string>LaunchScreen</string>
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
</plist>
`

const projPbxproj = `// !$*UTF8*$!
{
  archiveVersion = 1;
  classes = {
  };
  objectVersion = 46;
  objects = {

/* Begin PBXBuildFile section */
    254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = 254BB84E1B1FD08900C56DE9 /* Images.xcassets */; };
    254BB8681B1FD16500C56DE9 /* main in Resources */ = {isa = PBXBuildFile; fileRef = 254BB8671B1FD16500C56DE9 /* main */; };
    25FB30331B30FDEE0005924C /* assets in Resources */ = {isa = PBXBuildFile; fileRef = 25FB30321B30FDEE0005924C /* assets */; };
/* End PBXBuildFile section */

/* Begin PBXFileReference section */
    254BB83E1B1FD08900C56DE9 /* main.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = main.app; sourceTree = BUILT_PRODUCTS_DIR; };
    254BB8421B1FD08900C56DE9 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
    254BB84E1B1FD08900C56DE9 /* Images.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Images.xcassets; sourceTree = "<group>"; };
    254BB8671B1FD16500C56DE9 /* main */ = {isa = PBXFileReference; lastKnownFileType = "compiled.mach-o.executable"; path = main; sourceTree = "<group>"; };
    25FB30321B30FDEE0005924C /* assets */ = {isa = PBXFileReference; lastKnownFileType = folder; name = assets; path = main/assets; sourceTree = "<group>"; };
/* End PBXFileReference section */

/* Begin PBXGroup section */
    254BB8351B1FD08900C56DE9 = {
      isa = PBXGroup;
      children = (
        25FB30321B30FDEE0005924C /* assets */,
        254BB8401B1FD08900C56DE9 /* main */,
        254BB83F1B1FD08900C56DE9 /* Products */,
      );
      sourceTree = "<group>";
      usesTabs = 0;
    };
    254BB83F1B1FD08900C56DE9 /* Products */ = {
      isa = PBXGroup;
      children = (
        254BB83E1B1FD08900C56DE9 /* main.app */,
      );
      name = Products;
      sourceTree = "<group>";
    };
    254BB8401B1FD08900C56DE9 /* main */ = {
      isa = PBXGroup;
      children = (
        254BB8671B1FD16500C56DE9 /* main */,
        254BB84E1B1FD08900C56DE9 /* Images.xcassets */,
        254BB8411B1FD08900C56DE9 /* Supporting Files */,
      );
      path = main;
      sourceTree = "<group>";
    };
    254BB8411B1FD08900C56DE9 /* Supporting Files */ = {
      isa = PBXGroup;
      children = (
        254BB8421B1FD08900C56DE9 /* Info.plist */,
      );
      name = "Supporting Files";
      sourceTree = "<group>";
    };
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
    254BB83D1B1FD08900C56DE9 /* main */ = {
      isa = PBXNativeTarget;
      buildConfigurationList = 254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */;
      buildPhases = (
        254BB83C1B1FD08900C56DE9 /* Resources */,
      );
      buildRules = (
      );
      dependencies = (
      );
      name = main;
      productName = main;
      productReference = 254BB83E1B1FD08900C56DE9 /* main.app */;
      productType = "com.apple.product-type.application";
    };
/* End PBXNativeTarget section */

/* Begin PBXProject section */
    254BB8361B1FD08900C56DE9 /* Project object */ = {
      isa = PBXProject;
      attributes = {
        LastUpgradeCheck = 0630;
        ORGANIZATIONNAME = Developer;
        TargetAttributes = {
          254BB83D1B1FD08900C56DE9 = {
            CreatedOnToolsVersion = 6.3.1;
          };
        };
      };
      buildConfigurationList = 254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "main" */;
      compatibilityVersion = "Xcode 3.2";
      developmentRegion = English;
      hasScannedForEncodings = 0;
      knownRegions = (
        en,
        Base,
      );
      mainGroup = 254BB8351B1FD08900C56DE9;
      productRefGroup = 254BB83F1B1FD08900C56DE9 /* Products */;
      projectDirPath = "";
      projectRoot = "";
      targets = (
        254BB83D1B1FD08900C56DE9 /* main */,
      );
    };
/* End PBXProject section */

/* Begin PBXResourcesBuildPhase section */
    254BB83C1B1FD08900C56DE9 /* Resources */ = {
      isa = PBXResourcesBuildPhase;
      buildActionMask = 2147483647;
      files = (
        25FB30331B30FDEE0005924C /* assets in Resources */,
        254BB8681B1FD16500C56DE9 /* main in Resources */,
        254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */,
      );
      runOnlyForDeploymentPostprocessing = 0;
    };
/* End PBXResourcesBuildPhase section */

/* Begin XCBuildConfiguration section */
    254BB85F1B1FD08900C56DE9 /* Debug */ = {
      isa = XCBuildConfiguration;
      buildSettings = {
        ALWAYS_SEARCH_USER_PATHS = NO;
        CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
        CLANG_CXX_LIBRARY = "libc++";
        CLANG_ENABLE_MODULES = YES;
        CLANG_ENABLE_OBJC_ARC = YES;
        CLANG_WARN_BOOL_CONVERSION = YES;
        CLANG_WARN_CONSTANT_CONVERSION = YES;
        CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
        CLANG_WARN_EMPTY_BODY = YES;
        CLANG_WARN_ENUM_CONVERSION = YES;
        CLANG_WARN_INT_CONVERSION = YES;
        CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
        CLANG_WARN_UNREACHABLE_CODE = YES;
        CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
        "CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
        COPY_PHASE_STRIP = NO;
        DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
        ENABLE_STRICT_OBJC_MSGSEND = YES;
        GCC_C_LANGUAGE_STANDARD = gnu99;
        GCC_DYNAMIC_NO_PIC = NO;
        GCC_NO_COMMON_BLOCKS = YES;
        GCC_OPTIMIZATION_LEVEL = 0;
        GCC_PREPROCESSOR_DEFINITIONS = (
          "DEBUG=1",
          "$(inherited)",
        );
        GCC_SYMBOLS_PRIVATE_EXTERN = NO;
        GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
        GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
        GCC_WARN_UNDECLARED_SELECTOR = YES;
        GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
        GCC_WARN_UNUSED_FUNCTION = YES;
        GCC_WARN_UNUSED_VARIABLE = YES;
        IPHONEOS_DEPLOYMENT_TARGET = 8.3;
        MTL_ENABLE_DEBUG_INFO = YES;
        ONLY_ACTIVE_ARCH = YES;
        SDKROOT = iphoneos;
        TARGETED_DEVICE_FAMILY = "1,2";
      };
      name = Debug;
    };
    254BB8601B1FD08900C56DE9 /* Release */ = {
      isa = XCBuildConfiguration;
      buildSettings = {
        ALWAYS_SEARCH_USER_PATHS = NO;
        CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
        CLANG_CXX_LIBRARY = "libc++";
        CLANG_ENABLE_MODULES = YES;
        CLANG_ENABLE_OBJC_ARC = YES;
        CLANG_WARN_BOOL_CONVERSION = YES;
        CLANG_WARN_CONSTANT_CONVERSION = YES;
        CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
        CLANG_WARN_EMPTY_BODY = YES;
        CLANG_WARN_ENUM_CONVERSION = YES;
        CLANG_WARN_INT_CONVERSION = YES;
        CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
        CLANG_WARN_UNREACHABLE_CODE = YES;
        CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
        "CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
        COPY_PHASE_STRIP = NO;
        DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
        ENABLE_NS_ASSERTIONS = NO;
        ENABLE_STRICT_OBJC_MSGSEND = YES;
        GCC_C_LANGUAGE_STANDARD = gnu99;
        GCC_NO_COMMON_BLOCKS = YES;
        GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
        GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
        GCC_WARN_UNDECLARED_SELECTOR = YES;
        GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
        GCC_WARN_UNUSED_FUNCTION = YES;
        GCC_WARN_UNUSED_VARIABLE = YES;
        IPHONEOS_DEPLOYMENT_TARGET = 8.3;
        MTL_ENABLE_DEBUG_INFO = NO;
        SDKROOT = iphoneos;
        TARGETED_DEVICE_FAMILY = "1,2";
        VALIDATE_PRODUCT = YES;
      };
      name = Release;
    };
    254BB8621B1FD08900C56DE9 /* Debug */ = {
      isa = XCBuildConfiguration;
      buildSettings = {
        ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
        INFOPLIST_FILE = main/Info.plist;
        LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
        PRODUCT_NAME = "$(TARGET_NAME)";
      };
      name = Debug;
    };
    254BB8631B1FD08900C56DE9 /* Release */ = {
      isa = XCBuildConfiguration;
      buildSettings = {
        ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
        INFOPLIST_FILE = main/Info.plist;
        LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
        PRODUCT_NAME = "$(TARGET_NAME)";
      };
      name = Release;
    };
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
    254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "main" */ = {
      isa = XCConfigurationList;
      buildConfigurations = (
        254BB85F1B1FD08900C56DE9 /* Debug */,
        254BB8601B1FD08900C56DE9 /* Release */,
      );
      defaultConfigurationIsVisible = 0;
      defaultConfigurationName = Release;
    };
    254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */ = {
      isa = XCConfigurationList;
      buildConfigurations = (
        254BB8621B1FD08900C56DE9 /* Debug */,
        254BB8631B1FD08900C56DE9 /* Release */,
      );
      defaultConfigurationIsVisible = 0;
      defaultConfigurationName = Release;
    };
/* End XCConfigurationList section */
  };
  rootObject = 254BB8361B1FD08900C56DE9 /* Project object */;
}
`

const contentsJSON = `{
  "images" : [
    {
      "idiom" : "iphone",
      "size" : "29x29",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "29x29",
      "scale" : "3x"
    },
    {
      "idiom" : "iphone",
      "size" : "40x40",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "40x40",
      "scale" : "3x"
    },
    {
      "idiom" : "iphone",
      "size" : "60x60",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "60x60",
      "scale" : "3x"
    },
    {
      "idiom" : "ipad",
      "size" : "29x29",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "29x29",
      "scale" : "2x"
    },
    {
      "idiom" : "ipad",
      "size" : "40x40",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "40x40",
      "scale" : "2x"
    },
    {
      "idiom" : "ipad",
      "size" : "76x76",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "76x76",
      "scale" : "2x"
    }
  ],
  "info" : {
    "version" : 1,
    "author" : "xcode"
  }
}
`
