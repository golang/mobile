package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// General mobile build environment. Initialized by envInit.
var (
	gomobilepath string // $GOPATH/pkg/gomobile

	androidEnv map[string][]string // android arch -> []string

	appleEnv map[string][]string

	androidArmNM string
	appleNM      string
)

func isAndroidPlatform(platform string) bool {
	return platform == "android"
}

func isApplePlatform(platform string) bool {
	return contains(applePlatforms, platform)
}

var applePlatforms = []string{"ios", "iossimulator", "macos", "maccatalyst"}

func platformArchs(platform string) []string {
	switch platform {
	case "ios":
		return []string{"arm64"}
	case "iossimulator":
		return []string{"arm64", "amd64"}
	case "macos", "maccatalyst":
		return []string{"arm64", "amd64"}
	case "android":
		return []string{"arm", "arm64", "386", "amd64"}
	default:
		panic(fmt.Sprintf("unexpected platform: %s", platform))
	}
}

func isSupportedArch(platform, arch string) bool {
	return contains(platformArchs(platform), arch)
}

// platformOS returns the correct GOOS value for platform.
func platformOS(platform string) string {
	switch platform {
	case "android":
		return "android"
	case "ios", "iossimulator":
		return "ios"
	case "macos", "maccatalyst":
		// For "maccatalyst", Go packages should be built with GOOS=darwin,
		// not GOOS=ios, since the underlying OS (and kernel, runtime) is macOS.
		// We also apply a "macos" or "maccatalyst" build tag, respectively.
		// See below for additional context.
		return "darwin"
	default:
		panic(fmt.Sprintf("unexpected platform: %s", platform))
	}
}

