package bind

import (
	"bytes"
	"flag"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/mobile/internal/importers"
	"golang.org/x/mobile/internal/importers/java"
	"golang.org/x/mobile/internal/importers/objc"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var updateFlag = flag.Bool("update", false, "Update the golden files.")

var tests = []string{
	"testdata/basictypes.go",
	"testdata/structs.go",
	"testdata/interfaces.go",
	"testdata/issue10788.go",
	"testdata/issue12328.go",
	"testdata/issue12403.go",
	"testdata/keywords.go",
	"testdata/try.go",
	"testdata/vars.go",
	"testdata/ignore.go",
}

var javaTests = []string{
	"testdata/java.go",
	"testdata/classes.go",
}

var objcTests = []string{
	"testdata/objc.go",
	"testdata/objcw.go",
}

var fset = token.NewFileSet()

func fileRefs(t *testing.T, filename string, pkgPrefix string) *importers.References {
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("%s: %v", filename, err)
	}
	refs, err := importers.AnalyzeFile(f, pkgPrefix)
	if err != nil {
		t.Fatalf("%s: %v", filename, err)
	}
	return refs
}

func typeCheck(t *testing.T, filename string, gopath string) *types.Package {
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("%s: %v", filename, err)
	}

	pkgName := filepath.Base(filename)
	pkgName = strings.TrimSuffix(pkgName, ".go")

	// typecheck and collect typechecker errors
	var conf types.Config
	conf.Error = func(err error) {
		t.Error(err)
	}
	if gopath != "" {
		conf.Importer = importer.Default()
		oldDefault := build.Default
		defer func() { build.Default = oldDefault }()
		build.Default.GOPATH = gopath
	}
	pkg, err := conf.Check(pkgName, fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

// diff runs the command "diff a b" and returns its output
func diff(a, b string) string {
	var buf bytes.Buffer
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "plan9":
		cmd = exec.Command("/bin/diff", "-c", a, b)
	default:
		cmd = exec.Command("/usr/bin/diff", "-u", a, b)
	}
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Run()
	return buf.String()
}

