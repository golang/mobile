package bind

import (
	"bytes"
	"flag"
	"go/ast"
	"go/build"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
	"", // The universe package with the error type.
	"testdata/basictypes.go",
	"testdata/structs.go",
	"testdata/interfaces.go",
	"testdata/issue10788.go",
	"testdata/issue12328.go",
	"testdata/issue12403.go",
	"testdata/issue29559.go",
	"testdata/keywords.go",
	"testdata/try.go",
	"testdata/vars.go",
	"testdata/ignore.go",
	"testdata/doc.go",
	"testdata/underscores.go",
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
	fakePath := path.Dir(filename)
	for i := range refs.Embedders {
		refs.Embedders[i].PkgPath = fakePath
	}
	return refs
}

func typeCheck(t *testing.T, filename string, gopath string) (*types.Package, *ast.File) {
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors|parser.ParseComments)
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
	return pkg, f
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
		var pkg *types.Package
		var file *ast.File
		if filename != "" {
			pkg, file = typeCheck(t, filename, "")
		}

		var buf bytes.Buffer
		g := &ObjcGen{
			Generator: &Generator{
				Printer: &Printer{Buf: &buf, IndentEach: []byte("\t")},
				Fset:    fset,
				Files:   []*ast.File{file},
				Pkg:     pkg,
			},
		}
		if pkg != nil {
			g.AllPkg = []*types.Package{pkg}
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
			var golden string
			if filename != "" {
				golden = filename[:len(filename)-len(".go")]
			} else {
				golden = "testdata/universe"
			}
			golden += tc.suffix
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

func genObjcPackages(t *testing.T, dir string, cg *ObjcWrapper) {
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
		cg.Buf.Reset()
		cg.GenPackage(i)
		if err := ioutil.WriteFile(pkgFile, cg.Buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cg.Buf.Reset()
	cg.GenInterfaces()
	clsFile := filepath.Join(pkgBase, "interfaces.go")
	if err := ioutil.WriteFile(clsFile, cg.Buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	gocmd := filepath.Join(runtime.GOROOT(), "bin", "go")
	cmd := exec.Command(
		gocmd,
		"install",
		"-pkgdir="+filepath.Join(dir, "pkg", build.Default.GOOS+"_"+build.Default.GOARCH),
		"ObjC/...",
	)
	cmd.Env = append(os.Environ(), "GOPATH="+dir, "GO111MODULE=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go install the generated ObjC wrappers: %v: %s", err, string(out))
	}
}

func genJavaPackages(t *testing.T, dir string, cg *ClassGen) {
	buf := cg.Buf
	cg.Buf = new(bytes.Buffer)
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
		cg.Buf.Reset()
		cg.GenPackage(i)
		if err := ioutil.WriteFile(pkgFile, cg.Buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		io.Copy(buf, cg.Buf)
	}
	cg.Buf.Reset()
	cg.GenInterfaces()
	clsFile := filepath.Join(pkgBase, "interfaces.go")
	if err := ioutil.WriteFile(clsFile, cg.Buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	io.Copy(buf, cg.Buf)
	cg.Buf = buf

	gocmd := filepath.Join(runtime.GOROOT(), "bin", "go")
	cmd := exec.Command(
		gocmd,
		"install",
		"-pkgdir="+filepath.Join(dir, "pkg", build.Default.GOOS+"_"+build.Default.GOARCH),
		"Java/...",
	)
	cmd.Env = append(os.Environ(), "GOPATH="+dir, "GO111MODULE=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go install the generated Java wrappers: %v: %s", err, string(out))
	}
}

func TestGenJava(t *testing.T) {
	allTests := tests
	if java.IsAvailable() {
		allTests = append(append([]string{}, allTests...), javaTests...)
	}
	for _, filename := range allTests {
		var pkg *types.Package
		var file *ast.File
		var buf bytes.Buffer
		var cg *ClassGen
		var classes []*java.Class
		if filename != "" {
			refs := fileRefs(t, filename, "Java/")
			imp := &java.Importer{}
			var err error
			classes, err = imp.Import(refs)
			if err != nil {
				t.Fatal(err)
			}
			tmpGopath := ""
			if len(classes) > 0 {
				tmpGopath, err = ioutil.TempDir(os.TempDir(), "gomobile-bind-test-")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(tmpGopath)
				cg = &ClassGen{
					Printer: &Printer{
						IndentEach: []byte("\t"),
						Buf:        new(bytes.Buffer),
					},
				}
				cg.Init(classes, refs.Embedders)
				genJavaPackages(t, tmpGopath, cg)
				cg.Buf = &buf
			}
			pkg, file = typeCheck(t, filename, tmpGopath)
		}
		g := &JavaGen{
			Generator: &Generator{
				Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
				Fset:    fset,
				Files:   []*ast.File{file},
				Pkg:     pkg,
			},
		}
		if pkg != nil {
			g.AllPkg = []*types.Package{pkg}
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
			var golden string
			if filename != "" {
				golden = filename[:len(filename)-len(".go")]
			} else {
				golden = "testdata/universe"
			}
			golden += tc.suffix
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
		var pkg *types.Package
		if filename != "" {
			pkg, _ = typeCheck(t, filename, "")
		}
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
		imp := &java.Importer{}
		classes, err := imp.Import(refs)
		if err != nil {
			t.Fatal(err)
		}
		tmpGopath, err := ioutil.TempDir(os.TempDir(), "gomobile-bind-test-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpGopath)
		cg := &ClassGen{
			Printer: &Printer{
				IndentEach: []byte("\t"),
				Buf:        &buf,
			},
		}
		cg.Init(classes, refs.Embedders)
		genJavaPackages(t, tmpGopath, cg)
		pkg, _ := typeCheck(t, filename, tmpGopath)
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
		cg := &ObjcWrapper{
			Printer: &Printer{
				IndentEach: []byte("\t"),
				Buf:        &buf,
			},
		}
		var genNames []string
		for _, emb := range refs.Embedders {
			genNames = append(genNames, emb.Name)
		}
		cg.Init(types, genNames)
		genObjcPackages(t, tmpGopath, cg)
		pkg, _ := typeCheck(t, filename, tmpGopath)
		cg.GenGo()
		testGenGo(t, filename, &buf, pkg)
	}
}

func testGenGo(t *testing.T, filename string, buf *bytes.Buffer, pkg *types.Package) {
	conf := &GeneratorConfig{
		Writer: buf,
		Fset:   fset,
		Pkg:    pkg,
	}
	if pkg != nil {
		conf.AllPkg = []*types.Package{pkg}
	}
	if err := GenGo(conf); err != nil {
		t.Errorf("%s: %v", filename, err)
		return
	}
	// TODO(hyangah): let GenGo format the generated go files.
	out := writeTempFile(t, "go", gofmt(t, buf.Bytes()))
	defer os.Remove(out)

	golden := filename
	if golden == "" {
		golden = "testdata/universe"
	}
	golden += ".golden"

	goldenContents, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// format golden file using the current go version's formatting rule.
	formattedGolden := writeTempFile(t, "go", gofmt(t, goldenContents))
	defer os.Remove(formattedGolden)

	if diffstr := diff(formattedGolden, out); diffstr != "" {
		t.Errorf("%s: does not match Go golden:\n%s", filename, diffstr)

		if *updateFlag {
			t.Logf("Updating %s...", golden)
			if err := exec.Command("/bin/cp", out, golden).Run(); err != nil {
				t.Errorf("Update failed: %s", err)
			}
		}
	}
}

// gofmt formats the collection of Go source files auto-generated by gobind.
func gofmt(t *testing.T, src []byte) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	mark := []byte(gobindPreamble)
	for i, c := range bytes.Split(src, mark) {
		if i == 0 {
			buf.Write(c)
			continue
		}
		tmp := append(mark, c...)
		out, err := format.Source(tmp)
		if err != nil {
			t.Fatalf("failed to format Go file: error=%v\n----\n%s\n----", err, tmp)
		}
		if _, err := buf.Write(out); err != nil {
			t.Fatalf("failed to write formatted file to buffer: %v", err)
		}
	}
	return buf.Bytes()
}

func TestCustomPrefix(t *testing.T) {
	const datafile = "testdata/customprefix.go"
	pkg, file := typeCheck(t, datafile, "")

	type testCase struct {
		golden string
		gen    func(w io.Writer) error
	}
	var buf bytes.Buffer
	jg := &JavaGen{
		JavaPkg: "com.example",
		Generator: &Generator{
			Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
			Fset:    fset,
			AllPkg:  []*types.Package{pkg},
			Files:   []*ast.File{file},
			Pkg:     pkg,
		},
	}
	jg.Init(nil)
	testCases := []testCase{
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
	}
	for _, pref := range []string{"EX", ""} {
		og := &ObjcGen{
			Prefix: pref,
			Generator: &Generator{
				Printer: &Printer{Buf: &buf, IndentEach: []byte("    ")},
				Fset:    fset,
				AllPkg:  []*types.Package{pkg},
				Pkg:     pkg,
			},
		}
		og.Init(nil)
		testCases = append(testCases, []testCase{
			{
				"testdata/customprefix" + pref + ".objc.go.h.golden",
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
				"testdata/customprefix" + pref + ".objc.h.golden",
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
				"testdata/customprefix" + pref + ".objc.m.golden",
				func(w io.Writer) error {
					buf.Reset()
					if err := og.GenM(); err != nil {
						return err
					}
					_, err := io.Copy(w, &buf)
					return err
				},
			},
		}...)
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

// Test that typeName work for anonymous qualified fields.
func TestSelectorExprTypeName(t *testing.T) {
	e, err := parser.ParseExprFrom(fset, "", "struct { bytes.Buffer }", 0)
	if err != nil {
		t.Fatal(err)
	}
	ft := e.(*ast.StructType).Fields.List[0].Type
	if got, want := typeName(ft), "Buffer"; got != want {
		t.Errorf("got: %q; want %q", got, want)
	}
}