func platformTags(platform string) []string {
	switch platform {
	case "android":
		return []string{"android"}
	case "ios", "iossimulator":
		return []string{"ios"}
	case "macos":
		return []string{"macos"}
	case "maccatalyst":
		// Mac Catalyst is a subset of iOS APIs made available on macOS
		// designed to ease porting apps developed for iPad to macOS.
		// See https://developer.apple.com/mac-catalyst/.
		// Because of this, when building a Go package targeting maccatalyst,
		// GOOS=darwin (not ios). To bridge the gap and enable maccatalyst
		// packages to be compiled, we also specify the "ios" build tag.
		// To help discriminate between darwin, ios, macos, and maccatalyst
		// targets, there is also a "maccatalyst" tag.
		// Some additional context on this can be found here:
		// https://stackoverflow.com/questions/12132933/preprocessor-macro-for-os-x-targets/49560690#49560690
		// TODO(ydnar): remove tag "ios" when cgo supports Catalyst
		// See golang.org/issues/47228
		return []string{"ios", "macos", "maccatalyst"}
	default:
		panic(fmt.Sprintf("unexpected platform: %s", platform))
	}
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func buildEnvInit() (cleanup func(), err error) {
	// Find gomobilepath.
	gopath := goEnv("GOPATH")
	for _, p := range filepath.SplitList(gopath) {
		gomobilepath = filepath.Join(p, "pkg", "gomobile")
		if _, err := os.Stat(gomobilepath); buildN || err == nil {
			break
		}
	}

	if buildX {
		fmt.Fprintln(xout, "GOMOBILE="+gomobilepath)
	}

	// Check the toolchain is in a good state.
	// Pick a temporary directory for assembling an apk/app.
	if gomobilepath == "" {
		return nil, errors.New("toolchain not installed, run `gomobile init`")
	}

	cleanupFn := func() {
		if buildWork {
			fmt.Printf("WORK=%s\n", tmpdir)
			return
		}
		removeAll(tmpdir)
	}
	if buildN {
		tmpdir = "$WORK"
		cleanupFn = func() {}
	} else {
		tmpdir, err = ioutil.TempDir("", "gomobile-work-")
		if err != nil {
			return nil, err
		}
	}
	if buildX {
		fmt.Fprintln(xout, "WORK="+tmpdir)
	}

	if err := envInit(); err != nil {
		return nil, err
	}

	return cleanupFn, nil
}

func envInit() (err error) {
	// Setup the cross-compiler environments.
	if ndkRoot, err := ndkRoot(); err == nil {
		androidEnv = make(map[string][]string)
		if buildAndroidAPI < minAndroidAPI {
			return fmt.Errorf("gomobile requires Android API level >= %d", minAndroidAPI)
		}
		for arch, toolchain := range ndk {
			clang := toolchain.Path(ndkRoot, "clang")
			clangpp := toolchain.Path(ndkRoot, "clang++")
			if !buildN {
				tools := []string{clang, clangpp}
				if runtime.GOOS == "windows" {
					// Because of https://github.com/android-ndk/ndk/issues/920,
					// we require r19c, not just r19b. Fortunately, the clang++.cmd
					// script only exists in r19c.
					tools = append(tools, clangpp+".cmd")
				}
				for _, tool := range tools {
					_, err = os.Stat(tool)
					if err != nil {
						return fmt.Errorf("No compiler for %s was found in the NDK (tried %s). Make sure your NDK version is >= r19c. Use `sdkmanager --update` to update it.", arch, tool)
					}
				}
			}
			androidEnv[arch] = []string{
				"GOOS=android",
				"GOARCH=" + arch,
				"CC=" + clang,
				"CXX=" + clangpp,
				"CGO_ENABLED=1",
			}
			if arch == "arm" {
				androidEnv[arch] = append(androidEnv[arch], "GOARM=7")
			}
		}
	}

	if !xcodeAvailable() {
		return nil
	}

	appleNM = "nm"
	appleEnv = make(map[string][]string)
	for _, platform := range applePlatforms {
		for _, arch := range platformArchs(platform) {
			var env []string
			var goos, sdk, clang, cflags string
			var err error
			switch platform {
			case "ios":
				goos = "ios"
				sdk = "iphoneos"
				clang, cflags, err = envClang(sdk)
				cflags += " -miphoneos-version-min=" + buildIOSVersion
				cflags += " -fembed-bitcode"
			case "iossimulator":
				goos = "ios"
				sdk = "iphonesimulator"
				clang, cflags, err = envClang(sdk)
				cflags += " -mios-simulator-version-min=" + buildIOSVersion
				cflags += " -fembed-bitcode"
			case "maccatalyst":
				// Mac Catalyst is a subset of iOS APIs made available on macOS
				// designed to ease porting apps developed for iPad to macOS.
				// See https://developer.apple.com/mac-catalyst/.
				// Because of this, when building a Go package targeting maccatalyst,
				// GOOS=darwin (not ios). To bridge the gap and enable maccatalyst
				// packages to be compiled, we also specify the "ios" build tag.
				// To help discriminate between darwin, ios, macos, and maccatalyst
				// targets, there is also a "maccatalyst" tag.
				// Some additional context on this can be found here:
				// https://stackoverflow.com/questions/12132933/preprocessor-macro-for-os-x-targets/49560690#49560690
				goos = "darwin"
				sdk = "macosx"
				clang, cflags, err = envClang(sdk)
				// TODO(ydnar): the following 3 lines MAY be needed to compile
				// packages or apps for maccatalyst. Commenting them out now in case
				// it turns out they are necessary. Currently none of the example
				// apps will build for macos or maccatalyst because they have a
				// GLKit dependency, which is deprecated on all Apple platforms, and
				// broken on maccatalyst (GLKView isn’t available).
				// sysroot := strings.SplitN(cflags, " ", 2)[1]
				// cflags += " -isystem " + sysroot + "/System/iOSSupport/usr/include"
				// cflags += " -iframework " + sysroot + "/System/iOSSupport/System/Library/Frameworks"
				switch arch {
				case "amd64":
					cflags += " -target x86_64-apple-ios" + buildIOSVersion + "-macabi"
				case "arm64":
					cflags += " -target arm64-apple-ios" + buildIOSVersion + "-macabi"
					cflags += " -fembed-bitcode"
				}
			case "macos":
				goos = "darwin"
				sdk = "macosx" // Note: the SDK is called "macosx", not "macos"
				clang, cflags, err = envClang(sdk)
				if arch == "arm64" {
					cflags += " -fembed-bitcode"
				}
			default:
				panic(fmt.Errorf("unknown Apple target: %s/%s", platform, arch))
			}

			if err != nil {
				return err
			}

			env = append(env,
				"GOOS="+goos,
				"GOARCH="+arch,
				"GOFLAGS="+"-tags="+strings.Join(platformTags(platform), ","),
				"CC="+clang,
				"CXX="+clang+"++",
				"CGO_CFLAGS="+cflags+" -arch "+archClang(arch),
				"CGO_CXXFLAGS="+cflags+" -arch "+archClang(arch),
				"CGO_LDFLAGS="+cflags+" -arch "+archClang(arch),
				"CGO_ENABLED=1",
				"DARWIN_SDK="+sdk,
			)
			appleEnv[platform+"/"+arch] = env
		}
	}

	return nil
}

func ndkRoot() (string, error) {
	if buildN {
		return "$NDK_PATH", nil
	}

	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome != "" {
		ndkRoot := filepath.Join(androidHome, "ndk-bundle")
		_, err := os.Stat(ndkRoot)
		if err == nil {
			return ndkRoot, nil
		}
	}

	ndkRoot := os.Getenv("ANDROID_NDK_HOME")
	if ndkRoot != "" {
		_, err := os.Stat(ndkRoot)
		if err == nil {
			return ndkRoot, nil
		}
	}

	return "", fmt.Errorf("no Android NDK found in $ANDROID_HOME/ndk-bundle nor in $ANDROID_NDK_HOME")
}

func envClang(sdkName string) (clang, cflags string, err error) {
	if buildN {
		return sdkName + "-clang", "-isysroot " + sdkName, nil
	}
	cmd := exec.Command("xcrun", "--sdk", sdkName, "--find", "clang")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("xcrun --find: %v\n%s", err, out)
	}
	clang = strings.TrimSpace(string(out))

	cmd = exec.Command("xcrun", "--sdk", sdkName, "--show-sdk-path")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("xcrun --show-sdk-path: %v\n%s", err, out)
	}
	sdk := strings.TrimSpace(string(out))
	return clang, "-isysroot " + sdk, nil
}