func writeTempFile(t *testing.T, name string, contents []byte) string {
	f, err := ioutil.TempFile("", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestGenObjc(t *testing.T) {
	for _, filename := range tests {
		pkg := typeCheck(t, filename, "")

		var buf bytes.Buffer
		g := &ObjcGen{
			Generator: &Generator{
				Printer: &Printer{Buf: &buf, IndentEach: []byte("\t")},
				Fset:    fset,
				AllPkg:  []*types.Package{pkg},
				Pkg:     pkg,
			},
		}
		g.Init(nil)

		testcases := []struct {
			suffix string
			gen    func() error
		}{
			{
				".objc.h.golden",
				g.GenH,
			},
			{
				".objc.m.golden",
				g.GenM,
			},
			{
				".objc.go.h.golden",
				g.GenGoH,
			},
		}
		for _, tc := range testcases {
			buf.Reset()
			if err := tc.gen(); err != nil {
				t.Errorf("%s: %v", filename, err)
				continue
			}
			out := writeTempFile(t, "generated"+tc.suffix, buf.Bytes())
			defer os.Remove(out)
			golden := filename[:len(filename)-len(".go")] + tc.suffix
			if diffstr := diff(golden, out); diffstr != "" {
				t.Errorf("%s: does not match Objective-C golden:\n%s", filename, diffstr)
				if *updateFlag {
					t.Logf("Updating %s...", golden)
					err := exec.Command("/bin/cp", out, golden).Run()
					if err != nil {
						t.Errorf("Update failed: %s", err)
					}
				}
			}
		}
	}
}

func genObjcPackages(t *testing.T, dir string, types []*objc.Named, buf *bytes.Buffer) *ObjcWrapper {
	cg := &ObjcWrapper{
		Printer: &Printer{
			IndentEach: []byte("\t"),
			Buf:        buf,
		},
	}
	cg.Init(types)
	pkgBase := filepath.Join(dir, "src", "ObjC")
	if err := os.MkdirAll(pkgBase, 0700); err != nil {
		t.Fatal(err)
	}
	for i, jpkg := range cg.Packages() {
		pkgDir := filepath.Join(pkgBase, jpkg)
		if err := os.MkdirAll(pkgDir, 0700); err != nil {
			t.Fatal(err)
		}
		pkgFile := filepath.Join(pkgDir, "package.go")
		buf.Reset()
		cg.GenPackage(i)
		if err := ioutil.WriteFile(pkgFile, buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	buf.Reset()
	cg.GenInterfaces()
	clsFile := filepath.Join(pkgBase, "interfaces.go")
	if err := ioutil.WriteFile(clsFile, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(
		"go",
		"install",
		"-pkgdir="+filepath.Join(dir, "pkg", build.Default.GOOS+"_"+build.Default.GOARCH),
		"ObjC/...",
	)
	cmd.Env = append(cmd.Env, "GOPATH="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go install the generated ObjC wrappers: %v: %s", err, string(out))
	}
	return cg
}

func genJavaPackages(t *testing.T, dir string, classes []*java.Class, buf *bytes.Buffer) *ClassGen {
	cg := &ClassGen{
		Printer: &Printer{
			IndentEach: []byte("\t"),
			Buf:        buf,
		},
	}
	cg.Init(classes)
	pkgBase := filepath.Join(dir, "src", "Java")
	if err := os.MkdirAll(pkgBase, 0700); err != nil {
		t.Fatal(err)
	}
	for i, jpkg := range cg.Packages() {
		pkgDir := filepath.Join(pkgBase, jpkg)
		if err := os.MkdirAll(pkgDir, 0700); err != nil {
			t.Fatal(err)
		}
		pkgFile := filepath.Join(pkgDir, "package.go")
		buf.Reset()
		cg.GenPackage(i)
		if err := ioutil.WriteFile(pkgFile, buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	buf.Reset()
	cg.GenInterfaces()
	clsFile := filepath.Join(pkgBase, "interfaces.go")
	if err := ioutil.WriteFile(clsFile, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(
		"go",
		"install",
		"-pkgdir="+filepath.Join(dir, "pkg", build.Default.GOOS+"_"+build.Default.GOARCH),
		"Java/...",
	)
	cmd.Env = append(cmd.Env, "GOPATH="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go install the generated Java wrappers: %v: %s", err, string(out))
	}
	return cg
}

func TestGenJava(t *testing.T) {
	allTests := tests
	if java.IsAvailable() {
		allTests = append(append([]string{}, allTests...), javaTests...)
	}
	for _, filename := range allTests {
		refs := fileRefs(t, filename, "Java/")
		classes, err := java.Import("", "", refs)
		if err != nil {
			t.Fatal(err)
		}
		var cg *ClassGen
		tmpGopath := ""
		var buf bytes.Buffer
		if len(classes) > 0 {
			tmpGopath, err = ioutil.TempDir(os.TempDir(), "gomobile-bind-test-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpGopath)
			cg = genJavaPackages(t, tmpGopath, classes, new(bytes.Buffer))
			cg.Buf = &buf
		}
		pkg := typeCheck(t, filename, tmpGopath)
		g := &JavaGen{
			Generator: &Generator{
				Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
				Fset:    fset,
				AllPkg:  []*types.Package{pkg},
				Pkg:     pkg,
			},
		}
		g.Init(classes)
		testCases := []struct {
			suffix string
			gen    func() error
		}{
			{
				".java.golden",
				func() error {
					for i := range g.ClassNames() {
						if err := g.GenClass(i); err != nil {
							return err
						}
					}
					return g.GenJava()
				},
			},
			{
				".java.c.golden",
				func() error {
					if cg != nil {
						cg.GenC()
					}
					return g.GenC()
				},
			},
			{
				".java.h.golden",
				func() error {
					if cg != nil {
						cg.GenH()
					}
					return g.GenH()
				},
			},
		}

		for _, tc := range testCases {
			buf.Reset()
			if err := tc.gen(); err != nil {
				t.Errorf("%s: %v", filename, err)
				continue
			}
			out := writeTempFile(t, "generated"+tc.suffix, buf.Bytes())
			defer os.Remove(out)
			golden := filename[:len(filename)-len(".go")] + tc.suffix
			if diffstr := diff(golden, out); diffstr != "" {
				t.Errorf("%s: does not match Java golden:\n%s", filename, diffstr)

				if *updateFlag {
					t.Logf("Updating %s...", golden)
					if err := exec.Command("/bin/cp", out, golden).Run(); err != nil {
						t.Errorf("Update failed: %s", err)
					}
				}

			}
		}
	}
}

func TestGenGo(t *testing.T) {
	for _, filename := range tests {
		var buf bytes.Buffer
		pkg := typeCheck(t, filename, "")
		testGenGo(t, filename, &buf, pkg)
	}
}

func TestGenGoJavaWrappers(t *testing.T) {
	if !java.IsAvailable() {
		t.Skipf("java is not available")
	}
	for _, filename := range javaTests {
		var buf bytes.Buffer
		refs := fileRefs(t, filename, "Java/")
		classes, err := java.Import("", "", refs)
		if err != nil {
			t.Fatal(err)
		}
		tmpGopath, err := ioutil.TempDir(os.TempDir(), "gomobile-bind-test-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpGopath)
		cg := genJavaPackages(t, tmpGopath, classes, &buf)
		pkg := typeCheck(t, filename, tmpGopath)
		cg.GenGo()
		testGenGo(t, filename, &buf, pkg)
	}
}

func TestGenGoObjcWrappers(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skipf("can only generate objc wrappers on darwin")
	}
	for _, filename := range objcTests {
		var buf bytes.Buffer
		refs := fileRefs(t, filename, "ObjC/")
		types, err := objc.Import(refs)
		if err != nil {
			t.Fatal(err)
		}
		tmpGopath, err := ioutil.TempDir(os.TempDir(), "gomobile-bind-test-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpGopath)
		cg := genObjcPackages(t, tmpGopath, types, &buf)
		pkg := typeCheck(t, filename, tmpGopath)
		cg.GenGo()
		testGenGo(t, filename, &buf, pkg)
	}
}

func testGenGo(t *testing.T, filename string, buf *bytes.Buffer, pkg *types.Package) {
	conf := &GeneratorConfig{
		Writer: buf,
		Fset:   fset,
		Pkg:    pkg,
		AllPkg: []*types.Package{pkg},
	}
	if err := GenGo(conf); err != nil {
		t.Errorf("%s: %v", filename, err)
		return
	}
	out := writeTempFile(t, "go", buf.Bytes())
	defer os.Remove(out)
	golden := filename + ".golden"
	if diffstr := diff(golden, out); diffstr != "" {
		t.Errorf("%s: does not match Go golden:\n%s", filename, diffstr)

		if *updateFlag {
			t.Logf("Updating %s...", golden)
			if err := exec.Command("/bin/cp", out, golden).Run(); err != nil {
				t.Errorf("Update failed: %s", err)
			}
		}
	}
}

func TestCustomPrefix(t *testing.T) {
	const datafile = "testdata/customprefix.go"
	const isHeader = true
	pkg := typeCheck(t, datafile, "")

	var buf bytes.Buffer
	jg := &JavaGen{
		JavaPkg: "com.example",
		Generator: &Generator{
			Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
			Fset:    fset,
			AllPkg:  []*types.Package{pkg},
			Pkg:     pkg,
		},
	}
	jg.Init(nil)
	og := &ObjcGen{
		Prefix: "EX",
		Generator: &Generator{
			Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
			Fset:    fset,
			AllPkg:  []*types.Package{pkg},
			Pkg:     pkg,
		},
	}
	og.Init(nil)
	testCases := []struct {
		golden string
		gen    func(w io.Writer) error
	}{
		{
			"testdata/customprefix.java.golden",
			func(w io.Writer) error {
				buf.Reset()
				for i := range jg.ClassNames() {
					if err := jg.GenClass(i); err != nil {
						return err
					}
				}
				if err := jg.GenJava(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
		{
			"testdata/customprefix.java.h.golden",
			func(w io.Writer) error {
				buf.Reset()
				if err := jg.GenH(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
		{
			"testdata/customprefix.java.c.golden",
			func(w io.Writer) error {
				buf.Reset()
				if err := jg.GenC(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
		{
			"testdata/customprefix.objc.go.h.golden",
			func(w io.Writer) error {
				buf.Reset()
				if err := og.GenGoH(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
		{
			"testdata/customprefix.objc.h.golden",
			func(w io.Writer) error {
				buf.Reset()
				if err := og.GenH(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
		{
			"testdata/customprefix.objc.m.golden",
			func(w io.Writer) error {
				buf.Reset()
				if err := og.GenM(); err != nil {
					return err
				}
				_, err := io.Copy(w, &buf)
				return err
			},
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer
		if err := tc.gen(&buf); err != nil {
			t.Errorf("generating %s: %v", tc.golden, err)
			continue
		}
		out := writeTempFile(t, "generated", buf.Bytes())
		defer os.Remove(out)
		if diffstr := diff(tc.golden, out); diffstr != "" {
			t.Errorf("%s: generated file does not match:\b%s", tc.golden, diffstr)
			if *updateFlag {
				t.Logf("Updating %s...", tc.golden)
				err := exec.Command("/bin/cp", out, tc.golden).Run()
				if err != nil {
					t.Errorf("Update failed: %s", err)
				}
			}
		}
	}
}

func TestLowerFirst(t *testing.T) {
	testCases := []struct {
		in, want string
	}{
		{"", ""},
		{"Hello", "hello"},
		{"HelloGopher", "helloGopher"},
		{"hello", "hello"},
		{"ID", "id"},
		{"IDOrName", "idOrName"},
		{"ΓειαΣας", "γειαΣας"},
	}

	for _, tc := range testCases {
		if got := lowerFirst(tc.in); got != tc.want {
			t.Errorf("lowerFirst(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