func archClang(goarch string) string {
	switch goarch {
	case "arm":
		return "armv7"
	case "arm64":
		return "arm64"
	case "386":
		return "i386"
	case "amd64":
		return "x86_64"
	default:
		panic(fmt.Sprintf("unknown GOARCH: %q", goarch))
	}
}

// environ merges os.Environ and the given "key=value" pairs.
// If a key is in both os.Environ and kv, kv takes precedence.
func environ(kv []string) []string {
	cur := os.Environ()
	new := make([]string, 0, len(cur)+len(kv))

	envs := make(map[string]string, len(cur))
	for _, ev := range cur {
		elem := strings.SplitN(ev, "=", 2)
		if len(elem) != 2 || elem[0] == "" {
			// pass the env var of unusual form untouched.
			// e.g. Windows may have env var names starting with "=".
			new = append(new, ev)
			continue
		}
		if goos == "windows" {
			elem[0] = strings.ToUpper(elem[0])
		}
		envs[elem[0]] = elem[1]
	}
	for _, ev := range kv {
		elem := strings.SplitN(ev, "=", 2)
		if len(elem) != 2 || elem[0] == "" {
			panic(fmt.Sprintf("malformed env var %q from input", ev))
		}
		if goos == "windows" {
			elem[0] = strings.ToUpper(elem[0])
		}
		envs[elem[0]] = elem[1]
	}
	for k, v := range envs {
		new = append(new, k+"="+v)
	}
	return new
}

func getenv(env []string, key string) string {
	prefix := key + "="
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			return kv[len(prefix):]
		}
	}
	return ""
}

func archNDK() string {
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		return "windows"
	} else {
		var arch string
		switch runtime.GOARCH {
		case "386":
			arch = "x86"
		case "amd64":
			arch = "x86_64"
		case "arm64":
			// Android NDK does not contain arm64 toolchains (until and
			// including NDK 23), use use x86_64 instead. See:
			// https://github.com/android/ndk/issues/1299
			if runtime.GOOS == "darwin" {
				arch = "x86_64"
				break
			}
			fallthrough
		default:
			panic("unsupported GOARCH: " + runtime.GOARCH)
		}
		return runtime.GOOS + "-" + arch
	}
}

type ndkToolchain struct {
	arch        string
	abi         string
	minAPI      int
	toolPrefix  string
	clangPrefix string
}

func (tc *ndkToolchain) ClangPrefix() string {
	if buildAndroidAPI < tc.minAPI {
		return fmt.Sprintf("%s%d", tc.clangPrefix, tc.minAPI)
	}
	return fmt.Sprintf("%s%d", tc.clangPrefix, buildAndroidAPI)
}

func (tc *ndkToolchain) Path(ndkRoot, toolName string) string {
	var pref string
	switch toolName {
	case "clang", "clang++":
		pref = tc.ClangPrefix()
	default:
		pref = tc.toolPrefix
	}
	return filepath.Join(ndkRoot, "toolchains", "llvm", "prebuilt", archNDK(), "bin", pref+"-"+toolName)
}

type ndkConfig map[string]ndkToolchain // map: GOOS->androidConfig.

func (nc ndkConfig) Toolchain(arch string) ndkToolchain {
	tc, ok := nc[arch]
	if !ok {
		panic(`unsupported architecture: ` + arch)
	}
	return tc
}

var ndk = ndkConfig{
	"arm": {
		arch:        "arm",
		abi:         "armeabi-v7a",
		minAPI:      16,
		toolPrefix:  "arm-linux-androideabi",
		clangPrefix: "armv7a-linux-androideabi",
	},
	"arm64": {
		arch:        "arm64",
		abi:         "arm64-v8a",
		minAPI:      21,
		toolPrefix:  "aarch64-linux-android",
		clangPrefix: "aarch64-linux-android",
	},

	"386": {
		arch:        "x86",
		abi:         "x86",
		minAPI:      16,
		toolPrefix:  "i686-linux-android",
		clangPrefix: "i686-linux-android",
	},
	"amd64": {
		arch:        "x86_64",
		abi:         "x86_64",
		minAPI:      21,
		toolPrefix:  "x86_64-linux-android",
		clangPrefix: "x86_64-linux-android",
	},
}

func xcodeAvailable() bool {
	err := exec.Command("xcrun", "xcodebuild", "-version").Run()
	return err == nil
}
